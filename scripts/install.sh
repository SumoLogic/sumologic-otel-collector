#!/usr/bin/env bash

set -euo pipefail

############################ Functions

function check_dependencies() {
    local error
    error=0
    for cmd in echo sudo sed curl head less grep sort tac mv chmod; do
        if ! command -v "${cmd}" &> /dev/null; then
            echo "Command '${cmd}' not found. Please install it."
            error=1
        fi
    done

    if [[ "${error}" == "1" ]] ; then
        exit 1
    fi
}

function get_latest_version() {
    local versions
    readonly versions="${1}"
    # sed 's/ /\n/g' converts spaces to new lines
    echo "${versions}" | sed 's/ /\n/g' | head -n 1
}

# Get available versions of otelcol-sumo
# skip prerelease and draft releases
# sort it from last to first
# remove v from beginning of version
function get_versions() {
    curl \
    -sH "Accept: application/vnd.github.v3+json" \
    https://api.github.com/repos/SumoLogic/sumologic-otel-collector/releases \
    | grep -E '(tag_name|"(draft|prerelease)")' \
    | tac \
    | sed 'N;N;s/.*true.*//' \
    | grep -o 'v.*"' \
    | sort -r \
    | sed 's/^v//;s/"$//'
}

# Get versions from provided one to the latest
get_versions_from() {
    local versions
    readonly versions="${1}"

    local from
    readonly from="${2}"

    local line
    readonly line="$(( $(echo "${versions}" | sed 's/ /\n/g' | grep -n "${from}$" | sed 's/:.*//g') - 1 ))"

    echo "${versions}" | sed 's/ /\n/g' | head -n "${line}" | sort
}

# Get OS type (linux or darwin)
function get_os_type() {
    local os_type
    # Detect OS using uname
    case "$(uname)" in
    Darwin)
        os_type=darwin
        ;;
    Linux)
        os_type=linux
        ;;
    *)
        echo -e "Unsupported OS type:\t$(uname)"
        exit 1
        ;;
    esac
    echo "${os_type}"
}

# Get arch type (amd64 or arm64)
function get_arch_type() {
    local arch_type
    case "$(uname -m)" in
    x86_64)
        arch_type=amd64
        ;;
    aarch64_be | aarch64 | armv8b | armv8l)
        arch_type=arm64
        ;;
    *)
        echo -e "Unsupported architecture type:\t$(uname -m)"
        exit 1
        ;;
    esac
    echo "${arch_type}"
}

# Get installed version of otelcol-sumo
function get_installed_version() {
    if [[ -f "/usr/local/bin/otelcol-sumo" ]]; then
        set +o pipefail
        /usr/local/bin/otelcol-sumo --version | grep -o 'v[0-9].*$' | sed 's/v//'
        set -o pipefail
    fi
}

# Ask to continue and abort if not
function ask_to_continue() {
    local choice
    read -rp "Continue (y/n)?" choice
    case "${choice}" in
    y|Y ) ;;
    n|N | * )
        echo "Aborting..."
        exit 1
        ;;
    esac
}

# Get changelog for specific version
# Only version description and breaking changes are taken
function get_changelog() {
    local version
    readonly version="${1}"

    local notes
    notes="$(echo -e "$(curl -s "https://api.github.com/repos/SumoLogic/sumologic-otel-collector/releases/tags/v${version}" | grep -o "body.*"  | sed 's/body": "//;s/"$//')")"
    readonly notes

    local changelog
    # 's/\[\([^\[]*\)\]\[[^\[]*\]/\1/g' replaces [$1][*] with $1
    # 's/\[\([^\[]*\)\]([^\()]*)/\1/g' replaces [$1](*) with $1
    # sed '$ d' removes last line
    changelog="$(echo "${notes}" | sed -e "/## v${version}/,/###/!d" | sed '$ d' | sed 's/\[\([^\[]*\)\]\[[^\[]*\]/\1/g;s/\[\([^\[]*\)\]([^\()]*)/\1/g')"
    changelog="${changelog}\n### Release address\n\nhttps://github.com/SumoLogic/sumologic-otel-collector/releases/tag/v${version}\n"
    # 's/\[#.*//' remove everything starting from `[#`
    changelog="${changelog}\n$(echo "${notes}" | sed -e '/### Breaking changes/,/###/!d' | sed '$ d' | sed 's/\[#.*//')"
    echo -e "${changelog}"
}

############################ Main code

check_dependencies

OS_TYPE="$(get_os_type)"
ARCH_TYPE="$(get_arch_type)"
readonly OS_TYPE ARCH_TYPE

echo -e "Detected OS type:\t${OS_TYPE}"
echo -e "Detected architecture:\t${ARCH_TYPE}"

VERSIONS="$(get_versions)"
INSTALLED_VERSION="$(get_installed_version)"
VERSION="$(get_latest_version "${VERSIONS}")"
readonly VERSIONS VERSION INSTALLED_VERSION

echo -e "Installed version:\t${INSTALLED_VERSION}"
echo -e "Version to install:\t${VERSION}"

# Check if otelcol is already in newest version
if [[ "${INSTALLED_VERSION}" == "${VERSION}" ]]; then
    echo "OpenTelemetry collector is already in newest (${VERSION}) version"
    exit
elif [[ -n "${INSTALLED_VERSION}" ]]; then
    read -rp "Press enter to see changelog"
    # Take versions from installed up to the newest
    BETWEEN_VERSIONS="$(get_versions_from "${VERSIONS}" "${INSTALLED_VERSION}")"
    readonly BETWEEN_VERSIONS
    for version in ${BETWEEN_VERSIONS}; do
        # Print changelog for every version
        get_changelog "${version}"
    done | less
fi

readonly LINK="https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v${VERSION}/otelcol-sumo-${VERSION}-${OS_TYPE}_${ARCH_TYPE}"

ask_to_continue
echo -e "Downloading:\t\t${LINK}"
curl -L "${LINK}" --output otelcol-sumo --progress-bar

echo -e "Moving otelcol-sumo to /usr/local/bin"
sudo mv otelcol-sumo /usr/local/bin/otelcol-sumo
echo -e "Setting /usr/local/bin/otelcol-sumo to be executable"
sudo chmod +x /usr/local/bin/otelcol-sumo

OUTPUT="$(otelcol-sumo --version || true)"
readonly OUTPUT

if [[ -z "${OUTPUT}" ]]; then
    echo "Installation failed. Please try again"
    exit 1
fi

echo -e "Installation succeded:\t$(otelcol-sumo --version)"
