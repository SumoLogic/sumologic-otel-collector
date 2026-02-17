#!/usr/bin/env bash

set -euo pipefail

############################ DEPRECATION NOTICE ############################
#
# WARNING: This installation script has been deprecated and moved to:
# https://github.com/SumoLogic/sumologic-otel-collector-packaging
#
# Please use the installation scripts from the packaging repository instead.
# This script will be removed in a future release.
#
# For the latest installation instructions, please visit:
# - Linux: https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector-linux/
# - MacOS: https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector-macos/
#
############################################################################

############################ Static variables

ARG_SHORT_TOKEN='i'
ARG_LONG_TOKEN='installation-token'
DEPRECATED_ARG_LONG_TOKEN='installation-token'
ARG_SHORT_HELP='h'
ARG_LONG_HELP='help'
ARG_SHORT_API='a'
ARG_LONG_API='api'
ARG_SHORT_TAG='t'
ARG_LONG_TAG='tag'
ARG_SHORT_VERSION='v'
ARG_LONG_VERSION='version'
ARG_SHORT_FIPS='f'
ARG_LONG_FIPS='fips'
ARG_SHORT_YES='y'
ARG_LONG_YES='yes'
ARG_SHORT_SKIP_SYSTEMD='d'
ARG_LONG_SKIP_SYSTEMD='skip-systemd'
ARG_SHORT_SKIP_CONFIG='s'
ARG_LONG_SKIP_CONFIG='skip-config'
ARG_SHORT_UNINSTALL='u'
ARG_LONG_UNINSTALL='uninstall'
ARG_SHORT_PURGE='p'
ARG_LONG_PURGE='purge'
ARG_SHORT_SKIP_TOKEN='k'
ARG_LONG_SKIP_TOKEN='skip-installation-token'
DEPRECATED_ARG_LONG_SKIP_TOKEN='skip-install-token'
ARG_SHORT_DOWNLOAD='w'
ARG_LONG_DOWNLOAD='download-only'
ARG_SHORT_CONFIG_BRANCH='c'
ARG_LONG_CONFIG_BRANCH='config-branch'
ARG_SHORT_BINARY_BRANCH='e'
ARG_LONG_BINARY_BRANCH='binary-branch'
ENV_TOKEN="SUMOLOGIC_INSTALLATION_TOKEN"
DEPRECATED_ENV_TOKEN="SUMOLOGIC_INSTALL_TOKEN"
ARG_SHORT_BRANCH='b'
ARG_LONG_BRANCH='branch'
ARG_SHORT_KEEP_DOWNLOADS='n'
ARG_LONG_KEEP_DOWNLOADS='keep-downloads'
ARG_SHORT_INSTALL_HOSTMETRICS='H'
ARG_LONG_INSTALL_HOSTMETRICS='install-hostmetrics'
ARG_SHORT_REMOTELY_MANAGED='r'
ARG_LONG_REMOTELY_MANAGED='remotely-managed'
ARG_SHORT_EPHEMERAL='E'
ARG_LONG_EPHEMERAL='ephemeral'
ARG_SHORT_TIMEOUT='m'
ARG_LONG_TIMEOUT='download-timeout'

PACKAGE_GITHUB_ORG="SumoLogic"
PACKAGE_GITHUB_REPO="sumologic-otel-collector-packaging"

readonly ARG_SHORT_TOKEN ARG_LONG_TOKEN ARG_SHORT_HELP ARG_LONG_HELP ARG_SHORT_API ARG_LONG_API
readonly ARG_SHORT_TAG ARG_LONG_TAG ARG_SHORT_VERSION ARG_LONG_VERSION ARG_SHORT_YES ARG_LONG_YES
readonly ARG_SHORT_SKIP_SYSTEMD ARG_LONG_SKIP_SYSTEMD ARG_SHORT_UNINSTALL ARG_LONG_UNINSTALL
readonly ARG_SHORT_PURGE ARG_LONG_PURGE ARG_SHORT_DOWNLOAD ARG_LONG_DOWNLOAD
readonly ARG_SHORT_CONFIG_BRANCH ARG_LONG_CONFIG_BRANCH ARG_SHORT_BINARY_BRANCH ARG_LONG_CONFIG_BRANCH
readonly ARG_SHORT_BRANCH ARG_LONG_BRANCH ARG_SHORT_SKIP_CONFIG ARG_LONG_SKIP_CONFIG
readonly ARG_SHORT_SKIP_TOKEN ARG_LONG_SKIP_TOKEN ARG_SHORT_FIPS ARG_LONG_FIPS ENV_TOKEN
readonly ARG_SHORT_INSTALL_HOSTMETRICS ARG_LONG_INSTALL_HOSTMETRICS
readonly ARG_SHORT_REMOTELY_MANAGED ARG_LONG_REMOTELY_MANAGED
readonly ARG_SHORT_EPHEMERAL ARG_LONG_EPHEMERAL
readonly ARG_SHORT_TIMEOUT ARG_LONG_TIMEOUT
readonly DEPRECATED_ARG_LONG_TOKEN DEPRECATED_ENV_TOKEN DEPRECATED_ARG_LONG_SKIP_TOKEN
readonly PACKAGE_GITHUB_ORG PACKAGE_GITHUB_REPO

############################ Variables (see set_defaults function for default values)

# Support providing installation_token as env
set +u
if [[ -z "${SUMOLOGIC_INSTALLATION_TOKEN}" && -z "${SUMOLOGIC_INSTALL_TOKEN}" ]]; then
    SUMOLOGIC_INSTALLATION_TOKEN=""
elif [[ -z "${SUMOLOGIC_INSTALLATION_TOKEN}" ]]; then
    echo "${DEPRECATED_ENV_TOKEN} environmental variable is deprecated. Please use ${ENV_TOKEN} instead."
    SUMOLOGIC_INSTALLATION_TOKEN="${SUMOLOGIC_INSTALL_TOKEN}"
fi
set -u

API_BASE_URL=""
FIELDS=""
VERSION=""
FIPS=false
CONTINUE=false
HOME_DIRECTORY=""
CONFIG_DIRECTORY=""
USER_CONFIG_DIRECTORY=""
USER_ENV_DIRECTORY=""
SYSTEMD_CONFIG=""
UNINSTALL=""
SUMO_BINARY_PATH=""
SKIP_TOKEN=""
SKIP_CONFIG=false
CONFIG_PATH=""
COMMON_CONFIG_PATH=""
PURGE=""
DOWNLOAD_ONLY=""
INSTALL_HOSTMETRICS=false
REMOTELY_MANAGED=false
EPHEMERAL=false

LAUNCHD_CONFIG=""
LAUNCHD_ENV_KEY=""
LAUNCHD_TOKEN_KEY=""

USER_API_URL=""
USER_TOKEN=""
USER_FIELDS=""

ACL_LOG_FILE_PATHS="/var/log/ /srv/log/"

SYSTEM_USER="otelcol-sumo"

INDENTATION=""
EXT_INDENTATION=""

CONFIG_BRANCH=""
BINARY_BRANCH=""

KEEP_DOWNLOADS=false

CURL_MAX_TIME=1800

# set by check_dependencies therefore cannot be set by set_defaults
SYSTEMD_DISABLED=false

############################ Functions

function usage() {
  cat << EOF

Usage: bash install.sh [--${ARG_LONG_TOKEN} <token>] [--${ARG_LONG_TAG} <key>=<value> [ --${ARG_LONG_TAG} ...]] [--${ARG_LONG_API} <url>] [--${ARG_LONG_VERSION} <version>] \\
                       [--${ARG_LONG_YES}] [--${ARG_LONG_VERSION} <version>] [--${ARG_LONG_HELP}]

Supported arguments:
  -${ARG_SHORT_TOKEN}, --${ARG_LONG_TOKEN} <token>      Installation token. It has precedence over 'SUMOLOGIC_INSTALLATION_TOKEN' env variable.
  -${ARG_SHORT_SKIP_TOKEN}, --${ARG_LONG_SKIP_TOKEN}         Skips requirement for installation token.
                                        This option do not disable default configuration creation.
  -${ARG_SHORT_TAG}, --${ARG_LONG_TAG} <key=value>                 Sets tag for collector. This argument can be use multiple times. One per tag.
  -${ARG_SHORT_DOWNLOAD}, --${ARG_LONG_DOWNLOAD}                   Download new binary only and skip configuration part.

  -${ARG_SHORT_UNINSTALL}, --${ARG_LONG_UNINSTALL}                       Removes Sumo Logic Distribution for OpenTelemetry Collector from the system and
                                        disable Systemd service eventually.
                                        Use with '--purge' to remove all configurations as well.
  -${ARG_SHORT_PURGE}, --${ARG_LONG_PURGE}                           It has to be used with '--${ARG_LONG_UNINSTALL}'.
                                        It removes all Sumo Logic Distribution for OpenTelemetry Collector related configuration and data.

  -${ARG_SHORT_API}, --${ARG_LONG_API} <url>                       Api URL
  -${ARG_SHORT_SKIP_SYSTEMD}, --${ARG_LONG_SKIP_SYSTEMD}                    Do not install systemd unit.
  -${ARG_SHORT_SKIP_CONFIG}, --${ARG_LONG_SKIP_CONFIG}                     Do not create default configuration.
  -${ARG_SHORT_VERSION}, --${ARG_LONG_VERSION} <version>               Version of Sumo Logic Distribution for OpenTelemetry Collector to install, e.g. 0.57.2-sumo-1.
                                        By default it gets latest version.
  -${ARG_SHORT_FIPS}, --${ARG_LONG_FIPS}                            Install the FIPS 140-2 compliant binary on Linux.
  -${ARG_SHORT_INSTALL_HOSTMETRICS}, --${ARG_LONG_INSTALL_HOSTMETRICS}             Install the hostmetrics configuration to collect host metrics.
  -${ARG_SHORT_REMOTELY_MANAGED}, --${ARG_LONG_REMOTELY_MANAGED}                Remotely manage the collector configuration with Sumo Logic.
  -${ARG_SHORT_EPHEMERAL}, --${ARG_LONG_EPHEMERAL}                       Delete the collector from Sumo Logic after 12 hours of inactivity.
  -${ARG_SHORT_TIMEOUT}, --${ARG_LONG_TIMEOUT} <timeout>      Timeout in seconds after which download will fail. Default is ${CURL_MAX_TIME}.
  -${ARG_SHORT_YES}, --${ARG_LONG_YES}                             Disable confirmation asks.

  -${ARG_SHORT_HELP}, --${ARG_LONG_HELP}                            Prints this help and usage.

Supported env variables:
  ${ENV_TOKEN}=<token>       Installation token.'
EOF
}

