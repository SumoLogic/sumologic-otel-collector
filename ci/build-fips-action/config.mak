# https://github.com/richfelker/musl-cross-make configuration
# targets fe915821b652a7fa37b34a596f47d8e20bc72338
CONFIG_SUB_REV = 3d5db9ebe860
BINUTILS_VER = 2.33.1
GCC_VER = 9.4.0
MUSL_VER = 1.2.3

COMMON_CONFIG += CFLAGS="-g0 -Os" CXXFLAGS="-g0 -Os" LDFLAGS="-s"
