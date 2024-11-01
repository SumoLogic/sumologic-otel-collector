#!/bin/bash

declare -i major_version
declare -i minor_version
declare -i patch_version
declare -i sumo_version

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [-h] [core|sumo|productversion]

Detects the version to use for building otelcol-sumo using a combination of the
otelcol-builder config and Git tags. It can output version information in
different formats depending on the subcommand used.

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

parse_core_version() {
    config="otelcolbuilder/.otelcol-builder.yaml"
    regex='s/.*otelcol_version:[ ]+([0-9]+\.[0-9]+\.[0-9]+).*/\1/p'
    version="$(sed -En "${regex}" "${config}")"

    if [[ -z "${version}" ]]; then
        echo "Error: no otc version found in config: ${config}" >&2
        exit 1
    fi

    version_regex="^([0-9]+).([0-9]+).([0-9]+)$"

    if [[ $version =~ $version_regex ]]; then
        major_version="${BASH_REMATCH[1]}"
        minor_version="${BASH_REMATCH[2]}"
        patch_version="${BASH_REMATCH[3]}"
    else
        echo "Error: otc version does not match required version regex: ${version}" >&2
        exit 1
    fi
}

parse_sumo_version() {
    tags=()
    for tag in $(git tag -l "v$(version_core)-sumo-*"); do
        tags+=( "${tag}" )
    done

    for tag in "${tags[@]}"; do
        tag_regex="^v[0-9]+.[0-9]+.[0-9]+-sumo-([0-9])+$"
        if [[ $tag =~ $tag_regex ]]; then
            sumo_version="${BASH_REMATCH[1]}"
        fi
    done

    if [[ -z "${sumo_version}" ]]; then
        # No matching tags were found. Sumo version is 0.
        sumo_version="0"
    else
        # Matching tags found. Increment sumo_version by 1.
        (( sumo_version++ ))
    fi
}

# Validate version information using the Windows ProductVersion requirements.
validate() {
    if [[ -z "${major_version}" ]]; then
        echo "Major version cannot be empty" >&2
        exit 1
    fi

    if [[ $major_version -lt 0 ]]; then
        echo "Major version cannot be less than 0" >&2
        exit 1
    fi

    if [[ $major_version -gt 255 ]]; then
        echo "Major version cannot be greater than 255" >&2
        exit 1
    fi

    if [[ -z "${minor_version}" ]]; then
        echo "Minor version cannot be empty" >&2
        exit 1
    fi

    if [[ $minor_version -lt 0 ]]; then
        echo "Minor version cannot be less than 0" >&2
        exit 1
    fi

    if [[ $minor_version -gt 255 ]]; then
        echo "Minor version cannot be greater than 255" >&2
        exit 1
    fi

    # Patch version (known as the build version on Windows)
    if [[ -z "${patch_version}" ]]; then
        echo "Patch version cannot be empty" >&2
        exit 1
    fi

    if [[ $patch_version -lt 0 ]]; then
        echo "Patch version cannot be less than 0" >&2
        exit 1
    fi

    if [[ $patch_version -gt 65535 ]]; then
        echo "Patch version cannot be greater than 65,535" >&2
        exit 1
    fi

    # Sumo version (known as the internal version on Windows)
    if [[ -z "${sumo_version}" ]]; then
        echo "Sumo version cannot be empty" >&2
        exit 1
    fi
}

# Prints the semver version core. (e.g. A.B.C)
version_core() {
    echo "${major_version}.${minor_version}.${patch_version}"
}

# Prints the sumo version. (e.g. the X in A.B.C-sumo-X)
sumo_version() {
    echo "${sumo_version}"
}

# Convert the version to a Windows ProductVersion.
#
# https://learn.microsoft.com/en-us/windows/win32/msi/productversion
# MAJOR.MINOR.PATCH.SUMO -> MAJOR.MINOR.BUILD.INTERNAL
windows_product_version() {
    echo "${major_version}.${minor_version}.${patch_version}.${sumo_version}"
}

parse_params "$@"
parse_core_version
parse_sumo_version
validate

case "$1" in
    core) version_core ;;
    sumo) sumo_version ;;
    productversion) windows_product_version ;;
    *) bad_usage "Unknown argument: ${1}" ;;
esac
