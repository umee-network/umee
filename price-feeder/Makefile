BUILD_DIR ?= $(CURDIR)/build
COMMIT    := $(shell git log -1 --format='%H')

all: test-unit install

.PHONY: all

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
##                              Build / Install                              ##
###############################################################################

build_tags += $(BUILD_TAGS)

ldflags = -X github.com/umee-network/umee/price-feeder/cmd.Version=$(VERSION) \
		  -X github.com/umee-network/umee/price-feeder/cmd.Commit=$(COMMIT)

ldflags += $(LDFLAGS)
ifeq ($(LINK_STATICALLY),true)
  ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

build: go.sum
	@echo "--> Building..."
	go build -mod=readonly -o $(BUILD_DIR)/ $(BUILD_FLAGS) ./...

install: go.sum
	@echo "--> Installing..."
	go install -mod=readonly $(BUILD_FLAGS) ./...

.PHONY: build install

###############################################################################
##                              Tests & Linting                              ##
###############################################################################

test-unit:
	@echo "--> Running tests"
	@go test -mod=readonly -race ./... -v

.PHONY: test-unit

lint:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m

.PHONY: lint
