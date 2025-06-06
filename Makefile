GOLANGCI_LINT_VERSION ?= v1.59.1
PRETTIER_VERSION ?= 3.0.3
TOWNCRIER_VERSION ?= 23.6.0
SHELL := /usr/bin/env bash
GIT_SHA_CMD := git rev-parse HEAD

# Convert between Linux & Windows paths
ifeq ($(OS),Windows_NT)
	MAKE := "$(shell cygpath '$(MAKE)')"
endif

# Use GNU tools on macOS to ensure tool compatibility between platforms
ifeq ($(shell go env GOOS),darwin)
SED ?= gsed
DATE ?= gdate
else
SED ?= sed
DATE ?= date
endif

#################################################################################
# OpenTelemetry variables
#################################################################################

OT_CORE_VERSION := $(shell grep "version: .*" otelcolbuilder/.otelcol-builder.yaml | cut -f 4 -d " ")
OT_CONTRIB_VERSION := $(shell grep --max-count=1 '^  - gomod: github\.com/open-telemetry/opentelemetry-collector-contrib/' otelcolbuilder/.otelcol-builder.yaml | cut -d " " -f 6 | $(SED) "s/v//")

#################################################################################
# Go module variables
#################################################################################

ALL_GO_MODULES := $(shell find pkg/ -type f -name 'go.mod' -exec dirname {} \; | sort)
ALL_GO_MODULES += otelcolbuilder

EXPORTABLE_MODULES := $(filter pkg/%,$(ALL_GO_MODULES))
EXPORTABLE_MODULES := $(filter-out pkg/test,$(EXPORTABLE_MODULES))
EXPORTABLE_MODULES := $(filter-out pkg/tools/%,$(EXPORTABLE_MODULES))

TESTABLE_MODULES := $(filter-out pkg/tools/udpdemux,$(ALL_GO_MODULES))

#################################################################################
# Colour variables
#################################################################################

NC            := $(shell tput -Txterm sgr0)
Black         := $(shell tput -Txterm setaf 0)
Red           := $(shell tput -Txterm setaf 1)
Green         := $(shell tput -Txterm setaf 2)
Yellow        := $(shell tput -Txterm setaf 3)
Blue          := $(shell tput -Txterm setaf 4)
Magenta       := $(shell tput -Txterm setaf 5)
Cyan          := $(shell tput -Txterm setaf 6)
White         := $(shell tput -Txterm setaf 7)

#################################################################################
# Functions
#################################################################################

define target_echo
	echo -e "[$(Yellow)$(@D)$(NC)]: $1"
endef

define shell_run
	@start=$$($(DATE) +"%s%3N"); \
	$(call target_echo,Running command: $1); \
	$1 | awk '{print "[$(Yellow)$(@D)$(NC)]: " $$0}'; \
	end=$$($(DATE) +"%s%3N"); \
	duration=$$(( $$end - $$start )); \
	$(call target_echo,Completed in $$duration ms)
endef

#################################################################################
# Default target
#################################################################################

all: markdownlint yamllint

#################################################################################
# General wildcard targets
#################################################################################

.PHONY: %/print-directory
%/print-directory:
	@echo $(@D)

.PHONY: %/test
%/test:
	@$(call shell_run,cd "$(@D)" && make test)

.PHONY: %/lint
%/lint:
	@$(call shell_run,cd "$(@D)" && make lint)

#################################################################################
# System CLI tool targets
#################################################################################

.PHONY: install-gnu-sed
install-gnu-sed:
ifeq ($(shell go env GOOS),darwin)
	@which gsed || brew install gnu-sed
endif

.PHONY: install-coreutils
install-coreutils:
ifeq ($(shell go env GOOS),darwin)
	@which gdate || brew install coreutils
endif

#################################################################################
# Markdown targets
#################################################################################

.PHONY: install-markdownlint
install-markdownlint:
	npm install --global markdownlint-cli

.PHONY: markdownlint
markdownlint:
	markdownlint '**/*.md'

.PHONY: markdownlint-docker
markdownlint-docker:
	docker run --rm -v ${PWD}:/workdir ghcr.io/igorshubovych/markdownlint-cli:latest '**/*.md'

