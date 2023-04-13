#!/usr/bin/make -f

PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
export VERSION := $(shell echo $(shell git describe --tags --always --match "v*") | sed 's/^v//')
export TMVERSION := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::')
export COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build
DIST_DIR ?= $(CURDIR)/dist
MOCKS_DIR = $(CURDIR)/tests/mocks
HTTPS_GIT := https://github.com/umee-network/umee.git
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.0.0-rc8
PROJECT_NAME := umee

# RocksDB is a native dependency, so we don't assume the library is installed.
# Instead, it must be explicitly enabled and we warn when it is not.
ENABLE_ROCKSDB ?= false


###############################################################################
##                                   Build                                   ##
###############################################################################

build_tags = netgo

#  experimental feature
ifeq ($(EXPERIMENTAL),true)
	build_tags += experimental
endif

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

ifeq (secp,$(findstring secp,$(COSMOS_BUILD_OPTIONS)))
  build_tags += libsecp256k1_sdk
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
			-X github.com/tendermint/tendermint/version.TMCoreSemVer=$(TMVERSION)

ifeq ($(ENABLE_ROCKSDB),true)
  BUILD_TAGS += rocksdb
  test_tags += rocksdb
endif

# DB backend selection
ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += gcc
endif
ifeq (badgerdb,$(findstring badgerdb,$(COSMOS_BUILD_OPTIONS)))
  BUILD_TAGS += badgerdb
endif
# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  ifneq ($(ENABLE_ROCKSDB),true)
    $(error Cannot use RocksDB backend unless ENABLE_ROCKSDB=true)
  endif
  CGO_ENABLED=1
endif

ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

