BINARY_NAME ?= otelcol-sumo
BUILDER_VERSION ?= 0.36.0
BUILDER_REPO ?= github.com/open-telemetry/opentelemetry-collector-builder
BUILDER_BIN_PATH ?= ${HOME}/bin/opentelemetry-collector-builder
VERSION ?= "$(shell git describe --tags --abbrev=10)"
GO ?= go

# Builds for darwin need to be built with CGO_ENABLED set to 1 because some telegraf
# plugins that are used within the telegrafreceiver are implemented with CGO.
# Example of this might be the cpu input plugin using gopsutil to the cpu stats.
#
# https://github.com/shirou/gopsutil/blob/7ea80628/cpu/cpu_darwin_nocgo.go
# https://github.com/shirou/gopsutil/blob/7ea80628/cpu/cpu_darwin.go
ifeq ($(shell go env GOOS),darwin)
CGO_ENABLED ?= 1
else
# CGO_ENABLED=0 is used becuase we don't want to rely on c libs and opentelemetry
# also builds their binaries this way.
# ref: https://github.com/open-telemetry/opentelemetry-collector/blob/4c503ddc/Makefile#L254-L256
CGO_ENABLED ?= 0
endif

.PHONY: install
install:
	go install ${BUILDER_REPO}@v$(BUILDER_VERSION)

.PHONY: install-prego1.16
install-prego1.16:
	GO111MODULE=on go get ${BUILDER_REPO}@v$(BUILDER_VERSION)

.PHONY: _install-bin
_install-bin:
	mkdir -p ${HOME}/bin && \
	curl -L -o $(BUILDER_BIN_PATH) https://${BUILDER_REPO}/releases/download/v$(BUILDER_VERSION)/opentelemetry-collector-builder_$(BUILDER_VERSION)_$(PLATFORM)_amd64 && \
	chmod +x $(BUILDER_BIN_PATH)

.PHONY: install-bin-linux
install-bin-linux:
	$(MAKE) _install-bin PLATFORM=linux

.PHONY: install-bin-darwin
install-bin-darwin:
	$(MAKE) _install-bin PLATFORM=darwin

.PHONY: _builder
_builder:
# Need to specify go path because otherwise opentelemetry-collector-builder
# uses /usr/bin/go which on Github Actions is using preinstalled 1.15.12 by default.
	CGO_ENABLED=${CGO_ENABLED} opentelemetry-collector-builder \
		--go ${GO} \
		--version ${VERSION} \
		--config .otelcol-builder.yaml \
		--output-path ./cmd \
		--skip-compilation=${SKIP_COMPILATION} \
		--name $(BINARY_NAME)

.PHONY: build
build:
	$(MAKE) _builder SKIP_COMPILATION=false
	chmod +x ./cmd/$(BINARY_NAME)

.PHONY: generate-sources
generate-sources:
	$(MAKE) _builder SKIP_COMPILATION=true

.PHONY: test
test: install generate-sources
	$(MAKE) -C cmd test

.PHONY: lint
lint: install generate-sources
	$(MAKE) -C cmd lint

.PHONY: otelcol-sumo-all-sys
otelcol-sumo-all-sys: otelcol-sumo-darwin_amd64 otelcol-sumo-linux_amd64 otelcol-sumo-windows_amd64

.PHONY: otelcol-sumo-darwin_amd64
otelcol-sumo-darwin_amd64:
	GOOS=darwin  GOARCH=amd64 $(MAKE) build BINARY_NAME=otelcol-sumo-darwin_amd64

.PHONY: otelcol-sumo-linux_amd64
otelcol-sumo-linux_amd64:
	GOOS=linux   GOARCH=amd64 $(MAKE) build BINARY_NAME=otelcol-sumo-linux_amd64

.PHONY: otelcol-sumo-linux_arm64
otelcol-sumo-linux_arm64:
	GOOS=linux   GOARCH=arm64 $(MAKE) build BINARY_NAME=otelcol-sumo-linux_arm64

.PHONY: otelcol-sumo-linux_arm
otelcol-sumo-linux_arm:
	GOOS=linux   GOARCH=arm $(MAKE) build BINARY_NAME=otelcol-sumo-linux_arm

.PHONY: otelcol-sumo-windows_amd64
otelcol-sumo-windows_amd64:
	GOOS=windows GOARCH=amd64 $(MAKE) build BINARY_NAME=otelcol-sumo-windows_amd64.exe