markdown-links-lint:
	./ci/markdown_links_lint.sh

markdown-link-check:
	./ci/markdown_link_check.sh

#################################################################################
# YAML targets
#################################################################################

yamllint:
	yamllint -c .yamllint.yaml \
		otelcolbuilder/.otelcol-builder.yaml

#################################################################################
# Shell targets
#################################################################################

.PHONY: check-uniform-dependencies
check-uniform-dependencies:
	./ci/check_uniform_dependencies.sh

shellcheck:
	shellcheck --severity=info ci/*.sh

#################################################################################
# Commit-hook targets
#################################################################################

.PHONY: install-pre-commit
install-pre-commit:
	python3 -m pip install pre-commit

.PHONY: install-pre-commit-hook
install-pre-commit-hook:
	pre-commit install

# ref: https://pre-commit.com/
.PHONY: pre-commit-check
pre-commit-check:
	pre-commit run --all-files

#################################################################################
# Go targets
#################################################################################

.PHONY: %/mod-download-all
%/mod-download-all:
	@$(call shell_run,cd "$(@D)" && make mod-download-all)

.PHONY: %/test-junit
%/test-junit:
	make test

.PHONY: golint
golint: $(patsubst %,%/lint,$(ALL_GO_MODULES))

.PHONY: gomod-download-all
gomod-download-all: $(patsubst %,%/mod-download-all,$(ALL_GO_MODULES))

.PHONY: gotest
gotest: $(patsubst %,%/test,$(TESTABLE_GO_MODULES))

.PHONY: gotest-junit
gotest-junit: $(patsubst %,%/test-junit,$(TESTABLE_GO_MODULES))

.PHONY: install-go-junit-report
install-go-junit-report:
	go install github.com/jstemmer/go-junit-report@latest

.PHONY: install-staticcheck
install-staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: list-all-modules
list-all-modules: $(patsubst %,%/print-directory,$(ALL_GO_MODULES))

.PHONY: list-exportable-modules
list-exportable-modules: $(patsubst %,%/print-directory,$(EXPORTABLE_GO_MODULES))

.PHONY: list-testable-modules
list-testable-modules: $(patsubst %,%/print-directory,$(TESTABLE_GO_MODULES))

#################################################################################
# OpenTelemetry preparation targets
#################################################################################

# NOTE: This target can be used by setting OTC_CORE_NEW and OT_CONTRIB_NEW when
# calling the target.
#
# E.g. make update-ot OT_CORE_NEW=x.x.x OT_CONTRIB_NEW=y.y.y
.PHONY: update-ot
update-ot: install-gnu-sed
	@test $(OT_CORE_NEW)    || (echo "usage: make update-ot OT_CORE_NEW=x.x.x OT_CONTRIB_NEW=y.y.y"; exit 1);
	@test $(OT_CONTRIB_NEW) || (echo "usage: make update-ot OT_CORE_NEW=x.x.x OT_CONTRIB_NEW=y.y.y"; exit 1);
	@echo "Updating OT core from $(OT_CORE_VERSION) to $(OT_CORE_NEW) and OT contrib from $(OT_CONTRIB_VERSION) to $(OT_CONTRIB_NEW)"
	$(shell yq e ".dist.version = \"${OT_CORE_NEW}\"" -i otelcolbuilder/.otelcol-builder.yaml)
	$(shell yq -i '(.. | select(tag=="!!str")) |= sub("(go\.opentelemetry\.io/collector.*) v${OT_CORE_VERSION}", "$$1 v${OT_CORE_NEW}")'  otelcolbuilder/.otelcol-builder.yaml)
	$(shell yq -i '(.. | select(tag=="!!str")) |= sub("(github\.com/open-telemetry/opentelemetry-collector-contrib.*) v${OT_CONTRIB_VERSION}", "$$1 v${OT_CONTRIB_NEW}")'  otelcolbuilder/.otelcol-builder.yaml)
	$(SED) -i "s/$(OT_CORE_VERSION)/$(OT_CORE_NEW)/" otelcolbuilder/Makefile
	$(SED) -i "s/\(collector\/\(blob\|tree\)\/v\)$(OT_CORE_VERSION)/\1$(OT_CORE_NEW)/" \
		README.md \
		docs/configuration.md \
		docs/migration.md \
		docs/performance.md
	$(SED) -i "s/\(contrib\/\(blob\|tree\)\/v\)$(OT_CONTRIB_VERSION)/\1$(OT_CONTRIB_NEW)/" \
		README.md \
		docs/configuration.md \
		docs/migration.md \
		docs/performance.md \
		pkg/receiver/telegrafreceiver/README.md
	@find . -type f -name "go.mod" -exec $(SED) -i "s/\(go\.opentelemetry\.io\/collector.*\) v$(OT_CORE_VERSION)$$/\1 v$(OT_CORE_NEW)/" {} \;
	@find . -type f -name "go.mod" -exec $(SED) -i "s/\(github\.com\/open-telemetry\/opentelemetry-collector-contrib\/.*\) v$(OT_CONTRIB_VERSION)$$/\1 v$(OT_CONTRIB_NEW)/" {} \;
	@echo "building OT distro to check for breakage"
	make gomod-download-all
	pushd otelcolbuilder \
		&& make install-ocb \
		&& make build \
		&& popd

.PHONY: update-journalctl
update-journalctl: install-gnu-sed
	$(SED) -i "s/FROM debian.*/FROM debian:${DEBIAN_VERSION} as systemd/" Dockerfile*

