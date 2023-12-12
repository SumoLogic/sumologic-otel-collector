# Linux CGO Toolchain

This directory contains tools for building two C toolchains, aarch64-linux-musl
and x86_64-linux-musl. These toolchains are currently used in our linux builds
that require CGO (our FIPS builds) in order to produce statically linked and
portable executables.

## Background

By virtue of being pure go, our non-fips linux executables are portable across
their target platform (i.e. linux/amd64). Because of this our build
infrastructure is limited, largely focused on producing a single portable
statically linked binary per target.

Providing a similarly portable FIPS-capable linux executable presents some
challenges. These executables have a required CGO dependency (boringcrypto),
and in the past we have produced dynamically linked executables with a hidden
system dependency on glibc. Rather than expanding our build infrastructure to
produce distribution specific executables or building on an outdated
distribution to target a minimal glibc version, we have opted to produce static
executables. This is achieved by using the musl C library instead of the
standard glibc implementation (which does not allow for reliable static
linking.)

## Why not just Alpine Linux?

This would be ideal. Unfortunately the combination of our limited build
infrastructure (x86_64-only Github hosted runners) and the size and scale of
our build executable (emulation with qemu takes ~1.5 hours!) means that we are
forced to cross compile our aarch64 executables. Since alpine does not build a
suitable cross toolchain, we have opted to build it ourself until we can
provide a native build environment for our aarch64 builds.

## Building

This is a small warpper around
[musl-cross-make](https://github.com/richfelker/musl-cross-make). For more
detailed information see the documentation there. By default these targets will
build a toolchain under `./musl-cross-make/output`. To change the destination
use the `OUTPUT` variable.

```
# Example invocations
make toolchain-linux_arm64 OUTPUT=/opt/aarch64-toolchain -j8
make toolchain-linux_amd64 OUTPUT=/opt/x86_64-toolchain -j8
```

In order to build the toolchain you'll need `git`, `wget`, `xz` an existing gcc
toolchain and some other build essentials. This should be enough on ubuntu:
`apt-get install -y git wget xz-utils build-essential`.
