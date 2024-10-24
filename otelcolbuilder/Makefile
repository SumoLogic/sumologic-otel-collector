include ../toolchains/Makefile

BINARY_NAME ?= otelcol-sumo
BUILDER_VERSION ?= 0.109.0
FIPS_SUFFIX ?= ""
BUILDER_REPO ?= github.com/open-telemetry/opentelemetry-collector-releases
GO ?= go
GOTAGS ?=

ifeq ($(OS),Windows_NT)
	OS=windows
endif

ifeq ($(OS),windows)
	BUILDER_BIN_EXT ?= .exe
	MAKE := "$(shell cygpath '$(MAKE)')"
endif

OS ?= $(shell uname -s | tr A-Z a-z)
ARCH ?= $(shell uname -m | sed s/aarch64/arm64/ | sed s/x86_64/amd64/)

BUILDER_BIN_NAME ?= opentelemetry-collector-builder$(BUILDER_BIN_EXT)
BUILDER_BIN_PATH ?= $(HOME)/bin

INSTALLED_BUILDER_VERSION := $(shell $(BUILDER_BIN_NAME) version 2>&1)

# Settings for macOS builds
ifeq ($(OS),darwin)
	AC_USERNAME ?= otel-mac-certs@sumologic.com
	AC_PASSWORD ?= $(shell echo $$AC_PASSWORD)
	DEVELOPER_TEAM_ID ?= 8F28635Z7X
	DEVELOPER_SIGNING_IDENTITY ?= Developer ID Application: Sumo Logic, Inc. ($(DEVELOPER_TEAM_ID))
	DMG_VOLUME_NAME ?= Sumo Logic Distribution for OpenTelemetry Collector
endif

# Add additional LDFLAGS
# ex: to get a staticly executable with CGO use EXTRA_LDFLAGS="-linkmode external -extldflags '-static'"
LDFLAGS:=$(EXTRA_LDFLAGS)

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

.PHONY: _install-bin
_install-bin:
	@mkdir -p "$(BUILDER_BIN_PATH)"
	curl -L -o "$(BUILDER_BIN_PATH)/$(BUILDER_BIN_NAME)" https://$(BUILDER_REPO)/releases/download/cmd/builder/v$(BUILDER_VERSION)/ocb_$(BUILDER_VERSION)_$(PLATFORM)_$(ARCH)$(BUILDER_BIN_EXT)
	@chmod +x "$(BUILDER_BIN_PATH)/$(BUILDER_BIN_NAME)"

.PHONY: install-builder
install-builder:
	@echo "Installing $(BUILDER_REPO)/cmd/builder@v$(BUILDER_VERSION)... (in $(BUILDER_BIN_PATH))"
	@$(MAKE) _install-bin PLATFORM=$(OS)

.PHONY: _builder
_builder:
	$(eval VERSION ?= $(shell git describe --tags --abbrev=5 --match "v[0-9]*"))
# Need to specify go path because otherwise opentelemetry-collector-builder
# uses /usr/bin/go which on Github Actions is using preinstalled 1.15.12 by default.
	CGO_ENABLED=$(CGO_ENABLED) $(BUILDER_BIN_NAME) \
		--go $(GO) \
		--version "$(VERSION)$(FIPS_SUFFIX)" \
		--config .otelcol-builder.yaml \
		--output-path ./cmd \
		--skip-compilation=$(SKIP_COMPILATION)

.PHONY: _gobuild
_gobuild:
	(cd cmd && \
		CGO_ENABLED=$(CGO_ENABLED) GOEXPERIMENT=$(GOEXPERIMENT) go build -v \
		-tags "$(GOTAGS)" \
		-ldflags="$(LDFLAGS)" \
		-trimpath \
		-o ./$(BINARY_NAME)$(BUILDER_BIN_EXT) . \
	)
	chmod +x ./cmd/$(BINARY_NAME)$(BUILDER_BIN_EXT)