function set_defaults() {
    HOME_DIRECTORY="/var/lib/otelcol-sumo"
    FILE_STORAGE="${HOME_DIRECTORY}/file_storage"
    DOWNLOAD_CACHE_DIR="/var/cache/otelcol-sumo"  # this is in case we want to keep downloaded binaries
    CONFIG_DIRECTORY="/etc/otelcol-sumo"
    SYSTEMD_CONFIG="/etc/systemd/system/otelcol-sumo.service"
    SUMO_BINARY_PATH="/usr/local/bin/otelcol-sumo"
    USER_CONFIG_DIRECTORY="${CONFIG_DIRECTORY}/conf.d"
    REMOTE_CONFIG_DIRECTORY="${CONFIG_DIRECTORY}/opamp.d"
    USER_ENV_DIRECTORY="${CONFIG_DIRECTORY}/env"
    TOKEN_ENV_FILE="${USER_ENV_DIRECTORY}/token.env"
    CONFIG_PATH="${CONFIG_DIRECTORY}/sumologic.yaml"
    CONFIG_BAK_PATH="${CONFIG_PATH}.bak"
    COMMON_CONFIG_PATH="${USER_CONFIG_DIRECTORY}/common.yaml"
    COMMON_CONFIG_BAK_PATH="${USER_CONFIG_DIRECTORY}/common.yaml.bak"
    INDENTATION="  "
    EXT_INDENTATION="${INDENTATION}${INDENTATION}"

    LAUNCHD_CONFIG="/Library/LaunchDaemons/com.sumologic.otelcol-sumo.plist"
    LAUNCHD_ENV_KEY="EnvironmentVariables"
    LAUNCHD_TOKEN_KEY="${LAUNCHD_ENV_KEY}.${ENV_TOKEN}"

    # ensure the cache dir exists
    mkdir -p "${DOWNLOAD_CACHE_DIR}"
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
      "--${DEPRECATED_ARG_LONG_TOKEN}")
        echo "--${DEPRECATED_ARG_LONG_TOKEN}" is deprecated. Please use "--${ARG_LONG_TOKEN}" instead.
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
      "--${ARG_LONG_SKIP_CONFIG}")
        set -- "$@" "-${ARG_SHORT_SKIP_CONFIG}"
        ;;
      "--${ARG_LONG_VERSION}")
        set -- "$@" "-${ARG_SHORT_VERSION}"
        ;;
      "--${ARG_LONG_FIPS}")
        set -- "$@" "-${ARG_SHORT_FIPS}"
        ;;
      "--${ARG_LONG_SKIP_SYSTEMD}")
        set -- "$@" "-${ARG_SHORT_SKIP_SYSTEMD}"
        ;;
      "--${ARG_LONG_UNINSTALL}")
        set -- "$@" "-${ARG_SHORT_UNINSTALL}"
        ;;
      "--${ARG_LONG_PURGE}")
        set -- "$@" "-${ARG_SHORT_PURGE}"
        ;;
      "--${ARG_LONG_SKIP_TOKEN}")
        set -- "$@" "-${ARG_SHORT_SKIP_TOKEN}"
        ;;
      "--${DEPRECATED_ARG_LONG_SKIP_TOKEN}")
        echo "--${DEPRECATED_ARG_LONG_SKIP_TOKEN}" is deprecated. Please use "--${ARG_SHORT_SKIP_TOKEN}" instead.
        set -- "$@" "-${ARG_SHORT_SKIP_TOKEN}"
        ;;
      "--${ARG_LONG_DOWNLOAD}")
        set -- "$@" "-${ARG_SHORT_DOWNLOAD}"
        ;;
      "--${ARG_LONG_BRANCH}")
        set -- "$@" "-${ARG_SHORT_BRANCH}"
        ;;
      "--${ARG_LONG_BINARY_BRANCH}")
        set -- "$@" "-${ARG_SHORT_BINARY_BRANCH}"
        ;;
      "--${ARG_LONG_CONFIG_BRANCH}")
        set -- "$@" "-${ARG_SHORT_CONFIG_BRANCH}"
        ;;
      "--${ARG_LONG_KEEP_DOWNLOADS}")
        set -- "$@" "-${ARG_SHORT_KEEP_DOWNLOADS}"
        ;;
      "--${ARG_LONG_TIMEOUT}")
        set -- "$@" "-${ARG_SHORT_TIMEOUT}"
        ;;
      "-${ARG_SHORT_TOKEN}"|"-${ARG_SHORT_HELP}"|"-${ARG_SHORT_API}"|"-${ARG_SHORT_TAG}"|"-${ARG_SHORT_SKIP_CONFIG}"|"-${ARG_SHORT_VERSION}"|"-${ARG_SHORT_FIPS}"|"-${ARG_SHORT_YES}"|"-${ARG_SHORT_SKIP_SYSTEMD}"|"-${ARG_SHORT_UNINSTALL}"|"-${ARG_SHORT_PURGE}"|"-${ARG_SHORT_SKIP_TOKEN}"|"-${ARG_SHORT_DOWNLOAD}"|"-${ARG_SHORT_CONFIG_BRANCH}"|"-${ARG_SHORT_BINARY_BRANCH}"|"-${ARG_SHORT_BRANCH}"|"-${ARG_SHORT_KEEP_DOWNLOADS}"|"-${ARG_SHORT_TIMEOUT}"|"-${ARG_SHORT_INSTALL_HOSTMETRICS}"|"-${ARG_SHORT_REMOTELY_MANAGED}"|"-${ARG_SHORT_EPHEMERAL}")
        set -- "$@" "${arg}"
        ;;
      "--${ARG_LONG_INSTALL_HOSTMETRICS}")
        set -- "$@" "-${ARG_SHORT_INSTALL_HOSTMETRICS}"
        ;;
      "--${ARG_LONG_REMOTELY_MANAGED}")
        set -- "$@" "-${ARG_SHORT_REMOTELY_MANAGED}"
        ;;
      "--${ARG_LONG_EPHEMERAL}")
        set -- "$@" "-${ARG_SHORT_EPHEMERAL}"
        ;;
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
    getopts "${ARG_SHORT_HELP}${ARG_SHORT_TOKEN}:${ARG_SHORT_API}:${ARG_SHORT_TAG}:${ARG_SHORT_VERSION}:${ARG_SHORT_FIPS}${ARG_SHORT_YES}${ARG_SHORT_SKIP_SYSTEMD}${ARG_SHORT_UNINSTALL}${ARG_SHORT_PURGE}${ARG_SHORT_SKIP_TOKEN}${ARG_SHORT_SKIP_CONFIG}${ARG_SHORT_DOWNLOAD}${ARG_SHORT_KEEP_DOWNLOADS}${ARG_SHORT_CONFIG_BRANCH}:${ARG_SHORT_BINARY_BRANCH}:${ARG_SHORT_BRANCH}:${ARG_SHORT_EPHEMERAL}${ARG_SHORT_REMOTELY_MANAGED}${ARG_SHORT_INSTALL_HOSTMETRICS}${ARG_SHORT_TIMEOUT}:" opt
    set -e

    # Invalid argument catched, print and exit
    if [[ $? != 0 && ${OPTIND} -le $# ]]; then
      echo "Invalid argument:" "${@:${OPTIND}:1}"
      usage
      exit 1
    fi

    # Validate opt and set arguments
    case "$opt" in
      "${ARG_SHORT_HELP}")          usage; exit 0 ;;
      "${ARG_SHORT_TOKEN}")         SUMOLOGIC_INSTALLATION_TOKEN="${OPTARG}" ;;
      "${ARG_SHORT_API}")           API_BASE_URL="${OPTARG}" ;;
      "${ARG_SHORT_SKIP_CONFIG}")   SKIP_CONFIG=true ;;
      "${ARG_SHORT_VERSION}")       VERSION="${OPTARG}" ;;
      "${ARG_SHORT_FIPS}")          FIPS=true ;;
      "${ARG_SHORT_YES}")           CONTINUE=true ;;
      "${ARG_SHORT_SKIP_SYSTEMD}")       SYSTEMD_DISABLED=true ;;
      "${ARG_SHORT_UNINSTALL}")     UNINSTALL=true ;;
      "${ARG_SHORT_PURGE}")         PURGE=true ;;
      "${ARG_SHORT_SKIP_TOKEN}")    SKIP_TOKEN=true ;;
      "${ARG_SHORT_DOWNLOAD}")      DOWNLOAD_ONLY=true ;;
      "${ARG_SHORT_CONFIG_BRANCH}") CONFIG_BRANCH="${OPTARG}" ;;
      "${ARG_SHORT_BINARY_BRANCH}") BINARY_BRANCH="${OPTARG}" ;;
      "${ARG_SHORT_BRANCH}")
        if [[ -z "${BINARY_BRANCH}" ]]; then
            BINARY_BRANCH="${OPTARG}"
        fi
        if [[ -z "${CONFIG_BRANCH}" ]]; then
            CONFIG_BRANCH="${OPTARG}"
        fi ;;
      "${ARG_SHORT_INSTALL_HOSTMETRICS}") INSTALL_HOSTMETRICS=true ;;
      "${ARG_SHORT_REMOTELY_MANAGED}") REMOTELY_MANAGED=true ;;
      "${ARG_SHORT_EPHEMERAL}") EPHEMERAL=true ;;
      "${ARG_SHORT_KEEP_DOWNLOADS}") KEEP_DOWNLOADS=true ;;
      "${ARG_SHORT_TIMEOUT}") CURL_MAX_TIME="${OPTARG}" ;;
      "${ARG_SHORT_TAG}")
        if [[ "${OPTARG}" != ?*"="* ]]; then
            echo "Invalid tag: '${OPTARG}'. Should be in 'key=value' format"
            usage
            exit 1
        fi

        value="$(echo -e "${OPTARG}" | sed 's/.*=//')"
        key="$(echo -e "${OPTARG}" | sed 's/\(.*\)=.*/\1/')"
        line="${key}: $(escape_yaml_value "${value}")"

        # Cannot use `\n` and have to use `\\` as break line due to OSx sed implementation
        FIELDS="${FIELDS}\\
