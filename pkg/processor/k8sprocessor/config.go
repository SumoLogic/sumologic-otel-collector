// Copyright 2020 OpenTelemetry Authors
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

package k8sprocessor

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
)

// Config defines configuration for k8s attributes processor.
type Config struct {
	k8sconfig.APIConfig `mapstructure:",squash"`

	// Passthrough mode only annotates resources with the pod IP and
	// does not try to extract any other metadata. It does not need
	// access to the K8S cluster API. Agent/Collector must receive spans
	// directly from services to be able to correctly detect the pod IPs.
	Passthrough bool `mapstructure:"passthrough"`

	// OwnerLookupEnabled enables pulling owner data, which triggers
	// additional calls to Kubernetes API
	OwnerLookupEnabled bool `mapstructure:"owner_lookup_enabled"`

	// Extract section allows specifying extraction rules to extract
	// data from k8s pod specs
	Extract ExtractConfig `mapstructure:"extract"`

	// Filter section allows specifying filters to filter
	// pods by labels, fields, namespaces, nodes, etc.
	Filter FilterConfig `mapstructure:"filter"`

	// Association section allows to define rules for tagging spans, metrics,
	// and logs with Pod metadata.
	Association []PodAssociationConfig `mapstructure:"pod_association"`

	// Exclude section allows to define names of pod that should be
	// ignored while tagging.
	Exclude ExcludeConfig `mapstructure:"exclude"`
}

func (cfg *Config) Validate() error {
	return cfg.APIConfig.Validate()
}

// ExtractConfig section allows specifying extraction rules to extract
// data from k8s pod specs.
type ExtractConfig struct {
	// Metadata allows to extract pod metadata from a list of metadata fields.
	// The field accepts a list of strings.
	//
	// Metadata fields supported right now are,
	//   namespace, podName, podUID, deployment, node and startTime
	//
	// Specifying anything other than these values will result in an error.
	// By default all of the fields are extracted and added to spans and metrics.
	Metadata []string `mapstructure:"metadata"`

	// Tags allow to specify output name used for each of the kubernetes tags
	// The field accepts a map of string->string. It is optional and if no values
	// are provided, defaults will be used
	Tags map[string]string `mapstructure:"tags"`

	// Annotations allows extracting data from pod annotations and record it
	// as resource attributes.
	// It is a list of FieldExtractConfig type. See FieldExtractConfig
	// documentation for more details.
	Annotations []FieldExtractConfig `mapstructure:"annotations"`

	// Labels allows extracting data from pod labels and record it
	// as resource attributes.
	// It is a list of FieldExtractConfig type. See FieldExtractConfig
	// documentation for more details.
	Labels []FieldExtractConfig `mapstructure:"labels"`

	// NamespaceAnnotations allows extracting data from namespace annotations and record it
	// as resource attributes.
	// It is a list of FieldExtractConfig type. See FieldExtractConfig
	// documentation for more details.
	NamespaceAnnotations []FieldExtractConfig `mapstructure:"namespace_annotations"`

	// NamespaceLabels allows extracting data from namespace labels and record it
	// as resource attributes.
	// It is a list of FieldExtractConfig type. See FieldExtractConfig
	// documentation for more details.
	NamespaceLabels []FieldExtractConfig `mapstructure:"namespace_labels"`

	// Delimiter is going to be used to join multiple values for metadata.
	// For example if given pod is associated with more than one service,
	// delimiter is going to separate them in string.
	Delimiter string `mapstructure:"delimiter"`
}

