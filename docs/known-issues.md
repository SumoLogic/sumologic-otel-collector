# Known Issues

- [Changes to collector properties are not applied](#changes-to-collector-properties-are-not-applied)
- [Collector fails to start when deleted from UI](#collector-fails-to-start-when-deleted-from-ui)
- [Enabling `clobber` property re-registers collector on every restart](#enabling-clobber-property-re-registers-collector-on-every-restart)
- [Cannot start reading file logs from specific point in time](#cannot-start-reading-file-logs-from-specific-point-in-time)

## Changes to collector properties are not applied

After running the collector for the first time, changes to collector properties
(e.g. collector description, category, fields) are not applied.

To work around this, you need to delete the existing collector registration
and register the collector again.
To do this, you need to do two things:

- Remove the collector in Sumo Logic UI
  - Log in to your Sumo Logic UI
  - Go to `Manage Data` - `Collection`
  - Find your collector
  - Click `Delete` on the right-hand side of the collector

- Delete local collector registration file in `~/.sumologic-otel-collector/`.

After that, the collector will register on next run.

If you delete the collector in the UI but not delete the local registration file,
the collector will fail to start - see [Collector fails to start when deleted from UI](#collector-fails-to-start-when-deleted-from-ui).

On the other hand, if you only delete the local registration file
and do not delete the collector in the UI,
a new collector will be created with current timestamp as a suffix,
to prevent overwriting the existing collector.

## Collector fails to start when deleted from UI

After successful registration of collector, if you delete the collector in Sumo Logic UI,
the collector will fail to start on next run.
The error message is similar to the below:

```console
2021-08-24T10:52:38.639Z  error  sumologicextension@v0.31.0/extension.go:373  Heartbeat error  {"kind": "extension", "name": "sumologic", "collector_name": "<your-collector-name>", "collector_id": "0000000001A2B3C4", "error": "collector heartbeat request failed, status code: 401, body: {\n\"servlet\":\"rest\",\n\"message\":\"Could not authenticate.\",\n\"url\":\"/api/v1/collector/heartbeat\",\n\"status\":\"401\"\n}"}
github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension.(*SumologicExtension).heartbeatLoop
  github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension@v0.31.0/extension.go:373
```

To work around this, delete the local collector registration file at `~/.sumologic-otel-collector/`.
The collector will re-register on next run.

## Enabling `clobber` property re-registers collector on every restart

If you set the `extensions.sumologic.clobber` property to `true`,
a new collector registration that replaces the previously existing registration
will be created on every run of the collector.

This affects the `_collectorId` attribute, which is different for every new collector registration.

To prevent this, remove the `extensions.sumologic.clobber` property or set it to `false`.

## Cannot start reading file logs from specific point in time

The [Filelog receiver][filelogreceiver_docs] currently supports only two modes of reading local files:

- `start_at: beginning`: Ingest the whole file from the beginning, or
- `start_at: end`: Only ingest newly added lines.

The `start_at` property is common to all the files read by the receiver -
it cannot be set to `end` for some files and to `beginning` for other files.
Note that this can be worked around by creating two separate Filelog receivers,
one reading from the beginning and one reading from the end.

The other problem is that it is not currently possible for Filelog receiver to start reading files at a specific point,
or only read files created or modified after a specific point in time.

There is currently no workaround for this.

[filelogreceiver_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.66.0/receiver/filelogreceiver/README.md