$(escape_sed "${line}")" ;;
    "?")                            ;;
      *)                            usage; exit 1 ;;
    esac

    # Exit loop as we iterated over all arguments
    if [[ "${OPTIND}" -gt $# ]]; then
      break
    fi
  done
}

# Get github rate limit
function github_rate_limit() {
    curl --retry 5 --connect-timeout 5 --max-time 30 --retry-delay 0 --retry-max-time 150 -X GET https://api.github.com/rate_limit -v 2>&1 | grep x-ratelimit-remaining | grep -oE "[0-9]+"
}

# This function is applicable to very few platforms/distributions.
function install_missing_dependencies() {
    local REQUIRED_COMMANDS
    local REQUIRED_PACKAGES
    REQUIRED_COMMANDS=()
    REQUIRED_PACKAGES=()
    if [[ -n "${BINARY_BRANCH}" ]]; then  # unzip is only necessary for downloading from GHA artifacts
        REQUIRED_COMMANDS+=(unzip)
        REQUIRED_PACKAGES+=(unzip)
    fi
    if [[ -f "/etc/redhat-release" ]]; then # this will install semanage, which is necessary for SELinux relabeling
        REQUIRED_COMMANDS+=(semanage)
        REQUIRED_PACKAGES+=(policycoreutils-python-utils)
    fi
    if [ "${#REQUIRED_COMMANDS[@]}" == 0 ]; then
        # not all bash versions handle empty array expansion correctly
        # therefore we guard against this explicitly here
        return
    fi
    for i in "${!REQUIRED_COMMANDS[@]}"; do
        cmd=${REQUIRED_COMMANDS[i]}
        pkg=${REQUIRED_PACKAGES[i]}
        if ! command -v "${cmd}" &> /dev/null; then
            # Attempt to install it via yum if on a RHEL distribution.
            if [[ -f "/etc/redhat-release" ]]; then
                echo "Command '${cmd}' not found. Attempting to install '${pkg}'..."
                # This only works if the tool/command matches the system package name.
                yum install -y "${pkg}"
            fi
        fi
    done
}

# Ensure TMPDIR is set to a directory where we can safely store temporary files
function set_tmpdir() {
    # generate a new tmpdir using mktemp
    # need to specify the template for some MacOS versions
    TMPDIR=$(mktemp -d -t 'sumologic-otel-collector-XXXX')
}

function check_dependencies() {
    local error
    error=0

    if [ "$EUID" -ne 0 ]; then
        echo "Please run this script as root."
        error=1
    fi

    REQUIRED_COMMANDS=(echo sed curl head grep sort mv chmod getopts hostname touch xargs)
    if [[ -n "${BINARY_BRANCH}" ]]; then  # unzip is only necessary for downloading from GHA artifacts
        REQUIRED_COMMANDS+=(unzip)
    fi

    for cmd in "${REQUIRED_COMMANDS[@]}"; do
        if ! command -v "${cmd}" &> /dev/null; then
            echo "Command '${cmd}' not found. Please install it."
            error=1
        fi
    done

    if [[ ! -d /run/systemd/system ]]; then
        SYSTEMD_DISABLED=true
    fi

    if [[ "${error}" == "1" ]] ; then
        exit 1
    fi
}

function get_latest_package_version() {
    local versions
    readonly versions="${1}"

    # get latest version directly from website if there is no versions from api
    if [[ -z "${versions}" ]]; then
        curl --retry 5 --connect-timeout 5 --max-time 30 --retry-delay 0 \
            --retry-max-time 150 -s \
            "https://github.com/${PACKAGE_GITHUB_ORG}/${PACKAGE_GITHUB_REPO}/releases" \
            | grep -oE '/'${PACKAGE_GITHUB_ORG}'/'${PACKAGE_GITHUB_REPO}'/releases/tag/(.*)"' \
            | head -n 1 \
            | sed 's%/'${PACKAGE_GITHUB_ORG}'/'${PACKAGE_GITHUB_REPO}'/releases/tag/v\([^"]*\)".*%\1%g'
    else
        # sed 's/ /\n/g' converts spaces to new lines
        echo "${versions}" | sed 's/ /\n/g' | head -n 1
    fi
}

function get_latest_version() {
    local versions
    readonly versions="${1}"

    # get latest version directly from website if there is no versions from api
    if [[ -z "${versions}" ]]; then
        curl --retry 5 --connect-timeout 5 --max-time 30 --retry-delay 5 --retry-max-time 150 -s https://github.com/SumoLogic/sumologic-otel-collector/releases \
        | grep -Eo '/SumoLogic/sumologic-otel-collector/releases/tag/v[0-9]+\.[0-9]+\.[0-9]+-sumo-[0-9]+[^-]' \
        | head -n 1 | sed 's%/SumoLogic/sumologic-otel-collector/releases/tag/v\([^"]*\)".*%\1%g'
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
    # returns empty in case we exceeded github rate limit. This can happen if we are running this script too many times in a short period.
    if [[ "$(github_rate_limit)" == "0" ]]; then
        return
    fi

    curl \
    --connect-timeout 5 \
    --max-time 30 \
    --retry 5 \
    --retry-delay 0 \
    --retry-max-time 150 \
    -sH "Accept: application/vnd.github.v3+json" \
    https://api.github.com/repos/SumoLogic/sumologic-otel-collector/releases \
    | grep -E '(tag_name|"(draft|prerelease)")' \
    | sed 'N;N;s/.*true.*//' \
    | grep -o 'v.*"' \
    | sort -rV \
    | sed 's/^v//;s/"$//'
}

function get_package_versions() {
    # returns empty in case we exceeded github rate limit
    if [[ "$(github_rate_limit)" == "0" ]]; then
        return
    fi

    curl \
    --connect-timeout 5 \
    --max-time 30 \
    --retry 5 \
    --retry-delay 0 \
    --retry-max-time 150 \
    -sH "Accept: application/vnd.github.v3+json" \
    https://api.github.com/repos/${PACKAGE_GITHUB_ORG}/${PACKAGE_GITHUB_REPO}/releases \
    | grep -E '(tag_name|"(draft|prerelease)")' \
    | sed 'N;N;s/.*true.*//' \
    | grep -o 'v.*"' \
    | sort -rV \
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

# Verify that the otelcol install is correct
function verify_installation() {
    local otel_command
    if command -v otelcol-sumo; then
        otel_command="otelcol-sumo"
    else
        echo "WARNING: ${SUMO_BINARY_PATH} is not in \$PATH"
        otel_command="${SUMO_BINARY_PATH}"
    fi
    echo "Running ${otel_command} --version to verify installation"
    OUTPUT="$(${otel_command} --version || true)"
    readonly OUTPUT

    if [[ -z "${OUTPUT}" ]]; then
        echo "Installation failed. Please try again"
        exit 1
    fi

    echo -e "Installation succeded:\t$(${otel_command} --version)"
}

# Get installed version of otelcol-sumo
function get_installed_version() {
    if [[ -f "${SUMO_BINARY_PATH}" ]]; then
        set +o pipefail
        "${SUMO_BINARY_PATH}" --version | grep -o 'v[0-9].*$' | sed 's/v//'
        set -o pipefail
    fi
}

# Ask to continue and abort if not
function ask_to_continue() {
    if [[ "${CONTINUE}" == true ]]; then
        return 0
    fi

    # Just fail if we're not running in uninteractive mode
    # TODO: Figure out a way to reliably ask for confirmation with stdin redirected

    echo "Please use the --yes flag to continue"
    exit 1

    # local choice
    # read -rp "Continue (y/N)? " choice
    # case "${choice}" in
    # y|Y ) ;;
    # n|N | * )
    #     echo "Aborting..."
    #     exit 1
    #     ;;
    # esac

}

# Print information about breaking changes
function print_breaking_changes() {
    local versions
    readonly versions="${1}"

    local changelog
    changelog="$(echo -e "$(curl --retry 5 --connect-timeout 5 --max-time 30 --retry-delay 0 --retry-max-time 150 -sS https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/main/CHANGELOG.md)")"
    declare -r changelog

    local is_breaking_change
    local message
    message=""

    for version in ${versions}; do
        # Print changelog for every version
        is_breaking_change=$(echo -e "${changelog}" | grep -E '^## |^### Breaking|breaking changes' | sed -e '/## \[v'"${version}"'/,/## \[v/!d' | grep -E 'Breaking|breaking' || echo "")

        if [[ -n "${is_breaking_change}" ]]; then
            if [[ -n "${message}" ]]; then
                message="${message}, "
            fi
            message="${message}v${version}"
        fi
    done

    if [[ -n "${message}" ]]; then
        echo "The following versions contain breaking changes: ${message}! Please make sure to read the linked Changelog file."
    fi
}

# set up configuration
function setup_config() {
    echo 'We are going to get and set up a default configuration for you'

    echo -e "Creating file_storage directory (${FILE_STORAGE})"
    mkdir -p "${FILE_STORAGE}"

    echo -e "Creating configuration directory (${CONFIG_DIRECTORY})"
    mkdir -p "${CONFIG_DIRECTORY}"

    echo -e "Creating user configurations directory (${USER_CONFIG_DIRECTORY})"
    mkdir -p "${USER_CONFIG_DIRECTORY}"

    echo -e "Creating user env directory (${USER_ENV_DIRECTORY})"
    mkdir -p "${USER_ENV_DIRECTORY}"

    echo 'Changing permissions for config files and storage'
    chmod 551 "${CONFIG_DIRECTORY}"  # config directory world traversable, as is the /etc/ standard

    echo 'Changing permissions for user env directory'
    chmod 550 "${USER_ENV_DIRECTORY}"
    chmod g+s "${USER_ENV_DIRECTORY}"

    echo "Generating configuration and saving as ${CONFIG_PATH}"

    CONFIG_URL="https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/${CONFIG_BRANCH}/examples/sumologic.yaml"
    if ! curl --retry 5 --connect-timeout 5 --max-time 30 --retry-delay 0 --retry-max-time 150 -f -s "${CONFIG_URL}" -o "${CONFIG_PATH}"; then
        echo "Cannot obtain configuration for '${CONFIG_BRANCH}' branch. Either '${CONFIG_URL}' is invalid, or the network connection is unstable."
        exit 1
    fi

    if [[ "${REMOTELY_MANAGED}" == "true" ]]; then
        echo "Warning: remote management is currently in beta."

        echo -e "Creating remote configurations directory (${REMOTE_CONFIG_DIRECTORY})"
        mkdir -p "${REMOTE_CONFIG_DIRECTORY}"

        write_sumologic_extension "${CONFIG_PATH}" "${INDENTATION}"
        write_opamp_extension "${CONFIG_PATH}" "${REMOTE_CONFIG_DIRECTORY}" "${INDENTATION}" "${EXT_INDENTATION}" "${API_BASE_URL}"

        if [[ -n "${SUMOLOGIC_INSTALLATION_TOKEN}" && "${SYSTEMD_DISABLED}" == "true" ]]; then
            write_installation_token "${SUMOLOGIC_INSTALLATION_TOKEN}" "${CONFIG_PATH}" "${EXT_INDENTATION}"
        fi

        if [[ "${EPHEMERAL}" == "true" ]]; then
            write_ephemeral_true "${CONFIG_PATH}" "${EXT_INDENTATION}"
        fi

        if [[ -n "${API_BASE_URL}" ]]; then
            write_api_url "${API_BASE_URL}" "${CONFIG_PATH}" "${EXT_INDENTATION}"
        fi

        if [[ -n "${FIELDS}" ]]; then
            write_tags "${FIELDS}" "${CONFIG_PATH}" "${INDENTATION}" "${EXT_INDENTATION}"
        fi

        rm -f "${CONFIG_BAK_PATH}"

        # Finish setting permissions after we're done creating config files
        chmod -R 440 "${CONFIG_DIRECTORY}"/*  # all files only readable by the owner
        find "${CONFIG_DIRECTORY}/" -mindepth 1 -type d -exec chmod 550 {} \;  # directories also traversable

        # Remote configuration directory must be writable
        chmod 750 "${REMOTE_CONFIG_DIRECTORY}"

        # Return/stop function execution
        return
    fi

    if [[ "${INSTALL_HOSTMETRICS}" == "true" ]]; then
        echo -e "Installing ${OS_TYPE} hostmetrics configuration"
        HOSTMETRICS_CONFIG_URL="https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/${CONFIG_BRANCH}/examples/conf.d/${OS_TYPE}.yaml"
        if ! curl --retry 5 --connect-timeout 5 --max-time 30 --retry-delay 0 --retry-max-time 150 -f -s "${HOSTMETRICS_CONFIG_URL}" -o "${CONFIG_DIRECTORY}/conf.d/hostmetrics.yaml"; then
            echo "Cannot obtain hostmetrics configuration for '${CONFIG_BRANCH}' branch. Either '${HOSTMETRICS_CONFIG_URL}' is invalid, or the network connection is unstable."
            exit 1
        fi
        if [[ "${OS_TYPE}" == "linux" ]]; then
            echo -e "Setting the CAP_DAC_READ_SEARCH Linux capability on the collector binary to allow it to read host metrics from /proc directory: setcap 'cap_dac_read_search=ep' \"${SUMO_BINARY_PATH}\""
            echo -e "You can remove it with the following command: sudo setcap -r \"${SUMO_BINARY_PATH}\""
            echo -e "Without this capability, the collector will not be able to collect some of the host metrics."
            setcap 'cap_dac_read_search=ep' "${SUMO_BINARY_PATH}"
        fi
    fi

    # Ensure that configuration is created
    if [[ -f "${COMMON_CONFIG_PATH}" ]]; then
        echo "User configuration (${COMMON_CONFIG_PATH}) already exist)"
    fi

    ## Check if there is anything to update in configuration
    if [[ ( -n "${SUMOLOGIC_INSTALLATION_TOKEN}" && "${SYSTEMD_DISABLED}" == "true" ) || -n "${API_BASE_URL}" || -n "${FIELDS}" || "${EPHEMERAL}" == "true" ]]; then
        create_user_config_file "${COMMON_CONFIG_PATH}"
        add_extension_to_config "${COMMON_CONFIG_PATH}"
        write_sumologic_extension "${COMMON_CONFIG_PATH}" "${INDENTATION}"

        if [[ -n "${SUMOLOGIC_INSTALLATION_TOKEN}" && -z "${USER_TOKEN}" && "${SYSTEMD_DISABLED}" == "true" ]]; then
            write_installation_token "${SUMOLOGIC_INSTALLATION_TOKEN}" "${COMMON_CONFIG_PATH}" "${EXT_INDENTATION}"
        fi

        if [[ "${EPHEMERAL}" == "true" ]]; then
            write_ephemeral_true "${COMMON_CONFIG_PATH}" "${EXT_INDENTATION}"
        fi

        # fill in api base url
        if [[ -n "${API_BASE_URL}" && -z "${USER_API_URL}" ]]; then
            write_api_url "${API_BASE_URL}" "${COMMON_CONFIG_PATH}" "${EXT_INDENTATION}"
        fi

        if [[ -n "${FIELDS}" && -z "${USER_FIELDS}" ]]; then
            write_tags "${FIELDS}" "${COMMON_CONFIG_PATH}" "${INDENTATION}" "${EXT_INDENTATION}"
        fi

        # clean up bak file
        rm -f "${COMMON_CONFIG_BAK_PATH}"
    fi
    # Finish setting permissions after we're done creating config files
    chmod -R 440 "${CONFIG_DIRECTORY}"/*  # all files only readable by the owner
    find "${CONFIG_DIRECTORY}/" -mindepth 1 -type d -exec chmod 550 {} \;  # directories also traversable
}

function setup_config_darwin() {
    local config_path
    config_path="${COMMON_CONFIG_PATH}"

    if [[ "${REMOTELY_MANAGED}" == "true" ]]; then
        config_path="${CONFIG_PATH}"
    fi

    readonly config_path

    create_user_config_file "${config_path}"
    add_extension_to_config "${config_path}"
    write_sumologic_extension "${config_path}" "${INDENTATION}"

    if [[ "${EPHEMERAL}" == "true" ]]; then
        write_ephemeral_true "${config_path}" "${EXT_INDENTATION}"
    fi

    # fill in api base url
    if [[ -n "${API_BASE_URL}" ]]; then
        write_api_url "${API_BASE_URL}" "${config_path}" "${EXT_INDENTATION}"
    fi

    if [[ -n "${FIELDS}" ]]; then
        write_tags "${FIELDS}" "${config_path}" "${INDENTATION}" "${EXT_INDENTATION}"
    fi

    if [[ "${REMOTELY_MANAGED}" == "true" ]]; then
        echo "Warning: remote management is currently in beta."

        echo -e "Creating remote configurations directory (${REMOTE_CONFIG_DIRECTORY})"
        mkdir -p "${REMOTE_CONFIG_DIRECTORY}"

        write_opamp_extension "${config_path}" "${REMOTE_CONFIG_DIRECTORY}" "${INDENTATION}" "${EXT_INDENTATION}" "${API_BASE_URL}"

        write_remote_config_launchd "${LAUNCHD_CONFIG}"

        # Remote configuration directory must be writable
        chmod 750 "${REMOTE_CONFIG_DIRECTORY}"

        # Remote configuration directory must be owned by the mac pkg service user
        chown _otelcol-sumo:_otelcol-sumo "${REMOTE_CONFIG_DIRECTORY}"
    fi

    # clean up bak files
    rm -f "${CONFIG_BAK_PATH}"
    rm -f "${COMMON_CONFIG_BAK_PATH}"
}

# uninstall otelcol-sumo
function uninstall() {
    case "${OS_TYPE}" in
    "darwin") uninstall_darwin ;;
    "linux") uninstall_linux ;;
    *)
      echo "Uninstallation is not supported by this script for OS: ${OS_TYPE}"
      exit 1
      ;;
    esac

    echo "Uninstallation completed"
}

# uninstall otelcol-sumo on darwin
function uninstall_darwin() {
    local UNINSTALL_SCRIPT_PATH
    UNINSTALL_SCRIPT_PATH="/Library/Application Support/otelcol-sumo/uninstall.sh"

    echo "Going to uninstall otelcol-sumo."

    if [[ "${PURGE}" == "true" ]]; then
        echo "WARNING: purge is not yet supported on darwin"
    fi

    ask_to_continue

    "${UNINSTALL_SCRIPT_PATH}"
}

# uninstall otelcol-sumo on linux
function uninstall_linux() {
    local MSG
    MSG="Going to remove Otelcol binary"

    if [[ "${PURGE}" == "true" ]]; then
        MSG="${MSG}, user, file storage and configurations"
    fi

    echo "${MSG}."
    ask_to_continue

    # disable systemd service
    if [[ -f "${SYSTEMD_CONFIG}" ]]; then
        systemctl stop otelcol-sumo || true
        systemctl disable otelcol-sumo || true
    fi

    # remove binary
    rm -f "${SUMO_BINARY_PATH}"

    if [[ "${PURGE}" == "true" ]]; then
        # remove configuration and data
        rm -rf "${CONFIG_DIRECTORY}" "${FILE_STORAGE}" "${SYSTEMD_CONFIG}"

        # remove user and group only if getent exists (it was required in order to create the user)
        if command -v "getent" &> /dev/null; then
            # remove user
            if getent passwd "${SYSTEM_USER}" > /dev/null; then
                userdel -r -f "${SYSTEM_USER}"
                groupdel "${SYSTEM_USER}" 2>/dev/null || true
            fi
        fi
    fi

    echo "Uninstallation completed"
}

function escape_sed() {
    local text
    readonly text="${1}"

    # replaces `\` with `\\` and `/` with `\/`
    echo "${text}" \
        | sed -e 's/\\/\\\\/g' \
        | sed -e 's|/|\\/|g'
}

function get_indentation() {
    local file
    readonly file="${1}"

    local default
    readonly default="${2}"

    if [[ ! -f "${file}" ]]; then
        echo "${default}"
        return
    fi

    local indentation

    # take indentation same as first extension
    indentation="$(sed -e '/^extensions/,/^[a-z]/!d' "${file}" \
        | grep -m 1 -E '^\s+[a-z]' \
        | grep -m 1 -oE '^\s+' \
    || echo "")"
    if [[ -n "${indentation}" ]]; then
        echo "${indentation}"
        return
    fi

    # otherwise take indentation from any other package
    indentation="$(grep -m 1 -E '^\s+[a-z]' "${file}" \
        | grep -m 1 -oE '^\s+' \
    || echo "")"
    if [[ -n "${indentation}" ]]; then
        echo "${indentation}"
        return
    fi

    # return default indentation
    echo "${default}"
}

function get_extension_indentation() {
    local file
    readonly file="${1}"

    local indentation="${2}"
    readonly indentation

    if [[ ! -f "${file}" ]]; then
        echo "${indentation}${indentation}"
        return
    fi

    local ext_indentation

    # take indentation same as properties of sumologic extension
    ext_indentation="$(sed -e "/^${indentation}sumologic:/,/^${indentation}[a-z]/!d" "${file}" \
        | grep -m 1 -E "^${indentation}\s+[a-z]" \
        | grep -m 1 -oE '^\s+' \
    || echo "")"

    if [[ -n "${ext_indentation}" ]]; then
        echo "${ext_indentation}"
        return
    fi

    # otherwise take indentation from properties of any other package
    ext_indentation="$(grep -m 1 -E "^${indentation}\s+[a-z]" "${file}" \
        | grep -m 1 -oE '^\s+' \
    || echo "")"

    if [[ -n "${ext_indentation}" ]]; then
        echo "${ext_indentation}"
        return
    fi

    # otherwise use double indentation
    echo "${indentation}${indentation}"
}

function get_user_config() {
    local file
    readonly file="${1}"

    if [[ ! -f "${file}" ]]; then
        return
    fi

    # extract installation_token and strip quotes
    # fallback to deprecated install_token
    grep -m 1 installation_token "${file}" \
        | sed 's/.*installation_token:[[:blank:]]*//' \
        | sed 's/[[:blank:]]*$//' \
        | sed 's/^"//' \
        | sed "s/^'//" \
        | sed 's/"$//' \
        | sed "s/'\$//" \
    || grep -m 1 install_token "${file}" \
        | sed 's/.*install_token:[[:blank:]]*//' \
        | sed 's/[[:blank:]]*$//' \
        | sed 's/^"//' \
        | sed "s/^'//" \
        | sed 's/"$//' \
        | sed "s/'\$//" \
    || echo ""
}

# remove quotes and double quotes from yaml `value`` for `key: value` form
function unescape_yaml() {
    local fields
    readonly fields="${1}"

    # Process the string line by line
    echo -e "${fields}" | while IFS= read -r line; do
        # strip `\` from the end of the line
        line="$(echo "${line}" | sed 's/\\$//')"
        # extract key
        key="$(echo -e "${line}" | sed 's/\(.*\):.*/\1/')"

        # extract value
        value="$(echo -e "${line}" | sed 's/.*:[[:blank:]]*//')"
        # remove quote, double quote and escapes
        value="$(unescape_yaml_value "${value}")"
        if [[ -n "${key}" && -n "${value}" ]]; then
            echo "${key}: ${value}"
        fi
    done
}

# escape yaml value by replacing `'` with `''` and adding surrounding `'`
function escape_yaml_value() {
    local value
    readonly value="${1}"

    echo "'$(echo -e "${value}" | sed "s/'/''/")'"
}

function unescape_yaml_value() {
    local value
    readonly value="${1}"

    if echo -e "${value}" | grep -oqE "^'"; then
        # remove `'` from beginning and end of the string
        # replace `''` with `'`
        echo -e "${value}" \
        | sed "s/'[[:blank:]]*$//" \
        | sed "s/^[[:blank:]]*'//" \
        | sed "s/''/'/"
    elif echo -e "${value}" | grep -oqE '^"'; then
        # remove `"` from beginning and end of the string
        # remove `'` from beginning and end of the string
        # replace `\"` with `"`
        echo -e "${value}" \
        | sed 's/"[[:blank:]]*$//' \
        | sed 's/^[[:blank:]]*"//' \
        | sed 's/\"/"/'
    else
        echo -e "${value}"
    fi
}

function get_user_env_config() {
    local file
    readonly file="${1}"

    if [[ ! -f "${file}" ]]; then
        return
    fi

    # extract install_token and strip quotes
    grep -m 1 "${ENV_TOKEN}" "${file}" \
        | sed "s/.*${ENV_TOKEN}=[[:blank:]]*//" \
        | sed 's/[[:blank:]]*$//' \
        | sed 's/^"//' \
        | sed "s/^'//" \
        | sed 's/"$//' \
        | sed "s/'\$//" \
    || grep -m 1 "${DEPRECATED_ENV_TOKEN}" "${file}" \
        | sed "s/.*${DEPRECATED_ENV_TOKEN}=[[:blank:]]*//" \
        | sed 's/[[:blank:]]*$//' \
        | sed 's/^"//' \
        | sed "s/^'//" \
        | sed 's/"$//' \
        | sed "s/'\$//" \
    || echo ""
}

function get_user_api_url() {
    local file
    readonly file="${1}"

    if [[ ! -f "${file}" ]]; then
        return
    fi

    # extract api_base_url and strip quotes
    grep -m 1 api_base_url "${file}" \
        | sed 's/.*api_base_url:[[:blank:]]*//' \
        | sed 's/[[:blank:]]*$//' \
        | sed 's/^"//' \
        | sed "s/^'//" \
        | sed 's/"$//' \
        | sed "s/'\$//" \
    || echo ""
}

function get_user_tags() {
    local file
    readonly file="${1}"

    local indentation
    readonly indentation="${2}"

    local ext_indentation
    readonly ext_indentation="${3}"

    if [[ ! -f "${file}" ]]; then
        return
    fi

    local fields
    fields="$(sed -e '/^extensions/,/^[a-z]/!d' "${file}" \
        | sed -e "/^${indentation}sumologic/,/^${indentation}[a-z]/!d" \
        | sed -e "/^${ext_indentation}collector_fields/,/^${ext_indentation}[a-z]/!d;" \
        | grep -vE "^${ext_indentation}\\S" \
        | sed -e 's/^[[:blank:]]*//' \
        || echo "")"
    unescape_yaml "${fields}" \
        | sort \
        || echo ""
}

function get_fields_to_compare() {
    local fields
    # replace \/ with /
    fields="$(echo "${FIELDS}" | sed -e 's|\\/|/|')"
    declare -r fields

    unescape_yaml "${fields}" \
        | grep -vE '^$' \
        | sort \
    || echo ""
}

function create_user_config_file() {
    local file
    readonly file="${1}"

    if [[ -f "${file}" ]]; then
        return
    fi

    touch "${file}"
    chmod 440 "${file}"
}

# write extensions section to user configuration file
function add_extension_to_config() {
    local file
    readonly file="${1}"

    if grep -q 'extensions:$' "${file}"; then
        return
    fi

    echo "extensions:" \
        | tee -a "${file}" > /dev/null 2>&1
}

# write sumologic extension to user configuration file
function write_sumologic_extension() {
    local file
    readonly file="${1}"

    local indentation
    readonly indentation="${2}"

    if sed -e '/^extensions/,/^[a-z]/!d' "${file}" | grep -qE '^\s+(sumologic|sumologic\/.*):\s*$'; then
        return
    fi

    # add sumologic extension on the top of the extensions
    sed -i.bak -e "s/extensions:/extensions:\\
${indentation}sumologic:/" "${file}"
}

# write installation token to user configuration file
function write_installation_token() {
    local token
    readonly token="${1}"

    local file
    readonly file="${2}"

    local ext_indentation
    readonly ext_indentation="${3}"

    # ToDo: ensure we override only sumologic `installation_token`
    if grep "installation_token" "${file}" > /dev/null; then
        # Do not expose token in sed command as it can be saw on processes list
        echo "s/installation_token:.*$/installation_token: $(escape_sed "${token}")/" | sed -i.bak -f - "${file}"

        return
    fi

    # ToDo: ensure we override only sumologic `install_token`
    if grep "install_token" "${file}" > /dev/null; then
        # Do not expose token in sed command as it can be saw on processes list
        echo "s/install_token:.*$/installation_token: $(escape_sed "${token}")/" | sed -i.bak -f - "${file}"
    else
        # write installation token on the top of sumologic: extension
        # Do not expose token in sed command as it can be saw on processes list
        echo "1,/sumologic:/ s/sumologic:/sumologic:\\
\\${ext_indentation}installation_token: $(escape_sed "${token}")/" | sed -i.bak -f - "${file}"
    fi
}

# write ${ENV_TOKEN}" to systemd env configuration file
function write_installation_token_env() {
    local token
    readonly token="${1}"

    local file
    readonly file="${2}"

    local token_name
    if (( MAJOR_VERSION == 0 && MINOR_VERSION <= 71 )); then
        token_name="${DEPRECATED_ENV_TOKEN}"
    else
        token_name="${ENV_TOKEN}"
    fi
    readonly token_name

    # ToDo: ensure we override only ${ENV_TOKEN}" env value
    if grep "${token_name}" "${file}" > /dev/null 2>&1; then
        # Do not expose token in sed command as it can be saw on processes list
        echo "s/${token_name}=.*$/${token_name}=$(escape_sed "${token}")/" | sed -i.bak -f - "${file}"
    else
        echo "${token_name}=${token}" > "${file}"
    fi
}

# write ${ENV_TOKEN} to launchd configuration file
function write_installation_token_launchd() {
    local token
    readonly token="${1}"

    local file
    readonly file="${2}"

    if [[ ! -f "${file}" ]]; then
        echo "The LaunchDaemon configuration file is missing: ${file}"
        exit 1
    fi

    # Create EnvironmentVariables key if it does not exist
    if ! plutil_key_exists "${file}" "${LAUNCHD_ENV_KEY}"; then
        plutil_create_key "${file}" "${LAUNCHD_ENV_KEY}" "xml" "<dict/>"
    fi

    # Replace EnvironmentVariables key if it has an incorrect type
    if ! plutil_key_is_type "${file}" "${LAUNCHD_ENV_KEY}" "dictionary"; then
        plutil_replace_key "${file}" "${LAUNCHD_ENV_KEY}" "xml" "<dict/>"
    fi

    # Create SUMOLOGIC_INSTALLATION_TOKEN key if it does not exist
    if ! plutil_key_exists "${file}" "${LAUNCHD_TOKEN_KEY}"; then
        plutil_create_key "${file}" "${LAUNCHD_TOKEN_KEY}" "string" "${SUMOLOGIC_INSTALLATION_TOKEN}"
    fi

    # Replace SUMOLOGIC_INSTALLATION_TOKEN key if it has an incorrect type
    if ! plutil_key_is_type "${LAUNCHD_CONFIG}" "${LAUNCHD_TOKEN_KEY}" "string"; then
        plutil_replace_key "${LAUNCHD_CONFIG}" "${LAUNCHD_TOKEN_KEY}" "string" "${SUMOLOGIC_INSTALLATION_TOKEN}"
    fi
}

function write_remote_config_launchd() {
    local file
    readonly file="${1}"

    if [[ ! -f "${file}" ]]; then
        echo "The LaunchDaemon configuration file is missing: ${file}"
        exit 1
    fi

    # Delete existing ProgramArguments
    plutil_delete_key "${file}" "ProgramArguments"

    # Create new ProgramArguments with --remote-config
    plutil_create_key "${file}" "ProgramArguments" "json" "[ \"/usr/local/bin/otelcol-sumo\", \"--remote-config\", \"opamp:${CONFIG_PATH}\" ]"
}

# write sumologic ephemeral: true to user configuration file
function write_ephemeral_true() {
    local file
    readonly file="${1}"

    local ext_indentation
    readonly ext_indentation="${2}"

    if grep "ephemeral:" "${file}" > /dev/null; then
        sed -i.bak -e "1,/ephemeral:/ s/ephemeral:.*$/ephemeral: true/" "${file}"
    else
        # write ephemeral: true on the top of sumologic: extension
        sed -i.bak -e "1,/sumologic:/ s/sumologic:/sumologic:\\
\\${ext_indentation}ephemeral: true/" "${file}"
    fi
}

# write api_url to user configuration file
function write_api_url() {
    local api_url
    readonly api_url="${1}"

    local file
    readonly file="${2}"

    local ext_indentation
    readonly ext_indentation="${3}"

    # ToDo: ensure we override only sumologic `api_base_url`
    if grep "api_base_url" "${file}" > /dev/null; then
        sed -i.bak -e "s/api_base_url:.*$/api_base_url: $(escape_sed "${api_url}")/" "${file}"
    else
        # write api_url on the top of sumologic: extension
        sed -i.bak -e "1,/sumologic:/ s/sumologic:/sumologic:\\
\\${ext_indentation}api_base_url: $(escape_sed "${api_url}")/" "${file}"
    fi
}

# write tags to user configuration file
function write_tags() {
    local fields
    readonly fields="${1}"

    local file
    readonly file="${2}"

    local indentation
    readonly indentation="${3}"

    local ext_indentation
    readonly ext_indentation="${4}"

    local fields_indentation
    readonly fields_indentation="${ext_indentation}${indentation}"

    local fields_to_write
    fields_to_write="$(echo "${fields}" | sed -e "s/^\\([^\\]\\)/${fields_indentation}\\1/")"
    readonly fields_to_write

    # ToDo: ensure we override only sumologic `collector_fields`
    if grep "collector_fields" "${file}" > /dev/null; then
        sed -i.bak -e "s/collector_fields:.*$/collector_fields: ${fields_to_write}/" "${file}"
    else
        # write installation token on the top of sumologic: extension
        sed -i.bak -e "1,/sumologic:/ s/sumologic:/sumologic:\\
\\${ext_indentation}collector_fields: ${fields_to_write}/" "${file}"
    fi
}

# configure and enable the opamp extension for remote management
function write_opamp_extension() {
    local file
    readonly file="${1}"

    local directory
    readonly directory="${2}"

    local indentation
    readonly indentation="${3}"

    local ext_indentation
    readonly ext_indentation="${4}"

    local api_url
    readonly api_url="${5}"

    # add opamp extension if its missing
    if ! grep "opamp:" "${file}" > /dev/null; then
        sed -i.bak -e "1,/extensions:/ s/extensions:/extensions:\\
${indentation}opamp:/" "${file}"
    fi

    # set the remote_configuration_directory
    if grep "remote_configuration_directory:" "${file}" > /dev/null; then
        sed -i.bak -e "s/remote_configuration_directory:.*$/remote_configuration_directory: $(escape_sed "${directory}")/" "${file}"
    else
        sed -i.bak -e "s/opamp:/opamp:\\
\\${ext_indentation}remote_configuration_directory: $(escape_sed "${directory}")/" "${file}"
    fi

    # if a different base url is specified, configure the corresponding opamp endpoint
    if [[ -n "${api_url}" ]]; then
        wss_url=${api_url/https:/"wss:"}
        wss_url=${wss_url%/}
        wss_url=${wss_url/open-events/"opamp-events"}
        wss_url=${wss_url/open-collectors/"opamp-collectors"}
        wss_url=${wss_url/\.net/".net/v1/opamp"}
        wss_url=${wss_url/\.com/".com/v1/opamp"}

        if grep "endpoint: wss:" "${file}" > /dev/null; then
            sed -i.bak -e "s/endpoint: wss:.*$/endpoint: $(escape_sed "${wss_url}")/" "${file}"
        else
            sed -i.bak -e "s/opamp:/opamp:\\
\\${ext_indentation}endpoint: $(escape_sed "${wss_url}")/" "${file}"
        fi
    fi

    # enable the opamp extension
    if ! grep "\- opamp" "${file}" > /dev/null; then
        sed -i.bak -e "s/${indentation}extensions:/${indentation}extensions:\\
\\${ext_indentation}- opamp/" "${file}"
    fi
}

function get_binary_from_branch() {
    local branch
    readonly branch="${1}"

    local name
    readonly name="${2}"


    local actions_url actions_output artifacts_link artifact_id
    readonly actions_url="https://api.github.com/repos/SumoLogic/sumologic-otel-collector/actions/runs?status=success&branch=${branch}&event=push&per_page=1"
    echo -e "Getting artifacts from latest CI run for branch \"${branch}\":\t\t${actions_url}"
    actions_output="$(curl -f -sS \
      --connect-timeout 5 \
      --max-time 30 \
      --retry 5 \
      --retry-delay 0 \
      --retry-max-time 150 \
      -H "Accept: application/vnd.github+json" \
      -H "Authorization: token ${GITHUB_TOKEN}" \
      "${actions_url}")"
    readonly actions_output

    # get latest action run
    artifacts_link="$(echo "${actions_output}" | grep '"url"' | grep -oE '"https.*collector/actions.*"' -m 1)"
    # strip first and last double-quote from $artifacts_link
    artifacts_link=${artifacts_link%\"}
    artifacts_link="${artifacts_link#\"}"
    artifacts_link="${artifacts_link}/artifacts"
    readonly artifacts_link

    echo -e "Getting artifact id for CI run:\t\t${artifacts_link}"
    artifact_id="$(curl -f -sS \
    --connect-timeout 5 \
    --max-time 30 \
    --retry 5 \
    --retry-delay 0 \
    --retry-max-time 150 \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: token ${GITHUB_TOKEN}" \
    "${artifacts_link}" \
        | grep -E '"(id|name)"' \
        | grep -B 1 "\"${name}\"" -m 1 \
        | grep -oE "[0-9]+" -m 1)"
    readonly artifact_id

    local artifact_url download_path curl_args
    readonly artifact_url="https://api.github.com/repos/SumoLogic/sumologic-otel-collector/actions/artifacts/${artifact_id}/zip"
    readonly download_path="${DOWNLOAD_CACHE_DIR}/${name}.zip"
    echo -e "Downloading binary from: ${artifact_url}"
    curl_args=(
        "-fL"
        "--connect-timeout" "5"
        "--max-time" "${CURL_MAX_TIME}"
        "--retry" "5"
        "--retry-delay" "0"
        "--retry-max-time" "150"
        "--output" "${download_path}"
        "--progress-bar"
    )
    if [ "${KEEP_DOWNLOADS}" == "true" ]; then
        curl_args+=("-z" "${download_path}")
    fi
    curl "${curl_args[@]}" \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: token ${GITHUB_TOKEN}" \
        "${artifact_url}"

    unzip -p "$download_path" "${name}" >"${TMPDIR}"/otelcol-sumo
    if [ "${KEEP_DOWNLOADS}" == "false" ]; then
        rm -f "${download_path}"
    fi
}

function get_binary_from_url() {
    local url download_filename download_path curl_args
    readonly url="${1}"
    echo -e "Downloading:\t\t${url}"

    download_filename=$(basename "${url}")
    readonly download_filename
    readonly download_path="${DOWNLOAD_CACHE_DIR}/${download_filename}"
    curl_args=(
        "-fL"
        "--connect-timeout" "5"
        "--max-time" "${CURL_MAX_TIME}"
        "--retry" "5"
        "--retry-delay" "0"
        "--retry-max-time" "150"
        "--output" "${download_path}"
        "--progress-bar"
    )
    if [ "${KEEP_DOWNLOADS}" == "true" ]; then
        curl_args+=("-z" "${download_path}")
    fi
    curl "${curl_args[@]}" "${url}"

    cp -f "${download_path}" "${TMPDIR}"/otelcol-sumo

    if [ "${KEEP_DOWNLOADS}" == "false" ]; then
        rm -f "${download_path}"
    fi
}

function get_package_from_branch() {
    local branch
    readonly branch="${1}"

    local name
    readonly name="${2}"

    local actions_url actions_output artifacts_link artifact_id
    readonly actions_url="https://api.github.com/repos/${PACKAGE_GITHUB_ORG}/${PACKAGE_GITHUB_REPO}/actions/runs?status=success&branch=${branch}&event=push&per_page=1"
    echo -e "Getting artifacts from latest CI run for branch \"${branch}\":\t\t${actions_url}"
    actions_output="$(curl -f -sS \
      --connect-timeout 5 \
      --max-time 30 \
      --retry 5 \
      --retry-delay 0 \
      --retry-max-time 150 \
      -H "Accept: application/vnd.github+json" \
      -H "Authorization: token ${GITHUB_TOKEN}" \
      "${actions_url}")"
    readonly actions_output

    # get latest action run
    artifacts_link="$(echo "${actions_output}" | grep '"url"' | grep -oE '"https.*packaging/actions.*"' -m 1)"

    # strip first and last double-quote from $artifacts_link
    artifacts_link=${artifacts_link%\"}
    artifacts_link="${artifacts_link#\"}"
    artifacts_link="${artifacts_link}/artifacts"
    readonly artifacts_link

    echo -e "Getting artifact id for CI run:\t\t${artifacts_link}"
    artifact_id="$(curl -f -sS \
    --connect-timeout 5 \
    --max-time 30 \
    --retry 5 \
    --retry-delay 0 \
    --retry-max-time 150 \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: token ${GITHUB_TOKEN}" \
    "${artifacts_link}" \
        | grep -E '"(id|name)"' \
        | grep -B 1 "\"${name}\"" -m 1 \
        | grep -oE "[0-9]+" -m 1)"
    readonly artifact_id

    echo "Artifact ID: ${artifact_id}"

    local artifact_url download_path curl_args
    readonly artifact_url="https://api.github.com/repos/${PACKAGE_GITHUB_ORG}/${PACKAGE_GITHUB_REPO}/actions/artifacts/${artifact_id}/zip"
    readonly download_path="${DOWNLOAD_CACHE_DIR}/${artifact_id}.zip"
    echo -e "Downloading package from: ${artifact_url}"
    curl_args=(
        "-fL"
        "--connect-timeout" "5"
        "--max-time" "${CURL_MAX_TIME}"
        "--retry" "5"
        "--retry-delay" "0"
        "--retry-max-time" "150"
        "--output" "${download_path}"
        "--progress-bar"
    )
    if [ "${KEEP_DOWNLOADS}" == "true" ]; then
        curl_args+=("-z" "${download_path}")
    fi
    curl "${curl_args[@]}" \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: token ${GITHUB_TOKEN}" \
        "${artifact_url}"

    unzip -p "$download_path" > "${TMPDIR}/otelcol-sumo.pkg"
    if [ "${KEEP_DOWNLOADS}" == "false" ]; then
        rm -f "${download_path}"
    fi
}

function get_package_from_url() {
    local url download_filename download_path curl_args
    readonly url="${1}"
    echo -e "Downloading:\t\t${url}"

    download_filename=$(basename "${url}")
    readonly download_filename
    readonly download_path="${DOWNLOAD_CACHE_DIR}/${download_filename}"
    curl_args=(
        "-fL"
        "--connect-timeout" "5"
        "--max-time" "${CURL_MAX_TIME}"
        "--retry" "5"
        "--retry-delay" "0"
        "--retry-max-time" "150"
        "--output" "${download_path}"
        "--progress-bar"
    )
    if [ "${KEEP_DOWNLOADS}" == "true" ]; then
        curl_args+=("-z" "${download_path}")
    fi
    curl "${curl_args[@]}" "${url}"

    cp -f "${download_path}" "${TMPDIR}/otelcol-sumo.pkg"

    if [ "${KEEP_DOWNLOADS}" == "false" ]; then
        rm -f "${download_path}"
    fi
}

function set_acl_on_log_paths() {
    if command -v setfacl &> /dev/null; then
        for log_path in ${ACL_LOG_FILE_PATHS}; do
            if [ -d "$log_path" ]; then
                echo -e "Running: setfacl -R -m d:u:${SYSTEM_USER}:r-x,u:${SYSTEM_USER}:r-x,g:${SYSTEM_USER}:r-x ${log_path}"
                setfacl -R -m d:u:${SYSTEM_USER}:r-x,d:g:${SYSTEM_USER}:r-x,u:${SYSTEM_USER}:r-x,g:${SYSTEM_USER}:r-x "${log_path}"
            fi
        done
    else
        echo ""
        echo "setfacl command not found, skipping ACL creation for system log file paths."
        echo -e "You can fix it manually by installing setfacl and executing the following commands:"
        for log_path in ${ACL_LOG_FILE_PATHS}; do
            if [ -d "$log_path" ]; then
                echo -e "-> setfacl -R -m d:u:${SYSTEM_USER}:r-x,d:g:${SYSTEM_USER}:r-x,u:${SYSTEM_USER}:r-x,g:${SYSTEM_USER}:r-x ${log_path}"
            fi
        done
        echo ""
    fi
}

function plutil_create_key() {
    local file key type value
    readonly file="${1}"
    readonly key="${2}"
    readonly type="${3}"
    readonly value="${4}"

    if ! plutil -insert "${key}" -"${type}" "${value}" "${file}"; then
        echo "plutil_create_key error: key=${key}, type=${type}, value=${value}, file=${file}"
        exit 1
    fi
}

function plutil_delete_key() {
    local file key
    readonly file="${1}"
    readonly key="${2}"

    if ! plutil -remove "${key}" "${file}"; then
        echo "plutil_delete_key error: key=${key}, file=${file}"
        exit 1
    fi
}

function plutil_extract_key() {
    local file key output
    readonly file="${1}"
    readonly key="${2}"

    if output="$(plutil -extract "${key}" raw -o - "${file}")"; then
        echo "${output}"
    else
        echo
    fi
}

function plutil_key_exists() {
    local file key
    readonly file="${1}"
    readonly key="${2}"

    plutil -type "${key}" "${file}" > /dev/null 2>&1
}

function plutil_key_is_type() {
    local file key type
    readonly file="${1}"
    readonly key="${2}"
    readonly type="${3}"

    plutil -type "${key}" -expect "${type}" "${file}" > /dev/null 2>&1
}

function plutil_replace_key() {
    local file key type value
    readonly file="${1}"
    readonly key="${2}"
    readonly type="${3}"
    readonly value="${4}"

    if ! plutil -replace "${key}" -"${type}" "${value}" "${file}"; then
        echo "plutil_replace_key error: key=${key}, file=${file}"
        exit 1
    fi
}

############################ Main code

# Display deprecation warning
echo "============================================================================="
echo "WARNING: This installation script is DEPRECATED"
echo ""
echo "This script has been moved to the sumologic-otel-collector-packaging repository:"
echo "https://github.com/SumoLogic/sumologic-otel-collector-packaging"
echo ""
echo "Please use the installation scripts from the packaging repository instead."
echo "This script will be removed in a future release."
echo ""
echo "For the latest installation instructions, visit:"
echo "  https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/"
echo "============================================================================="
echo ""

OS_TYPE="$(get_os_type)"
ARCH_TYPE="$(get_arch_type)"
readonly OS_TYPE ARCH_TYPE

echo -e "Detected OS type:\t${OS_TYPE}"
echo -e "Detected architecture:\t${ARCH_TYPE}"

set_defaults
parse_options "$@"
set_tmpdir
install_missing_dependencies
check_dependencies

readonly SUMOLOGIC_INSTALLATION_TOKEN API_BASE_URL FIELDS CONTINUE FILE_STORAGE CONFIG_DIRECTORY SYSTEMD_CONFIG UNINSTALL
readonly USER_CONFIG_DIRECTORY USER_ENV_DIRECTORY CONFIG_DIRECTORY CONFIG_PATH COMMON_CONFIG_PATH
readonly ACL_LOG_FILE_PATHS
readonly INSTALL_HOSTMETRICS
readonly REMOTELY_MANAGED
readonly CURL_MAX_TIME
readonly LAUNCHD_CONFIG LAUNCHD_ENV_KEY LAUNCHD_TOKEN_KEY

if [[ "${UNINSTALL}" == "true" ]]; then
    uninstall
    exit 0
fi

# Attempt to find a token from an existing installation
case "${OS_TYPE}" in
darwin)
  USER_TOKEN="$(plutil_extract_key "${LAUNCHD_CONFIG}" "${LAUNCHD_TOKEN_KEY}")"
  ;;
*)
  USER_TOKEN="$(get_user_config "${COMMON_CONFIG_PATH}")"

  # If Systemd is not disabled, try to extract token from systemd env file
  if [[ -z "${USER_TOKEN}" && "${SYSTEMD_DISABLED}" == "false" ]]; then
      USER_TOKEN="$(get_user_env_config "${TOKEN_ENV_FILE}")"
  fi
  ;;