.PHONY: _gobuild_debug
_gobuild_debug:
	(cd cmd && \
		CGO_ENABLED=$(CGO_ENABLED) GOEXPERIMENT=$(GOEXPERIMENT) go build -v \
		-tags "$(GOTAGS)" \
		-race \
		-gcflags "all=-N -l" \
		-o ./$(BINARY_NAME)-debug$(BUILDER_BIN_EXT) . \
	)
	chmod +x ./cmd/$(BINARY_NAME)-debug$(BUILDER_BIN_EXT)

.PHONY: build
build:
# Since builder doesn't allow adding build tags, let's just skip the compilation
# and run 'go build ...' by hand.
	@$(MAKE) generate-sources
	@$(MAKE) _gobuild

.PHONY: build-debug
build-debug:
# Since builder doesn't allow adding build tags, let's just skip the compilation
# and run 'go build ...' by hand.
	@$(MAKE) generate-sources
	@$(MAKE) _gobuild_debug

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
	@$(MAKE) _builder SKIP_COMPILATION=true
	# the changes below allow us to use a custom config provider
	patch cmd/main.go cmd/00_main.go.patch
	patch cmd/main.go cmd/01_main.go.patch
	patch cmd/main_windows.go cmd/main_windows.go.patch

.PHONY: test
test:
	@$(MAKE) install-builder
	@$(MAKE) generate-sources
	CGO_ENABLED=${CGO_ENABLED} GOEXPERIMENT=$(GOEXPERIMENT) $(MAKE) -C cmd test

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

.PHONY: darwin-sign
darwin-sign: BINARY_PATH = cmd/$(BINARY_NAME)
darwin-sign: DMG_PATH = $(BINARY_PATH).dmg
darwin-sign:
	$(eval DMG_SOURCE_DIR = $(shell mktemp -d -t "$(BINARY_NAME)-dmg"))
	codesign --timestamp --options=runtime -s "$(DEVELOPER_SIGNING_IDENTITY)" -v "$(BINARY_PATH)"
	cp "$(BINARY_PATH)" "$(DMG_SOURCE_DIR)"
	hdiutil create "$(DMG_PATH)" -ov -volname "$(DMG_VOLUME_NAME)" -fs APFS -format UDZO -srcfolder "$(DMG_SOURCE_DIR)"
	xcrun notarytool submit --apple-id "$(AC_USERNAME)" --password "$(AC_PASSWORD)" --team-id "$(DEVELOPER_TEAM_ID)" --progress --wait "$(DMG_PATH)"
	xcrun stapler staple "$(DMG_PATH)"

.PHONY: darwin_amd64-sign
darwin_amd64-sign: BINARY_NAME=otelcol-sumo-darwin_amd64
darwin_amd64-sign: darwin-sign

.PHONY: darwin_arm64-sign
darwin_arm64-sign: BINARY_NAME=otelcol-sumo-darwin_arm64
darwin_arm64-sign: darwin-sign

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
	GOOS=windows GOARCH=amd64 $(MAKE) build BINARY_NAME=$(BINARY_NAME)$(FIPS_SUFFIX)-windows_amd64

.PHONY: otelcol-sumo-linux-fips_arm64
otelcol-sumo-linux-fips_arm64:
	$(MAKE) toolchain-linux_arm64
	$(MAKE) otelcol-sumo-linux_arm64 \
		FIPS_SUFFIX="-fips" \
		CGO_ENABLED="1" \
		CC="$(TOOLCHAINS_DIR)/toolchain_arm64/bin/aarch64-linux-musl-gcc" \
		EXTRA_LDFLAGS="-linkmode external -extldflags '-static'"

.PHONY: otelcol-sumo-linux-fips_amd64
otelcol-sumo-linux-fips_amd64:
	$(MAKE) toolchain-linux_amd64
	$(MAKE) otelcol-sumo-linux_amd64 \
		FIPS_SUFFIX="-fips" \
		CGO_ENABLED="1" \
		CC="$(TOOLCHAINS_DIR)/toolchain_amd64/bin/x86_64-linux-musl-gcc" \
		EXTRA_LDFLAGS="-linkmode external -extldflags '-static'"
