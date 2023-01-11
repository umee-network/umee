# Fetch base packages
FROM golang:1.19-alpine AS builder
RUN apk add --no-cache make git libc-dev gcc linux-headers build-base

WORKDIR /src/
# optimization: if go.sum didn't change, docker will use cached image
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Cosmwasm - Download correct libwasmvm version
RUN WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | cut -d ' ' -f 2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a \
      -O /lib/libwasmvm_muslc.a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep $(uname -m) | cut -d ' ' -f 1)
# Build the binary
RUN cd price-feeder && LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make install

## Prepare the final clear binary
FROM alpine:3.17
EXPOSE 7171
STOPSIGNAL SIGTERM
CMD ["price-feeder"]

RUN apk add ca-certificates
COPY --from=builder /go/bin/price-feeder /usr/local/bin/