esac
readonly USER_TOKEN

# Exit if installation token is not set and there is no user configuration
if [[ -z "${SUMOLOGIC_INSTALLATION_TOKEN}" && "${SKIP_TOKEN}" != "true" && -z "${USER_TOKEN}" && -z "${DOWNLOAD_ONLY}" ]]; then
    echo "Installation token has not been provided. Please set the '${ENV_TOKEN}' environment variable."
    echo "You can ignore this requirement by adding '--${ARG_LONG_SKIP_TOKEN} argument."
    exit 1
fi

# verify if passed arguments are the same like in user's configuration
if [[ -z "${DOWNLOAD_ONLY}" ]]; then
    if [[ -n "${USER_TOKEN}" && -n "${SUMOLOGIC_INSTALLATION_TOKEN}" && "${USER_TOKEN}" != "${SUMOLOGIC_INSTALLATION_TOKEN}" ]]; then
        echo "You are trying to install with different token than in your configuration file!"
        exit 1
    fi

    if [[ -f "${COMMON_CONFIG_PATH}" ]]; then
        INDENTATION="$(get_indentation "${COMMON_CONFIG_PATH}" "${INDENTATION}")"
        EXT_INDENTATION="$(get_extension_indentation "${COMMON_CONFIG_PATH}" "${INDENTATION}")"
        readonly INDENTATION EXT_INDENTATION

        USER_API_URL="$(get_user_api_url "${COMMON_CONFIG_PATH}")"
        if [[ -n "${USER_API_URL}" && -n "${API_BASE_URL}" && "${USER_API_URL}" != "${API_BASE_URL}" ]]; then
            echo "You are trying to install with different api base url than in your configuration file!"
            exit 1
        fi

        USER_FIELDS="$(get_user_tags "${COMMON_CONFIG_PATH}" "${INDENTATION}" "${EXT_INDENTATION}")"
        FIELDS_TO_COMPARE="$(get_fields_to_compare "${FIELDS}")"

        if [[ -n "${USER_FIELDS}" && -n "${FIELDS_TO_COMPARE}" && "${USER_FIELDS}" != "${FIELDS_TO_COMPARE}" ]]; then
            echo "You are trying to install with different tags than in your configuration file!"
            exit 1
        fi
    fi
