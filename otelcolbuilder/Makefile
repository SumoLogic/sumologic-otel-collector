################################################################################
# Path variables & stop implicit rules
################################################################################

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(patsubst %/,%,$(dir $(mkfile_path)))
parent_dir := $(abspath $(current_dir)/..)
toolchains_dir := $(abspath $(parent_dir)/toolchains)

Makefile:
	@true

$(parent_dir)/common.mk:
	@true

$(toolchains_dir)/Makefile:
	@true

################################################################################
# Override variables & include makefile dependencies
################################################################################

# A list of one or more space-delimited paths to files that are used to determine
# if the cache should be invalidated. When these files change, the cache is
# invalidated. This is used by GitHub Actions.
CACHE_DEPENDENCY_PATHS ?= .otelcol-builder.yaml

# Include common.mk from the parent directory
include $(parent_dir)/common.mk

# Include the toolchains Makefile
include $(toolchains_dir)/Makefile

################################################################################
# Variables
################################################################################

BINARY_NAME ?= otelcol-sumo
OCB_VERSION ?= 0.134.0
FIPS_SUFFIX ?=
OTC_RELEASE_REPO ?= github.com/open-telemetry/opentelemetry-collector-releases
GO ?= go
GOTAGS ?=
BIN_PATH ?= $(HOME)/bin

# Check if ocb is installed & determine ocb version
OCB_BIN_NAME := ocb$(BINARY_EXT)
OCB := $(shell which $(OCB_BIN_NAME) || echo)
ifneq ($(OCB),)
	INSTALLED_OCB_VERSION := $(shell $(OCB) version 2>&1)
endif

# OCB Download URL
OCB_FILE := ocb_$(OCB_VERSION)_$(GOOS)_$(GOARCH)$(BINARY_EXT)
OCB_RELEASE := cmd/builder/v$(OCB_VERSION)
OCB_BASE_URL := https://$(OTC_RELEASE_REPO)/releases/download
OCB_URL := $(OCB_BASE_URL)/$(OCB_RELEASE)/$(OCB_FILE)

# Check if yq is installed
YQ_BIN_NAME := yq$(BINARY_EXT)
YQ := $(shell command -v $(YQ_BIN_NAME) 2>&1)

# Settings for macOS builds
ifeq ($(HOST_OS),Darwin)
	AC_USERNAME ?= otel-mac-certs@sumologic.com
	AC_PASSWORD ?= $(shell echo $$AC_PASSWORD)
	DEVELOPER_TEAM_ID ?= 8F28635Z7X
	DEVELOPER_SIGNING_IDENTITY ?= Developer ID Application: Sumo Logic, Inc. ($(DEVELOPER_TEAM_ID))
	DMG_VOLUME_NAME ?= Sumo Logic Distribution for OpenTelemetry Collector
endif

# Add additional LDFLAGS
# ex: to get a staticly executable with CGO use EXTRA_LDFLAGS="-linkmode external -extldflags '-static'"
LDFLAGS := $(EXTRA_LDFLAGS)

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

# We don't want to strip symbols for FIPS builds, as they're useful for verifying the build
ifeq ($(FIPS_SUFFIX),)
	GOEXPERIMENT=
	LDFLAGS+= -s -w
else
# We want to use boringcrypto for fips setup
	ifeq ($(shell go env GOOS),windows)
		GOEXPERIMENT=systemcrypto
		GOTAGS := $(GOTAGS) requirefips
		LDFLAGS+= -w
	else
		GOEXPERIMENT=boringcrypto
		LDFLAGS+= -w
	endif
endif

# We want to enforce go dns for all types of binaries. The FIPS binaries are built with CGO enabled and use
# the CGO resolver. This has caused a DNS resolution error for .local domains in our K8s containers.
# ref: https://pkg.go.dev/net#hdr-Name_Resolution
GOTAGS := $(GOTAGS) netgo

.PHONY: require-yq
require-yq:
ifeq ($(YQ),)
	@$(error "The yq command is required but was not found")
endif

.PHONY: _install-bin
_install-bin:
	@mkdir -p $(BIN_PATH)
	curl -L -o $(BIN_PATH)/$(OCB_BIN_NAME) $(OCB_URL)
	@chmod +x "$(BIN_PATH)/$(OCB_BIN_NAME)"
	@$(MAKE) ensure-correct-ocb-version

.PHONY: ocb-version
ocb-version:
	@echo $(OCB_VERSION)

.PHONY: install-ocb
install-ocb:
	@echo "Installing $(OTC_RELEASE_REPO)/cmd/builder@v$(OCB_VERSION)... (in $(BIN_PATH))"
	@$(MAKE) _install-bin

.PHONY: ensure-ocb-installed
ensure-ocb-installed:
ifeq ($(OCB),)
	@$(error ocb executable was not found in PATH. \
		Please check that "$(BIN_PATH)" can be found in your PATH \
		environment variable and that "$(BIN_PATH)/$(OCB_BIN_NAME)" \
		exists. If ocb is missing it can be installed by running \
		'make install-ocb' from the otelcolbuilder directory.\
	)
endif

.PHONY: ensure-correct-ocb-version
ensure-correct-ocb-version: ensure-ocb-installed
ifneq ($(lastword $(INSTALLED_OCB_VERSION)),$(OCB_VERSION))
	@$(error Installed ocb version \
		"$(INSTALLED_OCB_VERSION)" \
		does not match the requested "$(OCB_VERSION)". The correct version of \
		ocb can be installed by running 'make install-ocb' from the \
		otelcolbuilder directory.\
	)
else
	@printf "Installed ocb (%s) is at the correct version %s\n" \
		"$(shell type $(OCB) | cut -d' ' -f 3)" \
		"$(OCB_VERSION)"