# Check for debug option
ifeq (debug,$(findstring debug,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -gcflags "all=-N -l"
endif

all: tools build lint test

echo-build-tags:
	echo ${BUILD_TAGS}

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/
build-linux:
	@if [ "${ENABLE_ROCKSDB}" != "true" ]; then \
		echo "RocksDB support is disabled; to build and test with RocksDB support, set ENABLE_ROCKSDB=true"; fi
	GOOS=linux GOARCH=$(if $(findstring aarch64,$(shell uname -m)) || $(findstring arm64,$(shell uname -m)),arm64,amd64) LEDGER_ENABLED=false $(MAKE) build

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	@if [ "${ENABLE_ROCKSDB}" != "true" ]; then \
		echo "RocksDB support is disabled; to build and test with RocksDB support, set ENABLE_ROCKSDB=true"; fi
	go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

build-experimental: go.sum
	@echo "--> Building Experimental version..."
	EXPERIMENTAL=true $(MAKE) build

build-no_cgo:
	@echo "--> Building static binary with no CGO nor GLIBC dynamic linking..."
	ifneq ($(ENABLE_ROCKSDB),true)
		$(error RocksDB requires CGO)
	endif
	CGO_ENABLED=0 CGO_LDFLAGS="-static" $(MAKE) build


go-mod-tidy:
	@contrib/scripts/go-mod-tidy-all.sh

clean:
	@echo "--> Cleaning..."
	@rm -rf $(BUILD_DIR)/**  $(DIST_DIR)/**

.PHONY: build build-linux build-experimental build-no_cgo clean go-mod-tidy

###############################################################################
###                          Tools & Dependencies                           ###
###############################################################################

go.sum: go.mod
	echo "Ensure dependencies have not been modified ..." >&2
	go mod verify

###############################################################################
##                                  Docker                                   ##
###############################################################################

docker-build:
	@DOCKER_BUILDKIT=1 docker build -t umee-network/umeed-e2e -f contrib/images/umee.e2e.dockerfile .

docker-build-experimental:
	@DOCKER_BUILDKIT=1 docker build -t umee-network/umeed-e2e -f contrib/images/umee.e2e.dockerfile --build-arg EXPERIMENTAL=true .

docker-push-hermes:
	@cd tests/e2e/docker; docker build -t ghcr.io/umee-network/hermes-e2e:latest -f hermes.Dockerfile .; docker push ghcr.io/umee-network/hermes-e2e:latest

docker-push-gaia:
	@cd tests/e2e/docker; docker build -t ghcr.io/umee-network/gaia-e2e:latest -f gaia.Dockerfile .; docker push ghcr.io/umee-network/gaia-e2e:latest

.PHONY: docker-build docker-push-hermes docker-push-gaia

###############################################################################
##                                   Tests                                   ##
###############################################################################

PACKAGES_UNIT=$(shell go list ./... | grep -v -e '/tests/e2e' -e '/tests/simulation' -e '/tests/network')
TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-cover test-race
TEST_COVERAGE_PROFILE=coverage.txt

UNIT_TEST_TAGS = norace
TEST_RACE_TAGS = ""
TEST_E2E_TAGS = "e2e"
TEST_E2E_DEPS = docker-build

ifeq ($(EXPERIMENTAL),true)
	UNIT_TEST_TAGS += experimental
	TEST_RACE_TAGS += experimental
	TEST_E2E_TAGS += experimental
	TEST_E2E_DEPS = docker-build-experimental
endif

test-unit: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)'
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)
test-unit-cover: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)' -coverprofile=$(TEST_COVERAGE_PROFILE) -covermode=atomic
test-unit-cover: TEST_PACKAGES=$(PACKAGES_UNIT)
test-race: ARGS=-timeout=10m -race -tags='$(TEST_RACE_TAGS)'
test-race: TEST_PACKAGES=$(PACKAGES_UNIT)
$(TEST_TARGETS): run-tests

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	@echo "--> Running tests"
	@go test -mod=readonly -json $(ARGS) $(TEST_PACKAGES) | tparse
else
	@echo "--> Running tests"
	@go test -mod=readonly $(ARGS) $(TEST_PACKAGES)
endif

cover-html: test-unit-cover
	@echo "--> Opening in the browser"
	@go tool cover -html=$(TEST_COVERAGE_PROFILE)

.PHONY: cover-html run-tests $(TEST_TARGETS)

# NOTE: when building locally, we need to run: $(MAKE) docker-build
# however we should be able to optimize it:
# https://linear.app/umee/issue/UMEE-463/fix-docker-login-problem-when-running-e2e-tests
test-e2e: $(TEST_E2E_DEPS)
	go test ./tests/e2e/... -mod=readonly -timeout 30m -race -v -tags='$(TEST_E2E_TAGS)'

test-e2e-cov: $(TEST_E2E_DEPS)
	go test ./tests/e2e/... -mod=readonly -timeout 30m -race -v -tags='$(TEST_E2E_TAGS)' -coverpkg=./... -coverprofile=e2e-profile.out -covermode=atomic

test-e2e-clean:
	docker stop umee0 umee1 umee2 umee-gaia-relayer gaiaval0 umee-price-feeder
	docker rm umee0 umee1 umee2 umee-gaia-relayer gaiaval0 umee-price-feeder

$(MOCKS_DIR):
	mkdir -p $(MOCKS_DIR)
mocks: $(MOCKS_DIR)
	sh ./contrib/scripts/mockgen.sh
.PHONY: mocks


###############################################################################
##                                Simulations                                ##
###############################################################################

SIMAPP = ./tests/simulation

# Install the runsim binary
runsim: $(RUNSIM)
$(RUNSIM):
	@echo "Installing runsim..."
	@go install github.com/cosmos/tools/cmd/runsim@v1.0.0

test-app-non-determinism:
	@echo "Running non-determinism application simulations..."
	@go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h

test-app-multi-seed-short: runsim
	@echo "Running short multi-seed application simulations. This may take a while!"
	@runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 10 TestFullAppSimulation

test-app-multi-seed-long: runsim
	@echo "Running long multi-seed application simulation. This may take awhile!"
	@runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 500 50 TestFullAppSimulation

test-app-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	@runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 5 TestAppImportExport

test-app-after-import: runsim
	@echo "Running application simulation-after-import. This may take several minutes..."
	@runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 5 TestAppSimulationAfterImport

test-app-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	@go test -mod=readonly $(SIMAPP) -benchmem -bench=BenchmarkFullAppSimulation -run=NOOP \
	-Enabled=true -NumBlocks=1000 -BlockSize=200 -Period=1 -Commit=true -Seed=57 -v -timeout 24h

.PHONY: \
test-app-non-determinism \
test-app-multi-seed-short \
test-app-import-export \
test-app-after-import \
test-app-benchmark-invariants


###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_lint_cmd := go run github.com/golangci/golangci-lint/cmd/golangci-lint
revive_cmd := go run github.com/mgechev/revive

# note: on new OSX, might require brew install diffutils
lint:
	@echo "--> Running revive"
	@${revive_cmd} -config .revive.toml -formatter friendly ./...
# todo: many errors to fix in price-feeder
#	@cd price-feeder && $(revive_cmd) -formatter friendly ./...
	@echo "--> Running golangci_lint"
	@${golangci_lint_cmd} run
	@cd price-feeder && $(golangci_lint_cmd) run

lint-fix:
	@echo "--> Running linter to fix the lint issues"
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -path "./tests/mocks/*" -not -name "*.pb.go" -not -name "*.pb.gw.go" -not -name "*.pulsar.go" -not -path "./crypto/keys/secp256k1/*" | xargs gofumpt -w -l
	@${golangci_lint_cmd} run --fix --out-format=tab --issues-exit-code=0 --timeout=8m
	@cd price-feeder && $(golangci_lint_cmd) run --fix --out-format=tab --issues-exit-code=0 --timeout=8m

.PHONY: lint lint-fix

###############################################################################
##                                 Protobuf                                  ##
###############################################################################

DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.8.0

protoVer=v0.7
protoImageName=tendermintdev/sdk-proto-gen:$(protoVer)
containerProtoGen=$(PROJECT_NAME)-proto-gen-$(protoVer)
containerProtoGenSwagger=$(PROJECT_NAME)-proto-gen-swagger-$(protoVer)
containerProtoFmt=$(PROJECT_NAME)-proto-fmt-$(protoVer)

proto-all: proto-format proto-lint proto-gen proto-swagger-gen
.PHONY: proto-all proto-gen proto-lint proto-check-breaking proto-format proto-swagger-gen

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./contrib/scripts/protocgen.sh; fi

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGenSwagger}$$"; then docker start -a $(containerProtoGenSwagger); else docker run --name $(containerProtoGenSwagger) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./contrib/scripts/protoc-swagger-gen.sh; fi

proto-format:
	@echo "Formatting Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./ -name "*.proto" -exec sh -c 'clang-format -style=file -i {}' \; ; fi

proto-lint:
	@echo "Linting Protobuf files"
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@echo "Checking for breaking changes"
	@$(DOCKER_BUF) breaking --against $(HTTPS_GIT)#branch=main
