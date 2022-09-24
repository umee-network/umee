#!/bin/bash -eu

# Download the umeed binary with the same version as mainnet and unpack it

# USAGE: ./download-mainnet-umeed.sh

is_macos() {
  [[ "$OSTYPE" == "darwin"* ]]
}

architecture=$(uname -m)

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
MAINNET_VERSION=${MAINNET_VERSION:-"v1.1.2"}

download_mainnet_binary(){
  # Checks for the umeed v1 file
  if [ ! -f "$UMEED_BIN_MAINNET" ]; then
    echo "$UMEED_BIN_MAINNET doesn't exist"

    if [ -z $UMEED_BIN_MAINNET_URL_TARBALL ]; then
      echo You need to set the UMEED_BIN_MAINNET_URL_TARBALL variable
      exit 1
    fi

    UMEED_RELEASES_PATH=$CWD/umeed-releases
    mkdir -p $UMEED_RELEASES_PATH
    wget -c $UMEED_BIN_MAINNET_URL_TARBALL -O - | tar -xz -C $UMEED_RELEASES_PATH

    UMEED_BIN_MAINNET_BASENAME=$(basename $UMEED_BIN_MAINNET_URL_TARBALL .tar.gz)
    UMEED_BIN_MAINNET=$UMEED_RELEASES_PATH/$UMEED_BIN_MAINNET_BASENAME/umeed
  fi
}

mac_mainnet() {
  if [[ "$architecture" == "arm64" ]];then
    UMEED_BIN_MAINNET_URL_TARBALL=${UMEED_BIN_MAINNET_URL_TARBALL:-"https://github.com/umee-network/umee/releases/download/${MAINNET_VERSION}/umeed-${MAINNET_VERSION}-darwin-arm64.tar.gz"}
    UMEED_BIN_MAINNET=${UMEED_BIN_MAINNET:-"$CWD/umeed-releases/umeed-${MAINNET_VERSION}-darwin-arm64/umeed"}
  else
    UMEED_BIN_MAINNET_URL_TARBALL=${UMEED_BIN_MAINNET_URL_TARBALL:-"https://github.com/umee-network/umee/releases/download/${MAINNET_VERSION}/umeed-${MAINNET_VERSION}-darwin-amd64.tar.gz"}
    UMEED_BIN_MAINNET=${UMEED_BIN_MAINNET:-"$CWD/umeed-releases/umeed-${MAINNET_VERSION}-darwin-amd64/umeed"}
  fi
}

linux_mainnet(){
  if [[ "$architecture" == "arm64" ]];then
    UMEED_BIN_MAINNET_URL_TARBALL=${UMEED_BIN_MAINNET_URL_TARBALL:-"https://github.com/umee-network/umee/releases/download/${MAINNET_VERSION}/umeed-${MAINNET_VERSION}-linux-arm64.tar.gz"}
    UMEED_BIN_MAINNET=${UMEED_BIN_MAINNET:-"$CWD/umeed-releases/umeed-${MAINNET_VERSION}-linux-arm64/umeed"}
  else
    UMEED_BIN_MAINNET_URL_TARBALL=${UMEED_BIN_MAINNET_URL_TARBALL:-"https://github.com/umee-network/umee/releases/download/${MAINNET_VERSION}/umeed-${MAINNET_VERSION}-linux-amd64.tar.gz"}
    UMEED_BIN_MAINNET=${UMEED_BIN_MAINNET:-"$CWD/umeed-releases/umeed-${MAINNET_VERSION}-linux-amd64/umeed"}
  fi
}

if is_macos ; then
  mac_mainnet
  download_mainnet_binary
else
  linux_mainnet
  download_mainnet_binary
fi