fi

set +u
if [[ -n "${BINARY_BRANCH}" && -z "${GITHUB_TOKEN}" ]]; then
    echo "GITHUB_TOKEN env is required for '${ARG_LONG_BINARY_BRANCH}' option"
    exit 1
fi
set -u

# Disable systemd if token is not specified at all
if [[ -z "${SUMOLOGIC_INSTALLATION_TOKEN}" && -z "${USER_TOKEN}" ]]; then
    SYSTEMD_DISABLED=true
fi

readonly SYSTEMD_DISABLED

if [ "${FIPS}" == "true" ]; then
    case "${OS_TYPE}" in
    linux)
        if  [ "${ARCH_TYPE}" != "amd64" ] && [ "${ARCH_TYPE}" != "arm64" ]; then
            echo "Error: The FIPS-approved binary is only available for linux/amd64 and linux/arm64"
            exit 1
        fi
        ;;
    *)
        echo "Error: The FIPS-approved binary is only available for linux"
        exit 1
        ;;
    esac
fi

if [[ "${OS_TYPE}" == "darwin" ]]; then
    package_arch=""
    case "${ARCH_TYPE}" in
      "amd64") package_arch="intel" ;;
      "arm64") package_arch="apple" ;;
      *)
        echo "Unsupported architecture for darwin: ${ARCH_TYPE}"
        exit 1
        ;;
    esac
    readonly package_arch

    if [[ "${SKIP_CONFIG}" == "true" ]]; then
        echo "SKIP_CONFIG is not supported on darwin"
        exit 1
    fi

    echo -e "Getting versions..."
    # Get versions, but ignore errors as we fallback to other methods later
    VERSIONS="$(get_package_versions || echo "")"

    # Use user's version if set, otherwise get latest version from API (or website)
    if [[ -z "${VERSION}" ]]; then
        VERSION="$(get_latest_package_version "${VERSIONS}")"
    fi

    VERSION_PREFIX="${VERSION%.*}"       # cut off the suffix starting with the last stop
    MAJOR_VERSION="${VERSION_PREFIX%.*}" # take the prefix from before the first stop
    MINOR_VERSION="${VERSION_PREFIX#*.}" # take the suffix after the first stop

    readonly VERSIONS VERSION INSTALLED_VERSION VERSION_PREFIX MAJOR_VERSION MINOR_VERSION

    echo -e "Version to install:\t${VERSION}"

    package_suffix="${package_arch}.pkg"

    if [[ -n "${BINARY_BRANCH}" ]]; then
        artifact_name="otelcol-sumo_.*-${package_suffix}"
        get_package_from_branch "${BINARY_BRANCH}" "${artifact_name}"
    else
        artifact_name="otelcol-sumo_${VERSION}-${package_suffix}"
        readonly artifact_name

        LINK="https://github.com/${PACKAGE_GITHUB_ORG}/${PACKAGE_GITHUB_REPO}/releases/download/v${VERSION}/${artifact_name}"
        readonly LINK

        get_package_from_url "${LINK}"
    fi

    pkg="${TMPDIR}/otelcol-sumo.pkg"
    choices="${TMPDIR}/otelcol-sumo-choices.xml"
    readonly pkg choices

    if [[ "${DOWNLOAD_ONLY}" == "true" ]]; then
        echo "Package downloaded to: ${pkg}"
        exit 0
    fi

    # Extract choices xml from meta package, override the choices to enable
    # optional choices, and then install using the new choice selections
    installer -showChoiceChangesXML -pkg "${pkg}" -target / > "${choices}"

    # Determine how many installation choices exist
    choices_count=$(plutil -convert raw -o - "${choices}")
    readonly choices_count

    # Loop through each installation choice
    for (( i=0; i < "${choices_count}"; i++ )); do
        choice_id_key="${i}.choiceIdentifier"
        choice_attr_key="${i}.choiceAttribute"
        attr_setting_key="${i}.attributeSetting"

        # Skip if choiceAttribute does not equal selected
        choice_attr="$(plutil_extract_key "${choices}" "${choice_attr_key}")"
        if [ "$choice_attr" != "selected" ]; then
            continue
        fi

        # Get the choice identifier
        choice_id="$(plutil_extract_key "${choices}" "${choice_id_key}")"

        # Mark the choice as selected if the feature flag is true
        case "${choice_id}" in
        "otelcol-sumo-hostmetricsChoice")
          if [[ "${INSTALL_HOSTMETRICS}" == "true" ]]; then
              echo -e "Enabling ${OS_TYPE} hostmetrics install option"
              plutil_replace_key "${choices}" "${attr_setting_key}" "integer" 1
          fi
          ;;
        esac
    done

    installer -applyChoiceChangesXML "$choices" -pkg "$pkg" -target /

    if [[ -n "${SUMOLOGIC_INSTALLATION_TOKEN}" && -z "${USER_TOKEN}" && "${SKIP_TOKEN}" != "true" ]]; then
        echo "Writing installation token to launchd config"
        write_installation_token_launchd "${SUMOLOGIC_INSTALLATION_TOKEN}" "${LAUNCHD_CONFIG}"
    fi

    setup_config_darwin

    # Run an unload/load launchd config to pull in new changes & restart the service
    launchctl unload "${LAUNCHD_CONFIG}"
    launchctl load -w "${LAUNCHD_CONFIG}"

    echo "Waiting for otelcol to start"
    while ! launchctl print system/otelcol-sumo | grep -q "state = running"; do
        sleep 0.1
    done
    OTEL_EXITED_WITH_ERROR=false
    echo 'Checking otelcol status'
    for i in {1..15}; do
        if launchctl print system/otelcol-sumo | grep -q "last exit code = 1"; then
            OTEL_EXITED_WITH_ERROR=true
            break;
        fi
        sleep 1
    done
    if [[ "${OTEL_EXITED_WITH_ERROR}" == "true" ]]; then
        echo "Failed to launch otelcol"
        tail /var/log/otelcol-sumo/otelcol-sumo.log
        exit 1
    fi
    exit 0
