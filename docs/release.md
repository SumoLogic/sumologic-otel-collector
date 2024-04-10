# Releasing

- [How to release](#how-to-release)
  - [Check end-to-end tests](#check-end-to-end-tests)
  - [Update Changelog and upgrading guide](#update-changelog-and-upgrading-guide)
  - [Create and push Git tag](#create-and-push-git-tag)
    - [Remove tag in case of a failed release job](#remove-tag-in-case-of-a-failed-release-job)
  - [Publish GitHub release](#publish-github-release)
  - [Build packages](#build-packages)
- [Updating OT to next version](#updating-ot-to-next-version)
  - [Update OT version](#update-ot-version)
  - [Fix lint errors and tests](#fix-lint-errors-and-tests)
  - [Check out the developer changelog](#check-out-the-developer-changelog)
  - [Add missing upstream components](#add-missing-upstream-components)
    - [Adding components from scratch](#adding-components-from-scratch)
- [Running Tracing E2E tests](#running-tracing-e2e-tests)
- [Updating journalctl](#updating-journalctl)

## How to release

### Check end-to-end tests

Check if the Sumo internal e2e tests are passing.

### Update Changelog and upgrading guide

Update the [CHANGELOG.md][changelog] by running `make update-changelog VERSION=x.x.x-sumo-x`.

Make sure the [upgrading.md][upgrading] file has entries for the breaking changes for this release.

### Create and push Git tag

In order to release a new version of Sumo OT distro you'd export `TAG` env variable
and create a tag and push it.

This can be done using `add-tag` and `push-tag` `make` targets which will handle
that for you as well as pushing tags for all the plugins in this repo so that
they can be imported from other repositories.

```shell
export TAG=v0.75.0-sumo-0
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

#### Remove tag in case of a failed release job

Pushing a new version tag to GitHub starts the [release build](../.github/workflows/release_builds.yml) jobs.

If one of these jobs fails for whatever reason (real world example: failing to notarize the MacOS binary),
you might need to remove the created tags, perhaps change something, and create the tags again.

To delete the tags both locally and remotely, run the following commands:

```shell
export TAG=v0.75.0-sumo-0
make delete-tag delete-remote-tag
```

### Publish GitHub release

The GitHub release is created as draft by the [create-release](../.github/workflows/release_builds.yml) GitHub Action.

After the release draft is created, go to [GitHub releases](https://github.com/SumoLogic/sumologic-otel-collector/releases),
edit the release draft and fill in missing information:

- Specify versions for upstream OT core and contrib releases
- Copy and paste the Changelog entry for this release from [CHANGELOG.md][changelog]

After verifying that the release text and all links are good, publish the release.

### Build packages

1. Go to the releases section of the [sumologic-otel-collector-packaging repository][package_releases].

2. Find the draft release for the new version. After verifying that the release text and all links are good, publish the release.

## Updating OT to next version

Here are the steps to update OT to next version:

1. Update the version number where necessary
1. Verify that Sumo OT distro builds correctly
1. Fix lint and test errors
1. Add missing upstream components
1. Ensure all [milestone] tasks and issues have been done

[milestone]: https://github.com/SumoLogic/sumologic-otel-collector/milestones

### Update OT version

Take note of the [core][otcore_releases] and [contrib][otcontrib_releases] versions you want to update to.
They are often the same version (e.g. v0.77.0),
but for some releases, these versions are different (e.g. v0.76.1/v0.76.3 or v0.78.2/v0.78.0).

Run:

```shell
make update-ot OT_CORE_NEW=x.x.x OT_CONTRIB_NEW=y.y.y
```

This make target also builds the Sumo OT distro binary to check that there are no build errors.

### Fix lint errors and tests

Run:

```shell
make golint
make gotest
```

### Check out the developer changelog

Contrib maintains a [separate changelog][otelcol_api_changelog] for distribution maintainers and component authors.
Check this for any changes that may affect our components or distribution. One example of this is increases
to the minimal supported Go version.

### Add missing upstream components

We include all of the components from the following list which are at least in *alpha* stability level:

- [OpenTelemetry Collector][otelcol_components] extensions, receivers, processors, connectors
- [OpenTelemetry Collector Contrib][otelcol_contrib_components] extensions, receivers, processors, connectors

Additionally, the following components are supported:

- [logstransformprocessor]
- [windowseventlogreceiver]
- [db_storage][dbstorage]
- [docker_observer][dockerobserver]
- [ecs_observer][ecsobserver]
- [ecs_task_observer][ecstaskobserver]

As a fourth step, please check [OpenTelemetry Collector][otcore_releases] and [OpenTelemetry Collector Contrib][otcontrib_releases]
release pages for new components and update [builder configuration][builder_config] and [README.md] if they are any.
New exporters should not be added without a reason.
Please consider example pull request: [#604]

#### Adding components from scratch

This shouldn't be required as long as list of components is regularly updated,
but in case you want to generate full list of components, the following instruction can be helpful:

1. [update builder configuration][builder_config]
   You can use the following snippet inside [OpenTelemetry Contrib repository][otc_repository]
   in order to get list of components:

   ```bash
   export TAG=vx.y.z
   git fetch --all
   git checkout "${TAG}"
   for dir in receiver processor extension; do
     echo "###############${dir}s###############"
     for file in $(ls "${dir}"); do
       echo "  - gomod: \"github.com/open-telemetry/opentelemetry-collector-contrib/${dir}/${file} ${TAG}\"";
     done;
   done;
   ```

1. Run `make build` in order to generate updated components file (`otelcolbuilder/cmd/components.go`)
1. Update markdown table:

   1. Prepare `local/receiver.txt`, `local/exporter.txt`, `local/extension.txt` and `local/processor.txt`
      in [OpenTelemetry Contrib repository][otc_repository] based on components file.
      Example content of `local/extension.txt`:

      ```text
      sumologicextension.NewFactory(),
      ballastextension.NewFactory(),
      zpagesextension.NewFactory(),
      asapauthextension.NewFactory(),
      awsproxy.NewFactory(),
      basicauthextension.NewFactory(),
      bearertokenauthextension.NewFactory(),
      fluentbitextension.NewFactory(),
      healthcheckextension.NewFactory(),
      httpforwarderextension.NewFactory(),
      jaegerremotesampling.NewFactory(),
      oauth2clientauthextension.NewFactory(),
      dockerobserver.NewFactory(),
      ecsobserver.NewFactory(),
      ecstaskobserver.NewFactory(),
      hostobserver.NewFactory(),
      k8sobserver.NewFactory(),
      oidcauthextension.NewFactory(),
      pprofextension.NewFactory(),
      sigv4authextension.NewFactory(),
      filestorage.NewFactory(),
      dbstorage.NewFactory(),
      ```

   1. Run the following snippet in order to prepare markdown links for [README.md]:

      ```bash
      for kind in extension processor receiver exporter; do
        echo "#####${kind}#####"
        for component in $(cat local/${kind}.txt | sed "s/\..*//g"); do
          export dir=$(find ${kind} -name "${component}");
          NAME=$(grep -oRP '(?<=typeStr\s)(.*?)=(.*?)"(.*?)"' "${dir}" | grep -oP '\w+"$' | sed 's/"//g');
          echo "[${NAME}][${component}]";
        done 2>/dev/null
      done
      ```

   1. Copy and fix the output
   1. Copy fixed output to spreadsheet. Every component in separate column as in [README.md].
   1. Sort components and add placeholder to empty rows.
      Placeholders make it easier to format table as markdown in next steps.
   1. Copy table into text editor with support for regexp replacement.
   1. Prepare table in markdown format. We recommend to use regexp replacement.
   1. Prepare links for the table. You can use regexp replacement in order to generate it
   1. Carefully analyse `git diff` and fix unwanted changes.

## Running Tracing E2E tests

We currently have some legacy E2E tests ported from [our OT fork][ot_fork], which serve as a means of
verifying feature parity for tracing as we migrate it to this distribution. The tests
are run by CircleCI on demand for `main` and release branches, and are defined [here][tracing_tests].

In order to run the tests, go to the [CircleCI page][circleci], choose the branch you want, and manually
approve the workflow to run. Note that you need commiter rights in this repository to run the tests.

![Approving the workflow in CircleCI][circleci_approve]

## Updating journalctl

Journalctl is dependency for [journaldreceiver]. In order to update it, please update debian version in Dockerfile files.
Please see available [debian versions][debian_versions].

```bash
export DEBIAN_VERSION=11.3
make update-journalctl
```

[builder_config]: ../otelcolbuilder/.otelcol-builder.yaml
[release_job]: ../.github/workflows/release_builds.yml
[ot_fork]: https://github.com/SumoLogic/opentelemetry-collector-contrib
[tracing_tests]: ../.circleci/config.yml
[circleci]: https://app.circleci.com/pipelines/github/SumoLogic/sumologic-otel-collector
[circleci_approve]: ../images/circleci_approve_workflow.png
[changelog]: ../CHANGELOG.md
[upgrading]: ./upgrading.md
[journaldreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/receiver/journaldreceiver
[debian_versions]: https://hub.docker.com/_/debian/?tab=description
[otelcol_components]: https://github.com/open-telemetry/opentelemetry-collector-releases/blob/main/distributions/otelcol/manifest.yaml
[otelcol_contrib_components]: https://github.com/open-telemetry/opentelemetry-collector-releases/blob/main/distributions/otelcol-contrib/manifest.yaml
[otc_repository]: https://github.com/open-telemetry/opentelemetry-collector-contrib
[otelcol_api_changelog]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/CHANGELOG-API.md
[readme.md]: ../README.md
[#604]: https://github.com/SumoLogic/sumologic-otel-collector/pull/604/files
[otcontrib_releases]: https://github.com/open-telemetry/opentelemetry-collector-contrib/releases
[otcore_releases]: https://github.com/open-telemetry/opentelemetry-collector/releases
[windowseventlogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/receiver/windowseventlogreceiver
[logstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/processor/logstransformprocessor
[dbstorage]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/extension/storage/dbstorage
[dockerobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/extension/observer/dockerobserver
[ecsobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/extension/observer/ecsobserver
[ecstaskobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/extension/observer/ecstaskobserver
[package_releases]: https://github.com/SumoLogic/sumologic-otel-collector-packaging/releases
