FROM informalsystems/hermes:1.6.0 AS hermes-builder

FROM debian:buster-slim
USER root

COPY --from=hermes-builder /usr/bin/hermes /usr/local/bin/
RUN chmod +x /usr/local/bin/hermes

EXPOSE 3031
ENTRYPOINT ["hermes", "start"]