# Docker for e2e testing
# Creates dynamic binaries, by building from the latest version of:
# umeed and release version of peggo

FROM golang:1.19-bullseye AS builder

## Download Peggo
WORKDIR /src
RUN wget https://github.com/umee-network/peggo/releases/download/v1.4.0/peggo-v1.4.0-linux-amd64.tar.gz && \
    tar -xvf peggo-v*

## Download go module dependencies for umeed
WORKDIR /src/umee
COPY go.mod go.sum ./
RUN go mod download

## Download go module dependnecies for price-feeder
WORKDIR /src/umee/price-feeder
COPY price-feeder/go.mod price-feeder/go.sum ./
RUN go mod download

## Build umeed and price-feeder
WORKDIR /src/umee
COPY . .
RUN make install && \
    cd price-feeder && make install

## Prepare the final clear binary
FROM ubuntu:rolling
EXPOSE 26656 26657 1317 9090 7171
ENTRYPOINT ["umeed", "start"]

COPY --from=builder /go/pkg/mod/github.com/\!cosm\!wasm/wasmvm\@v*/internal/api/libwasmvm.*.so /usr/lib/
COPY --from=builder /go/bin/* /usr/local/bin/
COPY --from=builder /src/peggo-v*/peggo /usr/local/bin/
RUN apt-get update && apt-get install ca-certificates -y
