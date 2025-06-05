GOLANGCI_LINT_VERSION ?= v1.59.1
PRETTIER_VERSION ?= 3.0.3
TOWNCRIER_VERSION ?= 23.6.0
SHELL := /usr/bin/env bash

ifeq ($(OS),Windows_NT)
	MAKE := "$(shell cygpath '$(MAKE)')"
endif

GIT_SHA_CMD := git rev-parse HEAD

all: markdownlint yamllint

ifeq ($(shell go env GOOS),darwin)
SED ?= gsed
else
SED ?= sed
endif

.PHONY: install-gsed
install-gsed:
ifeq ($(shell go env GOOS),darwin)
	@which gsed || brew install gsed
endif

.PHONY: install-markdownlint
install-markdownlint:
	npm install --global markdownlint-cli

.PHONY: markdownlint
markdownlint:
	markdownlint '**/*.md'

.PHONY: markdownlint-docker
markdownlint-docker:
	docker run --rm -v ${PWD}:/workdir ghcr.io/igorshubovych/markdownlint-cli:latest '**/*.md'

yamllint:
	yamllint -c .yamllint.yaml \
		otelcolbuilder/.otelcol-builder.yaml

markdown-links-lint:
	./ci/markdown_links_lint.sh

markdown-link-check:
	./ci/markdown_link_check.sh

shellcheck:
	shellcheck --severity=info ci/*.sh

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

# ALL_MODULES includes ./* dirs (excludes . dir and example with go code)
ALL_MODULES := $(shell find ./pkg -type f -name "go.mod" -exec dirname {} \; | sort | egrep  '^./' )
ALL_MODULES += ./otelcolbuilder

ALL_EXPORTABLE_MODULES += $(shell find ./pkg -type f -name "go.mod" ! -path "*pkg/test/*" -exec dirname {} \; | sort )

.PHONY: list-modules
list-modules:
	$(MAKE) for-all CMD=""

.PHONY: gotest
gotest:
	echo "GOCACHE = $(shell go env GOCACHE)"
	echo "GOMODCACHE = $(shell go env GOMODCACHE)"
	echo "GOTMPDIR = $(shell go env GOTMPDIR)"
	@$(MAKE) for-all CMD="make test"

.PHONY: golint
golint:
	@$(MAKE) for-all CMD="make lint"

.PHONY: gomod-download-all
gomod-download-all:
	@$(MAKE) for-all CMD="make mod-download-all"

.PHONY: install-staticcheck
install-staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: print-all-modules
print-all-modules:
	echo $(ALL_EXPORTABLE_MODULES)

.PHONY: for-all
for-all:
	@echo "running $${CMD} in all modules..."
	@set -e; for dir in $(ALL_MODULES); do \
	  (cd "$${dir}" && \
	  	echo "running $${CMD} in $${dir}" && \
	 	$${CMD} ); \
	done

.PHONY: check-uniform-dependencies
check-uniform-dependencies:
	./ci/check_uniform_dependencies.sh

OT_CORE_VERSION := $(shell grep "version: .*" otelcolbuilder/.otelcol-builder.yaml | cut -f 4 -d " ")
OT_CONTRIB_VERSION := $(shell grep --max-count=1 '^  - gomod: github\.com/open-telemetry/opentelemetry-collector-contrib/' otelcolbuilder/.otelcol-builder.yaml | cut -d " " -f 6 | $(SED) "s/v//")
# usage: make update-ot OT_CORE_NEW=x.x.x OT_CONTRIB_NEW=y.y.y
.PHONY: update-ot
update-ot: install-gsed
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
update-journalctl: install-gsed
	$(SED) -i "s/FROM debian.*/FROM debian:${DEBIAN_VERSION} as systemd/" Dockerfile*

