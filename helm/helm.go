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

// Version determines the correct version for the chart
func (c *Chart) Version() (*string, error) {
	git, err := git.New(c.path)
	if err != nil {
		return nil, err
	}

	tag, err := git.Tag()
	if err != nil {
		tag = "0.1.0"
		log.Infof("unable to find any git tags using %s", tag)
	}

	commits, err := git.Commits()
	if err != nil {
		return nil, err
	}

	sha, err := git.Sha()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch git sha %s", err)
	}

	branch, err := git.Branch()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch git branch %s", err)
	}

	ver, err := semver.NewVersion(tag)
	if err != nil {
		return nil, fmt.Errorf("%s %s", tag, err)
	}

	version := *ver
	if commits != 0 || branch != "master" {
		version = version.IncPatch()
	}

	prerel := ""
	if branch != "master" {
		prerel = branch
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
	pathArray := strings.Split(c.TagPath, ".")
	for i := 0; i < len(pathArray); i++ {
		k := pathArray[i]
		objMap := obj.(map[interface{}]interface{})
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
