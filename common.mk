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

#################################################################################
# GitHub Actions targets
#################################################################################

.PHONY: list-cache-dependency-paths
list-cache-dependency-paths:
	@echo $(CACHE_DEPENDENCY_PATHS)

.PHONY: cache-dependency-paths-json
cache-dependency-paths-json: install-jq
cache-dependency-paths-json:
	@printf '"%s"\n' $(CACHE_DEPENDENCY_PATHS) | jq -scr
