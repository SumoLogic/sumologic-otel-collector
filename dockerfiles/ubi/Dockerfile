# Build RedHat compliant image
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.10
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG COLLECTOR_BIN=otelcol-sumo

ENV SUMMARY="UBI based sumologic-otel-collector" \
    DESCRIPTION="Sumo Logic Distribution for OpenTelemetry Collector is a Sumo Logic-supported distribution of the OpenTelemetry Collector. It is a single agent to send logs, metrics and traces to Sumo Logic."

LABEL name="sumologic-otel-collector" \
    vendor="Sumo Logic" \
    release="0" \
    summary="$SUMMARY" \
    description="$DESCRIPTION" \
    io.k8s.description="$DESCRIPTION" \
    maintainer="collection@sumologic.com"

RUN echo "I am running on ${BUILDPLATFORM}, building for ${TARGETPLATFORM}"

COPY artifacts/$TARGETPLATFORM/$COLLECTOR_BIN /otelcol-sumo

# This shouldn't be necessary but sometimes we end up with execution bit not set.
# ref: https://github.com/open-telemetry/opentelemetry-collector/issues/1317
RUN chmod 755 /otelcol-sumo

ADD https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/main/LICENSE \
    /licenses/LICENSE

# hadolint ignore=DL3041
RUN microdnf update && microdnf install -y systemd && microdnf clean all

ARG USER_UID=10001
USER ${USER_UID}
ENV HOME=/etc/otel/

ENTRYPOINT ["/otelcol-sumo"]
CMD ["--config", "/etc/otel/config.yaml"]
