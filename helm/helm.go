package helm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
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
	chart := findChart(dir)
	if chart == nil {
		return nil, errors.New("unable to find a Chart.yaml")
	}

	chart.TagPath = "image.tag"
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
