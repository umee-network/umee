#!/usr/bin/make -f

BRANCH         := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT         := $(shell git log -1 --format='%H')
BUILD_DIR      ?= $(CURDIR)/build
DIST_DIR       ?= $(CURDIR)/dist
LEDGER_ENABLED ?= true
TM_VERSION     := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::')
DOCKER         := $(shell which docker)
PROJECT_NAME   = $(shell git remote get-url origin | xargs basename -s .git)

TM_URL                = https://raw.githubusercontent.com/tendermint/tendermint/v0.34.19/proto/tendermint
GOGO_PROTO_URL        = https://raw.githubusercontent.com/regen-network/protobuf/cosmos
GOOGLE_PROTOBUF_URL   = https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf
GOOGLE_API_URL      = https://raw.githubusercontent.com/googleapis/googleapis/master/google/api
COSMOS_PROTO_URL      = https://raw.githubusercontent.com/regen-network/cosmos-proto/master
CONFIO_URL            = https://raw.githubusercontent.com/confio/ics23/v0.6.4

TM_CRYPTO_TYPES       = third_party/proto/tendermint/crypto
TM_ABCI_TYPES         = third_party/proto/tendermint/abci
TM_TYPES              = third_party/proto/tendermint/types
TM_VERSION            = third_party/proto/tendermint/version
TM_LIBS               = third_party/proto/tendermint/libs/bits
TM_P2P                = third_party/proto/tendermint/p2p

GOGO_PROTO_TYPES      = third_party/proto/gogoproto
GOOGLE_API_TYPES    = third_party/proto/google/api
GOOGLE_PROTOBUF_TYPES = third_party/proto/google/protobuf
COSMOS_PROTO_TYPES    = third_party/proto/cosmos_proto
# For some reason ibc expects confio proto files to be in the main folder
CONFIO_TYPES          = third_party/proto

###############################################################################
##                                  Version                                  ##
###############################################################################

ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

###############################################################################
##                                   Build                                   ##
###############################################################################

build_tags = netgo

ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=umee \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=umeed \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -X github.com/tendermint/tendermint/version.TMCoreSemVer=$(TM_VERSION)

ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

build: go.sum
	@echo "--> Building..."
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILD_DIR)/ ./...

install: go.sum
	@echo "--> Installing..."
	go install -mod=readonly $(BUILD_FLAGS) ./...

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

go-mod-tidy:
	@contrib/scripts/go-mod-tidy-all.sh

