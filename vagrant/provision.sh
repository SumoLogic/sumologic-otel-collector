#!/usr/bin/env bash

export BUILDER_VERSION=0.27.0
export GO_VERSION=1.17rc1

# Install opentelemetry-collector-builder
curl -LJ \
    "https://github.com/open-telemetry/opentelemetry-collector-builder/releases/download/v${BUILDER_VERSION}/opentelemetry-collector-builder_${BUILDER_VERSION}_linux_amd64" \
    -o /usr/local/bin/opentelemetry-collector-builder \
    && chmod +x /usr/local/bin/opentelemetry-collector-builder

sudo apt update -y
sudo apt install -y \
    make \
    gcc

# Install Go
curl -LJ "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o go.linux-amd64.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go.linux-amd64.tar.gz \
    && rm go.linux-amd64.tar.gz \
    && ln -s /usr/local/go/bin/go /usr/local/bin
