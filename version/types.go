package version

import (
	"errors"
	"strings"

	"github.com/Masterminds/semver"
)

type NextType string

const (
	Major NextType = "major"
	Minor NextType = "minor"
	Patch NextType = "patch"
)

var nextTypes = map[string]NextType{
	"major": Major,
	"minor": Minor,
	"patch": Patch,
}

func NewNextType(val string) *NextType {
	nextType := nextTypes[strings.ToLower(val)]
	if nextType != "" {
		return &nextType
	}
	return nil
}

// Getter for versions
type Getter interface {
	Get() (*semver.Version, error)
	NextVersion(nextType *NextType) (*semver.Version, error)
}

// Setter for versions
type Setter interface {
	Set(*semver.Version) error
}

// NextVersion takes a semver and updates it to the next version
func NextVersion(version *semver.Version, nextType *NextType) (*semver.Version, error) {
	msg := "major, minor, and patch are the only valid options for Next"

	var nextVersion semver.Version
	if nextType == nil {
		return nil, errors.New(msg)
	} else if *nextType == Major {
		nextVersion = version.IncMajor()
	} else if *nextType == Minor {
		nextVersion = version.IncMinor()
	} else if *nextType == Patch {
		nextVersion = version.IncPatch()
	} else {
		return nil, errors.New(msg)
	}
	return &nextVersion, nil
}
