FROM golang:1.17-alpine
ARG GAIA_VERSION=v5.0.7

ENV PACKAGES curl make git libc-dev bash gcc linux-headers
RUN apk add --no-cache $PACKAGES

ADD https://github.com/cosmos/gaia/releases/download/${GAIA_VERSION}/gaiad-${GAIA_VERSION}-linux-arm64 /usr/local/bin/gaiad
RUN chmod +x /usr/local/bin/gaiad

EXPOSE 26656 26657 1317 9090
ENTRYPOINT ["gaiad", "start"]
