#!/usr/bin/env bash
set -e

# This script updates version references in project files.
# Usage:
#   ./renovate.sh  [--from-ver "0.130.0"
# Options:
#   --from-ver      Hard coded form version to be used.

FROM_VERSION_FIXED=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --from-ver)
            FROM_VERSION_FIXED="$1"
            shift
            ;;
        *)
            shift
            ;;
    esac
done
make gomod-download-all
TO_VERSIONS=$(grep -o 'github.com/open-telemetry/opentelemetry-collector-contrib[^ ]* v[0-9.]\+' pkg/extension/opampextension/go.mod | awk '{print $2}' | sed 's/^v//' | sort -u)
TO_VERSION_COUNT=$(echo "$versions" | wc -l)
if [[ "$TO_VERSION_COUNT" -gt 1 ]]; then
  echo "Error: Multiple versions found: $TO_VERSIONS"
  exit 1
fi
if [[ -n "$FROM_VERSION_FIXED" ]]; then
  FROM_VERSIONS="$FROM_VERSION_FIXED"
else
  FROM_VERSIONS=$(grep -ho 'tree/v[0-9.]\+' docs/configuration.md | sed 's/tree\/v//' | sort -u | tr '\n' ' ')
fi
TO_VER=${TO_VERSIONS[0]}
echo "Initial versions: $FROM_VERSIONS, changing to: $TO_VER."

for FROM_VER in $FROM_VERSIONS; do
    echo "Processing version $FROM_VER"

    # Update version references in otelcol-builder.yaml
    yq e ".dist.version = \"${TO_VER}\"" -i otelcolbuilder/.otelcol-builder.yaml
    yq -i '(.. | select(tag=="!!str")) |= sub("(go\.opentelemetry\.io/collector.*) v'"${FROM_VER}"'", "$1 v'"${TO_VER}"'")' otelcolbuilder/.otelcol-builder.yaml
    yq -i '(.. | select(tag=="!!str")) |= sub("(github\.com/open-telemetry/opentelemetry-collector-contrib.*) v'"${FROM_VER}"'", "$1 v'"${TO_VER}"'")' otelcolbuilder/.otelcol-builder.yaml

    # Update version references in otelcol-builder Makefile and all other md files
    sed -i "s/${FROM_VER}/${TO_VER}/" otelcolbuilder/Makefile
    sed -i "s/\(collector\/\(blob\|tree\)\/v\)${FROM_VER}/\1${TO_VER}/" \
        README.md \
        docs/configuration.md \
        docs/migration.md \
        docs/performance.md
    sed -i "s/\(contrib\/\(blob\|tree\)\/v\)${FROM_VER}/\1${TO_VER}/" \
        README.md \
        docs/configuration.md \
        docs/migration.md \
        docs/performance.md \
        pkg/receiver/telegrafreceiver/README.md
done
