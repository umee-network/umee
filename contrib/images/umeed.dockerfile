# Stage-1: build
# We use Debian Bullseye rather then Alpine because Alpine has problem building libwasmvm
# - requires to download libwasmvm_muslc from external source. Build with glibc is straightforward.
FROM golang:1.23-bookworm AS builder

WORKDIR /src/
# optimization: if go.sum didn't change, docker will use cached image
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN LEDGER_ENABLED=false BUILD_TAGS=badgerdb make install


# Stage-2: copy binary and required artifacts to a fresh image
# we need to use debian compatible system.
FROM ubuntu:24.04
EXPOSE 26656 26657 1317 9090
# Run umeed by default, omit entrypoint to ease using container with CLI
CMD ["umeed"]
STOPSIGNAL SIGTERM

RUN apt-get update && apt-get install ca-certificates -y \
    && groupadd umee && useradd -g umee -m umee

COPY --from=builder /go/bin/umeed /usr/local/bin/
COPY --from=builder /go/pkg/mod/github.com/\!cosm\!wasm/wasmvm\@v*/internal/api/libwasmvm.*.so /usr/lib/
