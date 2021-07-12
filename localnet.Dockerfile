FROM golang:1.16-alpine

RUN apk add --no-cache curl jq file make git libc-dev bash gcc linux-headers eudev-dev python3
VOLUME [ "/umeed" ]
WORKDIR /umeed
EXPOSE 26656 26657 1317 9090
COPY contrib/scripts/localnet-wrapper.sh /usr/bin/localnet-wrapper.sh
ENTRYPOINT ["/usr/bin/localnet-wrapper.sh"]
CMD ["start", "--x-crisis-skip-assert-invariants", "--log_format", "plain"]
