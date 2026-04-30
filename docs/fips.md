# FIPS 140-3 compliance

## Introduction

Sumo OT distro can be built using Go's native FIPS 140-3 cryptographic module, available since Go 1.24
via the `GOFIPS140` build setting. When enabled, Go uses its own pure-Go implementation of FIPS 140-3
approved algorithms (`crypto/internal/fips140/...`) instead of the standard library crypto packages.

The native module (version `v1.0.0`) has completed CAVP algorithm testing (certificate A6650) and
is currently undergoing [NIST CMVP review][fips_module_status]. See [FIPS 140-3 in Go][go_fips_doc]
for full details.

This document explains how to obtain FIPS-capable binaries and how to run them in FIPS-approved mode.

## Obtaining the binaries

### Linux

Refer to [FIPS section of installation documentation](https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector-linux/#fips)

### MacOS

We do not provide FIPS-compliant binary for macOS.

### Windows

Refer to [FIPS section of installation documentation](https://www.sumologic.com/help/docs/send-data/opentelemetry-collector/install-collector-windows/#fips)

### Docker images

Docker images containing the binaries have their tag suffixed with `-fips`. For example, where the normal image for
version `0.75.0-sumo-0` is:

```text
docker pull public.ecr.aws/sumologic/sumologic-otel-collector:0.75.0-sumo-0
```

The FIPS-approved version would be:

```text
docker pull public.ecr.aws/sumologic/sumologic-otel-collector:0.75.0-sumo-0-fips
```

## Running in FIPS-approved mode

The FIPS-compliant build statically enforces the use of FIPS-approved algorithms in TLS cipher suites, and doesn't require any
extra configuration.

For more information, see [FIPS 140-3 in Go][go_fips_doc].

> **Note**
> The above module guarantees not only the use of the right ciphers, but also FIPS-approved settings for each cipher.

## Validating use of native FIPS 140-3 symbols

Sumo OT distro can be built with debug symbols intact, which makes it possible to verify that it indeed
uses the native FIPS 140-3 cryptographic module:

```bash
$ go tool nm otelcol-sumo-fips-linux_amd64 | grep "crypto/internal/fips140"
  5a3d820 T crypto/internal/fips140.init
  5a3d860 T crypto/internal/fips140.Enabled
  5a3d8a0 T crypto/internal/fips140/aes.New
[...]
```

For more information on Go's native FIPS 140-3 support, see: <https://go.dev/doc/security/fips140>

## Validating TLS Handshake

Depending on configuration, Otelcol can act as a TLS client, server, or both. More specifically,
some receivers can accept data via a TLS-wrapped protocol like HTTP or GRPC. Similarly, some
exporters can send data using such a protocol. The configuration is standardized and can be
found in the [readme](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/configtls/README.md).

For the purpose of examining the TLS handshake in both of these situations, we'll use the following minimal
configuration:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        tls:
          cert_file: "/etc/otel/server.crt"
          key_file: "/etc/otel/server.key"
  hostmetrics:
    collection_interval: 10s
    scrapers:
      memory:

exporters:
  otlphttp:
    endpoint: https://localhost/

service:
  pipelines:
    metrics:
      receivers: [otlp, hostmetrics]
      exporters: [otlphttp]
```

> **Note**
> Valid TLS credentials need to be present under `/etc/otel/server.crt` and `/etc/otel/server.key` for this configuration to be valid.

### Validating the server configuration

We can use a tool like [cipherscan](https://github.com/mozilla/cipherscan) to enumerate all of a server's
supported ciphersuites.

### Validating the client configuration

A TLS client sends all of its supported ciphersuites as part of the handshake, in the `ClientHello` message.
Examining this message requires some additional work, as we need to set up a TLS server with a valid
certificate that will print it to stdout for us. For the purposes of this document, we're going to use a
small [Go utility](https://github.com/rgl/tls-dump-clienthello) to do this, but it's also possible with
OpenSSL's `s_server` command.

## Setting minimum TLS version

While [FIPS 140][fips] doesn't, strictly speaking, mandate a specific TLS version, it is nonetheless a good practice to use a recent version. Otelcol allows TLS to be configured for each component individually, in a standardized manner. You can find the details in the [readme](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/configtls/README.md).

By default, the minimum TLS version for otel components is `1.2`. This is a good default that is compliant with the standard. TLS 1.3 is also FIPS-approved under FIPS 140-3 when using the approved cipher suites (`TLS_AES_128_GCM_SHA256` and `TLS_AES_256_GCM_SHA384`), and Go's native FIPS module enforces this automatically.

If you want to use `1.3`, here's an example for the `otlp` exporter:

```yaml
exporters:
  otlp:
    endpoint: myserver.local:55690
    tls:
      min_version: "1.3"
```

[fips]: https://en.wikipedia.org/wiki/FIPS_140
[go_fips_doc]: https://go.dev/doc/security/fips140
[fips_module_status]: https://csrc.nist.gov/projects/cryptographic-module-validation-program/validated-modules/search?SearchMode=Basic&ModuleName=go&ValidationYear=0
