# Helm Release

[![CircleCI](https://circleci.com/gh/sstarcher/helm-release.svg?style=shield)](https://circleci.com/gh/sstarcher/helm-release)
[![GitHub release](https://img.shields.io/github/release/sstarcher/helm-release.svg)](https://github.com/sstarcher/helm-release/releases)

Provides simple semantic versioning based from previous git tags.

This allows you to run a single command and package the next version of your chart.  This project allows you to combine your Dockerfile and Helm Chart in a single repository and automatically version the Helm chart on a build.

## Install

You can install a specific release version: 

    $ helm plugin install https://github.com/sstarcher/helm-release --version 0.1.1

## Usage

* helm release CHART - Would determine the next tag for the chart and update the Chart.yaml
* helm release CHART -t 12345 - Would update Chart.yaml and modify values.yaml images.tag to equal 1235

## Release Logic

To describe the release naming process we will use the following nomenclature.
* LAST_TAG - finds the previous tag from the git history using `git describe --tags` 
* NEXT_TAG - uses LAST_TAG and increments the patch by 1
* COMMITS - finds the total number of commits using `git describe --tags`
* TAG - is used when COMMITS has a value of 0 as in the current commit has been specifically tagged
* SHA - uses a short git sha using `git rev-parse --short HEAD` with the `g` prefix removed
* BRANCH - finds the current branch name using `git rev-parse --abbrev-ref HEAD`
  * overridden using BRANCH_NAME environment variable
  * always converted to lowercase
  * strips any characters that don't match - https://semver.org/#spec-item-9

The default version for a branch is `NEXT_TAG-BRANCH-COMMITS+SHA`

### Tags

When COMMITS is equal to 0 we assume the intent is to do a release of the current commit and the version will be the tag itself `TAG+SHA`


### Master branch

The master branch is treated differently from the default and will be `NEXT_TAG-COMMITS+SHA`

### Integrated Support for Jenkins and PR branches

Jenkins uses the environment variable BRANCH_NAME with the value of the PR example `PR-97`.  This will result in a release version of `NEXT_TAG-pr-97-COMMITS+SHA`

## Uninstall

    $ helm plugin remove release

