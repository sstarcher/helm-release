package helm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"github.com/sstarcher/helm-release/git"
)

// DefaultTagPath is the default path to the image tag in a helm values.yaml
var DefaultTagPath = "image.tag"

// Chart defines a Helm Chart
type Chart struct {
	Name    string
	path    string
	TagPath string
	git     git.Gitter
}

// New finds the helm chart in the directory and returns a Chart object
func New(dir string) (*Chart, error) {
	charts := findCharts(dir)
	if len(charts) == 0 {
		return nil, errors.New("unable to find a Chart.yaml")
	} else if len(charts) > 1 {
		paths := []string{}
		for _, value := range charts {
			paths = append(paths, value.path)
		}
		return nil, fmt.Errorf("found more than a single chart in the following paths -  \n\t%s", strings.Join(paths, "\n\t"))
	}

	chart := charts[0]
	chart.TagPath = "image.tag"

	var err error
	chart.git, err = git.New(chart.path)
	if err != nil {
		return nil, err
	}
	return &chart, nil
}

// FindCharts looks for all helm charts under the given path
func findCharts(dir string) []Chart {
	charts := []Chart{}
	_ = filepath.Walk(dir, func(file string, f os.FileInfo, err error) error {
		if strings.HasSuffix(file, "Chart.yaml") {
			dir = path.Dir(file)
			charts = append(charts, Chart{
				path: dir,
				Name: path.Base(dir),
			})
		}
		return nil
	})
	return charts
}

func (c *Chart) gitSemVersion() (*semver.Version, error) {
	tag, err := c.git.Tag()
	if err != nil {
		tag = "0.0.1"
		log.Infof("unable to find any git tags using %s", tag)
	}

	tag = strings.TrimPrefix(tag, "v")
	tag = strings.TrimPrefix(tag, "r")
	ver, err := semver.NewVersion(tag)
	if err != nil {
		return nil, fmt.Errorf("%s %s", tag, err)
	}
	return ver, nil
}

// Version determines the correct version for the chart
func (c *Chart) Version() (*string, error) {
	commits, err := c.git.Commits()
	if err != nil {
		return nil, err
	}

	sha, err := c.git.Sha()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch git sha %s", err)
	}

	branch, err := c.git.Branch()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch git branch %s", err)
	}

	ver, err := c.gitSemVersion()
	if err != nil {
		return nil, err
	}

	version := *ver
	prerel := ""
	tagged := c.git.IsTagged()
	if !tagged {
		if branch == "head" && commits == 0 {
			return nil, errors.New("this is likely an light-weight git tag. please use a annotated tag for helm release to function properly")
		}
		version = version.IncPatch()
		if branch != "master" {
			prerel = "0." + branch
		}
	}

	if commits != 0 {
		if prerel != "" {
			prerel += "."
		}
		prerel += strconv.Itoa(commits)
	}

	if prerel != "" {
		version, err = version.SetPrerelease(prerel)
		if err != nil {
			return nil, err
		}
	}

	version, err = version.SetMetadata(sha)
	if err != nil {
		return nil, err
	}

	verStr := version.String()
	return &verStr, err
}

// UpdateChartVersion updates the version of the helm chart
func (c *Chart) UpdateChartVersion(chartVersion string) error {
	var config map[interface{}]interface{}
	source, err := ioutil.ReadFile(c.path + "/Chart.yaml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		return err
	}

	config["version"] = chartVersion

	out, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.path+"/Chart.yaml", out, 0644)
	if err != nil {
		return err
	}
	return nil
}

// UpdateImageVersion replaces the image tag in the values.yaml
func (c *Chart) UpdateImageVersion(dockerVersion string) error {
	var values interface{}
	valuesData, err := ioutil.ReadFile(c.path + "/values.yaml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(valuesData, &values)
	if err != nil {
		return err
	}

	obj := values
	if obj == nil {
		return errors.New("the values.yaml file is empty")
	}

	pathArray := strings.Split(c.TagPath, ".")
	for i := 0; i < len(pathArray); i++ {
		k := pathArray[i]
		switch objMap := obj.(type) {
		case map[interface{}]interface{}:
			if i == len(pathArray)-1 {
				if _, ok := objMap[k]; !ok {
					return fmt.Errorf("final key in path does not exist %s for path %s", k, c.TagPath)
				}
				objMap[k] = dockerVersion
				break
			}
			obj = objMap[k]
			if obj == nil {
				return fmt.Errorf("unable to process %s", c.TagPath)
			}
		default:
			return fmt.Errorf("while processing key[%s] for path[%s] expected a map, but got %T", k, c.TagPath, objMap)
		}

	}

	out, err := yaml.Marshal(&values)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.path+"/values.yaml", out, 0644)
	if err != nil {
		return err
	}

	return nil
}
