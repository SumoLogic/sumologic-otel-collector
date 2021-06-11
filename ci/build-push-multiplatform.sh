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

TAGS_IN_MANIFEST=()

# build builds a container image for a designated platform.
#
# First param is a platform for which to build the image as accepted by docker
# buildx build command.
# e.g.linux/amd64, linux/arm64, linux/ppc64le, linux/s390x, linux/386,
# linux/arm/v7, linux/arm/v6
function build() {
    local PLATFORM
    readonly PLATFORM="${1}"

    local BUILD_OS
    local BUILD_ARCH

    case "${PLATFORM}" in
    "linux/amd64")
        readonly BUILD_OS="linux"
        readonly BUILD_ARCH="amd64"
        ;;

    "linux/arm64")
        readonly BUILD_OS="linux"
        readonly BUILD_ARCH="arm64"
        ;;

    "linux/arm/v7")
        readonly BUILD_OS="linux"
        readonly BUILD_ARCH="arm_v7"
        ;;

    *)
        echo "Unsupported platform ${PLATFORM}"
        exit 1
        ;;
    esac

    local TAG="${REPO_URL}:${BUILD_TAG}-${BUILD_ARCH}"
    TAGS_IN_MANIFEST+=("${TAG}")


    echo "Building tag:${TAG}"
    docker buildx build \
        --push \
        --file "${DOCKERFILE}" \
        --progress tty \
        --platform="${PLATFORM}" \
        --tag "${TAG}" \
        . &

    case "${BUILD_ARCH}" in
    "amd64")
        readonly amd64_pid=$!
        ;;

    "arm64")
        readonly arm64_pid=$!
        ;;

    "arm_v7")
        readonly arm_v7_pid=$!
        ;;

    *)
        echo "Unsupported platform ${BUILD_ARCH}"
        exit 1
        ;;
    esac
}

build "linux/amd64"
build "linux/arm64"
build "linux/arm/v7"

wait "${amd64_pid}"
wait "${arm64_pid}"
wait "${arm_v7_pid}"

echo "Tags in the manifest:"
for tag in "${TAGS_IN_MANIFEST[@]}"
do
    echo "${tag}"
done

docker manifest create --amend \
    "${REPO_URL}:${BUILD_TAG}" \
    "${TAGS_IN_MANIFEST[@]}"

docker manifest push "${REPO_URL}:${BUILD_TAG}"
