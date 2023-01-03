# Installation script tests

This directory contains the tests for the [installation script](../../scripts/install.sh).

Run them with:

```bash
cd ./pkg/scripts_test
make test
```

These tests use `sudo` to run with root permissions,
because they actually install and uninstall the collector on the machine.
This is why you may need to enter your sudo password, depending on your account settings.

The tests take a while to complete (up to a couple minutes) and there's no console output while they are running, so be patient.

These tests are run in CI (GitHub Actions) as part of the global `make gotest` target,
as this is just another Go module in the `pkg/` directory.
This works because the GitHub Actions runners are configured to not require a password for `sudo`,
as described in the [docs](https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners#administrative-privileges):

> The Linux and macOS virtual machines both run using passwordless `sudo`.
> When you need to execute commands or install tools that require more privileges than the current user,
> you can use `sudo` without needing to provide a password.
