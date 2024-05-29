#!/bin/bash

declare -i major_version
declare -i minor_version
declare -i patch_version
declare build_version
declare build_windows_version
declare ot_channel
declare -i ot_channel_version
declare sumo_channel
declare -i sumo_channel_version

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [-h] [core|sumo|productversion]

Detects the latest version from Git tags and outputs it in different formats
depending on the subcommand used.

Available options:

-h, --help      Print this help and exit

Available commands:
  core            Prints the version core (e.g. 1.2.3)
  sumo            Prints the sumo version (e.g. the X in 1.2.3-sumo-X)
  productversion  Prints the Windows ProductVersion (e.g. 1.2.3.4)
EOF
  exit
}

bad_usage() {
    echo "${1}"
    echo
    usage
}

parse_params() {
    while :; do
        case "${1-}" in
            -h | --help) usage ;;
            -?*)
                bad_usage "Unknown option: ${1}" ;;
            *) break ;;
        esac
        shift
    done

    args=("$@")
    [[ ${#args[@]} -eq 0 ]] && usage && exit 1
    [[ ${#args[@]} -ne 1 ]] && bad_usage "Too many arguments"

    return 0
}

parse_version_tag() {
    # shellcheck disable=SC2153
    version_tag="${VERSION_TAG}"
    if [ -z "${version_tag}" ]; then
        version_tag=$(git describe --tags --abbrev=0 --match "v[0-9]*" | head -n 1)
    fi

    version_regex="^v([0-9]+).([0-9]+).([0-9]+)((-(alpha|beta|rc|sumo)[-.]([0-9]+))(-(alpha|beta|rc).([0-9])+)?)?$"

    if [[ $version_tag =~ $version_regex ]]; then
        major_version="${BASH_REMATCH[1]}"
        minor_version="${BASH_REMATCH[2]}"
        patch_version="${BASH_REMATCH[3]}"
        ot_channel="${BASH_REMATCH[6]}"
        ot_channel_version="${BASH_REMATCH[7]}"
        sumo_channel="${BASH_REMATCH[9]}"
        sumo_channel_version="${BASH_REMATCH[10]}"
    else
        echo "Error: Tag does not match required version regex: ${version_tag}" >&2
        exit 1
    fi

    if [[ $ot_channel == "sumo" ]]; then
        if [[ $sumo_channel != "" ]]; then
            build_version="${ot_channel_version}-${sumo_channel}.${sumo_channel_version}"
            build_windows_version="${ot_channel_version}"
        else
            build_version="${ot_channel_version}"
            build_windows_version="${ot_channel_version}"
        fi
    elif [[ $ot_channel != "" ]]; then
        build_version="${ot_channel_version}"
        build_windows_version="${ot_channel_version}"
    fi

    if [[ $OVERRIDE_BUILD_VERSION != "" ]]; then
        number_regex='^[0-9]+$'
        if ! [[ $OVERRIDE_BUILD_VERSION =~ $number_regex ]]; then
            echo "Error: OVERRIDE_BUILD_VERSION is not a number" >&2
            exit 1
        fi
        build_version="${OVERRIDE_BUILD_VERSION}"
        build_windows_version="${OVERRIDE_BUILD_VERSION}"
    fi
}

# Validate version information using the Windows ProductVersion requirements.
validate() {
    if [[ -z "${major_version}" ]]; then
        echo "Major version cannot be empty"
        exit 1
    fi

    if [[ $major_version -lt 0 ]]; then
        echo "Major version cannot be less than 0"
        exit 1
    fi

    if [[ $major_version -gt 255 ]]; then
        echo "Major version cannot be greater than 255"
        exit 1
    fi

    if [[ -z "${minor_version}" ]]; then
        echo "Minor version cannot be empty"
        exit 1
    fi

    if [[ $minor_version -lt 0 ]]; then
        echo "Minor version cannot be less than 0"
        exit 1
    fi

    if [[ $minor_version -gt 255 ]]; then
        echo "Minor version cannot be greater than 255"
        exit 1
    fi

    # Patch version is also known as the build version on Windows
    if [[ -z "${patch_version}" ]]; then
        echo "Patch version cannot be empty"
        exit 1
    fi

    if [[ $patch_version -lt 0 ]]; then
        echo "Patch version cannot be less than 0"
        exit 1
    fi

    if [[ $patch_version -gt 65535 ]]; then
        echo "Patch version cannot be greater than 65,535"
        exit 1
    fi

    # Build version is also known as the internal version on Windows
    if [[ -z "${build_version}" ]]; then
        echo "Build version cannot be empty"
        exit 1
    fi

    # Build version is also known as the internal version on Windows
    if [[ -z "${build_windows_version}" ]]; then
        echo "Windows Build version cannot be empty"
        exit 1
    fi

    if [[ $ot_channel_version -lt 0 ]]; then
        echo "Build version cannot be less than 0"
        exit 1
    fi

    if [[ $ot_channel_version -gt 65535 ]]; then
        echo "Build version cannot be greater than 65,535"
        exit 1
    fi
}

# Prints the semver version core. (e.g. A.B.C)
version_core() {
    echo "${major_version}.${minor_version}.${patch_version}"
}

# Prints the sumo version. (e.g. the X in A.B.C-sumo-X)
sumo_version() {
    echo "${build_version}"
}

# Convert the version to a Windows ProductVersion.
#
# https://learn.microsoft.com/en-us/windows/win32/msi/productversion
# MAJOR.MINOR.PATCH.BUILD -> MAJOR.MINOR.BUILD.INTERNAL
windows_product_version() {
    echo "${major_version}.${minor_version}.${patch_version}.${build_windows_version}"
}

parse_params "$@"
parse_version_tag
validate

case "$1" in
    core) version_core ;;
    sumo) sumo_version ;;
    productversion) windows_product_version ;;
    *) bad_usage "Unknown argument: ${1}" ;;
esac
