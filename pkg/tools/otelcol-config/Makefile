include ../../Makefile.Common
include ../../../toolchains/Makefile

TOOLCHAINS_DIR ?= "$(shell '$(MAKE)' print-toolchains-dir)"

BINARY_NAME ?= otelcol-config
GO ?= go

OS ?= $(shell uname -s | tr A-Z a-z)
ARCH ?= $(shell uname -m | sed s/aarch64/arm64/ | sed s/x86_64/amd64/)

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

# We want to enforce go dns for all types of binaries. The FIPS binaries are built with CGO enabled and use
# the CGO resolver. This has caused a DNS resolution error for .local domains in our K8s containers.
# ref: https://pkg.go.dev/net#hdr-Name_Resolution
GOTAGS := $(GOTAGS) netgo

################################################################################
# Functions
################################################################################

# Check that given variables are set and all have non-empty values,
# die with an error otherwise.
#
# PARAMS:
#   1. Variable name(s) to test.
#   2. (optional) Error message to print.
#
# EXAMPLE:
# @:$(call check_defined, ENV_REGION, you must set ENV_REGION=usc1|awsuse2)
#
check_defined = \
	$(strip $(foreach 1,$1, \
		$(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
	$(if $(value $1),, \
		$(error Undefined $1$(if $2, ($2))$(if $(value @), \
			required by target `$@')))

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
	$(MAKE) toolchain-linux_amd64
	$(MAKE) otelcol-config-linux_amd64 \
	FIPS_SUFFIX="-fips" \
	CGO_ENABLED="1" \
	CC="$(TOOLCHAINS_DIR)toolchain_amd64/bin/x86_64-linux-musl-gcc" \
	LDFLAGS="-linkmode external -extldflags '-static'"

.PHONY: otelcol-config-linux-fips_arm64
otelcol-config-linux-fips_arm64:
	$(MAKE) toolchain-linux_arm64
	$(MAKE) otelcol-config-linux_arm64 \
	FIPS_SUFFIX="-fips" \
	CGO_ENABLED="1" \
	CC="$(TOOLCHAINS_DIR)toolchain_arm64/bin/aarch64-linux-musl-gcc" \
	LDFLAGS="-linkmode external -extldflags '-static'"

################################################################################
# Disk imaging targets
################################################################################

# otelcol-sumo-dmg: APP_NAME=otelcol-sumo
# otelcol-sumo-dmg: DMG_VOLUME_NAME=Sumo Logic Distribution for OpenTelemetry Collector
# otelcol-sumo-dmg: DMG_PATH=test.dmg

.PHONY: otelcol-sumo-dmg
otelcol-sumo-dmg:
	$(call check_defined, OTELCOL_BIN_PATH, you must set OTELCOL_BIN_PATH=path/to/otelcol_sumo)
	$(call check_defined, OTELCOL_CONFIG_BIN_PATH, you must set OTELCOL_CONFIG_BIN_PATH=path/to/otelcol-config)
	$(eval TMP_DIR ?= $(shell mktemp -d))
	cp otelcol-sumo "$(DMG_SOURCE_DIR)"
#	$(eval DMG_SOURCE_DIR = $(shell $(MAKE) darwin-dmg-prepare))
#	$(MAKE) darwin-dmg-add-file DMG_SOURCE_DIR="$(DMG_SOURCE_DIR)" SOURCE_FILE_PATH="flag.go"
#	$(MAKE) darwin-dmg-add-file DMG_SOURCE_DIR="$(DMG_SOURCE_DIR)" SOURCE_FILE_PATH="flag_actions.go"
	hdiutil create "$(DMG_PATH)" -ov -volname "$(DMG_VOLUME_NAME)" -fs APFS -format UDZO -srcfolder "$(DMG_SOURCE_DIR)"

################################################################################
# Code signing targets
################################################################################

.PHONY: darwin-sign-binary
darwin-sign-binary:
darwin-sign-binary:
	codesign --timestamp --options=runtime -s "$(DEVELOPER_SIGNING_IDENTITY)" -v "$(BINARY_PATH)"

.PHONY: darwin-sign-dmg
darwin-sign: BINARY_PATH = cmd/$(BINARY_NAME)
darwin-sign: DMG_PATH = $(BINARY_PATH).dmg
darwin-sign:
	xcrun notarytool submit --apple-id "$(AC_USERNAME)" --password "$(AC_PASSWORD)" --team-id "$(DEVELOPER_TEAM_ID)" --progress --wait "$(DMG_PATH)"
	xcrun stapler staple "$(DMG_PATH)"

.PHONY: darwin_amd64-sign
darwin_amd64-sign: BINARY_NAME=otelcol-config-darwin_amd64
darwin_amd64-sign: darwin-dmg-and-sign

.PHONY: darwin_arm64-sign
darwin_arm64-sign: BINARY_NAME=otelcol-config-darwin_arm64
darwin_arm64-sign: darwin-dmg-and-sign
