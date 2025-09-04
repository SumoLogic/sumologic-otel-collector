ARG ALPINE_VERSION=3.22.1
ARG GOMPLATE_VERSION=v4.3.3-alpine
ARG OCB_VERSION=0.133.0

# NOTE: The next few build stages only exists to allow the contrib context to be
# passed in separately from the component's build context. It is passed in by
# Docker Compose's additional_contexts feature. While it is not strictly
# required to define them here, it makes the linter happy and serves as a good
# place in the to document this behaviour.
FROM scratch AS ca-certificates
FROM scratch AS contrib

# Process the OpenTelemetry Collector build manifest template
FROM gomplate/gomplate:${GOMPLATE_VERSION} AS tmpl
ARG OCB_VERSION
WORKDIR /
COPY builder-config.tmpl.yaml .
ENV OTEL_VERSION=${OCB_VERSION}
RUN gomplate -f builder-config.tmpl.yaml -o builder-config.yaml

# Build the OpenTelemetry Collector
FROM otel/opentelemetry-collector-builder:${OCB_VERSION} AS ocb
WORKDIR /
# NOTE: The contrib build stage is passed in by Docker Compose.
# # hadolint ignore=DL3022
COPY --from=contrib . contrib
COPY --from=tmpl /builder-config.yaml .
RUN ocb --config builder-config.yaml

# The final image
FROM scratch
ARG USER_UID=10001
ARG USER_GID=10001
USER ${USER_UID}:${USER_GID}
COPY --from=ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=ocb --chmod=755 /tmp/otelcol/otelcol /otelcol
ENTRYPOINT ["/otelcol"]
CMD ["--config", "/etc/otelcol/config.yaml"]
EXPOSE 4317 4318 55679
