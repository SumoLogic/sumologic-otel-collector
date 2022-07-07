BINARY_NAME ?= otelcol-sumo
BUILDER_VERSION ?= 0.55.0
FIPS_SUFFIX ?= ""
BUILDER_REPO ?= github.com/open-telemetry/opentelemetry-collector
BUILDER_BIN_NAME ?= opentelemetry-collector-builder
BUILDER_BIN_PATH ?= $(HOME)/bin
INSTALLED_BUILDER_VERSION := $(shell opentelemetry-collector-builder version 2>&1)
GO ?= go
OS ?= $(shell uname -s | tr A-Z a-z)

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

.PHONY: _install-bin
_install-bin:
	@mkdir -p /$(HOME)/bin
	curl -L -o $(BUILDER_BIN_PATH)/$(BUILDER_BIN_NAME) https://$(BUILDER_REPO)/releases/download/v$(BUILDER_VERSION)/ocb_$(BUILDER_VERSION)_$(PLATFORM)_amd64
	@chmod +x $(BUILDER_BIN_PATH)/$(BUILDER_BIN_NAME)
	@$(MAKE) ensure-correct-builder-version

.PHONY: install-builder
install-builder:
	@echo "Installing $(BUILDER_REPO)/cmd/builder@v$(BUILDER_VERSION)... (in $(BUILDER_BIN_PATH))"
	@$(MAKE) _install-bin PLATFORM=$(OS)


.PHONY: ensure-correct-builder-version
ensure-correct-builder-version:
ifneq ($(lastword $(INSTALLED_BUILDER_VERSION)),$(BUILDER_VERSION))
	@$(error Installed opentelemetry-collector-builder version \
		"$(INSTALLED_BUILDER_VERSION)" \
		does not match the requested "$(BUILDER_VERSION)" \
		Please check if "$(BUILDER_BIN_PATH)" can be found in your PATH \
		and if not, then install it using 'make install-builder' from otelcolbuilder's directory\
	)
else
	@printf "Installed opentelemetry-collector-builder (%s) is at the correct version %s\n" \
		"$(shell type opentelemetry-collector-builder | cut -d' ' -f 3)" \
		"$(BUILDER_VERSION)"
endif

.PHONY: _builder
_builder:
	$(eval VERSION ?= $(shell git describe --tags --abbrev=10 --match "v[0-9]*"))
# Need to specify go path because otherwise opentelemetry-collector-builder
# uses /usr/bin/go which on Github Actions is using preinstalled 1.15.12 by default.
	CGO_ENABLED=$(CGO_ENABLED) $(BUILDER_BIN_PATH)/$(BUILDER_BIN_NAME) \
		--go $(GO) \
		--version "$(VERSION)$(FIPS_SUFFIX)" \
		--config .otelcol-builder.yaml \
		--output-path ./cmd \
		--skip-compilation=$(SKIP_COMPILATION)

.PHONY: _gobuild
_gobuild:
	(cd cmd && \
		CGO_ENABLED=$(CGO_ENABLED) go build -v \
		-tags enable_unstable \
		-ldflags="-s -w" \
		-trimpath \
		-o ./$(BINARY_NAME) . \
	)
	chmod +x ./cmd/$(BINARY_NAME)

.PHONY: _gobuild_debug
_gobuild_debug:
	(cd cmd && \
		CGO_ENABLED=$(CGO_ENABLED) go build -v \
		-tags enable_unstable \
		-race \
		-gcflags "all=-N -l" \
		-o ./$(BINARY_NAME)-debug . \
	)
	chmod +x ./cmd/$(BINARY_NAME)-debug

.PHONY: build
build: ensure-correct-builder-version
# Since builder doesn't allow adding build tags, let's just skip the compilation
# and run 'go build ...' by hand.
	@$(MAKE) generate-sources
	@$(MAKE) _gobuild

.PHONY: build-debug
build-debug: ensure-correct-builder-version
# Since builder doesn't allow adding build tags, let's just skip the compilation
# and run 'go build ...' by hand.
	@$(MAKE) generate-sources
	@$(MAKE) _gobuild_debug

.PHONY: generate-sources
generate-sources:
	@$(MAKE) _builder SKIP_COMPILATION=true

.PHONY: test
test:
	@$(MAKE) ensure-correct-builder-version || $(MAKE) install-builder
	@$(MAKE) generate-sources
	@$(MAKE) -C cmd test

.PHONY: lint
lint: install-builder generate-sources
	@$(MAKE) -C cmd lint

.PHONY: mod-download-all
mod-download-all:
	@echo "No need to run go mod download or tidy on generated sources in $(PWD)"

.PHONY: otelcol-sumo-all-sys
otelcol-sumo-all-sys: otelcol-sumo-darwin_amd64 otelcol-sumo-darwin_arm64 otelcol-sumo-linux_amd64 otelcol-sumo-windows_amd64

.PHONY: otelcol-sumo-darwin_amd64
otelcol-sumo-darwin_amd64:
	GOOS=darwin  GOARCH=amd64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-darwin_amd64

.PHONY: otelcol-sumo-darwin_arm64
otelcol-sumo-darwin_arm64:
	GOOS=darwin  GOARCH=arm64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-darwin_arm64

.PHONY: darwin_amd64-sign
darwin_amd64-sign:
	gon -log-level=debug .gon_config_amd64.json

.PHONY: fips-darwin_amd64-sign
fips-darwin_amd64-sign:
	gon -log-level=debug .gon_config_fips_amd64.json

.PHONY: darwin_arm64-sign
darwin_arm64-sign:
	gon -log-level=debug .gon_config_arm64.json

.PHONY: fips-darwin_arm64-sign
fips-darwin_arm64-sign:
	gon -log-level=debug .gon_config_fips_arm64.json

.PHONY: otelcol-sumo-linux_amd64
otelcol-sumo-linux_amd64:
	GOOS=linux   GOARCH=amd64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-linux_amd64

.PHONY: otelcol-sumo-linux_arm64
otelcol-sumo-linux_arm64:
	GOOS=linux   GOARCH=arm64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-linux_arm64

.PHONY: otelcol-sumo-linux_arm
otelcol-sumo-linux_arm:
	GOOS=linux   GOARCH=arm $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-linux_arm

.PHONY: otelcol-sumo-windows_amd64
otelcol-sumo-windows_amd64:
	GOOS=windows GOARCH=amd64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-windows_amd64.exe
