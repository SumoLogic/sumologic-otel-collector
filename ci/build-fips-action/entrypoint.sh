#!/usr/bin/env sh

PLATFORM=$1

git config --global --add safe.directory /github/workspace

# Install builder
cd otelcolbuilder || exit 1
mkdir "${HOME}/bin"
export PATH="${HOME}/bin:${PATH}"
make install-builder

# Detect if cross CC needs set
case $PLATFORM in
'linux_amd64')
        ARCH='x86_64';
        ;;
'linux_arm64')
        ARCH='aarch64';
        ;;
esac

CC="/opt/${ARCH}-linux-musl/bin/${ARCH}-linux-musl-gcc"

if [ ! -f "$CC" ]; then
    echo "$CC not found.";
	exit 1;
fi

# Build otelcol-sumo
make otelcol-sumo-"${PLATFORM}" \
		FIPS_SUFFIX="-fips" \
		CGO_ENABLED="1" \
		CC="$CC" \
		EXTRA_LDFLAGS="-linkmode external -extldflags '-static'"
