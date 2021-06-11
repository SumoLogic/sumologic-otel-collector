#!/bin/bash

set -eo pipefail

if ! docker buildx ls | grep -q arm ; then
    echo "Your Buildx seems to lack ARM architecture support"
    echo
    docker buildx ls
    exit 1
fi

if [[ -z "${BUILD_TAG}" ]]; then
    echo "No BUILD_TAG passed in, using 'latest' as default"
    BUILD_TAG="latest"
fi

if [[ -z "${REPO_URL}" ]]; then
    echo "No REPO_URL passed in"
    exit 1
fi

if [[ -z "${PLATFORM}" ]]; then
    echo "No PLATFORM passed in"
    exit 1
fi

# build builds a container image for a designated platform.
#
# First param is a platform for which to build the image as accepted by docker
# buildx build command.
# e.g.linux/amd64, linux/arm64, linux/ppc64le, linux/s390x, linux/386,
# linux/arm/v7, linux/arm/v6
function build() {
    local BUILD_ARCH

    case "${PLATFORM}" in
    "linux/amd64")
        readonly BUILD_ARCH="amd64"
        ;;

    "linux/arm64")
        readonly BUILD_ARCH="arm64"
        ;;

    "linux/arm/v7")
        readonly BUILD_ARCH="arm_v7"
        ;;

    *)
        echo "Unsupported platform ${PLATFORM}"
        exit 1
        ;;
    esac

    local TAG
    readonly TAG="${REPO_URL}:${BUILD_TAG}-${BUILD_ARCH}"

    echo "Building tag:${TAG}"
    docker buildx build \
        --push \
        --file "${DOCKERFILE}" \
		--build-arg BUILD_TAG="${BUILD_TAG}" \
		--build-arg BUILDKIT_INLINE_CACHE=1 \
        --platform="${PLATFORM}" \
        --tag "${TAG}" \
        .
}

build "${PLATFORM}"
