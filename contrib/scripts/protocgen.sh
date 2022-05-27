#!/usr/bin/env bash

set -eo pipefail

cd proto
proto_dirs=$(find ./umee -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep go_package $file &> /dev/null ; then
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

cd ..

# after the proto files have been generated add them to the the repo
# in the proper location. Then, remove the ephemeral tree used for generation
cp -r github.com/umee-network/umee/v2/* .
rm -rf github.com

go mod tidy
