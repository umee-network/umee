# Docker for e2e testing
# Creates dynamic binaries, by building from the latest version of umeed

FROM golang:1.19-bullseye AS builder
ARG EXPERIMENTAL=false
ENV EXPERIMENTAL $EXPERIMENTAL

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
RUN if [ "$EXPERIMENTAL" = "true" ] ; then echo "Installing experimental build";else echo "Installing stable build";fi
RUN make install 
RUN cd price-feeder && make install

## Prepare the final clear binary
FROM ubuntu:rolling
EXPOSE 26656 26657 1317 9090 7171
ENTRYPOINT ["umeed", "start"]

COPY --from=builder /go/pkg/mod/github.com/\!cosm\!wasm/wasmvm\@v*/internal/api/libwasmvm.*.so /usr/lib/
COPY --from=builder /go/bin/* /usr/local/bin/
RUN apt-get update && apt-get install ca-certificates -y
