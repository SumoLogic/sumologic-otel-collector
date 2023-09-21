#!/usr/bin/env sh

git config --global --add safe.directory /github/workspace

# Install Go
url="https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz"
echo "Downloading ${url}"
curl -Lo go.tar.gz "$url"
tar -zxvf go.tar.gz -C /usr/local
export PATH="/usr/local/go/bin:${PATH}"

# Install builder
cd otelcolbuilder || exit 1
mkdir "${HOME}/bin"
export PATH="${HOME}/bin:${PATH}"
make install-builder

# Build otelcol-sumo
make otelcol-sumo-linux_amd64 FIPS_SUFFIX="-fips" CGO_ENABLED="1"
