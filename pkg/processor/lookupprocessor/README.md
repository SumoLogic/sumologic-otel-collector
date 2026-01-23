# Lookup Processor

| Status | |
| ------ | ----- |
| Stability | [alpha]: logs |
| Distributions | [] |

[alpha]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#alpha

## Description

The lookup processor enriches telemetry signals by performing external lookups to retrieve additional data. It reads an attribute value, uses it as a key to query a lookup source, and sets the result as a new attribute.

Currently supports logs, with metrics and traces support planned.

## Configuration

```yaml
processors:
  lookup:
    source:
      type: dns
      record_type: A
      timeout: 5s
    cache:
      enabled: true
      size: 1000
      ttl: 5m
      negative_ttl: 1m
    attributes:
      - key: server.ip
        from_attribute: server.hostname
        default: "unknown"
        action: upsert
        context: record
```

### Full Configuration

| Field | Description | Default |
| ----- | ----------- | ------- |
| `source.type` | The source type identifier (`noop`, `yaml`, `dns`) | `noop` |
| `attributes` | List of attribute enrichment rules (required) | - |
| `cache.enabled` | Enable caching of lookup results | `false` |
| `cache.size` | Maximum number of entries in the cache | `1000` |
| `cache.ttl` | Time-to-live for cached successful results (0 = no expiration) | `0` |
| `cache.negative_ttl` | Time-to-live for cached not-found results (0 = don't cache) | `0` |

### Attribute Configuration

Each entry in `attributes` defines a lookup rule:

| Field | Description | Default |
| ----- | ----------- | ------- |
| `key` | Name of the attribute to set with the lookup result (required) | - |
| `from_attribute` | Name of the attribute containing the lookup key (required) | - |
| `default` | Value to use when lookup returns no result | - |
| `action` | How to handle the result: `insert`, `update`, `upsert` | `upsert` |
| `context` | Where to read/write attributes: `record`, `resource` | `record` |

### Actions

- **insert**: Only set the attribute if it doesn't already exist
- **update**: Only set the attribute if it already exists
- **upsert**: Always set the attribute (default)

### Context

- **record**: Read from and write to record-level attributes (log records, spans, metric data points) (default)
- **resource**: Read from and write to resource attributes

## Built-in Sources

### noop

A no-operation source that always returns "not found". Useful for testing.

```yaml
processors:
  lookup:
    source:
      type: noop
    attributes:
      - key: result
        from_attribute: key
        default: "not-found"
```

### yaml

Loads key-value mappings from a YAML file. The file should contain a flat map of string keys to values.

| Field | Description | Default |
| ----- | ----------- | ------- |
| `path` | Path to the YAML file (required) | - |

```yaml
processors:
  lookup:
    source:
      type: yaml
      path: /etc/otel/mappings.yaml
    attributes:
      - key: service.display_name
        from_attribute: service.name
```

Example mappings file (`mappings.yaml`):

```yaml
svc-frontend: "Frontend Web App"
svc-backend: "Backend API Service"
svc-worker: "Background Worker"
```

### dns

Performs DNS lookups to resolve hostnames to IP addresses or IP addresses to hostnames based on DNS record type.

| Field | Description | Default |
| ----- | ----------- | ------- |
| `record_type` | DNS record type: `A` (hostname to IPv4), `AAAA` (hostname to IPv6), or `PTR` (IP to hostname) | `A` |
| `timeout` | Maximum time to wait for DNS resolution | `5s` |
| `resolver` | Custom DNS server (format: "host:port", e.g., "8.8.8.8:53"). If empty, uses system default | - |
| `multiple_results` | If true, returns all results as comma-separated string; if false, returns first result only | `false` |

**A record lookup** (hostname to IPv4):

```yaml
processors:
  lookup:
    source:
      type: dns
      record_type: A
      timeout: 5s
    cache:
      enabled: true
      size: 1000
      ttl: 5m
    attributes:
      - key: server.ip
        from_attribute: server.hostname
```

Detailed documentation for sources can be found under the respective source's readme.

## Caching

All sources support caching to improve performance and reduce load on external systems. Configure caching at the processor level:

```yaml
processors:
  lookup:
    source:
      type: dns
      record_type: A
    cache:
      enabled: true
      size: 1000           # Max entries
      ttl: 5m              # Cache successful lookups for 5 minutes
      negative_ttl: 1m     # Cache failed lookups for 1 minute
    attributes:
      - key: ip
        from_attribute: hostname
```

**Cache Configuration:**

- `enabled`: Enable/disable caching (default: `false`)
- `size`: Maximum number of entries (default: `1000`)
- `ttl`: Time-to-live for successful lookups. Use `0` for no expiration (default: `0`)
- `negative_ttl`: Time-to-live for not-found results. Use `0` to not cache failures (default: `0`)

The cache uses an LRU (Least Recently Used) eviction policy when it reaches the size limit.

## Benchmarks

Run benchmarks with:

```bash
make benchmark
```

### Processor Performance

Measures the full processing pipeline including pdata operations, attribute iteration, value conversion, and telemetry. Uses noop source to isolate processor overhead from source implementation (Apple M4 Pro):

| Scenario | ns/op | B/op | allocs/op |
| -------- | ----- | ---- | --------- |
| 1 log, 1 attribute | 323 | 696 | 20 |
| 10 logs, 1 attribute | 1,274 | 3,216 | 74 |
| 100 logs, 1 attribute | 11,076 | 28,512 | 614 |
| 100 logs, 3 attributes | 21,604 | 54,113 | 1,014 |
| 1000 logs, 1 attribute | 122,447 | 280,617 | 6,014 |

Source-specific benchmarks are available in each source's README.

## Custom Sources

Custom lookup sources can be added using `WithSources`:

```go
import (
    "github.com/open-telemetry/opentelemetry-collector-contrib/processor/lookupprocessor"
    "github.com/example/httplookup"
)

factories.Processors[lookupprocessor.Type] = lookupprocessor.NewFactoryWithOptions(
    lookupprocessor.WithSources(httplookup.NewFactory()),
)
```

### Implementing a Source

```go
package mysource

import (
    "context"
    "errors"
    "time"

    "github.com/open-telemetry/opentelemetry-collector-contrib/processor/lookupprocessor/lookupsource"
)

type Config struct {
    Endpoint string        `mapstructure:"endpoint"`
    Timeout  time.Duration `mapstructure:"timeout"`
}

func (c *Config) Validate() error {
    if c.Endpoint == "" {
        return errors.New("endpoint is required")
    }
    return nil
}

func NewFactory() lookupsource.SourceFactory {
    return lookupsource.NewSourceFactory(
        "mysource",
        func() lookupsource.SourceConfig {
            return &Config{Timeout: 5 * time.Second}
        },
        createSource,
    )
}

func createSource(
    ctx context.Context,
    settings lookupsource.CreateSettings,
    cfg lookupsource.SourceConfig,
) (lookupsource.Source, error) {
    c := cfg.(*Config)

    // Define the lookup function
    lookupFunc := func(ctx context.Context, key string) (any, bool, error) {
        // Perform lookup - return (value, found, error)
        return "result", true, nil
    }

    // Wrap with cache if enabled
    if settings.Cache.Enabled {
        cache := lookupsource.NewCache(settings.Cache)
        cache.SetLogger(settings.TelemetrySettings.Logger)
        lookupFunc = lookupsource.WrapWithCache(cache, lookupFunc)
    }

    return lookupsource.NewSource(
        lookupFunc,
        func() string { return "mysource" },
        nil, // start function (optional)
        nil, // shutdown function (optional)
    ), nil
}
```
