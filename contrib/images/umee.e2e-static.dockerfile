# Docker for e2e testing
# Creates static binaries, by building from the latest version of:
# umeed, price-feeder.

FROM golang:1.20-alpine AS builder
ENV PACKAGES make git gcc linux-headers build-base curl
RUN apk add --no-cache $PACKAGES

## Build umeed
WORKDIR /src/umee
# optimization: if go.sum didn't change, docker will use cached image
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Cosmwasm - Download correct libwasmvm version
RUN WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | cut -d ' ' -f 2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a -O /lib/libwasmvm_muslc.a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep $(uname -m) | cut -d ' ' -f 1)

RUN LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make install && \
    cd price-feeder && LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make install


## Prepare the final clear binary
FROM alpine:latest
EXPOSE 26656 26657 1317 9090 7171
ENTRYPOINT ["umeed", "start"]

# no need to copy libwasmvm_muslc.a because we created static
COPY --from=builder /go/bin/* /usr/local/bin/
RUN apk add ca-certificates