fi

echo -e "Getting installed version..."
INSTALLED_VERSION="$(get_installed_version)"
echo -e "Installed version:\t${INSTALLED_VERSION:-none}"

echo -e "Getting versions..."
# Get versions, but ignore errors as we fallback to other methods later
VERSIONS="$(get_versions || echo "")"

# Use user's version if set, otherwise get latest version from API (or website)
if [[ -z "${VERSION}" ]]; then
    VERSION="$(get_latest_version "${VERSIONS}")"
fi

VERSION_PREFIX="${VERSION%.*}"       # cut off the suffix starting with the last stop
MAJOR_VERSION="${VERSION_PREFIX%.*}" # take the prefix from before the first stop
MINOR_VERSION="${VERSION_PREFIX#*.}" # take the suffix after the first stop


readonly VERSIONS VERSION INSTALLED_VERSION VERSION_PREFIX MAJOR_VERSION MINOR_VERSION

echo -e "Version to install:\t${VERSION}"

if [[ -z "${CONFIG_BRANCH}" ]]; then
    # Remove glob for versions up to 0.57
    if (( MAJOR_VERSION == 0 && MINOR_VERSION <= 57 )); then
        CONFIG_BRANCH="9e06ada346b5e7fb3df582f28e582e07730899de"
    else
        CONFIG_BRANCH="v${VERSION}"
    fi
