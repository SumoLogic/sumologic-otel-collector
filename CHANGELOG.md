# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- docs: clarify status of sumologicextension [#553][#553]
- chore(deps): bump golang from 1.18 to 1.18.1 [#546][#546]
- chore: bump OT core to v0.49.0 [#550][#550]

[Unreleased]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.48.0-sumo-0...main
[#546]: https://github.com/SumoLogic/sumologic-otel-collector/pull/546
[#550]: https://github.com/SumoLogic/sumologic-otel-collector/pull/550
[#553]: https://github.com/SumoLogic/sumologic-otel-collector/pull/553

## [v0.48.0-sumo-0]

### Released 2022-04-12

### Changed

- chore: bump OT core to v0.48.0 [#534][#534]
- chore(deps): bump alpine from 3.15.3 to 3.15.4 [#525][#525]

### Fixed

- fix(sumologicexporter): treat resource attributes as fields for otlp #536

### Other

- refactor(sumologicexporter): use golang.org/x/exp/slices for sorting fields [#519][#519]
- refactor(sumologicextension): use bytes slices and strings.Builder to decrease allocations [#530][#530]

[v0.48.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/releases/tag/v0.48.0-sumo-0
[#519]: https://github.com/SumoLogic/sumologic-otel-collector/pull/519
[#525]: https://github.com/SumoLogic/sumologic-otel-collector/pull/525
[#530]: https://github.com/SumoLogic/sumologic-otel-collector/pull/530
[#534]: https://github.com/SumoLogic/sumologic-otel-collector/pull/534
[#536]: https://github.com/SumoLogic/sumologic-otel-collector/pull/536

## [v0.47.0-sumo-0]

### Released 2022-04-05

Welcome to the Sumo Logic OT Distro Collector!

With this release, we are officially out of beta status and in GA, as in General Availability. 🎉

This means the software is ready to be used by all customers (without signing up for beta program)
and is commercially fully supported by Sumo Logic in production environments.

Starting with this release, we are using upstream [OpenTelemetry Collector][otc] version numbers
as the base for the OT Distro version numbers.
This means that Sumo Logic OT Distro Collector `v0.47.0-sumo-0` is based on `v0.47.0`
of the OpenTelemetry Collector [core][otc_v0_47_0] and [contrib][contrib_v0_47_0] packages.

[otc]: https://github.com/open-telemetry/opentelemetry-collector
[otc_v0_47_0]: https://github.com/open-telemetry/opentelemetry-collector/releases/v0.47.0
[contrib_v0_47_0]: https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/v0.47.0

### Fixed

- fix(k8sprocessor): fix metadata dependencies by @astencel-sumo [#513]

### Other

- refactor(sumologicexporter): optimize fields stringification by @pmalek-sumo [#517]
- refactor(sumologicexporter): optimize compressor using sync.Pool by @pmalek-sumo [#518]

All changes: [v0.0.58-beta.0...v0.47.0-sumo-0]

[v0.47.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.0.58-beta.0...v0.47.0-sumo-0
[#513]: https://github.com/SumoLogic/sumologic-otel-collector/pull/513
[#517]: https://github.com/SumoLogic/sumologic-otel-collector/pull/517
[#518]: https://github.com/SumoLogic/sumologic-otel-collector/pull/518
[v0.0.58-beta.0...v0.47.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.0.58-beta.0...v0.47.0-sumo-0
