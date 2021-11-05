ARG IMG_TAG=latest

# Compile the umeed binary
FROM golang:1.17-alpine AS umeed-builder
WORKDIR /src/app/
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3
RUN apk add --no-cache $PACKAGES
RUN make install

# Fetch peggo (gravity bridge) binary
FROM golang:1.17-alpine AS peggo-builder
# TODO: Use semantic version once tagged & released
ARG PEGGO_VERSION=419921c8f7af4d3e89cad84e2d2b49def0465c1c
ENV PACKAGES make git libc-dev gcc linux-headers
RUN apk add --no-cache $PACKAGES
WORKDIR /downloads/
RUN git clone https://github.com/umee-network/peggo.git
RUN cd peggo && git checkout ${PEGGO_VERSION} && make build && cp ./build/peggo /usr/local/bin/

# Add to a distroless container
FROM gcr.io/distroless/cc:$IMG_TAG
ARG IMG_TAG
COPY --from=umeed-builder /go/bin/umeed /usr/local/bin/
COPY --from=peggo-builder /usr/local/bin/peggo /usr/local/bin/
EXPOSE 26656 26657 1317 9090

ENTRYPOINT ["umeed", "start"]
