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

FROM scratch
ARG BUILD_TAG=latest
ENV TAG $BUILD_TAG
ARG USER_UID=10001
USER ${USER_UID}
ENV HOME /etc/otel/

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=otelcol /otelcol-sumo /otelcol-sumo
COPY --from=directories --chown=${USER_UID}:${USER_UID} /etc/otel/ /etc/otel/

# journaldreceiver dependencies
COPY --from=systemd /bin/journalctl /bin/journalctl
COPY --from=systemd /lib/systemd/libsystemd-shared-247.so /lib/systemd/libsystemd-shared-247.so
COPY --from=systemd /lib/x86_64-linux-gnu/libaudit.so.1 /lib/x86_64-linux-gnu/libaudit.so.1
COPY --from=systemd /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=systemd /lib/x86_64-linux-gnu/libcap-ng.so.0 /lib/x86_64-linux-gnu/libcap-ng.so.0
COPY --from=systemd /lib/x86_64-linux-gnu/libcap.so.2 /lib/x86_64-linux-gnu/libcap.so.2
COPY --from=systemd /lib/x86_64-linux-gnu/libcrypt.so.1 /lib/x86_64-linux-gnu/libcrypt.so.1
COPY --from=systemd /lib/x86_64-linux-gnu/libdl.so.2 /lib/x86_64-linux-gnu/libdl.so.2
COPY --from=systemd /lib/x86_64-linux-gnu/libgpg-error.so.0 /lib/x86_64-linux-gnu/libgpg-error.so.0
COPY --from=systemd /lib/x86_64-linux-gnu/liblzma.so.5 /lib/x86_64-linux-gnu/liblzma.so.5
COPY --from=systemd /lib/x86_64-linux-gnu/libpam.so.0 /lib/x86_64-linux-gnu/libpam.so.0
COPY --from=systemd /lib/x86_64-linux-gnu/libpthread.so.0 /lib/x86_64-linux-gnu/libpthread.so.0
COPY --from=systemd /lib/x86_64-linux-gnu/librt.so.1 /lib/x86_64-linux-gnu/librt.so.1
COPY --from=systemd /lib/x86_64-linux-gnu/libselinux.so.1 /lib/x86_64-linux-gnu/libselinux.so.1
COPY --from=systemd /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2
COPY --from=systemd /usr/lib/x86_64-linux-gnu/libacl.so.1 /usr/lib/x86_64-linux-gnu/libacl.so.1
COPY --from=systemd /usr/lib/x86_64-linux-gnu/libblkid.so.1 /usr/lib/x86_64-linux-gnu/libblkid.so.1
COPY --from=systemd /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1 /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1
COPY --from=systemd /usr/lib/x86_64-linux-gnu/libgcrypt.so.20 /usr/lib/x86_64-linux-gnu/libgcrypt.so.20
COPY --from=systemd /usr/lib/x86_64-linux-gnu/libip4tc.so.2 /usr/lib/x86_64-linux-gnu/libip4tc.so.2
COPY --from=systemd /usr/lib/x86_64-linux-gnu/libkmod.so.2 /usr/lib/x86_64-linux-gnu/libkmod.so.2
COPY --from=systemd /usr/lib/x86_64-linux-gnu/liblz4.so.1 /usr/lib/x86_64-linux-gnu/liblz4.so.1
COPY --from=systemd /usr/lib/x86_64-linux-gnu/libmount.so.1 /usr/lib/x86_64-linux-gnu/libmount.so.1
COPY --from=systemd /usr/lib/x86_64-linux-gnu/libpcre2-8.so.0 /usr/lib/x86_64-linux-gnu/libpcre2-8.so.0
COPY --from=systemd /usr/lib/x86_64-linux-gnu/libseccomp.so.2 /usr/lib/x86_64-linux-gnu/libseccomp.so.2
COPY --from=systemd /usr/lib/x86_64-linux-gnu/libzstd.so.1 /usr/lib/x86_64-linux-gnu/libzstd.so.1

EXPOSE 55680 55679
ENTRYPOINT ["/otelcol-sumo"]
CMD ["--config", "/etc/otel/config.yaml"]
