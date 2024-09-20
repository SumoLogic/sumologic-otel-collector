#!/usr/bin/env sh

# Mac security suite is interferring, can't build in bind mount workspace
cp -r /root/workspace/* /root/build/
cd /root/build

# Build otelcol-config
cd pkg/tools/otelcol-config || exit 1
make otelcol-config-linux-fips_amd64
make otelcol-config-linux-fips_arm64

# Copy produced binaries to bind mount workspace
cp otelcol-config-fips-linux_* /root/workspace/pkg/tools/otelcol-config
