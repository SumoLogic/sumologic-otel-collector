FROM mcr.microsoft.com/windows/nanoserver:ltsc2025
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG COLLECTOR_BIN="otelcol-sumo"

# TODO: This appears to be failing intermittently, perhaps as buildkit is still
# buggy on Windows. Attempt to remove this sometime in the future.
#RUN echo "I am running on ${BUILDPLATFORM}, building for ${TARGETPLATFORM}"

COPY /artifacts/$TARGETPLATFORM/$COLLECTOR_BIN.exe /otelcol-sumo.exe
ENTRYPOINT ["/otelcol-sumo.exe"]
CMD ["--config", "/etc/otel/config.yaml"]
