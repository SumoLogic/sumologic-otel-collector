#!/bin/bash

declare -i major_version
declare -i minor_version
declare -i patch_version
declare -i sumo_version

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [-h] [otc|sumo] [path/to/otelcol-sumo]

Retrieves the version of an otelcol-sumo binary by parsing the output of
otelcol-sumo's version flag.

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
    [[ ${#args[@]} -ne 2 ]] && bad_usage "Too many arguments"

    return 0
}

parse_version() {
    binary_path="$1"

    output="$(${binary_path} --version)"

    regex=".* v([0-9]+)\.([0-9]+)\.([0-9]+)\-sumo\-([0-9]+).*"

    if [[ $output =~ $regex ]]; then
        major_version="${BASH_REMATCH[1]}"
        minor_version="${BASH_REMATCH[2]}"
        patch_version="${BASH_REMATCH[3]}"
        sumo_version="${BASH_REMATCH[4]}"
    else
        echo "Error: version output does not match required regex: ${output}" >&2
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

# Prints the version. (e.g. A.B.C-sumo-X)
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
parse_version "$2"

case "$1" in
    core) version_core ;;
    sumo) sumo_version ;;
    productversion) windows_product_version ;;
    *) bad_usage "Unknown argument: ${1}" ;;
esac
