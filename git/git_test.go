package git

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dir = "../tests/tags"
var noTagsDir = "../tests/notags"

func TestValidGitRepo(t *testing.T) {
	assert := assert.New(t)

	git := Git{
		directory: dir,
	}

	tag, err := git.tag()
	assert.Nil(err)
	assert.NotNil(tag)

	commits, err := git.commits()
	assert.Nil(err)
	assert.NotNil(commits)

	sha, err := git.sha()
	assert.Len(sha, 7)
}

func TestValidRepoSha(t *testing.T) {
	assert := assert.New(t)

	git := Git{
		directory: dir,
	}

	result, err := git.sha()
	assert.Nil(err)
	assert.NotNil(result)
	assert.Len(result, 7)
}

func TestValidRepoBranch(t *testing.T) {
	assert := assert.New(t)

	git := Git{
		directory: dir,
	}
	result, err := git.branch()
	assert.Nil(err)
	assert.NotEmpty(result)
}

func TestNoGitRepo(t *testing.T) {
	assert := assert.New(t)

	git, err := New("cmd")
	assert.NotNil(err)
	assert.Nil(git)
}

var branchTests = []struct {
	branch   string // input
	expected string // expected result
}{
	{"PR-2", "pr-2"},
	{"test_hello*hmm!!/value", "test.hello.hmm.value"},
}

func TestBranches(t *testing.T) {
	assert := assert.New(t)

	git := Git{
		directory: dir,
	}
	for _, tt := range branchTests {
		os.Setenv("BRANCH_NAME", tt.branch)
		actual, err := git.branch()
		assert.Nil(err)
		assert.Equal(tt.expected, actual)
		os.Unsetenv("BRANCH_NAME")
	}
}

func TestNoTags(t *testing.T) {
	assert := assert.New(t)

	git := Git{
		directory: noTagsDir,
	}

	branch, err := git.branch()
	assert.Nil(err)
	assert.NotNil(branch)

	commits, err := git.commits()
	assert.Nil(err)
	assert.Equal(2, commits)

	tag, err := git.tag()
	assert.NotNil(err)
	assert.Empty(tag)
}