.PHONY: update-docs
update-docs: install-gnu-sed
update-docs: LATEST_OT_VERSION = $(shell git describe --match 'v*' --abbrev=0 | cut -c2-)
update-docs: PREVIOUS_OT_VERSION = $(shell git describe --match 'v*' --abbrev=0 `git describe --match 'v*' --abbrev=0`^ | cut -c2-)
update-docs: PREVIOUS_CORE_VERSION = $(shell echo ${PREVIOUS_OT_VERSION} | sed -e 's/-sumo-.*//')
update-docs:
	@find docs/ -type f \( -name "*.md" ! -name "upgrading.md" \) -exec $(SED) -i 's#$(PREVIOUS_CORE_VERSION)#$(OT_CORE_VERSION)#g' {} \;
	@find docs/ -type f \( -name "*.md" ! -name "upgrading.md" \) -exec $(SED) -i 's#$(PREVIOUS_OT_VERSION)#$(LATEST_OT_VERSION)#g' {} \;

#################################################################################
# OpenTelemetry build targets
#################################################################################

.PHONY: build
build:
	@$(MAKE) -C ./otelcolbuilder/ build
	@$(MAKE) -C ./pkg/tools/otelcol-config/ build

.PHONY: install-ocb
install-ocb:
	@$(MAKE) -C ./otelcolbuilder/ install-ocb

#################################################################################
# Container targets
#################################################################################

.PHONY: _promote-container-image
_promote-container-image:
	docker buildx imagetools create "$(SRC_URL):$(TAG)" -t "$(DST_URL):$(TAG)"

.PHONY: _create-container-image-alias
_create-container-image-alias:
	docker buildx imagetools create "$(URL):$(SRC_TAG)" -t "$(URL):$(DST_TAG)"

.PHONY: _promote-container-image-standard
_promote-container-image-standard: GIT_SHA = $(shell $(GIT_SHA_CMD))
_promote-container-image-standard:
	$(MAKE) _promote-container-image TAG="$(GIT_SHA)"
	$(MAKE) _create-container-image-alias SRC_TAG="$(GIT_SHA)" DST_TAG="latest"

.PHONY: _promote-container-image-standard-fips
_promote-container-image-standard-fips: GIT_SHA = $(shell $(GIT_SHA_CMD))
_promote-container-image-standard-fips:
	$(MAKE) _promote-container-image TAG="$(GIT_SHA)-fips"
	$(MAKE) _create-container-image-alias SRC_TAG="$(GIT_SHA)" DST_TAG="latest-fips"

.PHONY: _promote-container-image-ubi
_promote-container-image-ubi: GIT_SHA = $(shell $(GIT_SHA_CMD))
_promote-container-image-ubi:
	$(MAKE) _promote-container-image TAG="$(GIT_SHA)-ubi"
	$(MAKE) _create-container-image-alias SRC_TAG="$(GIT_SHA)" DST_TAG="latest-ubi"

