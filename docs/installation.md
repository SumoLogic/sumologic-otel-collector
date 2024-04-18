# Installation

The Sumo Logic Distribution for OpenTelemetry Collector can be run using either the binary file available in [Github releases][github_releases] or
the container images stored in AWS Public ECR under the following repositories:

- [public.ecr.aws/sumologic/sumologic-otel-collector](https://gallery.ecr.aws/sumologic/sumologic-otel-collector)
- [sumologic/sumologic-otel-collector](https://hub.docker.com/repository/docker/sumologic/sumologic-otel-collector)

- [Linux][linux_installation]
- [Windows][windows_installation]
- [MacOS][macos_installation]
- [Container image](#container-image)
  - [Important note about local state files when using `sumologicextension`](#important-note-about-local-state-files-when-using-sumologicextension)
- [Ansible](#ansible)
- [Puppet](#puppet)
- [Chef](#chef)

[linux_installation]: https://help.sumologic.com/docs/send-data/opentelemetry-collector/install-collector-linux/
[macos_installation]: https://help.sumologic.com/docs/send-data/opentelemetry-collector/install-collector-macos/
[windows_installation]:https://help.sumologic.com/docs/send-data/opentelemetry-collector/install-collector-windows/
[github_releases]: https://github.com/SumoLogic/sumologic-otel-collector/releases

## Container image

To run the Sumo Logic Distribution for OpenTelemetry Collector in a container, you only need to run the container
using the image available in the one of the following repositories:

- [public.ecr.aws/sumologic/sumologic-otel-collector](https://gallery.ecr.aws/sumologic/sumologic-otel-collector)
- [sumologic/sumologic-otel-collector](https://hub.docker.com/repository/docker/sumologic/sumologic-otel-collector)

1. Set the release version variable:

   ```bash
   export RELEASE_VERSION=0.75.0-sumo-0
   ```

1. Prepare the configuration according to [this](configuration.md) document and save it in `config.yaml`.

   > **IMPORTANT NOTE**:
   > It is recommended to limit access to the configuration file as it contains sensitive information.
   > You can change access permissions to the configuration file using:
   >
   > ```bash
   > chmod 640 config.yaml
   > ```

1. Run the Sumo Logic Distribution for OpenTelemetry Collector in container, e.g.

    ```bash
    docker run --rm -ti --name sumologic-otel-collector \
       -v "$(pwd)/config.yaml:/etc/otel/config.yaml" \
       "public.ecr.aws/sumologic/sumologic-otel-collector:${RELEASE_VERSION}"
    ```

### Important note about local state files when using `sumologicextension`

> **IMPORTANT NOTE**:
>
> When running Sumo Logic Distribution for OpenTelemetry Collector in a container with [`sumologicextension`][sumologicextension],
> one needs to take into account the local state files which are being used locally
> in order to store collector credentials upon successful collector registration.
>
> When the collector is being run with `sumologicextension` (which manages collector
> registration), local directory (which is configured via `collector_credentials_directory`
> in `sumologicextension`, and which is by default set to `$HOME/.sumologic-otel-collector`)
> will be used to store the aforementioned state files.
> Without any mounts defined on the container the collector will register itself
> every time it starts up, creating clutter on Sumo Logic Collector Management page.
>
> In order to avoid that, use volume mounts or any other mechanism to mount
> the collector credentials directory to the container to persist the state
> or use `clobber` configuration option from `sumologicextension` to force collector
> re-registering under the same name, every time is starts up.
>
> One can read more about the above described mechanism in
> [`sumologicextension` README.md][sumologicextension_storing_credentials].

[sumologicextension]: ./../pkg/extension/sumologicextension/README.md
[sumologicextension_storing_credentials]: ./../pkg/extension/sumologicextension/README.md#Storing-credentials

## Ansible

Example installation of Sumo Logic Distribution for OpenTelemetry Collector with Ansible is described in
[examples/ansible](https://help.sumologic.com/docs/send-data/opentelemetry-collector/install-collector/ansible/).

## Puppet

Example installation of Sumo Logic Distribution for OpenTelemetry Collector with Puppet is described in
[examples/puppet](https://help.sumologic.com/docs/send-data/opentelemetry-collector/install-collector/puppet/).

## Chef

Example installation of Sumo Logic Distribution for OpenTelemetry Collector with Chef is described in
[examples/chef](https://help.sumologic.com/docs/send-data/opentelemetry-collector/install-collector/chef/.
