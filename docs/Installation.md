# Installation

The Sumo Logic OT Distro can be run using either the binary file available in [Github releases][github_releases] or
the container images stored in AWS Public ECR under the following repository:
[public.ecr.aws/sumologic/sumologic-otel-collector](https://gallery.ecr.aws/sumologic/sumologic-otel-collector).

- [Standalone](#standalone)
  - [Linux on amd64 (x86-64)](#linux-on-amd64-x86-64)
  - [Linux on arm64](#linux-on-arm64)
  - [MacOS on amd64 (x86-64)](#macos-on-amd64-x86-64)
  - [MacOS on arm64 (Apple M1)](#macos-on-arm64-apple-m1-x86-64)
  - [Upgrading standalone installation](#upgrading-standalone-installation)
- [Container image](#container-image)
- [Systemd service](#systemd-service)
- [Ansible](#ansible)
- [Puppet](#puppet)
- [Chef](#chef)

## Standalone

The Sumo Logic OT Distro is a static Go binary.
To run it as a standalone process you only need to run the binary file downloaded from
[Github releases][github_releases] with an appropriate configuration.

Follow the steps for your platform below.

### Linux on amd64 (x86-64)

1. Download the release binary:

    ```bash
    curl -sLo otelcol-sumo "https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v0.47.0-sumo-0/otelcol-sumo-0.47.0-sumo-0-linux_amd64"
    ```

1. Install the release binary in your `PATH`:

    ```bash
    chmod +x otelcol-sumo
    sudo mv otelcol-sumo /usr/local/bin/otelcol-sumo
    ```

1. Verify the installation:

    ```bash
    otelcol-sumo --version
    ```

1. Prepare the configuration according to [this](Configuration.md) document and save it in `config.yaml`.

> **IMPORTANT NOTE**:
> It is recommended to limit access to the configuration file as it contains sensitive information.
> You can change access permissions to the configuration file using:
>
> ```bash
> chmod 640 config.yaml
> ```

1. Run Sumo Logic OT Distro:

   ```bash
   otelcol-sumo --config config.yaml
   ```

### Linux on arm64

1. Download the release binary:

    ```bash
    curl -sLo otelcol-sumo "https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v0.47.0-sumo-0/otelcol-sumo-0.47.0-sumo-0-linux_arm64"
    ```

1. Install the release binary in your `PATH`:

    ```bash
    chmod +x otelcol-sumo
    sudo mv otelcol-sumo /usr/local/bin/otelcol-sumo
    ```

1. Verify the installation:

    ```bash
    otelcol-sumo --version
    ```

1. Prepare the configuration according to [this](Configuration.md) document and save it in `config.yaml`.

> **IMPORTANT NOTE**:
> It is recommended to limit access to the configuration file as it contains sensitive information.
> You can change access permissions to the configuration file using:
>
> ```bash
> chmod 640 config.yaml
> ```

1. Run Sumo Logic OT Distro:

   ```bash
   otelcol-sumo --config config.yaml
   ```

### MacOS on amd64 (x86-64)

1. Download the release binary:

    ```bash
    curl -sLo otelcol-sumo "https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v0.47.0-sumo-0/otelcol-sumo-0.47.0-sumo-0-darwin_amd64"
    ```

1. Install the release binary in your `PATH`:

    ```bash
    chmod +x otelcol-sumo
    sudo mv otelcol-sumo /usr/local/bin/otelcol-sumo
    ```

1. Verify the installation:

    ```bash
    otelcol-sumo --version
    ```

1. Prepare the configuration according to [this](Configuration.md) document and save it in `config.yaml`.

> **IMPORTANT NOTE**:
> It is recommended to limit access to the configuration file as it contains sensitive information.
> You can change access permissions to the configuration file using:
>
> ```bash
> chmod 640 config.yaml
> ```

1. Run Sumo Logic OT Distro:

   ```bash
   otelcol-sumo --config config.yaml
   ```

### MacOS on arm64 (Apple M1) (x86-64)

This is currently not a supported platform. Manually building for it might work, but we haven't tested it.

### Upgrading standalone installation

To upgrade, simply perform the above installation steps again,
overwriting the `otelcol-sumo` binary with newer version.

Before running the newer version, make sure to check the [release notes][github_releases]
for potential breaking changes that would require manual migration steps.

[github_releases]: https://github.com/SumoLogic/sumologic-otel-collector/releases

## Container image

To run the Sumo Logic OT Distro in a container, you only need to run the container
using the image available in the
[public.ecr.aws/sumologic/sumologic-otel-collector](https://gallery.ecr.aws/sumologic/sumologic-otel-collector)
repository.

1. Set the release version variable:

   ```bash
   export RELEASE_VERSION=0.47.0-sumo-0
   ```

1. Prepare the configuration according to [this](Configuration.md) document and save it in `config.yaml`.

> **IMPORTANT NOTE**:
> It is recommended to limit access to the configuration file as it contains sensitive information.
> You can change access permissions to the configuration file using:
>
> ```bash
> chmod 640 config.yaml
> ```

1. Run the Sumo Logic OT Distro in container, e.g.

    ```bash
    $ docker run --rm -ti --name sumologic-otel-collector -v "$(pwd)/config.yaml:/etc/config.yaml" "public.ecr.aws/sumologic/sumologic-otel-collector:${RELEASE_VERSION}" --config /etc/config.yaml
    2021-07-06T10:31:17.477Z      info      service/application.go:277      Starting otelcol-sumo-linux_amd64...    {"Version": "v0.0.10", "NumCPU": 4}
    2021-07-06T10:31:17.477Z      info      service/application.go:185      Setting up own telemetry...
    2021-07-06T10:31:17.478Z      info      service/telemetry.go:98 Serving Prometheus metrics      {"address": ":8888", "level": 0, "service.instance.id": "596814dd-d8ad-4a4f-b2e9-106c29c416a0"}
    2021-07-06T10:31:17.478Z      info      service/application.go:220      Loading configuration...
    2021-07-06T10:31:17.479Z      info      service/application.go:236      Applying configuration...
    2021-07-06T10:31:17.479Z      info      builder/exporters_builder.go:274        Exporter was built.     {"kind": "exporter", "exporter": "sumologic"}
    2021-07-06T10:31:17.479Z      info      builder/pipelines_builder.go:204        Pipeline was built.     {"pipeline_name": "metrics/1", "pipeline_datatype": "metrics"}
    2021-07-06T10:31:17.479Z      info      builder/receivers_builder.go:230        Receiver was built.     {"kind": "receiver", "name": "telegraf", "datatype": "metrics"}
    2021-07-06T10:31:17.479Z      info      service/service.go:155  Starting extensions...
    2021-07-06T10:31:17.479Z      info      builder/extensions_builder.go:53        Extension is starting...        {"kind": "extension", "name": "sumologic"}
    2021-07-06T10:31:17.479Z      info      sumologicextension@v0.27.0/extension.go:128     Locally stored credentials not found, registering the collector {"kind": "extension", "name": "sumologic"}
    2021-07-06T10:31:17.480Z      info      sumologicextension@v0.27.0/credentials.go:142   Calling register API    {"kind": "extension", "name": "sumologic", "URL": "https://collectors.sumologic.com/api/v1/collector/register"}
    ...
    ```

### Important note about local state files when using `sumologicextension`

> **IMPORTANT NOTE**:
>
> When running Sumo Logic OT distro in a container with [`sumologicextension`][sumologicextension],
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

## Systemd Service

> **IMPORTANT NOTE**:
>
> Make sure that the user that will run the `otelcol-sumo` process has access to
> any directories within your filesystem that have been used in you configuration.
>
> For instance, using [filestorage extension][filestorage_help] in your configuration
> like so:
>
> ```yaml
> extensions:
>   file_storage/custom_settings:
>     directory: /var/lib/otelcol/mydir
>     timeout: 1s
> ```
>
> will require that the user running the process has access to `/var/lib/otelcol/mydir`.

[filestorage_help]: ./Configuration.md#file-storage-extension

To run opentelemetry collector as Systemd Service please apply following steps:

1. Ensure that `otelcol-sumo` [has been installed](#linux-on-amd64-x86-64) into `/usr/local/bin/otelcol-sumo`:

   ```bash
   /usr/local/bin/otelcol-sumo --version
   ```

1. Create configuration file and save it as `/etc/otelcol-sumo/config.yaml`.

> **IMPORTANT NOTE**:
> It is recommended to limit access to the configuration file as it contains sensitive information.
> You can change access permissions to the configuration file using:
>
> ```bash
> chmod 640 config.yaml
> ```

1. Create `user` and `group` to run opentelemetry by:

   ```bash
   sudo useradd -rUs /bin/false opentelemetry
   ```

1. Ensure that `/etc/otelcol-sumo/config.yaml` can be accessed by `opentelemetry` user
   which will be used to run the service.

   ```bash
   $ ls -la /etc/otelcol-sumo/config.yaml
   -rw-r--r-- 1 opentelemetry daemon 0 Feb 16 16:23 /etc/otelcol-sumo/config.yaml
   ```

1. Verify if opentelemetry collector runs without errors:

   ```bash
   sudo su -s /bin/bash opentelemetry -c '/usr/local/bin/otelcol-sumo --config /etc/otelcol-sumo/config.yaml'
   ```

1. Create service file: `/etc/systemd/system/otelcol-sumo.service`:

   ```conf
   [Unit]
   Description=Sumologic Opentelemetry Collector

   [Service]
   ExecStart=/usr/local/bin/otelcol-sumo --config /etc/otelcol-sumo/config.yaml
   User=opentelemetry
   Group=opentelemetry
   MemoryHigh=200M
   MemoryMax=300M
   TimeoutStopSec=20

   [Install]
   WantedBy=multi-user.target
   ```

   _Note: adjust memory configuration to your setup._

1. Enable autostart of the service:

   ```bash
   sudo systemctl enable otelcol-sumo
   ```

1. Start service and check status:

   ```bash
   sudo systemctl start otelcol-sumo
   sudo systemctl status otelcol-sumo  # checks status
   sudo journalctl -u otelcol-sumo  # checks logs
   ```

## Ansible

Example installation of Sumo Logic Distro of OpenTelemetry Collector with Ansible is described in
[examples/ansible](../examples/ansible/README.md).

## Puppet

Example installation of Sumo Logic Distro of OpenTelemetry Collector with Puppet is described in
[examples/puppet](../examples/puppet/README.md).

## Chef

Example installation of Sumo Logic Distro of OpenTelemetry Collector with Chef is described in
[examples/chef](../examples/chef/README.md).
