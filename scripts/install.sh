#!/usr/bin/env bash

set -euo pipefail

############################ Static variables

ARG_SHORT_TOKEN='i'
ARG_LONG_TOKEN='installation-token'
ARG_SHORT_HELP='h'
ARG_LONG_HELP='help'
ARG_SHORT_API='a'
ARG_LONG_API='api'
ARG_SHORT_TAG='t'
ARG_LONG_TAG='tag'
ARG_SHORT_VERSION='v'
ARG_LONG_VERSION='version'
ARG_SHORT_YES='y'
ARG_LONG_YES='yes'
ARG_SHORT_CONFIG='c'
ARG_LONG_CONFIG='config'
ARG_SHORT_STORAGE='s'
ARG_LONG_STORAGE='storage'
ARG_SHORT_COLLECTOR='n'
ARG_LONG_COLLECTOR='collector-name'
ARG_SHORT_SYSTEMD='d'
ARG_LONG_SYSTEMD='disable-systemd-installation'

############################ Variables (see set_defaults function for default values)

# Support providing install_token as env
set +u
if [[ -z "${INSTALL_TOKEN}" ]]; then
    INSTALL_TOKEN=""
fi
set -u

API_BASE_URL=""
FIELDS=""
COLLECTOR_NAME=""
VERSION=""
CONTINUE=false
FILE_STORAGE=""
CONFIG_DIRECTORY=""
SYSTEMD_CONFIG=""

# set by dependency_check
SYSTEMD_DISABLED=false

############################ Functions

function usage() {
  cat << EOF

Usage: bash install.sh [--${ARG_LONG_TOKEN} <token>] [--${ARG_LONG_COLLECTOR} name] [--${ARG_LONG_TAG} <key>=<value> [ --${ARG_LONG_TAG} ...]] [--${ARG_LONG_API} <url>] [--${ARG_LONG_CONFIG} <config dir path>] [--${ARG_LONG_STORAGE} <storage dir path>] [--${ARG_LONG_VERSION} <version>] [--${ARG_LONG_YES}] [--${ARG_LONG_VERSION} <version>] [--${ARG_LONG_HELP}]
  -${ARG_SHORT_TOKEN}, --${ARG_LONG_TOKEN} <token>     Installation token
  -${ARG_SHORT_COLLECTOR}, --${ARG_LONG_COLLECTOR} <name>          Collector name (default is your hostname)
  -${ARG_SHORT_TAG}, --${ARG_LONG_TAG} <key=value>                Tag in format key=value

  -${ARG_SHORT_API}, --${ARG_LONG_API} <url>                      Api URL
  -${ARG_SHORT_CONFIG}, --${ARG_LONG_CONFIG} <config dir path>       Path to the configuration directory (default is '/etc/otelcol-sumo')
  -${ARG_SHORT_SYSTEMD}, --${ARG_LONG_SYSTEMD}   Do not set up systemd service
  -${ARG_SHORT_STORAGE}, --${ARG_LONG_STORAGE} <storage dir path>     Path to the storage directory (default is '/var/lib/sumologic/file_storage')
  -${ARG_SHORT_VERSION}, --${ARG_LONG_VERSION} <version>              Manually specified version, e.g. 0.55.0-sumo-0
  -${ARG_SHORT_YES}, --${ARG_LONG_YES}                            Do not ask for confirmation

  -${ARG_SHORT_HELP}, --${ARG_LONG_HELP}                           Prints this help

Supported env variables:
  INSTALL_TOKEN=<token>                Equivalent of '--${ARG_LONG_TOKEN} <token>'
EOF
}

function set_defaults() {
    COLLECTOR_NAME="$(hostname)"
    FILE_STORAGE="/var/lib/sumologic/file_storage"
    CONFIG_DIRECTORY="/etc/otelcol-sumo"
    SYSTEMD_CONFIG="/etc/systemd/system/otelcol-sumo.service"
}

