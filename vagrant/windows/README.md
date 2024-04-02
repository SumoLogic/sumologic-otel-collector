# Vagrant Windows environments

These environments are intended for development and troubleshooting of Windows-specific functionality.

## Prerequisites

See [the following document](../README.md) on how to install Vagrant.

## Structure

There's a Vagrantfile provided for each supported Windows version. To use a particular image, just `cd` into the
directory of your choice.

## Setting up

You can set up the Vagrant environment with just one command:

```bash
vagrant up
```

After successfull installation you can ssh to the virtual machine with:

```bash
vagrant ssh
```

> [!NOTE]:
> The directory with sumologic-otel-collector repository on the host is synced with the `C:\sumologic\` directory on the virtual machine.

### Shell

The default shell is Powershell. Bash is also available, and required if you actually want to build Otelcol in the
Windows VM. You can start `bash` by running:

```powershell
bash
```

### ARM hosts (Apple M1, and so on)

You'll need to use QEMU instead of VirtualBox to use Vagrant on ARM. The following instructions will assume an M1 Mac as the host:

1. Install QEMU: `brew install qemu`
2. Install the QEMU vagrant provider: `vagrant plugin install vagrant-qemu`
3. Use the `windows-server-2022` directory
3. Provision the VM with the provider: `vagrant up --provider=qemu`

## Building the application

After starting `bash` as per the previous section, you can use the same command as on Unix. So:

```bash
make install-builder
make build
```

to build the binary.

## Building Windows containers

Use the [docker](docker/) machine, as it comes with Docker for Windows installed and configured. After building
the application, you can build a Windows container the following way:

```bash
cp otelcolbuilder/cmd/otelcol-sumo.exe .
make build-push-container-windows-dev PLATFORM=windows/amd64/ltsc2022
docker run public.ecr.aws/sumologic/sumologic-otel-collector-dev:latest-windows-amd64-ltsc2022 --version
```
