# Vagrant

## Prerequisites

Please install the following:

- [VirtualBox](https://www.oracle.com/virtualization/technologies/vm/downloads/virtualbox-downloads.html)
- [Vagrant](https://www.vagrantup.com/)
- [vagrant-disksize](https://github.com/sprotheroe/vagrant-disksize) plugin

### MacOS

```bash
brew cask install virtualbox
brew cask install vagrant
vagrant plugin install vagrant-disksize
```

## Setting up

You can set up the Vagrant environment with just one command:

```bash
make vagrant-up
```

After successfull installation you can ssh to the virtual machine with:

```bash
make vagrant-ssh
```

NOTICE: The directory with sumologic-otel-collector repository on the host is synced with `/sumologic/` directory on the virtual machine.

## Build

In order to build `otelcol-sumo` binary, simply execute the following command:

```bash
make build
```

If you see the following or similar error:

```text
vagrant@sumologic-otel-collector:/sumologic$ make build
make[1]: Entering directory '/sumologic/otelcolbuilder'
/bin/sh: 1: ocb: not found
Makefile:41: *** Installed ocb version "" does not match the requested "0.53.0" Please check if "/home/vagrant/bin" can be found in your PATH and if not, then install it using 'make install-ocb' from otelcolbuilder's directory .  Stop.
make[1]: Leaving directory '/sumologic/otelcolbuilder'
make: *** [Makefile:177: build] Error 2
```

run the following command:

```bash
make install-ocb build
```

The `otelcol-sumo` binary is going to be placed at `/sumologic/otelcolbuilder/cmd/otelcol-sumo`.

## Build docker image

In order to build docker image, run the following command:

```bash
make build-container-local
```

## Useful targets

Here is the list of useful targets for [make](../Makefile):

- `install-staticcheck` - to install staticcheck
- `gotest` - to run unittests
- `golint` - to run go code linter
