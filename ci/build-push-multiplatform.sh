#!/bin/bash

set -eo pipefail

if echo "${PLATFORM}" | grep -v windows; then

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

if [[ -n "${BASE_IMAGE_TAG}" ]]; then
    BASE_IMAGE_TAG="-${BASE_IMAGE_TAG}"
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
    local BASE_IMAGE_TAG_SUFFIX
    set -x

    case "${PLATFORM}" in
    "linux/amd64"|"linux_amd64")
        readonly BUILD_ARCH="amd64"
        readonly BUILD_PLATFORM="linux"
        PLATFORM="linux/amd64"
        ;;

    "linux/arm64"|"linux_arm64")
        readonly BUILD_ARCH="arm64"
        readonly BUILD_PLATFORM="linux"
        PLATFORM="linux/arm64"
        ;;

    "windows/amd64"|"windows_amd64")
        readonly BUILD_ARCH="amd64"
        readonly BASE_IMAGE_TAG_SUFFIX="windows"
        PLATFORM="windows/amd64"
        ;;

    "windows/amd64/ltsc2022"|"windows_amd64_ltsc2022")
        readonly BUILD_ARCH="amd64"
        readonly BUILD_PLATFORM="windows"
        readonly BASE_IMAGE_TAG_SUFFIX="-ltsc2022"
        readonly BASE_IMAGE_TAG="ltsc2022"
        PLATFORM="windows/amd64"
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
    readonly TAG="${REPO_URL}:${BUILD_TAG}${BUILD_TYPE_SUFFIX}-${BUILD_PLATFORM}-${BUILD_ARCH}${BASE_IMAGE_TAG_SUFFIX}"
    local LATEST_TAG
    readonly LATEST_TAG="${REPO_URL}:latest${BUILD_TYPE_SUFFIX}-${BUILD_PLATFORM}-${BUILD_ARCH}${BASE_IMAGE_TAG_SUFFIX}"

    # --provenance=false for docker buildx ensures that we create manifest instead of manifest list
    if [[ "${PUSH}" == true ]]; then
        echo "Building tags: ${TAG}, ${LATEST_TAG}"

        if [[ "${BUILD_PLATFORM}" == "windows" ]]; then
            docker build \
                --file "${DOCKERFILE}" \
                --build-arg BUILD_TAG="${BUILD_TAG}" \
                --build-arg BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
                --build-arg BUILDKIT_INLINE_CACHE=1 \
                --platform="${PLATFORM}" \
                --tag "${LATEST_TAG}" \
                .

            docker tag "${LATEST_TAG}" "${TAG}"

            docker push "${LATEST_TAG}"
            docker push "${TAG}"
        else
            docker buildx build \
                --push \
                --file "${DOCKERFILE}" \
                --build-arg BUILD_TAG="${BUILD_TAG}" \
                --build-arg BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
                --build-arg BUILDKIT_INLINE_CACHE=1 \
                --platform="${PLATFORM}" \
                --tag "${LATEST_TAG}" \
                --tag "${TAG}" \
                --provenance=false \
                .
        fi
    else
        echo "Building tag: latest${BUILD_TYPE_SUFFIX}"
        if [[ "${BUILD_PLATFORM}" == "windows" ]]; then
            docker build \
                --file "${DOCKERFILE}" \
                --build-arg BUILD_TAG="latest${BUILD_TYPE_SUFFIX}" \
                --build-arg BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
                --build-arg BUILDKIT_INLINE_CACHE=1 \
                --platform="${PLATFORM}" \
                --tag "${REPO_URL}:latest${BUILD_TYPE_SUFFIX}" \
                .
        else
            # load flag is needed so that docker loads this image
            # for subsequent steps on github actions
            docker buildx build \
                --file "${DOCKERFILE}" \
                --build-arg BUILD_TAG="latest${BUILD_TYPE_SUFFIX}" \
                --build-arg BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
                --build-arg BUILDKIT_INLINE_CACHE=1 \
                --platform="${PLATFORM}" \
                --load \
                --tag "${REPO_URL}:latest${BUILD_TYPE_SUFFIX}" \
                --provenance=false \
                .
        fi
    fi
}

build_push "${PUSH}"