endif

.PHONY: _ocb_build
_ocb_build: ensure-correct-ocb-version require-yq
_ocb_build: SKIP_COMPILATION ?= false
_ocb_build:
	$(eval VERSION ?= $(shell BUILDER_CONFIG=.otelcol-builder.yaml ../ci/get_version.sh core))
	$(shell yq e ".dist.version = \"${VERSION}${FIPS_SUFFIX}\"" -i .otelcol-builder.yaml)
	$(eval SUMOVERSION ?= $(shell yq '.dist.version' .otelcol-builder.yaml))
	@echo "Building otelcol-sumo version: $(SUMOVERSION)"
# Need to specify go path because otherwise opentelemetry-collector-builder
# uses /usr/bin/go which on Github Actions is using preinstalled 1.15.12 by default.
	CGO_ENABLED=$(CGO_ENABLED) $(OCB) \
		--config .otelcol-builder.yaml \
		--skip-compilation=$(SKIP_COMPILATION)

.PHONY: _gobuild
_gobuild:
	(cd cmd && \
		CGO_ENABLED=$(CGO_ENABLED) GOEXPERIMENT=$(GOEXPERIMENT) go build -v \
		-tags "$(GOTAGS)" \
		-ldflags="$(LDFLAGS)" \
		-trimpath \
		-o ./$(BINARY_NAME)$(BINARY_EXT) . \
	)
	chmod +x ./cmd/$(BINARY_NAME)$(BINARY_EXT)

.PHONY: _gobuild_debug
_gobuild_debug:
	(cd cmd && \
		CGO_ENABLED=$(CGO_ENABLED) GOEXPERIMENT=$(GOEXPERIMENT) go build -v \
		-tags "$(GOTAGS)" \
		-race \
		-gcflags "all=-N -l" \
		-o ./$(BINARY_NAME)-debug$(BINARY_EXT) . \
	)
	chmod +x ./cmd/$(BINARY_NAME)-debug$(BINARY_EXT)

.PHONY: build
build:
# Since ocb doesn't allow adding build tags, let's just skip the compilation
# and run 'go build ...' by hand.
	$(MAKE) generate-sources
	$(MAKE) _gobuild

.PHONY: build-debug
build-debug:
# Since ocb doesn't allow adding build tags, let's just skip the compilation
# and run 'go build ...' by hand.
	$(MAKE) generate-sources
	$(MAKE) _gobuild_debug

.PHONY: build-fips
build-fips: docker-golang-cross-builder
	docker run -it --rm \
	-v $(DOCKER_CROSS_BUILDER_VOLUME):/root/build \
	-v $(PWD)/../:/root/workspace \
	-w /root/workspace \
	--tmpfs /tmp \
	--entrypoint /root/workspace/otelcolbuilder/build-fips/entrypoint.sh \
	golang-cross-builder

.PHONY: generate-sources
generate-sources:
	@$(MAKE) _ocb_build SKIP_COMPILATION=true
# the changes below allow us to use a custom config provider
	patch cmd/main.go cmd/00_main.go.patch
	patch cmd/main.go cmd/01_main.go.patch
	patch cmd/main_windows.go cmd/main_windows.go.patch

.PHONY: test
test:
	@$(MAKE) generate-sources
	CGO_ENABLED=${CGO_ENABLED} GOEXPERIMENT=$(GOEXPERIMENT) $(MAKE) -C cmd test

.PHONY: lint
lint: ensure-correct-ocb-version generate-sources
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

.PHONY: otelcol-sumo-linux_amd64
otelcol-sumo-linux_amd64:
	GOOS=linux   GOARCH=amd64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-linux_amd64

.PHONY: otelcol-sumo-linux_arm
otelcol-sumo-linux_arm:
	GOOS=linux   GOARCH=arm $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-linux_arm

.PHONY: otelcol-sumo-linux_arm64
otelcol-sumo-linux_arm64:
	GOOS=linux   GOARCH=arm64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-linux_arm64

.PHONY: otelcol-sumo-windows_amd64
otelcol-sumo-windows_amd64:
	GOOS=windows GOARCH=amd64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-windows_amd64

.PHONY: otelcol-sumo-linux-fips_arm64
otelcol-sumo-linux-fips_arm64:
	$(MAKE) otelcol-sumo-linux_arm64 \
		FIPS_SUFFIX=-fips \
		CGO_ENABLED=1 \
		CC="$(toolchains_dir)/toolchain_arm64/bin/aarch64-linux-musl-gcc" \
		EXTRA_LDFLAGS="-linkmode external -extldflags '-static'"

.PHONY: otelcol-sumo-linux-fips_amd64
otelcol-sumo-linux-fips_amd64:
	$(MAKE) otelcol-sumo-linux_amd64 \
		FIPS_SUFFIX=-fips \
		CGO_ENABLED=1 \
		CC="$(toolchains_dir)/toolchain_amd64/bin/x86_64-linux-musl-gcc" \
		EXTRA_LDFLAGS="-linkmode external -extldflags '-static'"

################################################################################
# Signing targets
################################################################################

.PHONY: darwin-sign-otelcol
darwin-sign-otelcol: BIN_PATH = cmd/$(BINARY_NAME)
darwin-sign-otelcol: DMG_BIN_FILENAME = otelcol-sumo
darwin-sign-otelcol: darwin-sign

.PHONY: darwin_amd64-sign
darwin_amd64-sign: BINARY_NAME=otelcol-sumo-darwin_amd64
darwin_amd64-sign: darwin-sign-otelcol

.PHONY: darwin_arm64-sign
darwin_arm64-sign: BINARY_NAME=otelcol-sumo-darwin_arm64
darwin_arm64-sign: darwin-sign-otelcol
