GOLANGCI_LINT_VERSION ?= v1.45.2
SHELL := /usr/bin/env bash

all: markdownlint yamllint

.PHONY: markdownlint
markdownlint: mdl

MD_FILES := $(shell find ./pkg -type f -name "*.md")

ifeq ($(shell go env GOOS),darwin)
SED ?= gsed
else
SED ?= sed
endif

.PHONY: mdl
mdl:
	mdl --style .markdownlint/style.rb \
		$(MD_FILES) \
		otelcolbuilder/README.md \
		docs

yamllint:
	yamllint -c .yamllint.yaml \
		otelcolbuilder/.otelcol-builder.yaml

markdown-links-lint:
	./ci/markdown_links_lint.sh

markdown-link-check:
	./ci/markdown_link_check.sh

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
	@$(MAKE) for-all CMD="make test"

.PHONY: golint
golint:
	@$(MAKE) for-all CMD="make lint"

.PHONY: gomod-download-all
gomod-download-all:
	@$(MAKE) for-all CMD="make mod-download-all"

.PHONY: install-golint
install-golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

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

OT_CORE_VERSION := $(shell grep "otelcol_version: .*" otelcolbuilder/.otelcol-builder.yaml | cut -f 4 -d " ")
# usage: make update-ot-core OT_CORE_NEW_VERSION=x.x.x
.PHONY: update-ot-core
update-ot-core:
	@test $(OT_CORE_NEW_VERSION) || (echo "usage: make update-otc-core OT_CORE_NEW_VERSION=x.x.x"; exit 1);
	@echo "updating OT core from $(OT_CORE_VERSION) to $(OT_CORE_NEW_VERSION)"
	@sed -i "" "s/$(OT_CORE_VERSION)/$(OT_CORE_NEW_VERSION)/" otelcolbuilder/.otelcol-builder.yaml
	@sed -i "" "s/$(OT_CORE_VERSION)/$(OT_CORE_NEW_VERSION)/" otelcolbuilder/Makefile
	@sed -i "" "s/$(OT_CORE_VERSION)/$(OT_CORE_NEW_VERSION)/" README.md
	@sed -i "" "s/$(OT_CORE_VERSION)/$(OT_CORE_NEW_VERSION)/" docs/Configuration.md
	@sed -i "" "s/$(OT_CORE_VERSION)/$(OT_CORE_NEW_VERSION)/" docs/KnownIssues.md
	@sed -i "" "s/$(OT_CORE_VERSION)/$(OT_CORE_NEW_VERSION)/" pkg/receiver/telegrafreceiver/README.md
	@find . -type f -name "go.mod" -exec sed -i "" "s/$(OT_CORE_VERSION)/$(OT_CORE_NEW_VERSION)/" {} \;
	@echo "building OT distro to check for breakage"
	make gomod-download-all
	pushd otelcolbuilder \
		&& make install-builder \
		&& make build \
		&& popd

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

# Exemplar usage for the release:
#
# export TAG=v0.0.43-beta.0
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
	@git push upstream ${TAG}
	@set -e; for dir in $(ALL_EXPORTABLE_MODULES); do \
	  (echo Pushing tag "$${dir:2}/$${TAG}" && \
	 	git push upstream "$${dir:2}/$${TAG}"); \
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

.PHONY: prepare-tag
prepare-tag:
ifeq ($(shell go env GOOS),darwin)
	@which gsed || brew install gsed
endif
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	$(SED) -i 's#\(gomod: "github.com/SumoLogic/sumologic-otel-collector/.*\) v0.0.0-00010101000000-000000000000#\1 ${TAG}#g' \
		otelcolbuilder/.otelcol-builder.yaml
# Make sure to work with both tags starting not starting with v.
	$(SED) -i 's#\(gomod: "github.com/SumoLogic/sumologic-otel-collector/.*\) \([^v].*\)#\1 v\2#g' \
		otelcolbuilder/.otelcol-builder.yaml


################################################################################
# Build
################################################################################

.PHONY: build
build:
	@$(MAKE) -C ./otelcolbuilder/ build

BUILD_TAG ?= latest
BUILD_CACHE_TAG = latest-builder-cache
IMAGE_NAME = sumologic-otel-collector
IMAGE_NAME_DEV = sumologic-otel-collector-dev

OPENSOURCE_ECR_URL = public.ecr.aws/sumologic
OPENSOURCE_REPO_URL = $(OPENSOURCE_ECR_URL)/$(IMAGE_NAME)
OPENSOURCE_REPO_URL_DEV = $(OPENSOURCE_ECR_URL)/$(IMAGE_NAME_DEV)

.PHONY: _build
_build:
	DOCKER_BUILDKIT=1 docker build \
		--file $(DOCKERFILE) \
		--build-arg BUILD_TAG=$(TAG) \
		--build-arg BUILDKIT_INLINE_CACHE=1 \
		--tag $(IMG):$(TAG) \
		.

.PHONY: build-container-local
build-container-local:
	$(MAKE) _build \
		IMG="$(IMAGE_NAME)-local" \
		DOCKERFILE="Dockerfile_local" \
		TAG="$(BUILD_TAG)"

.PHONY: build-container-dev
build-container-dev:
	$(MAKE) _build \
		IMG="$(IMAGE_NAME)-dev" \
		DOCKERFILE="Dockerfile_dev" \
		TAG="$(BUILD_TAG)"

#-------------------------------------------------------------------------------

# dev

.PHONY: _build-container-multiplatform-dev
_build-container-multiplatform-dev:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(OPENSOURCE_REPO_URL_DEV)" \
		DOCKERFILE="Dockerfile_dev" \
		PLATFORM="$(PLATFORM)" \
		./ci/build-push-multiplatform.sh $(PUSH)

.PHONY: build-container-multiplatform-dev
build-container-multiplatform-dev:
	$(MAKE) _build-container-multiplatform-dev PUSH=

.PHONY: build-container-multiplatform-dev
build-push-container-multiplatform-dev:
	$(MAKE) _build-container-multiplatform-dev PUSH=--push

.PHONY: push-container-manifest-dev
push-container-manifest-dev:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(OPENSOURCE_REPO_URL_DEV)" \
		./ci/push_docker_multiplatform_manifest.sh $(PLATFORMS)

# release

.PHONY: _build-container-multiplatform
_build-container-multiplatform:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(OPENSOURCE_REPO_URL)" \
		DOCKERFILE="Dockerfile" \
		PLATFORM="$(PLATFORM)" \
		./ci/build-push-multiplatform.sh $(PUSH)

.PHONY: build-container-multiplatform
build-container-multiplatform:
	$(MAKE) _build-container-multiplatform PUSH=

.PHONY: build-push-container-multiplatform
build-push-container-multiplatform:
	$(MAKE) _build-container-multiplatform PUSH=--push

.PHONY: test-built-image
test-built-image:
	docker run --rm "$(OPENSOURCE_REPO_URL):$(BUILD_TAG)" --version

.PHONY: push-container-manifest
push-container-manifest:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(OPENSOURCE_REPO_URL)" \
		./ci/push_docker_multiplatform_manifest.sh $(PLATFORMS)

#-------------------------------------------------------------------------------

.PHONY: _login
_login:
	aws ecr-public get-login-password --region us-east-1 \
	| docker login --username AWS --password-stdin $(ECR_URL)

.PHONY: login
login:
	$(MAKE) _login \
		ECR_URL="$(OPENSOURCE_ECR_URL)"
