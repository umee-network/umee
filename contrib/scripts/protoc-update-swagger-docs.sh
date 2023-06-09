#!/usr/bin/env bash
set -eo pipefail

SWAGGER_DIR=./swagger
SWAGGER_UI_DIR=${SWAGGER_DIR}/swagger-ui

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
  # replacing default swagger with our generated swagger yaml file
  sed -i 's+https://petstore.swagger.io/v2/swagger.json+./swagger.yaml+g' ${SWAGGER_DIR}/swagger-ui/swagger-initializer.js
fi

# move generated swagger yaml file to swagger-ui directory
cp ${SWAGGER_DIR}/swagger.yaml ${SWAGGER_DIR}/swagger-ui/

# install statik if not present
go install github.com/rakyll/statik

# generate statik golang code using updated swagger-ui directory
statik -src=${SWAGGER_DIR}/swagger-ui -dest=${SWAGGER_DIR} -f -m

# log whether or not the swagger directory was updated
if [ -n "$(git status ${SWAGGER_DIR} --porcelain)" ]; then
  echo "Swagger is updated"
else
  echo "Swagger in sync"
fi