fi
readonly CONFIG_BRANCH BINARY_BRANCH

# Check if otelcol is already in newest version
if [[ "${INSTALLED_VERSION}" == "${VERSION}" && -z "${BINARY_BRANCH}" ]]; then
    echo -e "OpenTelemetry collector is already in newest (${VERSION}) version"
else

    # add newline before breaking changes and changelog
    echo ""
    if [[ -n "${INSTALLED_VERSION}" && -z "${BINARY_BRANCH}" ]]; then
        # Take versions from installed up to the newest
        BETWEEN_VERSIONS="$(get_versions_from "${VERSIONS}" "${INSTALLED_VERSION}")"
        readonly BETWEEN_VERSIONS
        print_breaking_changes "${BETWEEN_VERSIONS}"
    fi

    echo -e "Changelog:\t\thttps://github.com/SumoLogic/sumologic-otel-collector/blob/main/CHANGELOG.md"
    # add newline after breaking changes and changelog
    echo ""

    # Add -fips to the suffix if necessary
    binary_suffix="${OS_TYPE}_${ARCH_TYPE}"
    if [ "${FIPS}" == "true" ]; then
        echo "Getting FIPS-compliant binary"
        binary_suffix="fips-${binary_suffix}"
    fi

    if [[ -n "${BINARY_BRANCH}" ]]; then
        get_binary_from_branch "${BINARY_BRANCH}" "otelcol-sumo-${binary_suffix}"
    else
        LINK="https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v${VERSION}/otelcol-sumo-${VERSION}-${binary_suffix}"
        readonly LINK

        get_binary_from_url "${LINK}"
    fi

    echo -e "Moving otelcol-sumo to /usr/local/bin"
    mv "${TMPDIR}"/otelcol-sumo "${SUMO_BINARY_PATH}"
    echo -e "Setting ${SUMO_BINARY_PATH} to be executable"
    chmod +x "${SUMO_BINARY_PATH}"

    verify_installation
