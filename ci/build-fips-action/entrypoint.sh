#!/usr/bin/env sh

git config --global --add safe.directory /github/workspace

# Install builder
cd otelcolbuilder || exit 1
mkdir "${HOME}/bin"
export PATH="${HOME}/bin:${PATH}"
make install-builder

# Build otelcol-sumo
make otelcol-sumo-linux_amd64 FIPS_SUFFIX="-fips" CGO_ENABLED="1"
