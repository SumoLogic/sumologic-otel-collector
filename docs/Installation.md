# Installation

Sumo Logic OT Distro can be run either using binary file available in [Github releases][github_releases] or
using container images which are stored in AWS Public ECR under the following repository:
`public.ecr.aws/sumologic/sumologic-otel-collector`.

## Standalone

Sumo Logic OT Distro is a static Go binary.
To run it as standalone process you only need to run the binary file downloaded from
[Github releases][github_releases] with appropriate configuration.

1. Set the release version variable:

   ```bash
   export RELEASE_VERSION=0.0.12
   ```

1. Set the platform variable:

    ```bash
    export PLATFORM=linux_amd64
    ```

1. Download the release binary:

    ```bash
    curl -sLo otelcol-sumo "https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v${RELEASE_VERSION}/otelcol-sumo-${RELEASE_VERSION}-${PLATFORM}"
    ```

1. Install the release binary in your `PATH`:

    ```bash
    chmod +x otelcol-sumo
    sudo mv otelcol-sumo /usr/local/bin/otelcol-sumo
    ```

1. Verify installation:

    ```bash
    otelcol-sumo --version
    ```

1. Prepare configuration according to [this](Configuration.md) documentation and save it in `config.yaml`

1. Run Sumo Logic OT Distro:

   ```bash
   otelcol-sumo --config config.yaml
   ```

## Container image

In order to run Sumo Logic OT Distro in a container you only need to run the container
using the image available in `public.ecr.aws/sumologic/sumologic-otel-collector` repository.

1. Set the release version variable:

   ```bash
   export RELEASE_VERSION=0.0.12
   ```

1. Prepare configuration according to [this](Configuration.md) documentation and save it in `config.yaml`

1. Run Sumo Logic OT Distro in container, e.g.

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