function parse_options() {
  # Transform long options to short ones
  for arg in "$@"; do

    shift
    case "$arg" in
      "--${ARG_LONG_HELP}")
        set -- "$@" "-${ARG_SHORT_HELP}"
        ;;
      "--${ARG_LONG_TOKEN}")
        set -- "$@" "-${ARG_SHORT_TOKEN}"
        ;;
      "--${ARG_LONG_API}")
        set -- "$@" "-${ARG_SHORT_API}"
        ;;
      "--${ARG_LONG_TAG}")
        set -- "$@" "-${ARG_SHORT_TAG}"
        ;;
      "--${ARG_LONG_YES}")
        set -- "$@" "-${ARG_SHORT_YES}"
        ;;
      "--${ARG_LONG_VERSION}")
        set -- "$@" "-${ARG_SHORT_VERSION}"
        ;;
      "--${ARG_LONG_CONFIG}")
        set -- "$@" "-${ARG_SHORT_CONFIG}"
        ;;
      "--${ARG_LONG_STORAGE}")
        set -- "$@" "-${ARG_SHORT_STORAGE}"
        ;;
      "--${ARG_LONG_COLLECTOR}")
        set -- "$@" "-${ARG_SHORT_COLLECTOR}"
        ;;
      "--${ARG_LONG_SYSTEMD}")
        set -- "$@" "-${ARG_SHORT_SYSTEMD}"
        ;;
      "-${ARG_SHORT_TOKEN}"|"-${ARG_SHORT_HELP}"|"-${ARG_SHORT_API}"|"-${ARG_SHORT_TAG}"|"-${ARG_SHORT_VERSION}"|"-${ARG_SHORT_YES}"|"-${ARG_SHORT_CONFIG}"|"-${ARG_SHORT_STORAGE}"|"-${ARG_SHORT_COLLECTOR}"|"-${ARG_SHORT_SYSTEMD}")
        set -- "$@" "${arg}"   ;;
      -*)
        echo "Unknown option ${arg}"; usage; exit 1 ;;
      *)
        set -- "$@" "$arg" ;;
    esac
  done

  # Parse short options
  OPTIND=1

  while true; do
    set +e
    getopts "${ARG_SHORT_HELP}${ARG_SHORT_TOKEN}:${ARG_SHORT_API}:${ARG_SHORT_TAG}:${ARG_SHORT_VERSION}:${ARG_SHORT_YES}${ARG_SHORT_CONFIG}:${ARG_SHORT_STORAGE}:${ARG_SHORT_COLLECTOR}:${ARG_SHORT_SYSTEMD}" opt
    set -e

    # Invalid argument catched, print and exit
    if [[ $? != 0 && ${OPTIND} -le $# ]]; then
      echo "Invalid argument:" "${@:${OPTIND}:1}"
      usage
      exit 1
    fi

    # Validate opt and set arguments
    case "$opt" in
      "${ARG_SHORT_HELP}")      usage; exit 0 ;;
      "${ARG_SHORT_TOKEN}")     INSTALL_TOKEN="${OPTARG}" ;;
      "${ARG_SHORT_API}")       API_BASE_URL="${OPTARG}" ;;
      "${ARG_SHORT_CONFIG}")    CONFIG_DIRECTORY="${OPTARG}" ;;
      "${ARG_SHORT_STORAGE}")   FILE_STORAGE="${OPTARG}" ;;
      "${ARG_SHORT_VERSION}")   VERSION="${OPTARG}" ;;
      "${ARG_SHORT_COLLECTOR}") COLLECTOR_NAME="${OPTARG}" ;;
      "${ARG_SHORT_YES}")       CONTINUE=true ;;
      "${ARG_SHORT_SYSTEMD}")   SYSTEMD_DISABLED=true ;;
      "${ARG_SHORT_TAG}")
        if [[ "${OPTARG}" != ?*"="* ]]; then
            echo "Invalid tag: '${OPTARG}'. Should be in 'key=value' format"
            usage
            exit 1
        fi

        # Cannot use `\n` and have to use `\\` as break line due to OSx sed implementation
        FIELDS="${FIELDS}\\
      ${OPTARG/=/: }" ;;
    "?")                        ;;
      *)                        usage; exit 1 ;;
    esac

    # Exit loop as we iterated over all arguments
    if [[ "${OPTIND}" -gt $# ]]; then
      break
    fi
  done
}

# Get github rate limit
function github_rate_limit() {
    curl -X GET https://api.github.com/rate_limit -v 2>&1 | grep x-ratelimit-remaining | grep -oE "[0-9]+"
}

function check_dependencies() {
    local error
    error=0
    for cmd in echo sudo sed curl head grep sort tac mv chmod envsubst getopts hostname; do
        if ! command -v "${cmd}" &> /dev/null; then
            echo "Command '${cmd}' not found. Please install it."
            error=1
        fi
    done

    if ! command -v systemctl &> /dev/null; then
        SYSTEMD_DISABLED=true
    fi

    if [[ "${error}" == "1" ]] ; then
        exit 1
    fi
}

function get_latest_version() {
    local versions
    readonly versions="${1}"

    # get latest version directly from website if there is no versions from api
    if [[ -z "${versions}" ]]; then
        curl -s https://github.com/SumoLogic/sumologic-otel-collector/releases | grep -oE '/SumoLogic/sumologic-otel-collector/releases/tag/(.*)"' | head -n 1 | sed 's%/SumoLogic/sumologic-otel-collector/releases/tag/v\([^"]*\)".*%\1%g'
    else
        # sed 's/ /\n/g' converts spaces to new lines
        echo "${versions}" | sed 's/ /\n/g' | head -n 1
    fi
}

# Get available versions of otelcol-sumo
# skip prerelease and draft releases
# sort it from last to first
# remove v from beginning of version
function get_versions() {
    # returns empty in case we exceeded github rate limit
    if [[ "$(github_rate_limit)" == "0" ]]; then
        return
    fi

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

    # Return if there is no installed version
    if [[ "${from}" == "" ]]; then
        return 0
    fi

    local line
    readonly line="$(( $(echo "${versions}" | sed 's/ /\n/g' | grep -n "${from}$" | sed 's/:.*//g') - 1 ))"

    if [[ "${line}" -gt "0" ]]; then
        echo "${versions}" | sed 's/ /\n/g' | head -n "${line}" | sort
    fi
    return 0
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
    aarch64_be | aarch64 | armv8b | armv8l | arm64)
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
    if [[ "${CONTINUE}" == true ]]; then
        return 0
    fi

    local choice
    read -rp "Continue (y/N)?" choice
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

    # 's/\[\([^\[]*\)\]\[[^\[]*\]/\1/g' replaces [$1][*] with $1
    # 's/\[\([^\[]*\)\]([^\()]*)/\1/g' replaces [$1](*) with $1
    local notes
    notes="$(echo -e "$(curl -s "https://api.github.com/repos/SumoLogic/sumologic-otel-collector/releases/tags/v${version}" | grep -o "body.*"  | sed 's/body": "//;s/"$//' | sed 's/\[\([^\[]*\)\]\[[^\[]*\]/\1/g;s/\[\([^\[]*\)\]([^\()]*)/\1/g')")"
    readonly notes

    local changelog
    # sed '$ d' removes last line
    changelog="$(echo "${notes}" | sed -e "/## v${version}/,/###/!d" | sed '$ d')"
    changelog="${changelog}\n### Release address\n\nhttps://github.com/SumoLogic/sumologic-otel-collector/releases/tag/v${version}\n"
    # 's/\[#.*//' remove everything starting from `[#`
    # 's/\[\([^\[]*\)\]/\1/g' replaces [$1] with $1
    changelog="${changelog}\n$(echo "${notes}" | sed -e '/### Changelog/,/###/!d' | sed '$ d' | sed 's/\[#.*//;s/\[\([^\[]*\)\]/\1/g')"
    changelog="${changelog}\n$(echo "${notes}" | sed -e '/### Breaking changes/,/###/!d' | sed '$ d' | sed 's/\[#.*//;s/\[\([^\[]*\)\]/\1/g')"
    echo -e "${changelog}"
}