clean:
	@echo "--> Cleaning..."
	@rm -rf $(BUILD_DIR)/**  $(DIST_DIR)/**

.PHONY: install build build-linux clean

###############################################################################
##                                  Docker                                   ##
###############################################################################

docker-build:
	@docker build -t umeenet/umeed-e2e -f umee.e2e.Dockerfile .

docker-build-debug:
	@docker build -t umeenet/umeed-e2e --build-arg IMG_TAG=debug -f umee.e2e.Dockerfile .

docker-push-hermes:
	@cd tests/e2e/docker; docker build -t ghcr.io/umee-network/hermes-e2e:latest -f hermes.Dockerfile .; docker push ghcr.io/umee-network/hermes-e2e:latest

docker-push-gaia:
	@cd tests/e2e/docker; docker build -t ghcr.io/umee-network/gaia-e2e:latest -f gaia.Dockerfile .; docker push ghcr.io/umee-network/gaia-e2e:latest

.PHONY: docker-build docker-build-debug docker-push-hermes docker-push-gaia

###############################################################################
##                              Tests & Linting                              ##
###############################################################################

PACKAGES_UNIT=$(shell go list ./... | grep -v -e '/tests/e2e' -e '/tests/simulation' -e '/tests/network')
PACKAGES_E2E=$(shell go list ./... | grep '/e2e')
TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-cover test-race test-e2e
TEST_COVERAGE_PROFILE=coverage.txt

test-unit: ARGS=-timeout=10m -tags='norace'
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)
test-unit-cover: ARGS=-timeout=10m -tags='norace' -coverprofile=$(TEST_COVERAGE_PROFILE) -covermode=atomic
test-unit-cover: TEST_PACKAGES=$(PACKAGES_UNIT)
test-race: ARGS=-timeout=10m -race
test-race: TEST_PACKAGES=$(PACKAGES_UNIT)
test-e2e: ARGS=-timeout=25m -v
test-e2e: TEST_PACKAGES=$(PACKAGES_E2E)
$(TEST_TARGETS): run-tests

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	@echo "--> Running tests"
	@go test -mod=readonly -json $(ARGS) $(TEST_PACKAGES) | tparse
else
	@echo "--> Running tests"
	@go test -mod=readonly $(ARGS) $(TEST_PACKAGES)
endif

.PHONY: run-tests $(TEST_TARGETS)

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m

cover-html: test-unit-cover
	@echo "--> Opening in the browser"
	@go tool cover -html=$(TEST_COVERAGE_PROFILE)


.PHONY: lint

###############################################################################
##                                Simulations                                ##
###############################################################################

SIMAPP = ./tests/simulation

test-sim-non-determinism:
	@echo "Running non-determinism application simulations..."
	@go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h

test-sim-multi-seed-short:
	@echo "Running short multi-seed application simulations. This may take a while!"
	@runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 10 TestFullAppSimulation

test-sim-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	@go test -mod=readonly $(SIMAPP) -benchmem -bench=BenchmarkFullAppSimulation -run=NOOP \
	-Enabled=true -NumBlocks=1000 -BlockSize=200 -Period=1 -Commit=true -Seed=57 -v -timeout 24h

.PHONY: \
test-sim-non-determinism \
test-sim-multi-seed-short \
test-sim-benchmark-invariants

###############################################################################
##                                 Protobuf                                  ##
###############################################################################

DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf

protoVer=v0.2
protoImageName=tendermintdev/sdk-proto-gen:$(protoVer)
containerProtoGen=$(PROJECT_NAME)-proto-gen-$(protoVer)
containerProtoGenAny=$(PROJECT_NAME)-proto-gen-any-$(protoVer)
containerProtoGenSwagger=$(PROJECT_NAME)-proto-gen-swagger-$(protoVer)
containerProtoFmt=$(PROJECT_NAME)-proto-fmt-$(protoVer)

proto-all: proto-update-deps proto-format proto-lint proto-gen

proto-update-deps:
	@echo "Updating Protobuf deps"
	@mkdir -p $(GOGO_PROTO_TYPES)
	@curl -sSL $(GOGO_PROTO_URL)/gogoproto/gogo.proto > $(GOGO_PROTO_TYPES)/gogo.proto

	@mkdir -p $(GOOGLE_API_TYPES)
	@curl -sSL $(GOOGLE_API_URL)/annotations.proto > $(GOOGLE_API_TYPES)/annotations.proto
	@curl -sSL $(GOOGLE_API_URL)/http.proto > $(GOOGLE_API_TYPES)/http.proto

	@mkdir -p $(GOOGLE_PROTOBUF_TYPES)
	@curl -sSL $(GOOGLE_PROTOBUF_URL)/descriptor.proto > $(GOOGLE_PROTOBUF_TYPES)/descriptor.proto

	@mkdir -p $(COSMOS_PROTO_TYPES)
	@curl -sSL $(COSMOS_PROTO_URL)/cosmos.proto > $(COSMOS_PROTO_TYPES)/cosmos.proto


## Importing of tendermint protobuf definitions currently requires the
## use of `sed` in order to build properly with cosmos-sdk's proto file layout
## (which is the standard Buf.build FILE_LAYOUT)
## Issue link: https://github.com/tendermint/tendermint/issues/5021
	@mkdir -p $(TM_ABCI_TYPES)
	@curl -sSL $(TM_URL)/abci/types.proto > $(TM_ABCI_TYPES)/types.proto

	@mkdir -p $(TM_VERSION)
	@curl -sSL $(TM_URL)/version/types.proto > $(TM_VERSION)/types.proto

	@mkdir -p $(TM_TYPES)
	@curl -sSL $(TM_URL)/types/types.proto > $(TM_TYPES)/types.proto
	@curl -sSL $(TM_URL)/types/evidence.proto > $(TM_TYPES)/evidence.proto

	@curl -sSL $(TM_URL)/types/params.proto > $(TM_TYPES)/params.proto
	@curl -sSL $(TM_URL)/types/validator.proto > $(TM_TYPES)/validator.proto

	@curl -sSL $(TM_URL)/types/block.proto > $(TM_TYPES)/block.proto

	@mkdir -p $(TM_CRYPTO_TYPES)
	@curl -sSL $(TM_URL)/crypto/proof.proto > $(TM_CRYPTO_TYPES)/proof.proto
	@curl -sSL $(TM_URL)/crypto/keys.proto > $(TM_CRYPTO_TYPES)/keys.proto

	@mkdir -p $(TM_LIBS)
	@curl -sSL $(TM_URL)/libs/bits/types.proto > $(TM_LIBS)/types.proto

	@mkdir -p $(TM_P2P)
	@curl -sSL $(TM_URL)/p2p/types.proto > $(TM_P2P)/types.proto

	@mkdir -p $(CONFIO_TYPES)
	@curl -sSL $(CONFIO_URL)/proofs.proto > $(CONFIO_TYPES)/proofs.proto
## insert go package option into proofs.proto file
## Issue link: https://github.com/confio/ics23/issues/32
	@./contrib/scripts/sed.sh $(CONFIO_TYPES)/proofs.proto

	@./contrib/scripts/proto-copy-cosmos-sdk.sh

proto-gen:
	@echo "Generating Protobuf files"
	@DOCKER_BUILDKIT=1 docker build -t ${containerProtoGen} -f ./Dockerfile.protocgen .
	@$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(containerProtoGen) sh ./contrib/scripts/protocgen.sh


proto-format:
	@echo "Formatting Protobuf files"
	@$(DOCKER) run --rm -v $(CURDIR):/workspace \
	--workdir /workspace tendermintdev/docker-build-proto \
	find ./ -not -path "./third_party/*" -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@echo "Linting Protobuf files"
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against https://github.com/umee-network/umee.git#branch=main

.PHONY: proto-all proto-gen proto-format proto-lint proto-check-breaking proto-update-deps
