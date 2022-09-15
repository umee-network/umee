# Fetch base packages
FROM golang:1.19-alpine AS base-builder
ENV PACKAGES make git libc-dev gcc linux-headers
RUN apk add --no-cache $PACKAGES

# Fetch base umee packages
FROM base-builder AS umee-base-builder
ENV PACKAGES curl bash eudev-dev python3
RUN apk add --no-cache $PACKAGES

# Compile the umeed binary
FROM umee-base-builder AS umeed-builder
WORKDIR /src/app/
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 make install
RUN cd price-feeder && make install

# Fetch peggo (gravity bridge) binary
FROM base-builder AS peggo-builder
ARG PEGGO_VERSION=v0.3.0
WORKDIR /downloads/
RUN git clone https://github.com/umee-network/peggo.git
RUN cd peggo && git checkout ${PEGGO_VERSION} && make build && cp ./build/peggo /usr/local/bin/

# Add to a distroless container
FROM gcr.io/distroless/cc:debug
COPY --from=umeed-builder /go/bin/umeed /usr/local/bin/
COPY --from=umeed-builder /go/bin/price-feeder /usr/local/bin/
COPY --from=peggo-builder /usr/local/bin/peggo /usr/local/bin/
EXPOSE 26656 26657 1317 9090 7171

ENTRYPOINT ["umeed", "start"]
