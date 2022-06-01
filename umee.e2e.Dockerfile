ARG IMG_TAG=latest

# Compile the umeed binary
FROM golang:1.18-alpine AS umeed-builder
WORKDIR /src/app/
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3
RUN apk add --no-cache $PACKAGES
RUN CGO_ENABLED=0 make install
RUN cd price-feeder && make install

# Fetch peggo (gravity bridge) binary
FROM golang:1.18-alpine AS peggo-builder
ARG PEGGO_VERSION=v0.3.0
ENV PACKAGES make git libc-dev gcc linux-headers
RUN apk add --no-cache $PACKAGES
WORKDIR /downloads/
RUN git clone https://github.com/umee-network/peggo.git
RUN cd peggo && git checkout ${PEGGO_VERSION} && make build && cp ./build/peggo /usr/local/bin/

# Add to a distroless container
FROM gcr.io/distroless/cc:$IMG_TAG
ARG IMG_TAG
COPY --from=umeed-builder /go/bin/umeed /usr/local/bin/
COPY --from=umeed-builder /go/bin/price-feeder /usr/local/bin/
COPY --from=peggo-builder /usr/local/bin/peggo /usr/local/bin/
EXPOSE 26656 26657 1317 9090 7171

ENTRYPOINT ["umeed", "start"]
