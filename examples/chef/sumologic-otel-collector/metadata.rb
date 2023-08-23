name 'sumologic-otel-collector'
maintainer 'Sumo Logic'
maintainer_email 'collection@sumologic.com'
license 'Apache-2.0'
description 'Installs sumologic-otel-collector'
long_description IO.read(File.join(File.dirname(__FILE__), 'README.md'))
version '0.0.0'
chef_version '>= 11' if respond_to?(:chef_version)
