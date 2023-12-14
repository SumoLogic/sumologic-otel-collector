#!/usr/bin/env sh

cp -r /root/workspace /root/build
cd /root/build

# Install builder
cd otelcolbuilder || exit 1
mkdir "${HOME}/bin"
export PATH="${HOME}/bin:${PATH}"
make install-builder

# Build otelcol-sumo
make otelcol-sumo-linux-fips_amd64
make otelcol-sumo-linux-fips_arm64