fi

if [[ "${DOWNLOAD_ONLY}" == "true" ]]; then
    exit 0
fi

if [[ "${SKIP_CONFIG}" == "false" ]]; then
    setup_config
fi

if [[ "${SYSTEMD_DISABLED}" == "true" ]]; then
    COMMAND_FLAGS=""

    if [[ "${REMOTELY_MANAGED}" == "true" ]]; then
        COMMAND_FLAGS="--remote-config \"opamp:${CONFIG_PATH}\""
    else
        COMMAND_FLAGS="--config=${CONFIG_PATH}"
        # Add glob for versions above 0.57
        if (( MAJOR_VERSION >= 0 && MINOR_VERSION > 57 )); then
            COMMAND_FLAGS="${COMMAND_FLAGS} --config \"glob:${CONFIG_DIRECTORY}/conf.d/*.yaml\""
        else
            COMMAND_FLAGS="${COMMAND_FLAGS} --config ${COMMON_CONFIG_PATH}"
        fi
    fi

    echo ""
    echo Warning: running as a service is not supported on your operation system.
    echo "Please use 'sudo otelcol-sumo ${COMMAND_FLAGS}' to run Sumo Logic Distribution for OpenTelemetry Collector"
    exit 0
fi

echo 'We are going to set up a systemd service'

if [[ -n "${SUMOLOGIC_INSTALLATION_TOKEN}" && -z "${USER_TOKEN}" ]]; then
    echo 'Writing installation token to env file'
    write_installation_token_env "${SUMOLOGIC_INSTALLATION_TOKEN}" "${TOKEN_ENV_FILE}"
    chmod -R 440 "${TOKEN_ENV_FILE}"
fi

echo 'Creating user and group'
if getent passwd "${SYSTEM_USER}" > /dev/null; then
    echo 'User and group already created'
else
    ADDITIONAL_OPTIONS=""
    if [[ -d "${HOME_DIRECTORY}" ]]; then
        # do not create home directory as it already exists
        ADDITIONAL_OPTIONS="-M"
    else
        # create home directory
        ADDITIONAL_OPTIONS="-m"
    fi
    readonly ADDITIONAL_OPTIONS
    useradd "${ADDITIONAL_OPTIONS}" -rUs /bin/false -d "${HOME_DIRECTORY}" "${SYSTEM_USER}"
fi

echo 'Creating ACL grants on log paths'
set_acl_on_log_paths

if [[ "${SKIP_CONFIG}" == "false" ]]; then
    echo 'Changing ownership for config and storage'
    chown -R "${SYSTEM_USER}":"${SYSTEM_USER}" "${HOME_DIRECTORY}" "${CONFIG_DIRECTORY}"/*
    chown -R "${SYSTEM_USER}":"${SYSTEM_USER}" "${USER_ENV_DIRECTORY}"

    if [[ "${REMOTELY_MANAGED}" == "true" ]]; then
        chown -R "${SYSTEM_USER}":"${SYSTEM_USER}" "${REMOTE_CONFIG_DIRECTORY}"
    fi
fi

SYSTEMD_CONFIG_URL="https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/${CONFIG_BRANCH}/examples/systemd/otelcol-sumo.service"

TMP_SYSTEMD_CONFIG="${TMPDIR}/otelcol-sumo.service"
TMP_SYSTEMD_CONFIG_BAK="${TMP_SYSTEMD_CONFIG}.bak"
echo 'Getting service configuration'
curl --retry 5 --connect-timeout 5 --max-time 30 --retry-delay 0 --retry-max-time 150 -fL "${SYSTEMD_CONFIG_URL}" --output "${TMP_SYSTEMD_CONFIG}" --progress-bar
sed -i.bak -e "s%/etc/otelcol-sumo%${CONFIG_DIRECTORY}%" "${TMP_SYSTEMD_CONFIG}"
sed -i.bak -e "s%/etc/otelcol-sumo/env%${USER_ENV_DIRECTORY}%" "${TMP_SYSTEMD_CONFIG}"

# Remove glob for versions up to 0.57
if (( MAJOR_VERSION == 0 && MINOR_VERSION <= 57 )); then
    sed -i.bak -e "s% --config \"glob.*\"% --config ${COMMON_CONFIG_PATH}%" "${TMP_SYSTEMD_CONFIG}"
fi

if [[ "${REMOTELY_MANAGED}" == "true" ]]; then
    sed -i.bak -e "s% --config.*$% --remote-config \"opamp:${CONFIG_PATH}\"%" "${TMP_SYSTEMD_CONFIG}"
fi

# clean up bak file
rm -f "${TMP_SYSTEMD_CONFIG_BAK}"

mv "${TMP_SYSTEMD_CONFIG}" "${SYSTEMD_CONFIG}"

if command -v sestatus && sestatus; then
    echo "SELinux is enabled, relabeling binary and systemd unit file"

    if command -v semanage &> /dev/null; then
        # Check if there's already an fcontext record for the collector bin.
        if semanage fcontext -l | grep otelcol-sumo &> /dev/null; then
            # Modify the existing fcontext record.
            semanage fcontext -m -t bin_t /usr/local/bin/otelcol-sumo
        else
            # Add an fcontext record.
            semanage fcontext -a -t bin_t /usr/local/bin/otelcol-sumo
        fi
        restorecon -v "${SUMO_BINARY_PATH}"
        restorecon -v "${SYSTEMD_CONFIG}"
    else
        echo "semanage command not found, skipping SELinux relabeling"
    fi
fi

echo 'Reloading systemd'
systemctl daemon-reload

echo 'Enable otelcol-sumo service'
systemctl enable otelcol-sumo

echo 'Starting otelcol-sumo service'
systemctl restart otelcol-sumo

echo 'Waiting 10s before checking status'
sleep 10
if ! systemctl status otelcol-sumo --no-pager; then
    echo "Failed to launch otelcol"
    exit 1
fi

exit 0
