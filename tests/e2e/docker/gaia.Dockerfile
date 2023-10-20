FROM golang:1.20 AS gaiad-builder
ARG GAIA_VERSION=v13.0.0

# ENV PACKAGES curl make git libc-dev bash gcc linux-headers
# RUN apk add --no-cache $PACKAGES
RUN apt update && apt install curl make git gcc -y

WORKDIR /downloads/
RUN git clone --single-branch --depth 1 --branch ${GAIA_VERSION} https://github.com/cosmos/gaia.git
RUN cd gaia && CGO_ENABLED=0 make build && cp ./build/gaiad /usr/local/bin/

# Final image
FROM alpine:edge
# Install ca-certificates
RUN apk add --update ca-certificates bash
COPY --from=gaiad-builder /usr/local/bin/gaiad /usr/local/bin/
EXPOSE 26656 26657 1317 9090