#################################################################################
# Path variables & stop implicit rules
#################################################################################

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(patsubst %/,%,$(dir $(mkfile_path)))
parent_dir := $(abspath $(current_dir)/..)
pkg_dir := $(abspath $(parent_dir)/..)
root_dir := $(abspath $(pkg_dir)/..)
toolchains_dir := $(abspath $(root_dir)/toolchains)

Makefile:
	@true

$(pkg_dir)/common.mk:
	@true

$(toolchains_dir)/Makefile:
	@true

#################################################################################
# Override variables & include makefile dependencies
#################################################################################

# Include commmon.mk from the package directory
include $(pkg_dir)/common.mk

# Include Makefile from the toolchains directory
include $(toolchains_dir)/Makefile

#################################################################################
# Variables
#################################################################################

BINARY_NAME ?= otelcol-config
GO ?= go
FIPS_SUFFIX ?= ""

# Builds for darwin need to be built with CGO_ENABLED set to 1 because some
# telegraf plugins that are used within the telegrafreceiver are implemented with
# CGO. Example of this might be the cpu input plugin using gopsutil to the cpu
# stats.
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

# We don't want to strip symbols for FIPS builds, as they're useful for verifying
# the build
ifeq ($(FIPS_SUFFIX),"")
	GOEXPERIMENT=""
	LDFLAGS+= -s -w
# We want to use boringcrypto for fips setup
else
	ifeq ($(shell go env GOOS),windows)
		GOEXPERIMENT=systemcrypto
		GOTAGS := $(GOTAGS) requirefips
		LDFLAGS+= -w
	else
		GOEXPERIMENT=boringcrypto
		LDFLAGS+= -w
	endif
endif

# We want to enforce go dns for all types of binaries. The FIPS binaries are
# built with CGO enabled and use the CGO resolver. This has caused a DNS
# resolution error for .local domains in our K8s containers.
# ref: https://pkg.go.dev/net#hdr-Name_Resolution
GOTAGS := $(GOTAGS) netgo

# Settings for macOS builds
ifeq ($(HOST_OS),Darwin)
	AC_USERNAME ?= otel-mac-certs@sumologic.com
	AC_PASSWORD ?= $(shell echo $$AC_PASSWORD)
	DEVELOPER_TEAM_ID ?= 8F28635Z7X
	DEVELOPER_SIGNING_IDENTITY ?= Developer ID Application: Sumo Logic, Inc. ($(DEVELOPER_TEAM_ID))
	DMG_VOLUME_NAME ?= Sumo Logic OpenTelemetry Collector Config Tool
endif

################################################################################
# Build targets
################################################################################

.PHONY: build
build:
	CGO_ENABLED=$(CGO_ENABLED) \
	GOEXPERIMENT=$(GOEXPERIMENT) \
	$(GO) build -v \
	-tags "$(GOTAGS)" \
	-ldflags="$(LDFLAGS)" \
	-trimpath \
	-o $(BINARY_NAME) .
	chmod +x $(BINARY_NAME)

.PHONY: build-fips
build-fips: docker-golang-cross-builder
	docker run -it --rm \
	-v golang-cross-builder:/root/build \
	-v $(PWD)/../../../:/root/workspace \
	-w /root/workspace \
	--tmpfs /tmp \
	--entrypoint ./pkg/tools/otelcol-config/build-fips/entrypoint.sh \
	golang-cross-builder

################################################################################
# Binary targets
################################################################################

.PHONY: otelcol-config-darwin_amd64
otelcol-config-darwin_amd64:
	GOOS=darwin  GOARCH=amd64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-darwin_amd64

.PHONY: otelcol-config-darwin_arm64
otelcol-config-darwin_arm64:
	GOOS=darwin  GOARCH=arm64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-darwin_arm64

.PHONY: otelcol-config-linux_amd64
otelcol-config-linux_amd64:
	GOOS=linux   GOARCH=amd64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-linux_amd64

.PHONY: otelcol-config-linux_arm64
otelcol-config-linux_arm64:
	GOOS=linux   GOARCH=arm64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-linux_arm64

.PHONY: otelcol-config-linux_arm
otelcol-config-linux_arm:
	GOOS=linux   GOARCH=arm $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-linux_arm

.PHONY: otelcol-config-windows_amd64
otelcol-config-windows_amd64:
	GOOS=windows GOARCH=amd64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-windows_amd64

################################################################################
# FIPS binary targets
################################################################################

.PHONY: otelcol-config-linux-fips_amd64
otelcol-config-linux-fips_amd64:
	$(MAKE) otelcol-config-linux_amd64 \
	FIPS_SUFFIX="-fips" \
	CGO_ENABLED="1" \
	CC="$(toolchains_dir)toolchain_amd64/bin/x86_64-linux-musl-gcc" \
	LDFLAGS="-linkmode external -extldflags '-static'"

.PHONY: otelcol-config-linux-fips_arm64
otelcol-config-linux-fips_arm64:
	$(MAKE) otelcol-config-linux_arm64 \
	FIPS_SUFFIX="-fips" \
	CGO_ENABLED="1" \
	CC="$(toolchains_dir)toolchain_arm64/bin/aarch64-linux-musl-gcc" \
	LDFLAGS="-linkmode external -extldflags '-static'"

################################################################################
# Disk imaging targets
################################################################################

.PHONY: otelcol-config-dmg
otelcol-config-dmg:
	$(call check_defined, BIN_PATH, BIN_PATH must be set to the path of the otelcol-config binary)
	$(call check_defined, DMG_PATH, DMG_PATH must be set to the path of the DMG file that will be created)
	$(eval TMP_DIR ?= $(shell mktemp -d))
	cp "$(BIN_PATH)" "$(TMP_DIR)/otelcol-config"
	hdiutil create "$(DMG_PATH)" -ov -volname "$(DMG_VOLUME_NAME)" -fs APFS -format UDZO -srcfolder "$(TMP_DIR)"

################################################################################
# Signing targets
################################################################################

.PHONY: darwin-sign-otelcol-config
darwin-sign-otelcol-config: BIN_PATH = $(BINARY_NAME)
darwin-sign-otelcol-config: DMG_BIN_FILENAME = otelcol-config
darwin-sign-otelcol-config: darwin-sign

.PHONY: darwin_amd64-sign
darwin_amd64-sign: BINARY_NAME=otelcol-config-darwin_amd64
darwin_amd64-sign: darwin-sign-otelcol-config

.PHONY: darwin_arm64-sign
darwin_arm64-sign: BINARY_NAME=otelcol-config-darwin_arm64
darwin_arm64-sign: darwin-sign-otelcol-config
