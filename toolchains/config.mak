# https://github.com/richfelker/musl-cross-make configuration
# targets 2688cf868f1cdfff36ac70ac01f95b55e3bf508f
CONFIG_SUB_REV = 3d5db9ebe860
BINUTILS_VER = 2.36.0
GCC_VER = 11.4.0
MUSL_VER = 1.2.4

COMMON_CONFIG += CFLAGS="-g0 -Os" CXXFLAGS="-g0 -Os" LDFLAGS="-s"
# build c toolchain only
GCC_CONFIG += --enable-languages=c
# add -nv to download command for tidy CI output
DL_CMD = wget -nv -c -O

GNU_SITE = https://ftp.gnu.org/gnu
