# Our C Toolchain

By virtue of being pure go, our non-fips linux executables are currently
portable across their target platform (i.e. linux/amd64). Because of this our
build infrastructure is limited, largely focused on producing a single portable
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

##  Why not Alpine Linux + alpine-sdk?

This would be ideal. Unfortunately the limitation of our build infrastructure
(the x86_64-only Github hosted runners), combined with the need to build an
aarch64 executable and the size and complexity of the build that means it takes
~2 hours under emulation means that we need a cross compiler toolchain for arm.
Since alpine does not build such a toolchain, we have opted to build it ourself
until we can provide a native build environment for our aarch64 builds.

