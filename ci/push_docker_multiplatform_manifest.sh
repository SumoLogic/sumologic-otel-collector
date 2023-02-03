#!/bin/bash

set -eo pipefail

# check for arm support only if we try to build it
if echo "${PLATFORM}" | grep -q arm && ! docker buildx ls | grep -q arm ; then
    echo "Your Buildx seems to lack ARM architecture support"
    echo
    docker buildx ls
    exit 1
fi

if [[ -z "${BUILD_TAG}" ]]; then
    echo "No BUILD_TAG passed in, using 'latest' as default"
    BUILD_TAG="latest"
fi

if [[ -z "${LATEST_TAG_FIPS_SUFFIX}" ]]; then
    LATEST_TAG_FIPS_SUFFIX=""
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
            ;;

        "linux/arm64")
            BUILD_ARCH="arm64"
            ;;

        "linux/arm/v7")
            BUILD_ARCH="arm_v7"
            ;;

        *)
            echo "Unsupported platform ${platform}"
            exit 1
            ;;
        esac

        TAGS_IN_MANIFEST+=("${REPO_URL}:${BUILD_TAG}-${BUILD_ARCH}")
    done

    echo "Tags in the manifest:"
    for tag in "${TAGS_IN_MANIFEST[@]}"
    do
        echo "${tag}"
    done

    TAGS_WITH_AMEND_COMMAND=()

    for T in "$TAGS_IN_MANIFEST[@]"
    do
        TAGS_WITH_AMEND_COMMAND+=("--amend ${T}")
    done

    echo
    set -x
    docker manifest create \
        "${REPO_URL}:${BUILD_TAG}" \
        "${TAGS_WITH_AMEND_COMMAND[@]}"
    docker manifest push "${REPO_URL}:${BUILD_TAG}"

    docker manifest create \
        "${REPO_URL}:latest${LATEST_TAG_FIPS_SUFFIX}" \
        "${TAGS_WITH_AMEND_COMMAND[@]}"
    docker manifest push "${REPO_URL}:latest${LATEST_TAG_FIPS_SUFFIX}"
}

push_manifest
