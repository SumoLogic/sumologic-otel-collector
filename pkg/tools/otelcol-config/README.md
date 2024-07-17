# otelcol-config

`otelcol-config` manipulates files in /etc/otelcol-config/conf.d (or in a
user-specified config directory).

It is used by the install.sh script to configure the collector for first-time
use, and also to adjust the collector configuration after installation.

Run `otelcol-config --help` for usage information.

## Contributing

Please make every effort to keep this tool as simple as possible. Here are
some rules to consider when making contributions:

1. Use the standard library whenever possible. Do not add third-party dependencies
where they can be avoided. If third party dependencies cannot be avoided, please
prefer dependencies from golang.org/x over other libraries.
2. Do not write long-running tests. Tests that take longer than 1s to run
should not be added to this module.
3. Do write exhaustive unit tests. Quality is paramount and branch coverage
of business logic should be thorough.
