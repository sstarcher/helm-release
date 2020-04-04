package helm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/sstarcher/helm-release/version"
	"gopkg.in/yaml.v2"
)

// DefaultTagPath is the default path to the image tag in a helm values.yaml
var DefaultTagPath = "image.tag"

// ChartInterface for helm
type ChartInterface interface {
	version.Getter
	version.Setter
	UpdateImageVersion(string) error
	UpdateChart(*semver.Version, string) error
}

// Chart defines a Helm Chart
type Chart struct {
	Name    string
	path    string
	tagPath string
}

// New finds the helm chart in the directory and returns a Chart object
func New(dir string, tagPath *string) (ChartInterface, error) {
	chart := findChart(dir)
	if chart == nil {
		return nil, errors.New("unable to find a Chart.yaml")
	}

	chart.tagPath = DefaultTagPath
	if tagPath != nil {
		chart.tagPath = *tagPath
	}

	return chart, nil
}

// findChart looks for all helm charts under the given path
func findChart(dir string) (chart *Chart) {
	_ = filepath.Walk(dir, func(file string, f os.FileInfo, err error) error {
		if strings.HasSuffix(file, "Chart.yaml") {
			dir = path.Dir(file)
			chart = &Chart{
				path: dir,
				Name: path.Base(dir),
			}
			return errors.New("found first")
		}
		return nil
	})
	return chart
}

// Set updates the version of the helm chart
func (c *Chart) Set(version *semver.Version) error {
	return c.UpdateChart(version, "")
}

// UpdateChart updates the version of the helm chart and appVersion
func (c *Chart) UpdateChart(version *semver.Version, imageVersion string) error {
	fmt.Println("running stuff")
	var config map[interface{}]interface{}
	source, err := ioutil.ReadFile(c.path + "/Chart.yaml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", config)

	config["version"] = version.String()
	if imageVersion != "" {
		if _, ok := config["appVersion"]; ok {
			config["appVersion"] = imageVersion
		}
	}

	fmt.Printf("%v\n", config)

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
func (c *Chart) UpdateImageVersion(imageVersion string) error {
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

	pathArray := strings.Split(c.tagPath, ".")
	for i := 0; i < len(pathArray); i++ {
		k := pathArray[i]
		switch objMap := obj.(type) {
		case map[interface{}]interface{}:
			if i == len(pathArray)-1 {
				if _, ok := objMap[k]; !ok {
					return fmt.Errorf("final key in path does not exist %s for path %s", k, c.tagPath)
				}
				objMap[k] = imageVersion
				break
			}
			obj = objMap[k]
			if obj == nil {
				return fmt.Errorf("unable to process %s", c.tagPath)
			}
		default:
			return fmt.Errorf("while processing key[%s] for path[%s] expected a map, but got %T", k, c.tagPath, objMap)
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

// Get version from Chart.yaml
func (c *Chart) Get() (*semver.Version, error) {
	var config map[interface{}]interface{}
	source, err := ioutil.ReadFile(c.path + "/Chart.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		return nil, err
	}

	if config["version"] == nil {
		return nil, fmt.Errorf("%s/Chart.yaml is missing a version", c.path)
	}

	return semver.NewVersion(config["version"].(string))
}

// NextVersion from current version
func (c *Chart) NextVersion(nextType *version.NextType) (*semver.Version, error) {
	ver, err := c.Get()
	if err != nil {
		return nil, err
	}

	return version.NextVersion(ver, nextType)
}
