MUSL_X_MAKE_REPO=https://github.com/richfelker/musl-cross-make.git
MUSL_X_MAKE_REF=fe915821b652a7fa37b34a596f47d8e20bc72338

.PHONY: toolchain-linux_arm64
toolchain-linux_arm64: musl-cross-make
	TARGET=aarch64-linux-musl $(MAKE) -C musl-cross-make install

.PHONY: toolchain-linux_amd64
toolchain-linux_amd64: musl-cross-make
	TARGET=x86_64-linux-musl $(MAKE) -C musl-cross-make install

musl-cross-make:
	git clone $(MUSL_X_MAKE_REPO);
	cd musl-cross-make && git checkout $(MUSL_X_MAKE_REF);
	cp config.mak ./musl-cross-make/config.mak;
