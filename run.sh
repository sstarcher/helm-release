#!/usr/bin/env bash
set -euo pipefail

CHART=''
VERSION=''
function finish {
    if [ -n "${CHART+x}" ] && [ -n "${VERSION+x}" ]; then
        rm "${CHART}-${VERSION}".tgz > /dev/null 2>&1 || true
    fi
}
trap finish EXIT

main(){
    CHART_PATH=${1:-'../'}
    if [ "$CHART_PATH" == "." ]; then
        CHART_PATH=".."
    fi
    CHART_FILE=$(find ${CHART_PATH} -name Chart.yaml)
    
    DIR=$(dirname "${CHART_FILE}")
    CHART=$(basename "${DIR}")
    cd ${DIR}

    # Check if we are in a git repo
    if git rev-parse --git-dir; then
        echo 'yaay get'
    else
        echo 'booh get'
    fi
    
    echo 'hmm'
    GIT_DESCRIBE=$(git describe --tags || true)
    echo 'after'

    if [ "$GIT_DESCRIBE" == "fatal: No names found, cannot describe anything." ]; then
        echo 'No tags so using 0.1.0'
        LAST_TAG='0.1.0'
        GIT_SHA=$(git rev-parse --short HEAD)
        COMMITS='0'
    else
        GIT_SHA=$(git describe --tags | rev | cut -d'-' -f1 | rev)
        GIT_SHA=${GIT_SHA#?} #strip leading g
        COMMITS=$(git describe --tags | rev | cut -d'-' -f2 | rev)
        LAST_TAG=$(git describe --tags | rev | cut -d'-' -f3 | rev)
    fi
    BRANCH_NAME=${BRANCH_NAME:-$(git rev-parse --abbrev-ref HEAD)}
    # If is PR-X change to lowercase pr.x
    # If is master publish next chart
    # If is develop publish next chart?

    VERSION=""
    if [ "$COMMITS" -eq "1" ]; then
        VERSION=${LAST_TAG}
    else
        VERSION="$(semver bump patch ${GIT_SHA} ${COMMITS} ${LAST_TAG})"
    fi


    echo "CHART $CHART"
    echo $LAST_TAG
    echo $COMMITS
    echo $BRANCH_NAME
    echo $VERSION

    ${HELM_BIN} lint .
    ${HELM_BIN} package . --version ${VERSION} --dependency-update
    echo ${HELM_BIN} s3 push ${CHART}-${VERSION}.tgz syapse
}
main "$@"
