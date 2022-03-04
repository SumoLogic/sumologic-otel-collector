# Developer guide

- [How to release](#how-to-release)
- [Updating OT core](#updating-ot-core)

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

1. Updating the version number where necessary
1. Verifying that Sumo OT distro builds correctly
1. Fixing lint errors from deprecations

The first two steps of this list are covered by the `update-ot-core` Makefile target. Run:

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

[builder_config]: ../otelcolbuilder/.otelcol-builder.yaml
[release_job]: ../.github/workflows/release_builds.yml
