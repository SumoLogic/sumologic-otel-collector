#################################################################################
# Default variables
#################################################################################

SHELL := /usr/bin/env bash

#################################################################################
# Host platform detection
#################################################################################

ifeq ($(OS),Windows_NT)
	HOST_OS ?= Windows_NT
	ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
		HOST_ARCH ?= AMD64
	else
		ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
			HOST_ARCH ?= AMD64
		endif
		ifeq ($(PROCESSOR_ARCHITECTURE),x86)
			HOST_ARCH ?= x86
		endif
	endif

	ifeq ($(HOST_ARCH),)
		ifneq ($(PROCESSOR_ARCHITEW6432),)
			UNSUPPORTED_ARCH := $(PROCESSOR_ARCHITEW6432)
		else
			UNSUPPORTED_ARCH := $(PROCESSOR_ARCHITECTURE)
		endif
	endif
else
	UNAME_S ?= $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		HOST_OS ?= Linux
	endif
	ifeq ($(UNAME_S),Darwin)
		HOST_OS ?= Darwin
	endif

	UNAME_M := $(shell uname -m)
	ifeq ($(UNAME_M),x86_64)
		HOST_ARCH ?= AMD64
	endif
	ifeq ($(UNAME_M),aarch64)
		HOST_ARCH ?= ARM64
	endif
	ifeq ($(UNAME_M),arm64)
		HOST_ARCH ?= ARM64
	endif

	ifeq ($(HOST_ARCH),)
		UNSUPPORTED_OS := $(UNAME_S)
	endif
	ifeq ($(HOST_ARCH),)
		UNSUPPORTED_ARCH := $(UNAME_M)
	endif
endif

ifeq ($(HOST_OS),)
$(error Unsupported operating system: $(UNSUPPORTED_OS))
endif
ifeq ($(HOST_ARCH),)
$(error Unsupported architecture: $(UNSUPPORTED_ARCH))
endif

ifneq ($(DEBUG),)
$(info Platform detected: $(HOST_OS)/$(HOST_ARCH))
endif

#################################################################################
# Platform-specific setup
#################################################################################

ifeq ($(HOST_OS),Windows_NT)
	MAKE := $(shell cygpath '$(MAKE)')
	BINARY_EXT ?= .exe
endif

# Set the default GOOS based on HOST_OS
ifeq ($(HOST_OS),Darwin)
	GOOS ?= darwin
endif
ifeq ($(HOST_OS),Linux)
	GOOS ?= linux
endif
ifeq ($(HOST_OS),Windows_NT)
	GOOS ?= windows
endif

# Set the default GOARCH based on HOST_ARCH
ifeq ($(HOST_ARCH),AMD64)
	GOARCH ?= amd64
endif
ifeq ($(HOST_ARCH),ARM64)
	GOARCH ?= arm64
endif

#################################################################################
# GitHub Actions variables
#################################################################################

# A list of one or more space-delimited paths to files that are used to determine
# if the cache should be invalidated. When these files change, the cache is
# invalidated. This is used by GitHub Actions.
CACHE_DEPENDENCY_PATHS ?= go.mod

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
			required by target '$@')))

#################################################################################
# System CLI tool targets
#################################################################################

.PHONY: install-coreutils
install-coreutils:
ifeq ($(HOST_OS),Darwin)
	@which gdate > /dev/null || brew install coreutils
endif

.PHONY: install-gnu-sed
install-gnu-sed:
ifeq ($(HOST_OS),Darwin)
	@which gsed > /dev/null || brew install gnu-sed
endif

.PHONY: install-jq
install-jq:
ifeq ($(HOST_OS),Darwin)
	@which jq > /dev/null || brew install jq
else
	@which jq > /dev/null || echo "Please install jq manually"
endif

.PHONY: install-staticcheck
install-staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: install-dependencies
install-dependencies: install-coreutils
install-dependencies: install-gnu-sed
install-dependencies: install-jq
install-dependencies: install-staticcheck
install-dependencies:
	@echo "All dependencies installed successfully."

################################################################################
# GitHub Actions targets
################################################################################

.PHONY: list-cache-dependency-paths
list-cache-dependency-paths:
	@echo $(CACHE_DEPENDENCY_PATHS)

.PHONY: cache-dependency-paths-json
cache-dependency-paths-json: install-jq
cache-dependency-paths-json:
	@printf '"%s"\n' $(CACHE_DEPENDENCY_PATHS) | jq -scr

################################################################################
# Disk imaging targets
################################################################################

# Creates a DMG file to store a binary.
#
# 1. Attempt to remove DMG file with same name.
# 2. Create temporary directory for DMG contents.
# 3. Create DMG file using contents in temporary directory.
.PHONY: darwin-create-dmg
darwin-create-dmg:
	@$(call check_defined, \
		BIN_PATH \
		DMG_BIN_FILENAME \
		DMG_PATH \
		DMG_VOLUME_NAME \
	)

	@rm -f "$(DMG_PATH)" 2>&1 || true
	@$(eval DMG_SOURCE_DIR ?= $(shell mktemp -d))
	@echo "Using temp dir: $(DMG_SOURCE_DIR)"
	@cp "$(BIN_PATH)" "$(DMG_SOURCE_DIR)/$(DMG_BIN_FILENAME)"

	@hdiutil create "$(DMG_PATH)" \
		-ov -volname "$(DMG_VOLUME_NAME)" -fs APFS \
		-format UDZO -srcfolder "$(DMG_SOURCE_DIR)" 2>&1 || \
		(rm -rf "$(DMG_SOURCE_DIR)" && exit 1)

################################################################################
# Signing targets
################################################################################

# Check if a binary is signed. Exits 1 if signed or 0 if unsigned.
.PHONY: darwin-check-binary-signed
darwin-check-binary-signed:
	@$(call check_defined,
		BIN_PATH \
		DEVELOPER_TEAM_ID \
	)

	@codesign -dv "$(BIN_PATH)" 2>&1 | grep "$(DEVELOPER_TEAM_ID)" && \
		echo "$(BIN_PATH) is already signed" || \
		(echo "$(BIN_PATH) is not yet signed" && exit 1)

# Uses codesign to sign a binary. Skips binaries that are already signed.
.PHONY: darwin-sign-binary
darwin-sign-binary:
	@$(call check_defined, \
		BIN_PATH \
		DEVELOPER_SIGNING_IDENTITY \
	)

	@$(MAKE) darwin-check-binary-signed || \
		(echo "Signing $(BIN_PATH)..." && \
		codesign --timestamp --options=runtime \
			-s "$(DEVELOPER_SIGNING_IDENTITY)" \
			-v "$(BIN_PATH)")

	@echo "$(BIN_PATH) is signed"

.PHONY: darwin-notarize-dmg
darwin-notarize-dmg:
	@$(call check_defined, \
		AC_USERNAME \
		AC_PASSWORD \
		DEVELOPER_TEAM_ID \
		DMG_PATH \
	)

	@xcrun notarytool submit \
		--apple-id "$(AC_USERNAME)" \
		--password "$(AC_PASSWORD)" \
		--team-id "$(DEVELOPER_TEAM_ID)" \
		--progress \
		--wait \
		"$(DMG_PATH)"

	@xcrun stapler staple "$(DMG_PATH)"

.PHONY: darwin-sign
darwin-sign: DMG_PATH = $(BIN_PATH).dmg
darwin-sign: darwin-sign-binary
darwin-sign: darwin-create-dmg
darwin-sign: darwin-notarize-dmg
