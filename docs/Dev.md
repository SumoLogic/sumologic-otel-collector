# Developer guide

- [Setting up Go workspaces](#setting-up-go-workspaces)
- [How to release](#how-to-release)
- [Updating OT core](#updating-ot-core)
  - [Updating patched processors](#updating-patched-processors)
  - [Updating OT distro](#updating-ot-distro)
- [Running Tracing E2E tests](#running-tracing-e2e-tests)

## Setting up Go workspaces

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

## How to release

In order to release a new version of Sumo OT distro you'd export `TAG` env variable
and create a tag and push it.

This can be done using `add-tag` and `push-tag` `make` targets which will handle
that for you as well as pushing tags for all the plugins in this repo so that
they can be imported from other repositories.

```shell
export TAG=v0.0.1
make add-tag push-tag
```

> **NOTE**:
>
> [Release build CI job][release_job]
> is additionally using the `prepare-tag` `make` target in order to substitute
> the version placeholders in [opentelemetry collector builder config][builder_config]
> so that the versions in the released binary match the repository tag.
>
> If you'd like to build a binary to mimic the release binary you'd have to run
> that yourself like so:
>
> ```shell
> make prepare-tag TAG=$(git describe --tags --abbrev=10 --match "v[0-9]*")
> ```

## Updating OT core

Updating OT core involves:

1. Rebasing our upstream processor patches on the new core version
1. Updating the version number where necessary
1. Verifying that Sumo OT distro builds correctly
1. Fixing lint errors from deprecations

### Updating patched processors

We currently maintain patches for three upstream processors: `resourceprocessor`, `attributesprocessor` and `filterprocessor`.
The patches live in our [contrib fork repository][contrib_fork], on the `vX.X.X-filterprocessor` branch. See [comments][builder_config]
in the builder configuration for more details.

To update this patchset for the new OT core version:

1. Checkout the contrib fork repo, add upstream as a remote, and pull the new version tag.

   ```bash
   export CURRENT_VERSION=vX.X.X
   export NEW_VERSION=vY.Y.Y
   git clone https://github.com/SumoLogic/opentelemetry-collector-contrib && cd opentelemetry-collector-contrib
   git remote add upstream https://github.com/open-telemetry/opentelemetry-collector-contrib
   git pull upstream "${NEW_VERSION}" "${CURRENT_VERSION}"
   ```

1. Create a new branch for the patchset and rebase it on the new version

   ```bash
   git switch "${CURRENT_VERSION}-filterprocessor"
   git checkout -b "${NEW_VERSION}-filterprocessor"
   git rebase -i --onto "${NEW_VERSION}" "${CURRENT_VERSION}" "${NEW_VERSION}-filterprocessor"
   ```

1. Resolve conflicts and make sure tests and linters pass afterwards.
   You can run them by invoking the following in the project root:

   ```bash
   make install-tools
   make golint
   make gotest
   ```

   If the command `make gotest` fails on unrelated tests, like for example `kafkareceiver`,
   only run the tests for the changed modules:

   ```bash
   make -C internal/coreinternal test && make -C processor/attributesprocessor test && make -C processor/filterprocessor test && make -C processor/resourceprocessor test
   ```

1. Push the new branch to the fork repo and write down the commit SHA

   ```bash
   git push origin "${NEW_VERSION}-filterprocessor"
   ```

1. Update the [builder configuration][builder_config] with the new commit SHA

### Updating OT distro

The second and third step of this list are covered by the `update-ot-core` Makefile target. Run:

```shell
make update-ot-core OT_CORE_NEW_VERSION=x.x.x
```

to carry these steps out.
Afterwards, you can run tests and check for lint errors by running:

```shell
make golint
make gotest
```

in the repository root.

## Running Tracing E2E tests

We currently have some legacy E2E tests ported from [our OT fork][ot_fork], which serve as a means of
verifying feature parity for tracing as we migrate it to this distribution. The tests
are run by CircleCI on demand for `main` and release branches, and are defined [here][tracing_tests].

In order to run the tests, go to the [CircleCI page][circleci], choose the branch you want, and manually
approve the workflow to run. Note that you need commiter rights in this repository to run the tests.

![Approving the workflow in CircleCI][circleci_approve]

[builder_config]: ../otelcolbuilder/.otelcol-builder.yaml
[release_job]: ../.github/workflows/release_builds.yml
[ot_fork]: https://github.com/SumoLogic/opentelemetry-collector-contrib
[tracing_tests]: ../.circleci/config.yml
[circleci]: https://app.circleci.com/pipelines/github/SumoLogic/sumologic-otel-collector
[circleci_approve]: ../images/circleci_approve_workflow.png
[contrib_fork]: https://github.com/SumoLogic/opentelemetry-collector-contrib
