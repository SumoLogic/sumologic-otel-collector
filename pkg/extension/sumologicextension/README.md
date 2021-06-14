# Sumo Logic Extension

**This extension is experimental and may receive breaking changes at any time.**

## Example Config

```yaml
extensions:
  sumologic:
    access_id: aaa
    access_key: bbbbbbbbbbbbbbbbbbbbbb
    collector_name: cccccccc

receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      load:

processors:

exporters:
  logging:
    loglevel: debug
  sumologic:
    endpoint: http://localhost:3000/
    auth:
      authenticator: sumologic

service:
  extensions: [sumologic]
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: []
      exporters: [sumologic, logging]
```
