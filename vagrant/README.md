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
vagrant up
```

After successfull installation you can ssh to the virtual machine with:

```bash
vagrant ssh
```

NOTICE: The directory with sumologic-otel-collector repository on the host is synced with `/sumologic/` directory on the virtual machine.
