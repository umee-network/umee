#!/usr/bin/make -f

BRANCH         := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT         := $(shell git log -1 --format='%H')
BUILD_DIR      ?= $(CURDIR)/build
LEDGER_ENABLED ?= true
TM_VERSION     := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::')

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

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))
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

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

build: go.sum
	@echo "--> Building..."
	CGO_ENABLED=0 go build -mod=readonly -o $(BUILD_DIR)/ $(BUILD_FLAGS) ./...

install: go.sum
	@echo "--> Installing..."
	CGO_ENABLED=0 go install -mod=readonly $(BUILD_FLAGS) ./...

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

clean:
	@echo "--> Cleaning..."
	@rm -rf $(BUILD_DIR)/**

.PHONY: install build build-linux clean

###############################################################################
##                                  Docker                                   ##
###############################################################################

docker-build:
	@docker build -t umeenetwork/umeed .

docker-localnet-build:
	@docker build -t umeenetwork/umeed-localnet --file localnet.Dockerfile .

.PHONY: docker-build docker-localnet-build

###############################################################################
##                                 Localnet                                  ##
###############################################################################

localnet-start: build-linux localnet-stop
  # start a local Umee network if a network configuration does not already exist
	@if ! [ -f $(BUILD_DIR)/node0/umeed/config/genesis.json ]; then \
    docker run --rm -v $(BUILD_DIR):/umeed:Z umeenetwork/umeed-localnet \
    localnet --num-validators=4 --starting-ip-address=192.168.30.2 --keyring-backend=test -o .; \
  fi
	@docker-compose -f docker-compose.localnet.yaml up -d

localnet-stop:
	@docker-compose -f docker-compose.localnet.yaml down

.PHONY: localnet-stop localnet-start

###############################################################################
##                              Tests & Linting                              ##
###############################################################################

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run

.PHONY: lint
