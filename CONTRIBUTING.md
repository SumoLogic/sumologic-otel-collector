# Contributing Guide

- [Setting up development environment](#setting-up-development-environment)
  - [Preparing the development environment (Windows)](#preparing-the-development-environment-windows)
  - [How to build](#how-to-build)
  - [How to build packages (Windows MSI)](#how-to-build-packages-windows-msi)
  - [Running Tests](#running-tests)
  - [Setting up Go workspaces](#setting-up-go-workspaces)
- [Changelog management](#changelog-management)
  - [Installing Towncrier](#installing-towncrier)
  - [Adding a changelog entry](#adding-a-changelog-entry)
  - [My change doesn't need a changelog entry](#my-change-doesnt-need-a-changelog-entry)
  - [How do I update the changelog while releasing?](#how-do-i-update-the-changelog-while-releasing)
  - [How do I add multiple entries for the same PR?](#how-do-i-add-multiple-entries-for-the-same-pr)
  - [How do I add an entry with multiple PR links?](#how-do-i-add-an-entry-with-multiple-pr-links)
- [Releasing](./docs/release.md)

---

## Setting up development environment

### Preparing the development environment (Windows)

Several packages & environment variables are necessary to build in a Windows
environment:

1. Download & install the latest [Go MSI][go-msi].

1. Download & install [Make for Windows][make-for-windows].

1. Open `cmd` and add make to the user PATH:

  ```bat
  setx path "%PATH%;C:\Program Files (x86)\GnuWin32\bin"
  ```

1. Download & install [64-bit Git for Windows Setup][git-for-windows].

1. Open gitbash.

1. Navigate to the otelcolbuilder directory:

  ```bash
  cd sumologic-otel-collector/otelcolbuilder
  ```

1. Install ocb:

  ```bash
  make install-ocb
  ```

### How to build

```bash
$ cd otelcolbuilder && make build
Installed ocb (/Users/sumo/bin/ocb) is at the correct version: 0.124.0
Building otelcol-sumo version: 0.124.1
CGO_ENABLED=1 /Users/sumo/bin/ocb \
                --config .otelcol-builder.yaml \
                --skip-compilation=true
2021-05-24T16:29:03.494+0200    INFO    cmd/root.go:99  OpenTelemetry Collector distribution builder    {"version": "dev", "date": "unknown"}
2021-05-24T16:29:03.498+0200    INFO    builder/main.go:90      Sources created {"path": "./cmd"}
2021-05-24T16:29:03.612+0200    INFO    builder/main.go:126     Getting go modules
2021-05-24T16:29:03.957+0200    INFO    builder/main.go:107     Compiling
2021-05-24T16:29:09.770+0200    INFO    builder/main.go:113     Compiled        {"binary": "./cmd/otelcol-sumo"}
```

In order to build for a different platform one can use `otelcol-sumo-${platform}_${arch}`
make targets e.g.:

```bash
$ cd otelcolbuilder && make otelcol-sumo-linux_arm64
GOOS=linux   GOARCH=arm64 /Library/Developer/CommandLineTools/usr/bin/make build BINARY_NAME=otelcol-sumo-linux_arm64
Installed ocb (/Users/sumo/bin/ocb) is at the correct version: 0.124.0
Building otelcol-sumo version: 0.124.1
CGO_ENABLED=1 /Users/sumo/bin/ocb \
                --config .otelcol-builder.yaml \
                --skip-compilation=true
2021-05-24T16:32:11.963+0200    INFO    cmd/root.go:99  OpenTelemetry Collector distribution builder    {"version": "dev", "date": "unknown"}
2021-05-24T16:32:11.965+0200    INFO    builder/main.go:90      Sources created {"path": "./cmd"}
2021-05-24T16:32:12.066+0200    INFO    builder/main.go:126     Getting go modules
2021-05-24T16:32:12.376+0200    INFO    builder/main.go:107     Compiling
2021-05-24T16:32:37.326+0200    INFO    builder/main.go:113     Compiled        {"binary": "./cmd/otelcol-sumo-linux_arm64"}
```

### How to build packages (Windows MSI)

1. Install Visual Studio 2019 or newer with the  `.NET desktop development`
   workload selected in the installer.

1. Open `Developer Command Prompt for VS 2019/2022`.

1. Navigate to the `packaging/msi/wix` directory.

1. Fetch project dependencies & build the MSI:

  ```
  msbuild.exe /p:Configuration=Release /p:Platform=x64 -Restore
  ```

1. The MSI package can be found in the `packaging/msi/wix/bin/x64/en-US`
   directory.

### Running Tests

In order to run tests run `make gotest` in root directory of this repository.
This will run tests in every module from this repo by running `make test` in its
directory.

### Setting up Go workspaces

This repository contains multiple Go modules with their own dependencies. Some IDEs
(VS Code for example) do not like this kind of setup and demand that you work on each
module in a separate workspace. As of [Go 1.19](https://tip.golang.org/doc/go1.19#go-work)
this can be solved by configuring a single Go workspace covering all the modules.
This can be done by adding a `go.work` file to the repository root:

```go
go 1.21

use (
        ./otelcolbuilder/cmd
        ./pkg/test
        ./pkg/exporter/sumologicexporter
        ./pkg/processor/cascadingfilterprocessor
        ./pkg/processor/k8sprocessor
        ./pkg/processor/metricfrequencyprocessor
        ./pkg/processor/sourceprocessor
        ./pkg/processor/sumologicsyslogprocessor
        ./pkg/receiver/telegrafreceiver
        ./pkg/configprovider/globprovider
        ./pkg/configprovider/opampprovider
        ./pkg/tools/udpdemux
)
```

This will also cause Go to generate a `go.work.sum` file to match.

To contribute you will need to ensure you have the following setup:

- working Go environment
- installed `ocb`

  `ocb` can be installed using following command:

  ```bash
  make -C otelcolbuilder install-ocb
  ```

  Which will by default install the ocb binary in `${HOME}/bin/ocb`.
  You can customize it by providing the `BIN_PATH` argument.

  ```bash
  make -C otelcolbuilder install-ocb \
    BIN_PATH=/custom/dir/bin
  ```

[go-msi]: https://go.dev/dl/
[make-for-windows]: https://gnuwin32.sourceforge.net/downlinks/make.php
[git-for-windows]: https://git-scm.com/download/win

## Changelog management

We use [Towncrier](https://towncrier.readthedocs.io) for changelog management. We keep the changelog entries for currently unreleased
changed in the [.changelog] directory. The contents of this directory are consumed when the changelog is updated prior to a release.

### Installing Towncrier

Towncrier is written in Python and can be installed with [pip](https://pypi.org/project/pip/).

Prerequisites:

- [Python v3](https://www.python.org/)
- [pip](https://pypi.org/project/pip/)

```shell
make install-towncrier
```

### Adding a changelog entry

If you want to add a changelog entry for your PR, run:

```bash
make add-changelog-entry
```

You can also just create the file manually. The filename format is `<PR NUMBER>.<CHANGE TYPE>(.<FRAGMENT NUMBER>).txt`, and the content is
the entry text.

### My change doesn't need a changelog entry

Add a `skip-changelog` label to your pull request in GitHub.

### How do I update the changelog while releasing?

Apart from Towncrier, you'll also need [Prettier](https://prettier.io/) for this.

Prerequisites:

- [Node.js](https://nodejs.org/)

```shell
make install-prettier
```

After you have Towncrier and Prettier available in your console, run:

```shell
make update-changelog VERSION=x.x.x-sumo-x
```

### How do I add multiple entries for the same PR?

Run `make add-changelog-entry` again. Another change file will be created.

### How do I add an entry with multiple PR links?

Just add an entry with the same text for each PR, they will be grouped together.
