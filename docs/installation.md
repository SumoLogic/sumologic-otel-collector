# Installation

> **⚠️ DEPRECATION NOTICE**
>
> The installation scripts (`scripts/install.sh` and `scripts/install.ps1`) in this repository have been **deprecated** and moved to the [sumologic-otel-collector-packaging][packaging_repo] repository.
>
> **Please use the installation scripts and packages from the packaging repository instead.** These scripts will be removed from this repository in a future release.
>
> **Download the latest scripts:**
>
> - Linux/macOS: <https://download-otel.sumologic.com/latest/download/install.sh>
> - Windows: <https://download-otel.sumologic.com/latest/download/install.ps1>
>
> **For the latest installation instructions, please refer to the official documentation:**
>
> - [Linux Installation](https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector-linux/)
> - [MacOS Installation](https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector-macos/)
> - [Windows Installation](https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector-windows/)

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

[packaging_repo]: https://github.com/SumoLogic/sumologic-otel-collector-packaging
[linux_installation]: https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector-linux/
[macos_installation]: https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector-macos/
[windows_installation]:https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector-windows/
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

[sumologicextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.127.0/extension/sumologicextension/README.md
[sumologicextension_storing_credentials]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.127.0/extension/sumologicextension#storing-credentials

## Ansible

Example installation of Sumo Logic Distribution for OpenTelemetry Collector with Ansible is described in
[examples/ansible](https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector/ansible/).

## Puppet

Example installation of Sumo Logic Distribution for OpenTelemetry Collector with Puppet is described in
[examples/puppet](https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector/puppet/).

## Chef

Example installation of Sumo Logic Distribution for OpenTelemetry Collector with Chef is described in
[examples/chef](https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector/chef/.
