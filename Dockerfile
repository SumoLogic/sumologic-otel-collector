FROM alpine:3.16.0 as otelcol
COPY otelcol-sumo /
# This shouldn't be necessary but sometimes we end up with execution bit not set.
# ref: https://github.com/open-telemetry/opentelemetry-collector/issues/1317
RUN chmod 755 /otelcol-sumo

FROM alpine:3.16.0 as certs
RUN apk --update add ca-certificates

FROM alpine:3.16.0 as directories
RUN mkdir /etc/otel/

FROM debian:11.3 as systemd
RUN apt update && apt install -y systemd
# prepare package with journald and it's dependencies
# h stands for dereference of symbolic links
RUN tar czhf journalctl.tar.gz /bin/journalctl $(ldd /bin/journalctl | grep -oP "\/.*? ")
# extract package to /output so it can be take as base for scratch image
RUN mkdir /output && tar xf /journalctl.tar.gz --directory /output

FROM scratch
ARG BUILD_TAG=latest
ENV TAG $BUILD_TAG
ARG USER_UID=10001
USER ${USER_UID}
ENV HOME /etc/otel/

COPY --from=systemd /output/ /
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=otelcol /otelcol-sumo /otelcol-sumo
COPY --from=directories --chown=${USER_UID}:${USER_UID} /etc/otel/ /etc/otel/

ENTRYPOINT ["/otelcol-sumo"]
CMD ["--config", "/etc/otel/config.yaml"]
