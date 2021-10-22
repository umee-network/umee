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

ldflags = -X github.com/umee-network/umee/price-feeder/cmd.Version=$(VERSION) \
		  -X github.com/umee-network/umee/price-feeder/cmd.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(ldflags)'

build: go.sum
	@echo "--> Building..."
	CGO_ENABLED=0 go build -mod=readonly -o $(BUILD_DIR)/ $(BUILD_FLAGS) ./...

install: go.sum
	@echo "--> Installing..."
	CGO_ENABLED=0 go install -mod=readonly $(BUILD_FLAGS) ./...

.PHONY: build install

###############################################################################
##                              Tests & Linting                              ##
###############################################################################

test-unit:
	@echo "--> Running tests"
	@go test -mod=readonly -race ./... -v

.PHONY: test-unit
