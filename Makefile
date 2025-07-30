#################################################################################
# Path variables
#################################################################################

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(patsubst %/,%,$(dir $(mkfile_path)))

# Include common.mk
include $(realpath $(current_dir)/common.mk)

#################################################################################
# Variables
#################################################################################

PRETTIER_VERSION ?= 3.0.3
TOWNCRIER_VERSION ?= 23.6.0

#################################################################################
# Platform-specific setup
#################################################################################

# Convert between Linux & Windows paths
ifeq ($(HOST_OS),Windows_NT)
	MAKE := "$(shell cygpath '$(MAKE)')"
endif

# Use GNU tools on macOS to ensure tool compatibility between platforms
ifeq ($(HOST_OS),Darwin)
	SED ?= gsed
	DATE ?= gdate
else
	SED ?= sed
	DATE ?= date
endif

OP := $(shell which op 2> /dev/null)
ifneq ($(OP),)
	AWS ?= op plugin run -- aws
else
	AWS ?= aws
endif

#################################################################################
# OpenTelemetry variables
#################################################################################

OT_CORE_VERSION_CMD := grep "version: .*" otelcolbuilder/.otelcol-builder.yaml \
	| cut -f 4 -d " "

OT_CONTRIB_VERSION_CMD := grep --max-count=1 \
	'^  - gomod: github\.com/open-telemetry/opentelemetry-collector-contrib/' \
	otelcolbuilder/.otelcol-builder.yaml | cut -d " " -f 6 | $(SED) "s/v//"

#################################################################################
# Go module variables
#################################################################################

ALL_GO_MODULES := $(shell find pkg/ -type f -name 'go.mod' -exec dirname {} \; | sort)
ALL_GO_MODULES += otelcolbuilder

EXPORTABLE_GO_MODULES := $(filter pkg/%,$(ALL_GO_MODULES))
EXPORTABLE_GO_MODULES := $(filter-out pkg/test,$(EXPORTABLE_GO_MODULES))
EXPORTABLE_GO_MODULES := $(filter-out pkg/tools/%,$(EXPORTABLE_GO_MODULES))

TESTABLE_GO_MODULES := $(filter-out pkg/tools/otelcol-config,$(ALL_GO_MODULES))
TESTABLE_GO_MODULES := $(filter-out pkg/tools/udpdemux,$(TESTABLE_GO_MODULES))

#################################################################################
# GitHub Actions variables
#################################################################################

# Contains the base matrix for workflow-test-otelcol.yml
TEST_OTELCOL_BASE_MATRIX = ci/matrix/workflow-test-otelcol.json
TEST_OTELCOL_JQ_FILTER = ci/matrix/workflow-test-otelcol.jq

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
	$(call shell_run,cd "$(@D)" && $(MAKE) test)

.PHONY: %/lint
%/lint:
	$(call shell_run,cd "$(@D)" && $(MAKE) lint)

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
	docker run --rm -v ${PWD}:/workdir \
		ghcr.io/igorshubovych/markdownlint-cli:latest '**/*.md'

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
	$(call shell_run,cd "$(@D)" && $(MAKE) mod-download-all)

.PHONY: golint
golint: $(patsubst %,%/lint,$(ALL_GO_MODULES))

.PHONY: gomod-download-all
gomod-download-all: $(patsubst %,%/mod-download-all,$(ALL_GO_MODULES))

.PHONY: gotest
gotest: $(patsubst %,%/test,$(TESTABLE_GO_MODULES))

.PHONY: list-all-modules
list-all-modules: $(patsubst %,%/print-directory,$(ALL_GO_MODULES))

.PHONY: list-exportable-modules
list-exportable-modules: $(patsubst %,%/print-directory,$(EXPORTABLE_GO_MODULES))

.PHONY: list-testable-modules
list-testable-modules: $(patsubst %,%/print-directory,$(TESTABLE_GO_MODULES))

#################################################################################
# GitHub Actions targets
#################################################################################

.PHONY: testable-modules-json
testable-modules-json: install-jq
testable-modules-json:
	@printf '"%s"\n' $(TESTABLE_GO_MODULES) | jq -scr

.PHONY: test-otelcol-base-matrix
test-otelcol-base-matrix: install-jq
test-otelcol-base-matrix:
	@jq -cr . "$(TEST_OTELCOL_BASE_MATRIX)"

# Generate a matrix for the otelcol test workflow by combining the base matrix
# with the testable Go modules
.PHONY: test-otelcol-matrix
test-otelcol-matrix: install-jq
test-otelcol-matrix: BASE_MATRIX = $(shell $(MAKE) test-otelcol-base-matrix)
test-otelcol-matrix: PKGS = $(shell $(MAKE) testable-modules-json)
test-otelcol-matrix:
	@jq -ncr \
	--argjson base '$(BASE_MATRIX)' \
	--argjson pkgs '$(PKGS)' \
	-f "$(TEST_OTELCOL_JQ_FILTER)"

#################################################################################
# OpenTelemetry preparation targets
#################################################################################

# NOTE: This target can be used by setting OTC_CORE_NEW and OT_CONTRIB_NEW when
# calling the target.
#
# E.g. make update-ot OT_CORE_NEW=x.x.x OT_CONTRIB_NEW=y.y.y
.PHONY: update-ot
update-ot: install-gnu-sed
update-ot: OT_CORE_VERSION = $(shell $(OT_CORE_VERSION_CMD))
update-ot: OT_CONTRIB_VERSION = $(shell $(OT_CONTRIB_VERSION_CMD))
update-ot:
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
	$(MAKE) gomod-download-all
	pushd otelcolbuilder \
		&& $(MAKE) install-ocb \
		&& $(MAKE) build \
		&& popd

.PHONY: update-docs
update-docs: install-gnu-sed
update-docs: OT_CORE_VERSION = $(shell $(OT_CORE_VERSION_CMD))
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
