all: markdownlint yamllint

markdownlint: mdl

mdl:
	mdl --style .markdownlint/style.rb \
		README.md \
		docs

yamllint:
	yamllint -c .yamllint.yaml \
		otelcolbuilder/.otelcol-builder.yaml

markdown-links-lint:
	./ci/markdown_links_lint.sh

markdown-link-check:
	./ci/markdown_link_check.sh

# ALL_MODULES includes ./* dirs (excludes . dir and example with go code)
ALL_MODULES := $(shell find ./pkg -type f -name "go.mod" -exec dirname {} \; | sort | egrep  '^./' )

.PHONY: gotest
gotest:
	$(MAKE) for-all CMD="make test"

.PHONY: golint
golint:
	$(MAKE) for-all CMD="make lint"

.PHONY: gomod-download-all
gomod-download-all:
	$(MAKE) for-all CMD="make mod-download-all"

.PHONY: install-golint
install-golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
	sh -s -- -b $(shell go env GOPATH)/bin v1.42.1

.PHONY: for-all
for-all:
	@echo "running $${CMD} in root"
	@set -e; for dir in $(ALL_MODULES); do \
	  (cd "$${dir}" && \
	  	echo "running $${CMD} in $${dir}" && \
	 	$${CMD} ); \
	done

.PHONY: check-uniform-dependencies
check-uniform-dependencies:
	./ci/check_uniform_dependencies.sh

################################################################################
# Build
################################################################################

.PHONY: build
build:
	$(MAKE) -C ./otelcolbuilder/ build

BUILD_TAG ?= latest
BUILD_CACHE_TAG = latest-builder-cache
IMAGE_NAME = sumologic-otel-collector
IMAGE_NAME_DEV = sumologic-otel-collector-dev

LEGACY_ECR_URL = public.ecr.aws/sumologic
LEGACY_REPO_URL = $(LEGACY_ECR_URL)/$(IMAGE_NAME)
LEGACY_REPO_URL_DEV = $(LEGACY_ECR_URL)/$(IMAGE_NAME_DEV)

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

.PHONY: build-push-container-multiplatform-dev
build-push-container-multiplatform-dev:
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

.PHONY: build-push-container-multiplatform-legacy-dev
build-push-container-multiplatform-legacy-dev:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(LEGACY_REPO_URL_DEV)" \
		DOCKERFILE="Dockerfile_dev" \
		PLATFORM="$(PLATFORM)" \
		./ci/build-push-multiplatform.sh

.PHONY: push-container-manifest-legacy-dev
push-container-manifest-legacy-dev:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(LEGACY_REPO_URL_DEV)" \
		./ci/push_docker_multiplatform_manifest.sh $(PLATFORMS)

# release

.PHONY: build-push-container-multiplatform
build-push-container-multiplatform:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(OPENSOURCE_REPO_URL)" \
		DOCKERFILE="Dockerfile" \
		PLATFORM="$(PLATFORM)" \
		./ci/build-push-multiplatform.sh

.PHONY: build-push-container-multiplatform-legacy
build-push-container-multiplatform-legacy:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(LEGACY_REPO_URL)" \
		DOCKERFILE="Dockerfile" \
		PLATFORM="$(PLATFORM)" \
		./ci/build-push-multiplatform.sh

.PHONY: push-container-manifest
push-container-manifest:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(OPENSOURCE_REPO_URL)" \
		./ci/push_docker_multiplatform_manifest.sh $(PLATFORMS)

.PHONY: push-container-manifest-legacy
push-container-manifest-legacy:
	BUILD_TAG="$(BUILD_TAG)" \
		REPO_URL="$(LEGACY_REPO_URL)" \
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

.PHONY: login-legacy
login-legacy:
	$(MAKE) _login \
		ECR_URL="$(LEGACY_ECR_URL)"
