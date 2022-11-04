# Fetch base packages
FROM golang:1.19-alpine AS builder
ENV PACKAGES make git libc-dev gcc linux-headers
RUN apk add --no-cache $PACKAGES
WORKDIR /src/app/
COPY . .
# Build the binary
RUN CGO_ENABLED=0 make install


FROM alpine:3.14
RUN apk add bash curl jq
COPY --from=builder /go/bin/umeed /usr/local/bin/
EXPOSE 26656 26657 1317 9090
CMD ["umeed", "start"]
STOPSIGNAL SIGTERM
