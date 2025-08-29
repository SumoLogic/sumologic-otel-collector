# RabbitMQ Receiver

This directory contains tools to help develop and test the RabbitMQ Receiver in
the OpenTelemetry Collector Contrib repository.

## Getting Started

1. Clone / update opentelemetry-collector-contrib:

``` shell
make contrib
```

1. Make any desired changes to the contrib files in `../../contrib`.

1. Spin up the Docker Compose stack:

``` shell
make up
```

1. View Docker Compose logs with:

``` shell
make logs
```

or

``` shell
make logs-follow
```

1. Fetch metrics or metrics with values of `0`:

``` shell
make metrics
```

``` shell
make zero-metrics
```

1. Whenever files are changed inside of the contrib directory, spin down the
Docker Compose stack and spin it back up:

``` shell
make downup
```

1. When work is complete, destroy the Docker Compose stack:

``` shell
make down
```