# Get full changelog if there is no versions from API
function get_full_changelog() {
    local version
    readonly version="${1}"

    local notes
    notes="$(echo -e "$(curl -s "https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/v${version}/CHANGELOG.md")")"
    readonly notes

    local changelog
    # 's/\[\([^\[]*\)\]\[[^\[]*\]/\1/g' replaces [$1][*] with $1
    # 's/\[\([^\[]*\)\]([^\()]*)/\1/g' replaces [$1](*) with $1
    changelog="$(echo "${notes}" | sed 's/\[\([^\[]*\)\]\[[^\[]*\]/\1/g;s/\[\([^\[]*\)\]([^\()]*)/\1/g' | sed "s%# Changelog%# Changelog\n\nAddress: https://github.com/SumoLogic/sumologic-otel-collector/blob/v${version}/CHANGELOG.md%g")"
    # s/## \[\(.*\)\]/## \1/g changes `## [$1]` to `## $1`
    # 's/\[.*//' remove everything starting from `[#`
    # 's/\[\([^\[]*\)\]/\1/g' replaces [$1] with $1
    changelog="$(echo "${changelog}" | sed 's/^## \[\(.*\)\]/## \1/g' | sed '/^\[.*/d;;s/\[\([^\[]*\)\]/\1/g'))"
    echo -e "${changelog}"
}

