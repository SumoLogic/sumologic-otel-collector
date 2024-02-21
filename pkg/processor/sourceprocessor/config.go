// Copyright 2019 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sourceprocessor

// Config defines configuration for Source processor.
type Config struct {
	Collector                 string `mapstructure:"collector"`
	SourceHost                string `mapstructure:"source_host"`
	SourceName                string `mapstructure:"source_name"`
	SourceCategory            string `mapstructure:"source_category"`
	SourceCategoryPrefix      string `mapstructure:"source_category_prefix"`
	SourceCategoryReplaceDash string `mapstructure:"source_category_replace_dash"`

	// Exclude is a mapping of field names to exclusion regexes for those
	// particular fields.
	// Whenever a value for a particular field matches a corresponding regex,
	// the processed entry is dropped.
	Exclude map[string]string `mapstructure:"exclude"`

	AnnotationPrefix          string `mapstructure:"annotation_prefix"`
	NamespaceAnnotationPrefix string `mapstructure:"namespace_annotation_prefix"`
	PodKey                    string `mapstructure:"pod_key"`
	PodNameKey                string `mapstructure:"pod_name_key"`
	PodTemplateHashKey        string `mapstructure:"pod_template_hash_key"`

	ContainerAnnotations ContainerAnnotationsConfig `mapstructure:"container_annotations"`
}

type ContainerAnnotationsConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	ContainerNameKey string   `mapstructure:"container_name_key"`
	Prefixes         []string `mapstructure:"prefixes"`
}