//FieldExtractConfig allows specifying an extraction rule to extract a value from exactly one field.
//
// The field accepts a list FilterExtractConfig map. The map accepts three keys
//     tag_name, key and regex
//
// - tag_name represents the name of the tag that will be added to the span.
//   When not specified a default tag name will be used of the format:
//       k8s.pod.annotation.<annotation key>
//       k8s.pod.label.<label key>
//   For example, if tag_name is not specified and the key is git_sha,
//   then the attribute name will be `k8s.pod.annotation.git_sha`.
//
//- key represents the annotation name. This must exactly match an annotation name.
//  To capture all keys, `*` can be used
//
//- regex is an optional field used to extract a sub-string from a complex field value.
//  The supplied regular expression must contain one named parameter with the string "value"
//  as the name. For example, if your pod spec contains the following annotation,
//
//		kubernetes.io/change-cause: 2019-08-28T18:34:33Z APP_NAME=my-app GIT_SHA=58a1e39 CI_BUILD=4120
//
//  and you'd like to extract the GIT_SHA and the CI_BUILD values as tags, then you must
//  specify the following two extraction rules:
//
//  procesors:
//    k8s-tagger:
//      annotations:
//        - tag_name: git.sha
//          key: kubernetes.io/change-cause
//          regex: GIT_SHA=(?P<value>\w+)
//        - tag_name: ci.build
//	         key: kubernetes.io/change-cause
//          regex: JENKINS=(?P<value>[\w]+)
//
//  this will add the `git.sha` and `ci.build` tags to the spans.
//
//  It is also possible to generically fetch all keys and fill them into a template.
//  To substitute the original name, use `%s`. For example:
//
//  procesors:
//    k8s-tagger:
//      annotations:
//        - tag_name: k8s.annotation/%s
//          key: *

type FieldExtractConfig struct {
	TagName string `mapstructure:"tag_name"`
	Key     string `mapstructure:"key"`
	Regex   string `mapstructure:"regex"`
}

// FilterConfig section allows specifying filters to filter
// pods by labels, fields, namespaces, nodes, etc.
type FilterConfig struct {
	// Node represents a k8s node or host. If specified, any pods not running
	// on the specified node will be ignored by the tagger.
	Node string `mapstructure:"node"`

	// NodeFromEnv can be used to extract the node name from an environment
	// variable. The value must be the name of the environment variable.
	// This is useful when the node a Otel agent will run on cannot be
	// predicted. In such cases, the Kubernetes downward API can be used to
	// add the node name to each pod as an environment variable. K8s tagger
	// can then read this value and filter pods by it.
	//
	// For example, node name can be passed to each agent with the downward API as follows
	//
	// env:
	//   - name: K8S_NODE_NAME
	//     valueFrom:
	//       fieldRef:
	//         fieldPath: spec.nodeName
	//
	// Then the NodeFromEnv field can be set to `K8S_NODE_NAME` to filter all pods by the node that
	// the agent is running on.
	//
	// More on downward API here: https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/
	NodeFromEnvVar string `mapstructure:"node_from_env_var"`

	// Namespace filters all pods by the provided namespace. All other pods are ignored.
	Namespace string `mapstructure:"namespace"`

	// Fields allows to filter pods by generic k8s fields.
	// Only the following operations are supported:
	//    - equals
	//    - not-equals
	//
	// Check FieldFilterConfig for more details.
	Fields []FieldFilterConfig `mapstructure:"fields"`

	// Labels allows to filter pods by generic k8s pod labels.
	// Only the following operations are supported:
	//    - equals
	//    - not-equals
	//    - exists
	//    - not-exists
	//
	// Check FieldFilterConfig for more details.
	Labels []FieldFilterConfig `mapstructure:"labels"`
}

// FieldFilterConfig allows specifying exactly one filter by a field.
// It can be used to represent a label or generic field filter.
type FieldFilterConfig struct {
	// Key represents the key or name of the field or labels that a filter
	// can apply on.
	Key string `mapstructure:"key"`

	// Value represents the value associated with the key that a filter
	// operation specified by the `Op` field applies on.
	Value string `mapstructure:"value"`

	// Op represents the filter operation to apply on the given
	// Key: Value pair. The following operations are supported
	//   equals, not-equals, exists, does-not-exist.
	Op string `mapstructure:"op"`
}

// PodAssociationConfig contain single rule how to associate Pod metadata
// with logs, spans and metrics
type PodAssociationConfig struct {
	// From represents the source of the association.
	// Allowed values are "connection" and "labels".
	From string `mapstructure:"from"`

	// Name represents extracted key name.
	// e.g. ip, pod_uid, k8s.pod.ip
	Name string `mapstructure:"name"`
}

// DefaultDelimiter is default value for Delimiter for ExtractConfig
const DefaultDelimiter string = ", "

// ExcludeConfig represent a list of Pods to exclude
type ExcludeConfig struct {
	Pods []ExcludePodConfig `mapstructure:"pods"`
}

// ExcludePodConfig represent a Pod name to ignore
type ExcludePodConfig struct {
	Name string `mapstructure:"name"`
}
