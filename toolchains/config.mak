# https://github.com/richfelker/musl-cross-make configuration
# targets e5147dde912478dd32ad42a25003e82d4f5733aa
CONFIG_SUB_REV = 3d5db9ebe860
BINUTILS_VER = 2.44
GCC_VER = 15.1.0
MUSL_VER = 1.2.5

COMMON_CONFIG += CFLAGS="-g0 -Os" CXXFLAGS="-g0 -Os" LDFLAGS="-s"
# build c toolchain only
GCC_CONFIG += --enable-languages=c
# add -nv to download command for tidy CI output
DL_CMD = wget -nv -c -O

# GNU_SITE defaults to https://ftpmirror.gnu.org/gnu which will redirect to
# another mirror. At the time of this commit, one of the mirrors that is
# frequently being picked has an expired certificate. Remove or comment out the
# following line once https://mirror.us-midwest-1.nexcess.net has a valid
# certificate.
GNU_SITE = https://mirror.team-cymru.com/gnu
