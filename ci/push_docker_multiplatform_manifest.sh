#!/bin/bash

set -eo pipefail

# check for arm support only if we try to build it
if echo "${PLATFORMS}" | grep -q arm && ! docker buildx ls | grep -q arm ; then
    echo "Your Buildx seems to lack ARM architecture support"
    echo
    docker buildx ls
    exit 1
fi

if [[ -z "${BUILD_TAG}" ]]; then
    echo "No BUILD_TAG passed in, using 'latest' as default"
    BUILD_TAG="latest"
fi

if [[ -z "${BUILD_TYPE_SUFFIX}" ]]; then
    BUILD_TYPE_SUFFIX=""
fi

if [[ -z "${REPO_URL}" ]]; then
    echo "No REPO_URL passed in"
    exit 1
fi

PLATFORMS=("${@}")
if [[ 0 -eq "${#PLATFORMS[@]}" ]]; then
    echo "No PLATFORMS passed in as argument to this script"
    exit 1
fi

function push_manifest() {
    TAGS_IN_MANIFEST=()

    echo "Platforms:"
    for platform in "${PLATFORMS[@]}"
    do
        echo "${platform}"
        case "${platform}" in
        "linux/amd64")
            BUILD_ARCH="amd64"
            BUILD_PLATFORM="linux"
            ;;

        "linux/arm64")
            BUILD_ARCH="arm64"
            BUILD_PLATFORM="linux"
            ;;

        "linux/arm/v7")
            BUILD_ARCH="arm_v7"
            BUILD_PLATFORM="linux"
            ;;

        "windows/amd64/ltsc2022")
            BUILD_ARCH="amd64"
            BUILD_PLATFORM="windows"
            BASE_IMAGE_TAG_SUFFIX="-ltsc2022"
            ;;
        *)
            echo "Unsupported platform ${platform}"
            exit 1
            ;;
        esac

        TAGS_IN_MANIFEST+=("${REPO_URL}:${BUILD_TAG}${BUILD_TYPE_SUFFIX}-${BUILD_PLATFORM}-${BUILD_ARCH}${BASE_IMAGE_TAG_SUFFIX}")
    done

    echo "Tags in the manifest:"
    for tag in "${TAGS_IN_MANIFEST[@]}"
    do
        echo "${tag}"
    done

    echo
    set -x
    # Use docker manifest as docker buildx didn't create "${REPO_URL}:${BUILD_TAG}" correctly. It was containing only linux/amd64 image
    docker manifest create \
        "${REPO_URL}:${BUILD_TAG}${BUILD_TYPE_SUFFIX}" \
        "${TAGS_IN_MANIFEST[@]}"

    docker manifest push \
        "${REPO_URL}:${BUILD_TAG}${BUILD_TYPE_SUFFIX}"

    docker manifest create \
        "${REPO_URL}:latest${BUILD_TYPE_SUFFIX}" \
        "${TAGS_IN_MANIFEST[@]}"

    docker manifest push \
        "${REPO_URL}:latest${BUILD_TYPE_SUFFIX}"
}

push_manifest
