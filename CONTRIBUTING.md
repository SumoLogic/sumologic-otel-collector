# Contributing Guide

- [Setting up development environment](#setting-up-development-environment)
  - [How to build](#how-to-build)
  - [Running Tests](#running-tests)
  - [Setting up Go workspaces](#setting-up-go-workspaces)

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

1. Install otelcolbuilder:

  ```bash
  make install-builder
  ```

### How to build

```bash
$ cd otelcolbuilder && make build
opentelemetry-collector-builder \
                --config .otelcol-builder.yaml \
                --output-path ./cmd \
                --name otelcol-sumo
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
opentelemetry-collector-builder \
                --config .otelcol-builder.yaml \
                --output-path ./cmd \
                --name otelcol-sumo-linux_arm64
2021-05-24T16:32:11.963+0200    INFO    cmd/root.go:99  OpenTelemetry Collector distribution builder    {"version": "dev", "date": "unknown"}
2021-05-24T16:32:11.965+0200    INFO    builder/main.go:90      Sources created {"path": "./cmd"}
2021-05-24T16:32:12.066+0200    INFO    builder/main.go:126     Getting go modules
2021-05-24T16:32:12.376+0200    INFO    builder/main.go:107     Compiling
2021-05-24T16:32:37.326+0200    INFO    builder/main.go:113     Compiled        {"binary": "./cmd/otelcol-sumo-linux_arm64"}
```

### Running Tests

In order to run tests run `make gotest` in root directory of this repository.
This will run tests in every module from this repo by running `make test` in its
directory.

### Setting up Go workspaces

This repository contains multiple Go packages with their own dependencies. Some IDEs
(VS Code for example) do not like this kind of setup and demand that you work on each
package in a separate workspace. As of [Go 1.18](https://tip.golang.org/doc/go1.18#go-work)
this can be solved by configuring a single Go workspace covering all the packages.
This can be done by adding a `go.work` file to the repository root:

```go
go 1.18

use (
        ./otelcolbuilder/cmd
        ./pkg/test
        ./pkg/exporter/sumologicexporter
        ./pkg/extension/sumologicextension
        ./pkg/processor/cascadingfilterprocessor
        ./pkg/processor/k8sprocessor
        ./pkg/processor/metricfrequencyprocessor
        ./pkg/processor/sourceprocessor
        ./pkg/processor/sumologicschemaprocessor
        ./pkg/processor/sumologicsyslogprocessor
        ./pkg/receiver/telegrafreceiver
)
```

This will also cause Go to generate a `go.work.sum` file to match.

To contribute you will need to ensure you have the following setup:

- working Go environment
- installed `opentelemetry-collector-builder`

  `opentelemetry-collector-builder` can be installed using following command:

  ```bash
  make -C otelcolbuilder install-builder
  ```

  Which will by default install the builder binary in `${HOME}/bin/opentelemetry-collector-builder`.
  You can customize it by providing the `BUILDER_BIN_PATH` argument.

  ```bash
  make -C otelcolbuilder install-builder \
    BUILDER_BIN_PATH=/custom/dir/bin/opentelemetry-collector-builder
  ```

[go-msi]: https://go.dev/dl/
[make-for-windows]: https://gnuwin32.sourceforge.net/downlinks/make.php
[windows-terminal]: https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701
[git-for-windows]: https://git-scm.com/download/win
