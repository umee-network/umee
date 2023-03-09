#!/usr/bin/make -f


BRANCH         := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT         := $(shell git log -1 --format='%H')
BUILD_DIR      ?= $(CURDIR)/build
DIST_DIR       ?= $(CURDIR)/dist
LEDGER_ENABLED ?= true
TM_VERSION     := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::')
DOCKER         := $(shell which docker)
PROJECT_NAME   := umee
HTTPS_GIT      := https://github.com/umee-network/umee.git
MOCKS_DIR = $(CURDIR)/tests/mocks

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

whitespace :=
whitespace := $(whitespace) $(whitespace)
comma := ,

build_tags += $(BUILD_TAGS)
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=umee \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=umeed \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -X github.com/tendermint/tendermint/version.TMCoreSemVer=$(TM_VERSION)

ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif

ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

build: go.sum
	@echo "--> Building..."
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILD_DIR)/ ./...

build-experimental: go.sum
	@echo "--> Building Experimental version..."
	EXPERIMENTAL=true $(MAKE) build

build-no_cgo:
	@echo "--> Building static binary with no CGO nor GLIBC dynamic linking..."
	CGO_ENABLED=0 CGO_LDFLAGS="-static" $(MAKE) build

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

install: go.sum
	@echo "--> Installing..."
	go install -mod=readonly $(BUILD_FLAGS) ./...

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
	@docker build -t umee-network/umeed-e2e -f contrib/images/umee.e2e.dockerfile .

docker-build-experimental:
	@docker build -t umee-network/umeed-e2e -f contrib/images/umee.e2e.dockerfile --build-arg EXPERIMENTAL=true . 

docker-push-hermes:
	@cd tests/e2e/docker; docker build -t ghcr.io/umee-network/hermes-e2e:latest -f hermes.Dockerfile .; docker push ghcr.io/umee-network/hermes-e2e:latest

docker-push-gaia:
	@cd tests/e2e/docker; docker build -t ghcr.io/umee-network/gaia-e2e:latest -f gaia.Dockerfile .; docker push ghcr.io/umee-network/gaia-e2e:latest

.PHONY: docker-build docker-push-hermes docker-push-gaia

###############################################################################
##                                   Tests                                   ##
###############################################################################

PACKAGES_UNIT=$(shell go list ./... | grep -v -e '/tests/e2e' -e '/tests/simulation' -e '/tests/network')
PACKAGES_E2E=$(shell go list ./... | grep '/e2e')
TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-cover test-race test-e2e
TEST_COVERAGE_PROFILE=coverage.txt

UNIT_TEST_TAGS = norace 
TEST_RACE_TAGS = ""
TEST_E2E_TAGS = ""

ifeq ($(EXPERIMENTAL),true)
	UNIT_TEST_TAGS	+= experimental
	TEST_RACE_TAGS 	+= experimental
	TEST_E2E_TAGS 	+= experimental
endif

test-unit: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)'
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)
test-unit-cover: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)' -coverprofile=$(TEST_COVERAGE_PROFILE) -covermode=atomic
test-unit-cover: TEST_PACKAGES=$(PACKAGES_UNIT)
test-race: ARGS=-timeout=10m -race -tags='$(TEST_RACE_TAGS)'
test-race: TEST_PACKAGES=$(PACKAGES_UNIT)
test-e2e: ARGS=-timeout=25m -v -tags='$(TEST_E2E_TAGS)'
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

cover-html: test-unit-cover
	@echo "--> Opening in the browser"
	@go tool cover -html=$(TEST_COVERAGE_PROFILE)

.PHONY: cover-html run-tests $(TEST_TARGETS)


$(MOCKS_DIR):
	mkdir -p $(MOCKS_DIR)
mocks: $(MOCKS_DIR)
	sh ./contrib/scripts/mockgen.sh
.PHONY: mocks

###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_lint_cmd=golangci-lint

lint:
	@echo "--> Running linter with revive"
	@go install github.com/mgechev/revive
	@revive -config .revive.toml -formatter friendly ./...
# note: on new OSX, might require brew install diffutils
	@echo "--> Running regular linter"
	@go install mvdan.cc/gofumpt
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint
	@$(golangci_lint_cmd) run
	@cd price-feeder && $(golangci_lint_cmd) run

lint-fix:
	@echo "--> Running linter to fix the lint issues"
	@go install mvdan.cc/gofumpt
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint
	@$(golangci_lint_cmd) run --fix --out-format=tab --issues-exit-code=0 --timeout=8m
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -path "./tests/mocks/*" -not -name "*.pb.go" -not -name "*.pb.gw.go" -not -name "*.pulsar.go" -not -path "./crypto/keys/secp256k1/*" | xargs gofumpt -w -l
	@cd price-feeder && $(golangci_lint_cmd) run --fix --out-format=tab --issues-exit-code=0 --timeout=8m

.PHONY: lint lint-fix

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
##                                 Protobuf                                  ##
###############################################################################

DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.8.0

containerProtoVer=v0.7
containerProtoImage=tendermintdev/sdk-proto-gen:$(containerProtoVer)
containerProtoGen=$(PROJECT_NAME)-proto-gen-$(containerProtoVer)
containerProtoFmt=$(PROJECT_NAME)-proto-fmt-$(containerProtoVer)
containerProtoGenSwagger=$(PROJECT_NAME)-proto-gen-swagger-$(containerProtoVer)

proto-all: proto-format proto-lint proto-gen proto-swagger-gen
.PHONY: proto-all proto-gen proto-lint proto-check-breaking proto-format proto-swagger-gen

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) \
		sh ./contrib/scripts/protocgen.sh; fi

proto-swagger-gen:
	@echo "Generating Swagger of Protobuf"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGenSwagger}$$"; then docker start -a $(containerProtoGenSwagger); else docker run --name $(containerProtoGenSwagger) -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) \
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
