FROM informalsystems/hermes:1.4.0

USER root

EXPOSE 3031
ENTRYPOINT ["hermes", "start"]