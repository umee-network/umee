ARG IMG_TAG=latest

# Compile the umeed binary
FROM golang:1.16-alpine AS umeed-builder
WORKDIR /src/app/
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3
RUN apk add --no-cache $PACKAGES
RUN make install

# Fetch gravity bridge binaries and contract
FROM alpine:3.14 as gravity-builder
ARG GRAVITY_VERSION=v0.1.21
# TODO: Enable checksum verification once version stabalizes
# ARG GRAVITY_CLIENT_HASH=b52634eda7cc3c053f266b9b6f772095f5db7848051266b350d68b8136ba2040
# ARG GRAVITY_CONTRACT_DEPLOYER_HASH=0177ed11f11a06d30fc72b6d930800a5e35481a7a8c63dae30ecaf4857c7ed27
# ARG GRAVITY_CONTRACT_HASH=2be39ac537979fc57eb979e4faabce8eec8816d3070ad0f8faa0ba7742d30c53
# ARG GRAVITY_GORC_HASH=74c7107e41270cea2eaf797b0bd0f97d97ef7774dad8a9f3dcf862584372250a
WORKDIR /downloads/
ADD https://github.com/PeggyJV/gravity-bridge/releases/download/${GRAVITY_VERSION}/client .
# RUN echo "$GRAVITY_CLIENT_HASH *client" | sha256sum -c -s
ADD https://github.com/PeggyJV/gravity-bridge/releases/download/${GRAVITY_VERSION}/contract-deployer .
# RUN echo "$GRAVITY_CONTRACT_DEPLOYER_HASH *contract-deployer" | sha256sum -c -s
ADD https://github.com/PeggyJV/gravity-bridge/releases/download/${GRAVITY_VERSION}/Gravity.json .
# RUN echo "$GRAVITY_CONTRACT_HASH *Gravity.json" | sha256sum -c -s
ADD https://github.com/PeggyJV/gravity-bridge/releases/download/${GRAVITY_VERSION}/gorc .
# RUN echo "$GRAVITY_GORC_HASH *gorc" | sha256sum -c -s
RUN chmod +x /downloads/*

# Add to a distroless container
FROM gcr.io/distroless/cc:$IMG_TAG
ARG IMG_TAG
COPY --from=gravity-builder /downloads/client /usr/local/bin/gravity-client
COPY --from=gravity-builder /downloads/contract-deployer /usr/local/bin/
COPY --from=gravity-builder /downloads/Gravity.json /var/data/
COPY --from=gravity-builder /downloads/gorc /usr/local/bin/
COPY --from=umeed-builder /go/bin/umeed /usr/local/bin/
EXPOSE 26656 26657 1317 9090

ENTRYPOINT ["umeed", "start"]