.PHONY: _promote-container-image-ubi-fips
_promote-container-image-ubi-fips: GIT_SHA = $(shell $(GIT_SHA_CMD))
_promote-container-image-ubi-fips:
	$(MAKE) _promote-container-image TAG="$(GIT_SHA)-ubi-fips"
	$(MAKE) _create-container-image-alias SRC_TAG="$(GIT_SHA)" DST_TAG="latest-ubi-fips"

# Promotes an image in the ECR ci-builds repository to the release-candidates
# repository
.PHONY: promote-ecr-image-ci-to-rc
promote-ecr-image-ci-to-rc: SRC_URL = "docker://$(ECR_URL_CI)"
promote-ecr-image-ci-to-rc: DST_URL = "docker://$(ECR_URL_RC)"
promote-ecr-image-ci-to-rc: _promote-container-manifest

# Promotes an image in the ECR release-candidates repository to the stable
# repository
.PHONY: promote-ecr-image-rc-to-stable
promote-ecr-image-rc-to-stable: GIT_SHA = $(shell $(GIT_SHA_CMD))
promote-ecr-image-rc-to-stable: SRC_URL = "docker://$(ECR_URL_RC):$(GIT_SHA)"
promote-ecr-image-rc-to-stable: DST_URL = "docker://$(ECR_URL_STABLE):$(GIT_SHA)"
promote-ecr-image-rc-to-stable: _promote-container-manifest

# Promotes an image in the Docker Hub ci-builds repository to the
# release-candidates repository
.PHONY: promote-dh-image-ci-to-rc
promote-dh-image-ci-to-rc: GIT_SHA = $(shell $(GIT_SHA_CMD))
promote-dh-image-ci-to-rc: SRC_URL = "$(DH_URL_CI):$(GIT_SHA)"
promote-dh-image-ci-to-rc: DST_URL = "$(DH_URL_RC):$(GIT_SHA)"
promote-dh-image-ci-to-rc: _promote-container-manifest

# Promotes an image in the Docker Hub release-candidates repository to the
# stable repository
.PHONY: promote-dh-image-rc-to-stable
promote-dh-image-rc-to-stable: GIT_SHA = $(shell $(GIT_SHA_CMD))
promote-dh-image-rc-to-stable: SRC_URL = "docker://$(DH_URL_RC):$(GIT_SHA)"
promote-dh-image-rc-to-stable: DST_URL = "docker://$(DH_URL_STABLE):$(GIT_SHA)"
promote-dh-image-rc-to-stable: _promote-container-manifest

#################################################################################
# Changelog targets
#
# NOTE: We use Towncrier (https://towncrier.readthedocs.io) for changelog
#       management.
#################################################################################

.PHONY: install-towncrier
install-towncrier:
	python3 -m pip install towncrier==$(TOWNCRIER_VERSION)

.PHONY: install-prettier
install-prettier:
	npm install --global prettier@$(PRETTIER_VERSION)

## Usage: make add-changelog-entry
.PHONY: add-changelog-entry
add-changelog-entry:
	./ci/add-changelog-entry.sh

## Consume the files in .changelog and update CHANGELOG.md
## We also format it afterwards to make sure it's consistent with our style
## Usage: make update-changelog VERSION=x.x.x-sumo-x
.PHONY: update-changelog
update-changelog:
ifndef VERSION
	$(error Usage: make update-changelog VERSION=x.x.x-sumo-x)
endif
	towncrier build --yes --version $(VERSION)
	prettier -w CHANGELOG.md
	git add CHANGELOG.md

## Check if the branch relative to main adds a changelog entry.
## This target is used in the CI `changelog` check.
.PHONY: check-changelog
check-changelog:
	towncrier check

#################################################################################
# Vagrant targets
#################################################################################

.PHONY: vagrant-up
vagrant-up:
	vagrant up

.PHONY: vagrant-ssh
vagrant-ssh:
	vagrant ssh -c 'cd /sumologic; exec "$$SHELL"'

.PHONY: vagrant-destroy
vagrant-destroy:
	vagrant destroy -f

.PHONY: vagrant-halt
vagrant-halt:
	vagrant halt
