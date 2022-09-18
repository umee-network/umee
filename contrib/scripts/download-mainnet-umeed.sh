#!/bin/bash -eu

# Download the umeed binary with the same version as mainnet and unpack it

# USAGE: ./download-mainnet-umeed.sh

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
UMEED_BIN_MAINNET_URL_TARBALL=${UMEED_BIN_MAINNET_URL_TARBALL:-"https://github.com/umee-network/umee/releases/download/v1.1.2/umeed-v1.1.2-linux-amd64.tar.gz"}
UMEED_BIN_MAINNET=${UMEED_BIN_MAINNET:-"$CWD/umeed-releases/umeed-v1.1.2-linux-amd64/umeed"}

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
