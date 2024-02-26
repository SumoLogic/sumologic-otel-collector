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

## Building the application

After starting `bash` as per the previous section, you can use the same command as on Unix. So:

```bash
make install-builder
make build
```

to build the binary.
