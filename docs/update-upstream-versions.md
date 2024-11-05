# Updating Upstream OT Version

- [Update OT version](#update-ot-version)
- [Fix lint errors and tests](#fix-lint-errors-and-tests)
- [Check out the developer changelog](#check-out-the-developer-changelog)
- [Add missing upstream components](#add-missing-upstream-components)
- [Adding components from scratch](#adding-components-from-scratch)

## Steps

Here are the steps to update OT to next version:

1. Update the version number where necessary
1. Verify that Sumo OT distro builds correctly
1. Fix lint and test errors
1. Add missing upstream components
1. Ensure all [milestone] tasks and issues have been done

## Update OT version

Take note of the [core][otcore_releases] and [contrib][otcontrib_releases]
versions you want to update to. They are often the same version (e.g. v0.77.0),
but for some releases, these versions are different (e.g. v0.76.1/v0.76.3 or
v0.78.2/v0.78.0).

Run:

```shell
make update-ot OT_CORE_NEW=x.x.x OT_CONTRIB_NEW=y.y.y
```

This make target also builds the Sumo OT distro binary to check that there are
no build errors.

### Fix lint errors and tests

Run:

```shell
make golint
make gotest
```

### Check out the developer changelog

Contrib maintains a [separate changelog][otelcol_api_changelog] for distribution
maintainers and component authors. Check this for any changes that may affect
our components or distribution. One example of this is increases to the minimal
supported Go version.

### Add missing upstream components

We include all of the components from the following list which are at least in
*alpha* stability level:

- [OpenTelemetry Collector][otelcol_components] extensions, receivers,
  processors, connectors
- [OpenTelemetry Collector Contrib][otelcol_contrib_components] extensions,
  receivers, processors, connectors

Additionally, the following components are supported:

- [logstransformprocessor]
- [windowseventlogreceiver]
- [db_storage][dbstorage]
- [docker_observer][dockerobserver]
- [ecs_observer][ecsobserver]
- [ecs_task_observer][ecstaskobserver]

As a fourth step, please check [OpenTelemetry Collector][otcore_releases] and
[OpenTelemetry Collector Contrib][otcontrib_releases] release pages for new
components and update [builder configuration][builder_config] and [README.md] if
they are any.

New exporters should not be added without a reason. Please consider example pull
request: [#604]

#### Adding components from scratch

This shouldn't be required as long as list of components is regularly updated,
but in case you want to generate full list of components, the following
instruction can be helpful:

1. [Update builder configuration][builder_config]

   You can use the following snippet inside [OpenTelemetry Contrib repository][otc_repository]
   in order to get list of components:

   ```bash
   export TAG=vx.y.z
   git fetch --all
   git checkout "${TAG}"
   contrib_repo="github.com/open-telemetry/opentelemetry-collector-contrib"
   for dir in receiver processor extension; do
     echo "###############${dir}s###############"
     for file in $(ls "${dir}"); do
       echo "  - gomod: \"${contrib_repo}/${dir}/${file} ${TAG}\"";
     done;
   done;
   ```

1. Run `make build` in order to generate updated components file (`otelcolbuilder/cmd/components.go`)
1. Update markdown table:

   1. Prepare `local/receiver.txt`, `local/exporter.txt`, `local/extension.txt` and
   `local/processor.txt`
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

[#604]: https://github.com/SumoLogic/sumologic-otel-collector/pull/604/files
[builder_config]: ../otelcolbuilder/.otelcol-builder.yaml
[milestone]: https://github.com/SumoLogic/sumologic-otel-collector/milestones
[otcontrib_releases]: https://github.com/open-telemetry/opentelemetry-collector-contrib/releases
[otcore_releases]: https://github.com/open-telemetry/opentelemetry-collector/releases
[otelcol_components]: https://github.com/open-telemetry/opentelemetry-collector-releases/blob/main/distributions/otelcol/manifest.yaml
[otelcol_contrib_components]: https://github.com/open-telemetry/opentelemetry-collector-releases/blob/main/distributions/otelcol-contrib/manifest.yaml
[otc_repository]: https://github.com/open-telemetry/opentelemetry-collector-contrib
[otelcol_api_changelog]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/CHANGELOG-API.md
[readme.md]: ../README.md
[dbstorage]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/extension/storage/dbstorage
[dockerobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/extension/observer/dockerobserver
[ecsobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/extension/observer/ecsobserver
[ecstaskobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/extension/observer/ecstaskobserver
[logstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/processor/logstransformprocessor
[windowseventlogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/receiver/windowseventlogreceiver
