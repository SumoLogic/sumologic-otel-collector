#!/bin/bash

set -eo pipefail

if ! type yq >/dev/null
then
    echo "yq (https://github.com/mikefarah/yq/) is not installed, please install it."
    exit 1
fi

if [[ ! $(yq --version | cut -f 4 -d ' ') =~ 4.[0-9]+.[0-9]+ ]]
then
    printf "yq (https://github.com/mikefarah/yq/) v4 should be installed."
    exit 1
fi

if [[ $(uname) == "Darwin" ]]; then
readonly GREP=ggrep
else
readonly GREP=grep
fi

readonly BUILDER_CONFIG="otelcolbuilder/.otelcol-builder.yaml"
OT_VERSION=$(yq e '.dist.version' "${BUILDER_CONFIG}" | cut -f1,2 -d'.')
readonly OT_VERSION
readonly CONTRIB_PLUGIN_HTTP_URL_REGEX="https://github.com/open-telemetry/opentelemetry-collector(-contrib)?/tree/(v[0-9]+.[0-9]+.[0-9]+)/(receiver|processor|exporter|extension)/([a-zA-Z]+)"
readonly CONTRIB_PLUGIN_REGEX="github.com/open-telemetry/opentelemetry-collector(-contrib)?/(receiver|processor|exporter|extension)/([a-zA-Z]+)"
BUILDER_PLUGINS=$(yq e '... comments="" | [.receivers[], .exporters[], .processors[], .extensions[]][] | to_entries | .[].value | match("[^\s]+") | .string' "${BUILDER_CONFIG}")
readonly BUILDER_PLUGINS

# For all plugins in README.md ...
for plugin_url in $(${GREP} -o -E "${CONTRIB_PLUGIN_HTTP_URL_REGEX}" README.md)
do
    # ... check if the version in the README.md is the same as in the builder config
    if [[ ${plugin_url} =~ ${CONTRIB_PLUGIN_HTTP_URL_REGEX} ]]
    then
        PLUGIN_VERSION_FROM_README=$(echo "${BASH_REMATCH[2]}" | cut -f1,2 -d'.' )
        if [[ ${PLUGIN_VERSION_FROM_README} != v${OT_VERSION} ]]
        then
            printf "There's an unexpected plugin version in README.md for %s (should be %s)\n" "${plugin_url}" "${OT_VERSION}.*"
            fail=1
        fi

        # ... and when the plugin is from contrib ... (core plugins are not listed in builder config) ...
        if [[ "${BASH_REMATCH[1]}" == "-contrib" ]]; then
            # ... check if the plugin from README.md is also in the builder config
            PLUGIN_NAME_FROM_README="${BASH_REMATCH[4]}"

            if ! ${GREP} -o "${PLUGIN_NAME_FROM_README}" "${BUILDER_CONFIG}" > /dev/null
            then
                printf "A plugin mentioned in README: %s (%s), is not in the builder config %s\n" "${PLUGIN_NAME_FROM_README}" "${BASH_REMATCH[0]}" "${BUILDER_CONFIG}"
                fail=1
            fi
        fi
    fi
done

# For all plugins in builder config ...
for BUILDER_PLUGIN in ${BUILDER_PLUGINS}
do
    # ... check if the plugin is also mentioned in the README.md
    if [[ ${BUILDER_PLUGIN} =~ ${CONTRIB_PLUGIN_REGEX} ]]
    then
        PLUGIN_NAME_FROM_BUILDER_CONFIG="${BASH_REMATCH[3]}"
        if ! ${GREP} -o "${PLUGIN_NAME_FROM_BUILDER_CONFIG}" README.md > /dev/null
        then
            printf "A plugin from the builder config: %s (%s), is not mentioned in the README.md\n" "${PLUGIN_NAME_FROM_BUILDER_CONFIG}" "${BASH_REMATCH[0]}"
            fail=1
        fi
    fi
done

if [[ ${fail} == 1 ]]; then exit 1; fi
