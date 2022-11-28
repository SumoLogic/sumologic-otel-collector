#!/bin/bash

set -eo pipefail

# Detects the latest version in git tags and converts it to a Windows
# ProductVersion.
#
# https://learn.microsoft.com/en-us/windows/win32/msi/productversion
# MAJOR.MINOR.BUILD.INTERNAL

declare -i major_version
declare -i minor_version
declare -i build_version
declare -i internal_version
#declare channel
declare ot_channel
declare -i ot_channel_version
#declare sumo_channel
declare -i sumo_channel_version

version_tag="${VERSION_TAG:-$(git tag -l --sort -version:refname | head -n 1)}"
version_regex="^v([0-9]+).([0-9]+).([0-9]+)((-(alpha|beta|rc|sumo)[-.]([0-9]+))(-(alpha|beta|rc).([0-9])+)?)?$"

if [[ $version_tag =~ $version_regex ]]; then
    major_version="${BASH_REMATCH[1]}"
    minor_version="${BASH_REMATCH[2]}"
    build_version="${BASH_REMATCH[3]}"
    ot_channel="${BASH_REMATCH[6]}"
    ot_channel_version="${BASH_REMATCH[7]}"
    sumo_channel="${BASH_REMATCH[9]}"
    sumo_channel_version="${BASH_REMATCH[10]}"
else
    echo "Error: Tag can not be converted to a Windows ProductVersion: ${version_tag}" >&2
    exit 1
fi

if [[ $ot_channel == "sumo" ]]; then
    #channel="${sumo_channel}"
    if [[ $sumo_channel != "" ]]; then
        internal_version="${sumo_channel_version}"
    else
        internal_version="${ot_channel_version}"
    fi
elif [[ $ot_channel != "" ]]; then
    #channel="${ot_channel}"
    internal_version="${ot_channel_version}"
fi

if [[ $OVERRIDE_INTERNAL_VERSION != "" ]]; then
    number_regex='^[0-9]+$'
    if ! [[ $OVERRIDE_INTERNAL_VERSION =~ $number_regex ]]; then
        echo "Error: OVERRIDE_INTERNAL_VERSION is not a number" >&2
        exit 1
    fi

    internal_version="${OVERRIDE_INTERNAL_VERSION}"
fi

# Validation
if [[ $major_version -gt 255 ]]; then
    echo "Major version cannot be greater than 255"
    exit 1
fi

if [[ $minor_version -gt 255 ]]; then
    echo "Minor version cannot be greater than 255"
    exit 1
fi

if [[ $build_version -gt 65535 ]]; then
    echo "Build version cannot be greater than 65,535"
    exit 1
fi

if [[ $internal_version -gt 65535 ]]; then
    echo "Internal version cannot be greater than 65,535"
    exit 1
fi

echo "${major_version}.${minor_version}.${build_version}.${internal_version}"
