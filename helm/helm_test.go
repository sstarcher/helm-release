package helm

import (
	"os"
	"strconv"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/sstarcher/helm-release/git"
	"github.com/stretchr/testify/assert"
)

var noTags = "../tests/notags"

func TestFindCharts(t *testing.T) {
	assert := assert.New(t)

	chart := findChart(noTags)
	assert.NotNil(chart)
	assert.Equal("notags", chart.Name)
}

func TestUpdateChart(t *testing.T) {
	assert := assert.New(t)

	chart, err := New(noTags, nil)
	assert.Nil(err)
	assert.NotNil(chart)

	ver, _ := semver.NewVersion("1.1.1")
	err = chart.Set(ver)
	assert.Nil(err)
}

func TestUpdateImage(t *testing.T) {
	assert := assert.New(t)

	chart, err := New(noTags, nil)
	assert.Nil(err)
	assert.NotNil(chart)

	err = chart.UpdateImageVersion("1.1.1")
	assert.Nil(err)
}

func TestUpdateImageInvalidPath(t *testing.T) {
	assert := assert.New(t)

	tag := "invalid"
	chart, err := New(noTags, &tag)
	assert.Nil(err)
	assert.NotNil(chart)

	err = chart.UpdateImageVersion("1.1.1")
	assert.NotNil(err)
}

var versionTests = []struct {
	branch   string
	tag      string
	sha      string
	commits  string
	tagged   bool
	expected string
}{
	{"master", "1.0.0", "0000001", "1", false, "1.0.1-1+0000001"},
	{"master", "1.0.0", "0000002", "0", true, "1.0.0+0000002"},
	{"master", "", "0000003", "1", false, "0.0.2-1+0000003"},
	{"otherBranch", "1.0.0", "0000010", "1", false, "1.0.1-0.otherbranch+0000010"},
	{"otherBranch", "1.0.0", "0000011", "0", true, "1.0.0+0000011"},
	{"weird/branch$$other", "0.1.2", "0000020", "1", false, "0.1.3-0.weird.branch.other+0000020"},
	{"noversion", "", "0000030", "0", true, "0.0.1+0000030"},
}

func TestVersions(t *testing.T) {
	assert := assert.New(t)

	chart, err := New(noTags, nil)
	assert.Nil(err)
	assert.NotNil(chart)

	for _, tt := range versionTests {
		os.Setenv("BRANCH_NAME", tt.branch)
		os.Setenv("LAST_TAG", tt.tag)
		os.Setenv("SHA", tt.sha)
		os.Setenv("COMMITS", tt.commits)
		os.Setenv("IS_TAGGED", strconv.FormatBool(tt.tagged))

		git, err := git.New(".")
		assert.Nil(err)

		actual, err := git.NextVersion(nil)
		assert.NotNil(actual)
		if actual != nil {
			assert.Equal(tt.expected, actual.String())
		} else {
			assert.Fail("nil results")
		}

		os.Unsetenv("BRANCH_NAME")
		os.Unsetenv("LAST_TAG")
		os.Unsetenv("SHA")
		os.Unsetenv("COMMITS")
		os.Unsetenv("IS_TAGGED")
	}
}
