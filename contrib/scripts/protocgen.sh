#!/usr/bin/env bash

set -eo pipefail

protoc_gen_go() {
  if ! grep "github.com/gogo/protobuf => github.com/umee-network/umee" go.mod &>/dev/null ; then
    echo -e "\tPlease run this command from somewhere inside the umee-core folder."
    return 1
  fi

  go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos 2>/dev/null
}

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
