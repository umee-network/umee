#!/usr/bin/env bash

set -eo pipefail

mkdir -p ./tmp-swagger-gen

# Get the path of the cosmos-sdk repo from go/pkg/mod
proto_dirs=$(find ./proto ./third_party/proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)

for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  proto_files=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ -z "$proto_files" ]]; then
    continue
  fi

  for proto_file in ${proto_files}; do
    buf protoc  \
    -I "proto" \
    -I "third_party/proto" \
    "$proto_file" \
    --swagger_out=./tmp-swagger-gen \
    --swagger_opt=logtostderr=true \
    --swagger_opt=fqn_for_swagger_name=true \
    --swagger_opt=simple_operation_ids=true
  done
done

cd ./client/docs
yarn install
yarn combine
yarn convert
yarn build
cd ../../

# clean swagger files
rm -rf ./tmp-swagger-gen
