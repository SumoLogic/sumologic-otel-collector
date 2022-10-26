#!/bin/bash

set -eo pipefail

while ! docker buildx ls; do
    echo "Cannot connect to docker daemon"
    sleep 1
done

DOCKER_BUILDX_LS_OUT=$(docker buildx ls <<-END
END
)
readonly DOCKER_BUILDX_LS_OUT

# check for arm support only if we try to build it
if echo "${PLATFORM}" | grep -q arm && ! grep -q arm <<< "${DOCKER_BUILDX_LS_OUT}"; then
    echo "Your Buildx seems to lack ARM architecture support"
    echo "${DOCKER_BUILDX_LS_OUT}"
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

if [[ -z "${PLATFORM}" ]]; then
    echo "No PLATFORM passed in"
    exit 1
fi

PUSH=""
if [[ $# -eq 1 ]] && [[ "${1}" == "--push" ]]; then
    PUSH="true"
fi

# build_push builds a container image for a designated platform and then pushes
# it to container repository specified by REPO_URL variable.
#
# PLATFORM variable is the platform for which to build the image as accepted
# by docker buildx build command.
# e.g.linux/amd64, linux/arm64, linux/ppc64le, linux/s390x, linux/386,
# linux/arm/v7, linux/arm/v6
function build_push() {
    local BUILD_ARCH
    set -x

    case "${PLATFORM}" in
    "linux/amd64"|"linux_amd64")
        readonly BUILD_ARCH="amd64"
        PLATFORM="linux/amd64"
        ;;

    "linux/arm64"|"linux_arm64")
        readonly BUILD_ARCH="arm64"
        PLATFORM="linux/arm64"
        ;;

    # Can't really enable it for now because:
    # !shopify/sarama@v1.29.0/gssapi_kerberos.go:62:10: constant 4294967295 overflows int
    # ref: https://github.com/SumoLogic/sumologic-otel-collector/runs/2805247906
    # If we'd like to support arm then we'd need to provide a patch in sarama.
    #
    # "linux/arm/v7"|"linux_arm_v7"|"linux/arm"|"linux_arm")
    #     readonly BUILD_ARCH="arm"
    #     PLATFORM="linux/arm/v7"
    #     ;;

    *)
        echo "Unsupported platform ${PLATFORM}"
        exit 1
        ;;
    esac

    local TAG
    readonly TAG="${REPO_URL}:${BUILD_TAG}-${BUILD_ARCH}"
    local LATEST_TAG
    readonly LATEST_TAG="${REPO_URL}:latest${LATEST_TAG_FIPS_SUFFIX}-${BUILD_ARCH}"

    if [[ "${PUSH}" == true ]]; then
        echo "Building tag: ${TAG}"
        docker buildx build \
            --push \
            --file "${DOCKERFILE}" \
            --build-arg BUILD_TAG="${BUILD_TAG}" \
            --build-arg BUILDKIT_INLINE_CACHE=1 \
            --platform="${PLATFORM}" \
            --tag "${TAG}" \
            .

        # This is needed on CI because the above build command does not include
        # --load flag, which is forbidded to be used together with --push, hence
        # the docker pull.
        docker pull "${TAG}"

        echo "Tagging: ${LATEST_TAG}"
        docker tag "${TAG}" "${LATEST_TAG}"
        docker push "${LATEST_TAG}"
    else
        echo "Building tag: latest${LATEST_TAG_FIPS_SUFFIX}"
        # load flag is needed so that docker loads this image
        # for subsequent steps on github actions
        docker buildx build \
            --file "${DOCKERFILE}" \
            --build-arg BUILD_TAG="latest${LATEST_TAG_FIPS_SUFFIX}" \
            --build-arg BUILDKIT_INLINE_CACHE=1 \
            --platform="${PLATFORM}" \
            --load \
            --tag "${REPO_URL}:latest${LATEST_TAG_FIPS_SUFFIX}" \
            .
    fi
}

build_push "${PUSH}"
