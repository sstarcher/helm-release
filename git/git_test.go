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

	git, err := New(dir)
	assert.Nil(err)

	tag, err := git.Tag()
	assert.Nil(err)
	assert.NotNil(tag)

	commits, err := git.Commits()
	assert.Nil(err)
	assert.NotNil(commits)

	sha, err := git.Sha()
	assert.Len(sha, 7)
}

func TestValidRepoSha(t *testing.T) {
	assert := assert.New(t)

	git, err := New(dir)
	assert.Nil(err)
	assert.NotNil(git)

	result, err := git.Sha()
	assert.Nil(err)
	assert.NotNil(result)
	assert.Len(result, 7)
}

func TestValidRepoBranch(t *testing.T) {
	assert := assert.New(t)

	git, err := New(dir)
	result, err := git.Branch()
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

	git, err := New(dir)
	assert.Nil(err)
	for _, tt := range branchTests {
		os.Setenv("BRANCH_NAME", tt.branch)
		actual, err := git.Branch()
		assert.Nil(err)
		assert.Equal(tt.expected, actual)
		os.Unsetenv("BRANCH_NAME")
	}
}

func TestNoTags(t *testing.T) {
	assert := assert.New(t)

	git, err := New(noTagsDir)

	branch, err := git.Branch()
	assert.Nil(err)
	assert.Equal("master", branch)

	commits, err := git.Commits()
	assert.Nil(err)
	assert.Equal(2, commits)

	tag, err := git.Tag()
	assert.NotNil(err)
	assert.Empty(tag)
}
