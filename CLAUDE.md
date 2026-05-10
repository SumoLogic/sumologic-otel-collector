# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Is

A Sumo Logic distribution of the OpenTelemetry Collector (`otelcol-sumo`). It bundles upstream OTel components with custom Sumo Logic-specific processors, receivers, extensions, and config providers, compiled into a single agent binary.

## Multi-Module Repository

This is a **multi-module Go repository**. Each component under `pkg/` has its own `go.mod`. The collector binary module is at `otelcolbuilder/cmd/go.mod`. When working across modules, create a `go.work` at the repo root including the modules you need.

## Build Commands

```bash
# Build everything (collector binary + otelcol-config tool)
make build

# Build only the collector binary
make -C otelcolbuilder build

# Cross-platform builds
make -C otelcolbuilder otelcol-sumo-linux_amd64
make -C otelcolbuilder otelcol-sumo-darwin_arm64
make -C otelcolbuilder otelcol-sumo-all-sys   # all platforms

# Install the OTel Collector Builder (ocb) — required before first build
make -C otelcolbuilder install-ocb            # installs to ~/bin/ocb
```

## Test Commands

```bash
# Run all tests across all modules
make gotest

# Run tests for a single module
cd pkg/processor/cascadingfilterprocessor && make test
# or from repo root:
make pkg/processor/cascadingfilterprocessor/test

# Run tests in otelcolbuilder/cmd
make -C otelcolbuilder test

# Run integration tests (requires the binary to exist, builds it if missing)
make gotest-integration
cd pkg/test && make test-integration
```

## Lint Commands

```bash
# Lint all modules (uses staticcheck)
make golint

# Lint a single module
cd pkg/processor/cascadingfilterprocessor && make lint

# Install staticcheck
make install-staticcheck

# Format code
cd <module-dir> && make fmt    # runs gofmt + goimports

# Other linters
make markdownlint
make yamllint
make shellcheck
make pre-commit-check          # runs all pre-commit hooks
```

## Architecture

### Build System
The binary is built using the **OpenTelemetry Collector Builder (ocb)** v0.151.0, configured by `otelcolbuilder/.otelcol-builder.yaml`. The builder generates Go source files in `otelcolbuilder/cmd/` — **these are committed but should not be edited directly**. After generation, two patches are applied (`otelcolbuilder/00_main.go.patch`, `01_main.go.patch`) to wire in custom config providers.

### Custom Components (`pkg/`)
Sumo Logic-specific components, each an independent Go module:
- **Processors**: `cascadingfilterprocessor`, `sourceprocessor`, `lookupprocessor`, `metricfrequencyprocessor`, `sumologicsyslogprocessor`
- **Receivers**: `activedirectoryinvreceiver`, `jobreceiver`, `rawk8seventsreceiver`
- **Extension**: `opampextension`
- **Config Providers**: `globprovider` (glob patterns), `opampprovider` (OpAMP-based)
- **Tools**: `otelcol-config`, `udpdemux`
- **Integration Tests**: `pkg/test`

### CGO and FIPS
- macOS and Windows: `CGO_ENABLED=1` (required for gopsutil/Telegraf plugins)
- Linux standard: `CGO_ENABLED=0`
- FIPS builds: `GOEXPERIMENT=boringcrypto` (Linux) or `GOEXPERIMENT=systemcrypto` with `-tags requirefips` (Windows), built with musl toolchains

## Development Setup

Required tools:
- **Go 1.25.0**
- **ocb 0.151.0**: `make -C otelcolbuilder install-ocb`
- **yq**: `brew install yq`
- **staticcheck**: `make install-staticcheck`
- **goimports**: `go install golang.org/x/tools/cmd/goimports@latest`
- **macOS only**: `brew install gnu-sed coreutils` (for `gsed`, `gdate` used in build scripts)

## Changelog

Uses [Towncrier](https://towncrier.readthedocs.io/). Add entries as `.changelog/<PR_NUMBER>.<type>.txt`.

```bash
make add-changelog-entry    # interactive entry creation
make check-changelog        # CI check for changelog entry
```

PRs not needing a changelog entry should have the `skip-changelog` GitHub label.
