package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"github.com/sstarcher/helm-release/version"
)

type Git struct {
	directory string
}

// New creates the structure
func New(directory string) (version.Getter, error) {
	err := validate(directory)
	if err != nil {
		return nil, err
	}
	return &Git{
		directory: directory,
	}, nil
}

// ~r4.8-40-g56a99c2~
func (g *Git) tag() (tag string, err error) {
	tag, exists := os.LookupEnv("LAST_TAG")
	if exists {
		return
	}

	s, err := g.run("describe", "--tags")
	if err != nil {
		return
	}

	tag, err = g.run("describe", "--tags", "--abbrev=0")
	if err != nil {
		return
	}

	if tag == s {
		return tag, nil
	}

	// TAG-COMMITS-gSHA
	items := strings.Split(s, "-")
	if len(items) < 3 {
		err = fmt.Errorf("unknown response from git describe --tags [%s]", s)
		return
	}

	if tag == "" {
		tag = strings.Join(items[0:len(items)-2], "-")
	}

	return
}

func (g *Git) isTagged() bool {
	tag := os.Getenv("IS_TAGGED")
	if tag != "" {
		b, err := strconv.ParseBool(tag)
		if err != nil {
			return false
		}
		return b
	}

	_, err := g.run("describe", "--exact-match")
	return err == nil
}

func (g *Git) commits() (commits int, err error) {
	commitStr := os.Getenv("COMMITS")
	commits = -1
	if commitStr != "" {
		commits, err = strconv.Atoi(commitStr)
		if err != nil {
			err = fmt.Errorf("expected COMMITS environment variable to be an integer instead of [%s]", commitStr)
			return
		}
	}

	if commits != -1 {
		return
	}

	s, err := g.run("describe", "--tags")
	if err != nil {
		s, err = g.run("rev-list", "--count", "HEAD")
		if err != nil {
			return
		}

		commits, err = strconv.Atoi(s)
		return
	}

	// TAG-COMMITS-gSHA
	items := strings.Split(s, "-")
	if len(items) < 3 {
		return 0, nil // When at the tag we won't have any splits
	}

	commits, err = strconv.Atoi(items[len(items)-2])
	if err != nil {
		return 0, nil // When at the tag we won't have the right format
	}

	return
}

// Sha returns the short git sha of the repo
func (g *Git) sha() (string, error) {
	sha := os.Getenv("SHA")
	if sha != "" {
		if len(sha) == 7 {
			return sha, nil
		}

		log.Warnf("ignoring the environment variable SHA it is not of length 7 [%s]", sha)
		sha = ""
	}

	return g.run("rev-parse", "--short", "HEAD")
}

// Branch returns the branch reference of the repo
func (g *Git) branch() (string, error) {
	branch := os.Getenv("BRANCH_NAME")

	if branch == "" {
		var err error
		branch, err = g.run("rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			return "", err
		}
	}

	reg, err := regexp.Compile("[^0-9A-Za-z-]+")
	if err != nil {
		return "", err
	}
	branch = reg.ReplaceAllString(branch, ".")

	return strings.ToLower(branch), err
}

// Get the semantic version from git
func (g *Git) Get() (*semver.Version, error) {
	tag, err := g.tag()
	if err != nil || tag == "" {
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

// NextVersion determines the correct version
func (g *Git) NextVersion(nextType *version.NextType) (*semver.Version, error) {
	ver, err := g.Get()
	if err != nil {
		return nil, err
	}

	if nextType == nil { // Determine from git history
		commits, err := g.commits()
		if err != nil {
			return nil, err
		}

		sha, err := g.sha()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch git sha %s", err)
		}

		branch, err := g.branch()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch git branch %s", err)
		}

		version := *ver
		prerel := ""
		tagged := g.isTagged()
		if !tagged {
			if branch == "head" && commits == 0 {
				return nil, errors.New("this is likely an light-weight git tag. please use a annotated tag for helm release to function properly")
			}
			version = version.IncPatch()
			if branch != "master" {
				prerel = "0." + branch
			}
		}

		if branch == "master" {
			if commits != 0 {
				if prerel != "" {
					prerel += "."
				}
				prerel += strconv.Itoa(commits)
			}
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
		return &version, err
	}
	return version.NextVersion(ver, nextType)
}

func validate(dir string) error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	err := cmd.Run()

	return err
}

func (g *Git) run(command ...string) (result string, err error) {
	cmd := exec.Command("git", command...)
	cmd.Dir = g.directory
	out, err := cmd.CombinedOutput()

	s := strings.TrimSpace(string(out))
	if s == "fatal: No names found, cannot describe anything." {
		err = errors.New("error processing git repo")
		return
	} else if err != nil {
		return
	}
	return s, err
}
