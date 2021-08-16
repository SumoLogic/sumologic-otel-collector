# Installation

The Sumo Logic OT Distro can be run using either the binary file available in [Github releases][github_releases] or
the container images stored in AWS Public ECR under the following repository:
[public.ecr.aws/sumologic/sumologic-otel-collector](https://gallery.ecr.aws/sumologic/sumologic-otel-collector).

- [Standalone](#standalone)
  - [Linux on amd64 (x86-64)](#linux-on-amd64-x86-64)
  - [Linux on arm64](#linux-on-arm64)
  - [MacOS on amd64 (x86-64)](#macos-on-amd64-x86-64)
  - [Upgrading standalone installation](#upgrading-standalone-installation)
- [Container image](#container-image)
- [Systemd service](#systemd-service)

## Standalone

The Sumo Logic OT Distro is a static Go binary.
To run it as a standalone process you only need to run the binary file downloaded from
[Github releases][github_releases] with an appropriate configuration.

Follow the steps for your platform below.

### Linux on amd64 (x86-64)

1. Download the release binary:

    ```bash
    curl -sLo otelcol-sumo "https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v0.0.18/otelcol-sumo-0.0.18-linux_amd64"
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

1. Run Sumo Logic OT Distro:

   ```bash
   otelcol-sumo --config config.yaml
   ```

### Linux on arm64

1. Download the release binary:

    ```bash
    curl -sLo otelcol-sumo "https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v0.0.18/otelcol-sumo-0.0.18-linux_arm64"
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

1. Run Sumo Logic OT Distro:

   ```bash
   otelcol-sumo --config config.yaml
   ```

### MacOS on amd64 (x86-64)

1. Download the release binary:

    ```bash
    curl -sLo otelcol-sumo "https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v0.0.18/otelcol-sumo-0.0.18-darwin_amd64"
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

1. Run Sumo Logic OT Distro:

   ```bash
   otelcol-sumo --config config.yaml
   ```

### Upgrading standalone installation

To upgrade, simply perform the above installation steps again,
overwriting the `otelcol-sumo` binary with newer version.

Before running the newer version, make sure to check the [release notes][github_releases]
for potential breaking changes that would require manual migration steps.

## Container image

To run the Sumo Logic OT Distro in a container, you only need to run the container
using the image available in the
[public.ecr.aws/sumologic/sumologic-otel-collector](https://gallery.ecr.aws/sumologic/sumologic-otel-collector)
repository.

1. Set the release version variable:

   ```bash
   export RELEASE_VERSION=0.0.18
   ```

1. Prepare the configuration according to [this](Configuration.md) document and save it in `config.yaml`.

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

[github_releases]: https://github.com/SumoLogic/sumologic-otel-collector/releases

## Systemd Service

To run opentelemetry collector as Systemd Service please apply following steps:

1. Ensure that `otelcol-sumo` [has been installed](#linux-on-amd64-x86-64) into `/usr/local/bin/otelcol-sumo`:

   ```bash
   /usr/local/bin/otelcol-sumo --version
   ```

1. Create configuration file and save it as `/etc/otelcol-sumo/config.yml`.

1. Create `user` and `group` to run opentelemetry by:

   ```bash
   sudo useradd -rUs /bin/false opentelemetry
   ```

1. Verify if opentelemetry collector runs without errors:

   ```bash
   sudo su -s /bin/bash opentelemetry -c '/usr/local/bin/otelcol-sumo --config /etc/otelcol-sumo/config.yml'
   ```

1. Create service file: `/etc/systemd/system/otelcol-sumo.service`:

   ```conf
   [Unit]
   Description=Sumologic Opentelemetry Collector

   [Service]
   ExecStart=/usr/local/bin/otelcol-sumo --config /etc/otelcol-sumo/config.yml
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