.PHONY: update-docs
update-docs: install-gsed
update-docs: LATEST_OT_VERSION = $(shell git describe --match 'v*' --abbrev=0 | cut -c2-)
update-docs: PREVIOUS_OT_VERSION = $(shell git describe --match 'v*' --abbrev=0 `git describe --match 'v*' --abbrev=0`^ | cut -c2-)
update-docs: PREVIOUS_CORE_VERSION = $(shell echo ${PREVIOUS_OT_VERSION} | sed -e 's/-sumo-.*//')
update-docs:
	@find docs/ -type f \( -name "*.md" ! -name "upgrading.md" \) -exec $(SED) -i 's#$(PREVIOUS_CORE_VERSION)#$(OT_CORE_VERSION)#g' {} \;
	@find docs/ -type f \( -name "*.md" ! -name "upgrading.md" \) -exec $(SED) -i 's#$(PREVIOUS_OT_VERSION)#$(LATEST_OT_VERSION)#g' {} \;

################################################################################
# Release
################################################################################
#
# These targets should be used for the release process in order to make the modules
# contained within this repo importable.
# This is required because as of now Go doesn't allow importing modules being
# defined in repository's sub directories without having this directory name set
# as prefix for the tag
#
# So when we want to make pkg/exporter/sumologicexporter with version v0.0.43-beta.0
# importable then we need to create the following tag:
# `pkg/exporter/sumologicexporter/v0.0.43-beta.0`
# in order for it to be importable.
#
# Related issue: https://github.com/golang/go/issues/34055

# Example usage for the release:
#
# export TAG=v0.98.0-sumo-0
# make add-tag push-tag

.PHONY: add-tag
add-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Adding tag ${TAG}"
	@git tag -a ${TAG} -s -m "${TAG}"
	@set -e; for dir in $(ALL_EXPORTABLE_MODULES); do \
	  (echo Adding tag "$${dir:2}/$${TAG}" && \
	 	git tag -a "$${dir:2}/$${TAG}" -s -m "${dir:2}/${TAG}" ); \
	done

.PHONY: push-tag
push-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Pushing tag ${TAG}"
	@git push origin ${TAG}
	@set -e; for dir in $(ALL_EXPORTABLE_MODULES); do \
	  (echo Pushing tag "$${dir:2}/$${TAG}" && \
	 	git push origin "$${dir:2}/$${TAG}"); \
	done

.PHONY: delete-tag
delete-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Deleting tag ${TAG}"
	@git tag -d ${TAG}
	@set -e; for dir in $(ALL_EXPORTABLE_MODULES); do \
	  (echo Deleting tag "$${dir:2}/$${TAG}" && \
	 	git tag -d "$${dir:2}/$${TAG}" ); \
	done

.PHONY: delete-remote-tag
delete-remote-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Deleting remote tag ${TAG}"
	@git push --delete origin ${TAG}
	@set -e; for dir in $(ALL_EXPORTABLE_MODULES); do \
		(echo Deleting remote tag "$${dir:2}/$${TAG}" && \
		git push --delete origin "$${dir:2}/$${TAG}"); \
	done

.PHONY: prepare-tag
prepare-tag: install-gsed
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	$(SED) -i 's#\(gomod: github.com/SumoLogic/sumologic-otel-collector/.*\) v0.0.0-00010101000000-000000000000#\1 ${TAG}#g' \
		otelcolbuilder/.otelcol-builder.yaml
# Make sure to work with both tags starting not starting with v.
	$(SED) -i 's#\(gomod: github.com/SumoLogic/sumologic-otel-collector/.*\) \([^v].*\)#\1 v\2#g' \
		otelcolbuilder/.otelcol-builder.yaml


################################################################################
# Build
################################################################################

.PHONY: build
build:
	@$(MAKE) -C ./otelcolbuilder/ build
	@$(MAKE) -C ./pkg/tools/otelcol-config/ build

.PHONY: install-ocb
install-ocb:
	@$(MAKE) -C ./otelcolbuilder/ install-ocb

#-------------------------------------------------------------------------------

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

#-------------------------------------------------------------------------------

# Changelog management
# We use Towncrier (https://towncrier.readthedocs.io) for changelog management.

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

#-------------------------------------------------------------------------------

# vagrant

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
