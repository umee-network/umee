FROM golang:1.18-alpine
ARG GAIA_VERSION=v5.0.7

ENV PACKAGES curl make git libc-dev bash gcc linux-headers
RUN apk add --no-cache $PACKAGES

WORKDIR /downloads/
RUN git clone https://github.com/cosmos/gaia.git
RUN cd gaia && git checkout ${GAIA_VERSION} && make build && cp ./build/gaiad /usr/local/bin/

EXPOSE 26656 26657 1317 9090
ENTRYPOINT ["gaiad", "start"]
