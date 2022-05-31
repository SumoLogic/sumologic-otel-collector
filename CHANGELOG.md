# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Breaking changes

- feat(sumologicexporter)!: remove support for Carbon2 metrics format [#590][#590]
- feat(sumologicexporter)!: remove support for Graphite metrics format [#592][#592]

[Unreleased]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.51.0-sumo-0...main
[#590]: https://github.com/SumoLogic/sumologic-otel-collector/pull/590
[#592]: https://github.com/SumoLogic/sumologic-otel-collector/pull/592

## [v0.51.0-sumo-0]

### Released 2022-05-19

See [Upgrade guide][upgrade_guide_v0_51_0] for the breaking changes in this version.

### Breaking changes

- fix(k8sprocessor)!: remove `clusterName` metadata extraction option [#578] ([upgrade guide][upgrade_guide_v0_51_0_cluster_name])
- feat(sumologicexporter)!: attribute translation: change `file.path.resolved` to `log.file.path_resolved` [#579] ([upgrade guide][upgrade_guide_v0_51_0_attribute_translation])

### Added

- feat: enable rawk8seventsreceiver [#576]
- feat: add sumo ic marshaler [#596]

### Changed

- chore(deps): update OT core to v0.51.0 [#580]
- chore(deps): update Telegraf to v1.22.0-sumo-4 [#580]

### Fixed

- fix: fix(cascadingfilterprocessor): do not attach sampling.rule attribute if trace accept rules are not specified [#575][#575]

[v0.51.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.50.0-sumo-0...v0.51.0-sumo-0
[upgrade_guide_v0_51_0]: ./docs/Upgrading.md#upgrading-to-v0510-sumo-0
[upgrade_guide_v0_51_0_cluster_name]: ./docs/Upgrading.md#k8s_tagger-processor-removed-clustername-metadata-extraction-option
[upgrade_guide_v0_51_0_attribute_translation]: ./docs/Upgrading.md#sumologic-exporter-metadata-translation-changed-the-attribute-that-is-translated-to-_sourcename-from-filepathresolved-to-logfilepath_resolved
[#576]: https://github.com/SumoLogic/sumologic-otel-collector/pull/576
[#575]: https://github.com/SumoLogic/sumologic-otel-collector/pull/575
[#578]: https://github.com/SumoLogic/sumologic-otel-collector/pull/578
[#579]: https://github.com/SumoLogic/sumologic-otel-collector/pull/579
[#580]: https://github.com/SumoLogic/sumologic-otel-collector/pull/580

## [v0.50.0-sumo-0]

### Released 2022-04-29

Aside from upstream changes, this release only contains a performance fix to metrics batching in the Sumo Logic exporter.
The performance improvement is very substantial, so we recommend upgrading to this version immediately after `0.49.0-sumo-0`.

### Changed

- chore: update OT core to 0.50.0 [#562][#562]

### Fixed

- fix: fix(sumologicexporter): batch metrics if source headers match [#561][#561]

[v0.50.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.49.0-sumo-0...v0.50.0-sumo-0
[#561]: https://github.com/SumoLogic/sumologic-otel-collector/pull/561
[#562]: https://github.com/SumoLogic/sumologic-otel-collector/pull/562

## [v0.49.0-sumo-0]

### Released 2022-04-26

This release includes two breaking changes. One is an upstream change to the configuration syntax of several
log receivers, most notably the `filelog` receiver. The other changes how the Sumo Logic exporter determines
metadata based on the attributes of a OpenTelemetry record. Please consult the upgrade guides linked below
for more details.

### Breaking changes

- chore: bump OT core to v0.49.0 [#550][#550] ([upgrade guide][upgrade-guide-log-collection])
- fix!(sumologicexporter): send resource attributes as fields for non-otlp, removing metadata_attributes [#549][#549] ([upgrade guide][upgrade-guide-metadata])

### Changed

- docs: clarify status of sumologicextension [#553][#553]
- chore(deps): bump golang from 1.18 to 1.18.1 [#546][#546]
- chore: bump Telegraf to v1.22.0-sumo-3 [#557][#557]

### Fixed

- fix(cascadingfilterprocessor): prevent overriding metrics in cascading filter processor - add processor tag [#539][#539]

[v0.49.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.48.0-sumo-0...v0.49.0-sumo-0
[upgrade-guide-log-collection]: docs/Upgrading.md#several-changes-to-receivers-using-opentelemetry-log-collection
[upgrade-guide-metadata]: docs/Upgrading.md#sumo-logic-exporter-metadata-handling
[#546]: https://github.com/SumoLogic/sumologic-otel-collector/pull/546
[#550]: https://github.com/SumoLogic/sumologic-otel-collector/pull/550
[#553]: https://github.com/SumoLogic/sumologic-otel-collector/pull/553
[#539]: https://github.com/SumoLogic/sumologic-otel-collector/pull/539
[#557]: https://github.com/SumoLogic/sumologic-otel-collector/pull/557
[#549]: https://github.com/SumoLogic/sumologic-otel-collector/pull/549

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
