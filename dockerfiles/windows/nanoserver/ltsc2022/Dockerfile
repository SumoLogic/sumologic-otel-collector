FROM mcr.microsoft.com/windows/servercore:ltsc2022 AS core
FROM mcr.microsoft.com/windows/nanoserver:ltsc2022
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG COLLECTOR_BIN="otelcol-sumo"

# netapi32.dll is not included in nanoserver:ltsc2022 and older. It is necessary
# for Go applications to run. Fortunately, the servercore images include
# netapi32.dll so we can copy it to our image based on nanoserver.
COPY --from=core /windows/system32/netapi32.dll /windows/system32/netapi32.dll

# TODO: This appears to be failing intermittently, perhaps as buildkit is still
# buggy on Windows. Attempt to remove this sometime in the future.
#RUN echo "I am running on ${BUILDPLATFORM}, building for ${TARGETPLATFORM}"

COPY /artifacts/$TARGETPLATFORM/$COLLECTOR_BIN.exe /otelcol-sumo.exe
ENTRYPOINT ["/otelcol-sumo.exe"]
CMD ["--config", "/etc/otel/config.yaml"]
