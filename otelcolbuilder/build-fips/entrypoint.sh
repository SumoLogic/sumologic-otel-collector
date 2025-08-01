#!/usr/bin/env sh

# Mac security suite is interferring, can't build in bind mount workspace
cp -r /root/workspace/* /root/build/
cd /root/build

# Install ocb
cd otelcolbuilder || exit 1
mkdir "${HOME}/bin"
export PATH="${HOME}/bin:${PATH}"
make install-ocb

# Build otelcol-sumo
make otelcol-sumo-linux-fips_amd64
make otelcol-sumo-linux-fips_arm64

# Copy produced binaries to bind mount workspace
cp cmd/otelcol-sumo-fips-linux_* /root/workspace/otelcolbuilder/cmd/
