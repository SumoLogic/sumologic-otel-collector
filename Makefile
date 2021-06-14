all: markdownlint yamllint

markdownlint: mdl

mdl:
	mdl --style .markdownlint/style.rb \
		README.md

yamllint:
	yamllint -c .yamllint.yaml \
		otelcolbuilder/.otelcol-builder.yaml

# ALL_MODULES includes ./* dirs (excludes . dir and example with go code)
ALL_MODULES := $(shell find ./pkg -type f -name "go.mod" -exec dirname {} \; | sort | egrep  '^./' )

.PHONY: gotest
gotest:
	$(MAKE) for-all CMD="make test"

.PHONY: golint
golint:
	$(MAKE) for-all CMD="make lint"

.PHONY: install-golint
install-golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
	sh -s -- -b $(shell go env GOPATH)/bin v1.40.1

.PHONY: for-all
for-all:
	@echo "running $${CMD} in root"
	@set -e; for dir in $(ALL_MODULES); do \
	  (cd "$${dir}" && \
	  	echo "running $${CMD} in $${dir}" && \
	 	$${CMD} ); \
	done

################################################################################
# Build
################################################################################

BUILD_TAG ?= latest
BUILD_CACHE_TAG = latest-builder-cache
IMAGE_NAME = sumologic-otel-collector
IMAGE_NAME_DEV = sumologic-otel-collector-dev

OPENSOURCE_ECR_URL = public.ecr.aws/a4t4y2n3
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

.PHONY: build-container
build-container-local:
	$(MAKE) _build \
		IMG="$(IMAGE_NAME)-local" \
		DOCKERFILE="Dockerfile_local" \
		TAG="$(BUILD_TAG)"

#-------------------------------------------------------------------------------

# dev

.PHONY: build-container-multiplatform-dev
build-container-multiplatform-dev:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(OPENSOURCE_REPO_URL_DEV)" \
		DOCKERFILE="Dockerfile_dev" \
		PLATFORM="$(PLATFORM)" \
		./ci/build-push-multiplatform.sh

.PHONY: push-container-manifest-dev
push-container-manifest-dev:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(OPENSOURCE_REPO_URL_DEV)" \
		./ci/push_docker_multiplatform_manifest.sh $(PLATFORMS)

# release

.PHONY: build-container-multiplatform
build-container-multiplatform:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(OPENSOURCE_REPO_URL)" \
		DOCKERFILE="Dockerfile" \
		PLATFORM="$(PLATFORM)" \
		./ci/build-push-multiplatform.sh

.PHONY: push-container-manifest
push-container-manifest:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(OPENSOURCE_REPO_URL)" \
		./ci/push_docker_multiplatform_manifest.sh $(PLATFORMS)

#-------------------------------------------------------------------------------

.PHONY: login
login:
	aws ecr-public get-login-password --region us-east-1 \
	| docker login --username AWS --password-stdin $(OPENSOURCE_ECR_URL)
