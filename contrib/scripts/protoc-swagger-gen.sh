#!/usr/bin/env bash

set -eo pipefail

if ! [ -x "$(command -v swagger-combine )" ]; then
  echo 'Error: swagger-combine is not installed. Install with $~ npm i -g swagger-combine' >&2
  exit 1
fi

mkdir -p ./tmp-swagger-gen
cd proto
proto_dirs=$(find ./umee -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    buf generate --template buf.gen.swagger.yaml $query_file
  fi
done

cd ..
# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./swagger/proto-config-gen.json -o ./swagger/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./tmp-swagger-gen
