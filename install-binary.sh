#!/usr/bin/env bash

# Based off of https://github.com/technosophos/helm-template/blob/master/install-binary.sh
PROJECT_NAME="helm-release"
PROJECT_GH="sstarcher/$PROJECT_NAME"

[ -n "${DEBUG+x}" ] && set -x

helm_version=$($HELM_BIN version --client --template '{{ .Version }}')
if [ "${helm_version}" == "<no value>" ]; then
  # Assume this means v2
  helm_version="v2"
fi

if [ "${helm_version:0:2}" == "v2" ]; then
  : "${HELM_PLUGIN_PATH:="$(helm home)/plugins/helm-release"}"

  # Convert the HELM_PLUGIN_PATH to unix if cygpath is
  # available. This is the case when using MSYS2 or Cygwin
  # on Windows where helm returns a Windows path but we
  # need a Unix path
  if type cygpath > /dev/null 2>&1; then
    HELM_PLUGIN_PATH=$(cygpath -u "${HELM_PLUGIN_PATH}")
  fi
elif [ "${helm_version:0:2}" == "v3" ]; then
  eval "$(helm env)"

  HELM_PLUGIN_PATH="${HELM_PLUGINS}"
else
  echo "helm version not supported or not found"
  exit 1;
fi

if [[ $SKIP_BIN_INSTALL == "1" ]]; then
  echo "Skipping binary install"
  exit
fi

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="armv7";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS=$(uname | tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Msys support
    msys*) OS='windows';;
    # Minimalist GNU for Windows
    mingw*) OS='windows';;
    darwin) OS='darwin';;
  esac
}

# verifySupported checks that the os/arch combination is supported for
# binary builds.
verifySupported() {
  local supported="linux-amd64\ndarwin-amd64\nwindows-amd64"
  if ! echo "${supported}" | grep -q "${OS}-${ARCH}"; then
    echo "No prebuild binary for ${OS}-${ARCH}."
    exit 1
  fi

  if ! type "curl" > /dev/null && ! type "wget" > /dev/null; then
    echo "Either curl or wget is required"
    exit 1
  fi
}

# getDownloadURL checks the latest available version.
getDownloadURL() {
  # Use the GitHub API to find the latest version for this project.
  local latest_url="https://api.github.com/repos/$PROJECT_GH/releases/latest"
  if type "curl" > /dev/null; then
    # This is so if you can see if you have hit githubs rate limits
    latest_url_payload=$(curl -s $latest_url)
    DOWNLOAD_URL=$(echo "${latest_url_payload}" | grep $OS | awk '/\"browser_download_url\":/{gsub( /[,\"]/,"", $2); print $2}')
  elif type "wget" > /dev/null; then
    DOWNLOAD_URL=$(wget -q -O - $latest_url | awk '/\"browser_download_url\":/{gsub( /[,\"]/,"", $2); print $2}')
  fi
}

# downloadFile downloads the latest binary package and also the checksum
# for that binary.
downloadFile() {
  PLUGIN_TMP_FILE="/tmp/${PROJECT_NAME}.tgz"
  echo "Downloading $DOWNLOAD_URL"
  if type "curl" > /dev/null; then
    curl -L "$DOWNLOAD_URL" -o "$PLUGIN_TMP_FILE"
  elif type "wget" > /dev/null; then
    wget -q -O "$PLUGIN_TMP_FILE" "$DOWNLOAD_URL"
  fi
}

# installFile verifies the SHA256 for the file, then unpacks and
# installs it.
installFile() {
  HELM_TMP="/tmp/$PROJECT_NAME"
  mkdir -p "$HELM_TMP"
  tar xf "$PLUGIN_TMP_FILE" -C "$HELM_TMP"
  HELM_TMP_BIN="$HELM_TMP/${PROJECT_NAME}"
  echo "Preparing to install into ${HELM_PLUGIN_PATH}"
  DST="$HELM_PLUGIN_PATH"
  if [ "${helm_version:0:2}" == "v3" ]; then
    DST="${HELM_PLUGIN_PATH}/${PROJECT_NAME}/"
  fi
  cp "$HELM_TMP_BIN"* "${DST}"
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    echo "Failed to install $PROJECT_NAME"
    echo -e "\tFor support, go to https://github.com/${PROJECT_GH}."
  fi
  exit $result
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
  set +e
  echo "$PROJECT_NAME installed into $HELM_PLUGIN_PATH/$PROJECT_NAME"
  if [ "${helm_version:0:2}" == "v2" ]; then
    # To avoid to keep track of the Windows suffix,
    # call the plugin assuming it is in the PATH
    PATH=$PATH:$HELM_PLUGIN_PATH
  elif [ "${helm_version:0:2}" == "v3" ]; then
    PATH=$PATH:$HELM_PLUGIN_PATH/$PROJECT_NAME
  else
    echo "helm version not supported or not found"
    exit 1;
  fi
  ${PROJECT_NAME} -h
  set -e
}

# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e
initArch
initOS
verifySupported
getDownloadURL
downloadFile
installFile
testVersion
