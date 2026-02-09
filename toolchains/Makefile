# Determine the absolute path of the toolchains directory that contains this
# Makefile.
TOOLCHAINS_DIR := $(realpath $(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

DOCKER_CROSS_BUILDER_NAME=golang-cross-builder
DOCKER_CROSS_BUILDER_VOLUME=golang-cross-builder

MUSL_X_MAKE_REPO=https://github.com/richfelker/musl-cross-make.git
MUSL_X_MAKE_REF=e5147dde912478dd32ad42a25003e82d4f5733aa

.PHONY: print-toolchains-dir
print-toolchains-dir:
	@echo $(TOOLCHAINS_DIR)

.PHONY: print-docker-cross-builder-name
print-docker-cross-builder-name:
	@echo $(DOCKER_CROSS_BUILDER_NAME)

.PHONY: print-docker-cross-builder-volume
print-docker-cross-builder-volume:
	@echo $(DOCKER_CROSS_BUILDER_VOLUME)

musl-cross-make:
	git clone $(MUSL_X_MAKE_REPO);
	cd musl-cross-make && git checkout $(MUSL_X_MAKE_REF);
	cp $(TOOLCHAINS_DIR)/config.mak musl-cross-make/config.mak;

.PHONY: toolchain-linux_arm64
toolchain-linux_arm64: OUTPUT=$(TOOLCHAINS_DIR)/toolchain_arm64
toolchain-linux_arm64: musl-cross-make
	TARGET=aarch64-linux-musl $(MAKE) -C musl-cross-make install

.PHONY: toolchain-linux_amd64
toolchain-linux_amd64: OUTPUT=$(TOOLCHAINS_DIR)/toolchain_amd64
toolchain-linux_amd64: musl-cross-make
	TARGET=x86_64-linux-musl $(MAKE) -C musl-cross-make install

.PHONY: docker-golang-cross-builder
docker-golang-cross-builder:
	docker volume create $(DOCKER_CROSS_BUILDER_VOLUME)
	docker build \
		-t $(DOCKER_CROSS_BUILDER_NAME) \
		-f $(TOOLCHAINS_DIR)/Dockerfile \
		$(TOOLCHAINS_DIR)
