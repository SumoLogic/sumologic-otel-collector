#!/usr/bin/env bash
set -e

# This script updates version references in project files.
#
# Core (go.opentelemetry.io/collector) and contrib
# (github.com/open-telemetry/opentelemetry-collector-contrib) are released
# separately, and contrib sometimes publishes a patch release that core (and the
# ocb builder, which is released together with core) does not have. They are
# therefore tracked separately: the core version is used for the ocb version, the
# distribution version, the core components in the builder config and the links to
# the core repository, while the contrib version is used for the contrib components
# and the links to the contrib repository.
#
# Usage:
#   ./renovate.sh  [--from-ver "0.130.0"] [--from-core-ver "0.130.0"]
# Options:
#   --from-ver        Hard coded from version to be used for both core and contrib.
#   --from-core-ver   Hard coded from version to be used for core only.

FROM_VERSION_FIXED=""
FROM_CORE_VERSION_FIXED=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --from-ver)
            shift
            FROM_VERSION_FIXED="$1"
            shift
            ;;
        --from-core-ver)
            shift
            FROM_CORE_VERSION_FIXED="$1"
            shift
            ;;
        *)
            shift
            ;;
    esac
done
make gomod-download-all

GO_MOD=pkg/extension/opampextension/go.mod

# Prints the single 0.x version used by the direct requirements of the given module
# prefix, or exits if there is none or more than one. Indirect requirements are
# ignored: they can be stuck on an old version, e.g. a module removed upstream that
# is still pulled in by a third party library.
single_version() {
    local name="$1" prefix="$2" versions count
    versions=$(grep -v '// indirect' "$GO_MOD" | grep -o "${prefix}[^ ]* v[0-9.]\+" | awk '{print $2}' | sed 's/^v//' | grep '^0\.' | sort -u)
    if [[ -z "$versions" ]]; then
        echo "Error: No $name versions found" >&2
        exit 1
    fi
    count=$(echo "$versions" | wc -l)
    if [[ "$count" -gt 1 ]]; then
        echo "Error: Multiple $name versions found: $(echo "$versions" | tr '\n' ' ')" >&2
        exit 1
    fi
    echo "$versions"
}

TO_CORE_VER=$(single_version core 'go.opentelemetry.io/collector')
TO_CONTRIB_VER=$(single_version contrib 'github.com/open-telemetry/opentelemetry-collector-contrib')

# Prints the versions of the links to the given repository ("opentelemetry-collector"
# or "opentelemetry-collector-contrib") found in docs/configuration.md.
doc_versions() {
    grep -ho "open-telemetry/$1/\(tree\|blob\)/v[0-9.]\+" docs/configuration.md | sed 's/.*\/v//' | sort -u | tr '\n' ' '
}

if [[ -n "$FROM_VERSION_FIXED" ]]; then
    FROM_CONTRIB_VERSIONS="$FROM_VERSION_FIXED"
    FROM_CORE_VERSIONS="${FROM_CORE_VERSION_FIXED:-$FROM_VERSION_FIXED}"
else
    FROM_CONTRIB_VERSIONS=$(doc_versions opentelemetry-collector-contrib)
    FROM_CORE_VERSIONS="${FROM_CORE_VERSION_FIXED:-$(doc_versions opentelemetry-collector)}"
fi
echo "Core: initial versions: $FROM_CORE_VERSIONS, changing to: $TO_CORE_VER."
echo "Contrib: initial versions: $FROM_CONTRIB_VERSIONS, changing to: $TO_CONTRIB_VER."

BUILDER_CONFIG=otelcolbuilder/.otelcol-builder.yaml
DOCS=(README.md docs/configuration.md docs/migration.md docs/performance.md)

# The distribution and the ocb version follow core.
yq e ".dist.version = \"${TO_CORE_VER}\"" -i "$BUILDER_CONFIG"

for FROM_VER in $FROM_CORE_VERSIONS; do
    echo "Processing core version $FROM_VER"

    yq -i '(.. | select(tag=="!!str")) |= sub("(go\.opentelemetry\.io/collector.*) v'"${FROM_VER}"'", "$1 v'"${TO_CORE_VER}"'")' "$BUILDER_CONFIG"
    sed -i "s/${FROM_VER}/${TO_CORE_VER}/" otelcolbuilder/Makefile
    sed -i "s/\(open-telemetry\/opentelemetry-collector\/\(blob\|tree\)\/v\)${FROM_VER}/\1${TO_CORE_VER}/" "${DOCS[@]}"
done

for FROM_VER in $FROM_CONTRIB_VERSIONS; do
    echo "Processing contrib version $FROM_VER"

    yq -i '(.. | select(tag=="!!str")) |= sub("(github\.com/open-telemetry/opentelemetry-collector-contrib.*) v'"${FROM_VER}"'", "$1 v'"${TO_CONTRIB_VER}"'")' "$BUILDER_CONFIG"
    sed -i "s/\(open-telemetry\/opentelemetry-collector-contrib\/\(blob\|tree\)\/v\)${FROM_VER}/\1${TO_CONTRIB_VER}/" "${DOCS[@]}"
done