############################ Main code

check_dependencies
set_defaults
parse_options "$@"

readonly INSTALL_TOKEN API_BASE_URL FIELDS COLLECTOR_NAME CONTINUE FILE_STORAGE CONFIG_DIRECTORY SYSTEMD_CONFIG SYSTEMD_DISABLED

OS_TYPE="$(get_os_type)"
ARCH_TYPE="$(get_arch_type)"
readonly OS_TYPE ARCH_TYPE

echo -e "Detected OS type:\t${OS_TYPE}"
echo -e "Detected architecture:\t${ARCH_TYPE}"

echo -e "Getting installed version..."
INSTALLED_VERSION="$(get_installed_version)"
echo -e "Installed version:\t${INSTALLED_VERSION}"

echo -e "Getting versions..."
VERSIONS="$(get_versions)"

# Use user's version if set, otherwise get latest version from API (or website)
if [[ -z "${VERSION}" ]]; then
    VERSION="$(get_latest_version "${VERSIONS}")"
fi

readonly VERSIONS VERSION INSTALLED_VERSION

echo -e "Version to install:\t${VERSION}"

# Check if otelcol is already in newest version
if [[ "${INSTALLED_VERSION}" == "${VERSION}" ]]; then
    echo -e "OpenTelemetry collector is already in newest (${VERSION}) version"
else
    if [[ -z "${INSTALLED_VERSION}" ]]; then
        # Take versions from installed up to the newest
        BETWEEN_VERSIONS="$(get_versions_from "${VERSIONS}" "${INSTALLED_VERSION}")"
        readonly BETWEEN_VERSIONS
        echo "${BETWEEN_VERSIONS}"

        # Get full changelog if we were unable to access github API
        if [[ -z "${BETWEEN_VERSIONS}" ]] || [[ "$(github_rate_limit)" < "$(echo BETWEEN_VERSIONS | wc -w)" ]]; then
            echo -e "Showing full changelog up to ${VERSION}"
            read -rp "Press enter to see changelog"
            get_full_changelog "${VERSION}"
        else
            read -rp "Press enter to see changelog"
            for version in ${BETWEEN_VERSIONS}; do
                # Print changelog for every version
                get_changelog "${version}"
            done
        fi
    fi

    readonly LINK="https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v${VERSION}/otelcol-sumo-${VERSION}-${OS_TYPE}_${ARCH_TYPE}"

    ask_to_continue
    echo -e "Downloading:\t\t${LINK}"
    curl -fL "${LINK}" --output otelcol-sumo --progress-bar

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
fi

