#!/usr/bin/env bash

set -eo pipefail

SWAGGER_DIR=./swagger
SWAGGER_UI_DIR=${SWAGGER_DIR}/swagger-ui

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

# TODO: needs to fix merge issue of cosmos swagger and ibc-go swagger
# SDK_VERSION=$(go list -m -f '{{ .Version }}' github.com/cosmos/cosmos-sdk)
# IBC_VERSION=$(go list -m -f '{{ .Version }}' github.com/cosmos/ibc-go/v6)

# SDK_RAW_URL=https://raw.githubusercontent.com/cosmos/cosmos-sdk/${SDK_VERSION}/client/docs/swagger-ui/swagger.yaml
# IBC_RAW_URL=https://raw.githubusercontent.com/cosmos/ibc-go/${IBC_VERSION}/docs/client/swagger-ui/swagger.yaml

# # download Cosmos SDK swagger yaml file
# echo "SDK version ${SDK_VERSION}"
# wget "${SDK_RAW_URL}" -O ./tmp-swagger-gen/cosmos-sdk-swagger.yaml

# # download IBC swagger yaml file
# echo "IBC version ${IBC_VERSION}"
# wget "${IBC_RAW_URL}" -O ./tmp-swagger-gen/ibc-go-swagger.yaml

# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./swagger/proto-config-gen.json -o ./swagger/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./tmp-swagger-gen

SWAGGER_UI_VERSION=4.18.3
SWAGGER_UI_DOWNLOAD_URL=https://github.com/swagger-api/swagger-ui/archive/refs/tags/v${SWAGGER_UI_VERSION}.zip
SWAGGER_UI_PACKAGE_NAME=${SWAGGER_DIR}/swagger-ui-${SWAGGER_UI_VERSION}

# if swagger-ui does not exist locally, download swagger-ui and move dist directory to
# swagger-ui directory, then remove zip file and unzipped swagger-ui directory
if [ ! -d ${SWAGGER_UI_DIR} ]; then
  # download swagger-ui
  wget ${SWAGGER_UI_DOWNLOAD_URL} -O ${SWAGGER_UI_PACKAGE_NAME}.zip
  # unzip swagger-ui package
  unzip ${SWAGGER_UI_PACKAGE_NAME}.zip -d ${SWAGGER_DIR}
  # move swagger-ui dist directory to swagger-ui directory
  mv ${SWAGGER_UI_PACKAGE_NAME}/dist ${SWAGGER_UI_DIR}
  # remove swagger-ui zip file and unzipped swagger-ui directory
  rm -rf ${SWAGGER_UI_PACKAGE_NAME}.zip ${SWAGGER_UI_PACKAGE_NAME}

  sed -i 's+https://petstore.swagger.io/v2/swagger.json+./swagger.yaml+g' ${SWAGGER_DIR}/swagger-ui/swagger-initializer.js
fi

# move generated swagger yaml file to swagger-ui directory
cp ${SWAGGER_DIR}/swagger.yaml ${SWAGGER_DIR}/swagger-ui/

go install github.com/rakyll/statik

# generate statik golang code using updated swagger-ui directory
statik -src=${SWAGGER_DIR}/swagger-ui -dest=${SWAGGER_DIR} -f -m

# log whether or not the swagger directory was updated
if [ -n "$(git status ${SWAGGER_DIR} --porcelain)" ]; then
  echo "Swagger updated"
else
  echo "Swagger in sync"
fi


