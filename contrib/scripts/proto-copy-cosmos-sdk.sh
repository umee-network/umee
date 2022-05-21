#!/usr/bin/env bash

cosmos_protos="$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk)/proto"

#copy the cosmos-sdk proto files from the version used in the go.mod into the third party folder
#for information on why find -exec is used in this way see https://github.com/koalaman/shellcheck/wiki/SC2156
find "$cosmos_protos" -maxdepth 1 -mindepth 1 \
  -exec sh -c 'file=$1; basename $1  | xargs -I % -n1 rm -rf "./third_party/proto/%"' _ {} \;
cp -r "$cosmos_protos" ./third_party

ibc_protos="$(go list -f '{{ .Dir }}' -m github.com/cosmos/ibc-go/v2)/proto"
find "$ibc_protos" -maxdepth 1 -mindepth 1 \
  -exec sh -c 'file=$1; basename $1  | xargs -I % -n1 rm -rf "./third_party/proto/%"' _ {} \;
cp -r "$ibc_protos" ./third_party

chmod -R +w ./third_party/proto
