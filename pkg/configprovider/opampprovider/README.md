# OpAMP configuration provider

The OpAMP configuration provider reads an initial configuration file that
contains a URL to an OpAMP server. The OpAMP server can push configuration
changes to the provider, which will update the configuration file.

To use the OpAMP configuration provider, it is necessary to use the
--opamp-config flag (and *not* the --config flag):

```bash
./otelcol --opamp-config "/etc/config/otel/integrations/opamp.yml"
```

Unlike other configuration providers which can be used alongside one another,
the OpAMP provider must be used exclusively.
