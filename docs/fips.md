# BoringCrypto and FIPS compliance

## Introduction

BoringCrypto is a cryptographic module [validated against FIPS 140][boringcrypto_certs].

Sumo OT distro can be built to use BoringCrypto instead of Go standard library crypto implementations.
This document explains how to obtain FIPS-capable binaries and how to run them in FIPS-approved mode.

## Obtaining the binaries

### Linux

Refer to [FIPS section of installation documentation](https://help.sumologic.com/docs/send-data/opentelemetry-collector/install-collector-linux/#fips)

### MacOS

We do not provde FIPS-compliant binary for macOS.

### Windows

Refer to [FIPS section of installation documentation](https://help.sumologic.com/docs/send-data/opentelemetry-collector/install-collector-windows/#fips)

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

> **Note**
> Currently FIPS-compliant builds are only available for Linux on amd64.

## Running in FIPS-approved mode

The FIPS-compliant build statically enforces the use of FIPS-approved algorithms in TLS cipher suites, and doesn't require any
extra configuration.

For more information, see the relevant [Go module](https://go.googlesource.com/go/+/dev.boringcrypto/src/crypto/tls/fipsonly/fipsonly.go).

> **Note**
> The above module guarantees not only the use of the right ciphers, but also FIPS-approved settings for each cipher.

## Validating use of BoringCrypto Symbols

Sumo OT distro can be built with debug symbols intact, which makes it possible to verify that it indeed
uses cryptographic functions from BoringCrypto:

```bash
$ go tool nm otelcol-sumo-fips-linux_amd64 | grep "_Cfunc__goboringcrypto_"
 42357d0 T _cgo_77e5233fcf12_Cfunc__goboringcrypto_AES_cbc_encrypt
 42357f0 T _cgo_77e5233fcf12_Cfunc__goboringcrypto_AES_ctr128_encrypt
 4235820 T _cgo_77e5233fcf12_Cfunc__goboringcrypto_AES_decrypt
 4235840 T _cgo_77e5233fcf12_Cfunc__goboringcrypto_AES_encrypt
 4235860 T _cgo_77e5233fcf12_Cfunc__goboringcrypto_AES_set_decrypt_key
 42358a0 T _cgo_77e5233fcf12_Cfunc__goboringcrypto_AES_set_encrypt_key
[...]
```

For more information on validating that a Go binary has been built against BoringCrypto, see the
following: <https://go.googlesource.com/go/+/dev.boringcrypto/misc/boring/>

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
supported ciphersuites. See the abridged cipherscan output for the test configuration below:

```bash
$ ./cipherscan localhost:4317
Warning: target is not a FQDN. SNI was disabled. Use a FQDN or '-servername <fqdn>'
[...]

Target: localhost:4317

prio  ciphersuite                  protocols              pfs                 curves
1     ECDHE-RSA-AES128-GCM-SHA256  TLSv1.2                ECDH,P-521,521bits  prime256v1,secp384r1,secp521r1
2     ECDHE-RSA-AES256-GCM-SHA384  TLSv1.2                ECDH,P-521,521bits  prime256v1,secp384r1,secp521r1
3     ECDHE-RSA-AES128-SHA         TLSv1,TLSv1.1,TLSv1.2  ECDH,P-521,521bits  prime256v1,secp384r1,secp521r1
4     ECDHE-RSA-AES256-SHA         TLSv1,TLSv1.1,TLSv1.2  ECDH,P-521,521bits  prime256v1,secp384r1,secp521r1
5     AES128-GCM-SHA256            TLSv1.2                None                None
6     AES256-GCM-SHA384            TLSv1.2                None                None
7     AES128-SHA                   TLSv1,TLSv1.1,TLSv1.2  None                None
8     AES256-SHA                   TLSv1,TLSv1.1,TLSv1.2  None                None
9     ECDHE-RSA-DES-CBC3-SHA       TLSv1,TLSv1.1,TLSv1.2  ECDH,P-521,521bits  prime256v1,secp384r1,secp521r1
10    DES-CBC3-SHA                 TLSv1,TLSv1.1,TLSv1.2  None                None
```

### Validating the client configuration

A TLS client sends all of its supported ciphersuites as part of the handshake, in the `ClientHello` message.
Examining this message requires some additional work, as we need to set up a TLS server with a valid
certificate that will print it to stdout for us. For the purposes of this document, we're going to use a
small [Go utility](https://github.com/rgl/tls-dump-clienthello) to do this, but it's also possible with
OpenSSL's `s_server` command.

```bash
$ ./tls-dump-clienthello_$(go env GOOS)_$(go env GOARCH)_$(go env GOAMD64)/tls-dump-clienthello

[...]

client version: TLSv1.3
client version: TLSv1.2
client cipher suite: TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 (0xc02b)
client cipher suite: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 (0xc02f)
client cipher suite: TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 (0xc02c)
client cipher suite: TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384 (0xc030)
client cipher suite: TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256 (0xcca9)
client cipher suite: TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256 (0xcca8)
client cipher suite: TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA (0xc009)
client cipher suite: TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA (0xc013)
client cipher suite: TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA (0xc00a)
client cipher suite: TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA (0xc014)
client cipher suite: TLS_RSA_WITH_AES_128_GCM_SHA256 (0x009c)
client cipher suite: TLS_RSA_WITH_AES_256_GCM_SHA384 (0x009d)
client cipher suite: TLS_RSA_WITH_AES_128_CBC_SHA (0x002f)
client cipher suite: TLS_RSA_WITH_AES_256_CBC_SHA (0x0035)
client cipher suite: TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA (0xc012)
client cipher suite: TLS_RSA_WITH_3DES_EDE_CBC_SHA (0x000a)
client cipher suite: TLS_AES_128_GCM_SHA256 (0x1301)
client cipher suite: TLS_AES_256_GCM_SHA384 (0x1302)
client cipher suite: TLS_CHACHA20_POLY1305_SHA256 (0x1303)
client curve: x25519 (29)
client curve: secp256r1 (23)
client curve: secp384r1 (24)
client curve: secp521r1 (25)
client point: uncompressed (0)

[...]
```

## Setting minimum TLS version

While [FIPS 140][fips] doesn't, strictly speaking, mandate a specific TLS version, it is nonetheless a good practice to use a recent version. Otelcol allows TLS to be configured for each component individually, in a standardized manner. You can find the details in the [readme](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/configtls/README.md).

By default, the minimum TLS version for otel components is `1.2`. This is a good default that is compliant with the standard. At the moment, there is some uncertainty as to
whether BoringCrypto is FIPS-compliant when using TLS `1.3`, so we recommend staying with the default.

If you nonetheless want to use `1.3`, here's an example for the `otlp` exporter:

```yaml
exporters:
  otlp:
    endpoint: myserver.local:55690
    tls:
      min_version: "1.3"
```

[fips]: https://en.wikipedia.org/wiki/FIPS_140
[boringcrypto_certs]: https://csrc.nist.gov/projects/cryptographic-module-validation-program/validated-modules/search?SearchMode=Basic&ModuleName=boring&CertificateStatus=Active&ValidationYear=0