# Exit if install token is not set
if [[ -z "${INSTALL_TOKEN}" ]]; then
    exit 0
fi

echo 'We are going to get and set up default configuration for you'
ask_to_continue
# Preparing default configuration
readonly CONFIG_PATH="${CONFIG_DIRECTORY}/config.yaml"

if [[ -f "${CONFIG_PATH}" ]]; then
    echo "Configuration (${CONFIG_PATH}) already exist)"
else

    echo -e "Creating file_storage directory (${FILE_STORAGE})"
    sudo mkdir -p "${FILE_STORAGE}"

    echo -e "Creating configuration directory (${CONFIG_DIRECTORY})"
    sudo mkdir -p "${CONFIG_DIRECTORY}"


    echo "Generating configuration and saving as ${CONFIG_PATH}"

    CONFIG_URL="https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/v${VERSION}/examples/default.yaml"

    # ToDo: remove this line after release
    CONFIG_URL="https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/5147e309c37ba543d76b4544fdedfbd4c77cc820/examples/default.yaml"

    # Generate template
    export FILE_STORAGE
    export COLLECTOR_NAME
    export INSTALL_TOKEN
    export API_BASE_URL

    curl -s "${CONFIG_URL}" | \
        if [[ -n "${API_BASE_URL}" ]]; then
            # add api_base_url after install_token
            sed "/^    install_token/a\\
    api_base_url: \${API_BASE_URL}
    "
        else
            cat -
        fi | \
        if [[ -n "${FIELDS}" ]]; then
            # add collector_fields after install_token
            sed "/^    install_token/a\\
    collector_fields:${FIELDS}
    "
        else
            cat -
        fi | \
        if [[ "${OS_TYPE}" == "darwin" ]]; then
            # adjust default configuration for macos
            sed '/^      process:/d'
        else
            cat -
        fi | \
        envsubst | sudo tee "${CONFIG_PATH}"

    echo 'Changing permissions for config file and storage'
    sudo chmod 640 "${CONFIG_PATH}" "${FILE_STORAGE}"
fi

if [[ "${SYSTEMD_DISABLED}" == "true" ]]; then
    echo "Use 'sudo otelcol-sumo --config=${CONFIG_PATH}' to run Sumo Logic Distribution for OpenTelemetry Collector"
    exit 0
fi

echo 'We are going to set up systemd service'
ask_to_continue

if [[ -f "${SYSTEMD_CONFIG}" ]]; then
    echo "Configuration for systemd service (${SYSTEMD_CONFIG}) already exist. Aborting"
    exit 0
fi

echo 'Creating user and group'
if getent passwd opentelemetry > /dev/null; then
    echo 'User and group already created'
else
    sudo useradd -mrUs /bin/false opentelemetry
fi

echo 'Changing ownership for config and storage'
sudo chown -R opentelemetry:opentelemetry "${CONFIG_PATH}" "${FILE_STORAGE}"

CONFIG_URL="https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/v${VERSION}/examples/systemd/otelcol-sumo.service"

# ToDo: remove this line after release
SYSTEMD_CONFIG="https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/1e677911ea866769b33fa259c68f4507369fc141/examples/systemd/otelcol-sumo.service"

echo 'Getting service configuration'
curl -fL "${SYSTEMD_CONFIG}" --output otelcol-sumo.service --progress-bar
sudo mv otelcol-sumo.service "${SYSTEMD_CONFIG}"

echo 'Enable otelcol-sumo service'
sudo systemctl enable otelcol-sumo

echo 'Starting otelcol-sumo service'
sudo systemctl start otelcol-sumo

echo 'Waiting 10s before checking status'
sleep 10
sudo systemctl status otelcol-sumo
