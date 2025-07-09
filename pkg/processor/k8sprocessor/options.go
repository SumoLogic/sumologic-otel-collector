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
	"fmt"
	"os"
	"regexp"
	"strings"

	conventions "go.opentelemetry.io/otel/semconv/v1.18.0"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/kube"
)

const (
	filterOPEquals       = "equals"
	filterOPNotEquals    = "not-equals"
	filterOPExists       = "exists"
	filterOPDoesNotExist = "does-not-exist"

	metadataContainerID     = "containerId"
	metadataContainerName   = "containerName"
	metadataContainerImage  = "containerImage"
	metadataCronJobName     = "cronJobName"
	metadataDaemonSetName   = "daemonSetName"
	metadataDeploymentName  = "deploymentName"
	metadataHostName        = "hostName"
	metadataJobName         = "jobName"
	metadataNamespace       = "namespace"
	metadataNodeName        = "nodeName"
	metadataPodID           = "podId"
	metadataPodName         = "podName"
	metadataReplicaSetName  = "replicaSetName"
	metadataServiceName     = "serviceName"
	metadataStartTime       = "startTime"
	metadataStatefulSetName = "statefulSetName"

	metadataOtelSemconvServiceName = "k8s.service.name"  // no semantic convention for service name as of right now, but this is reasonable
	metadataOtelPodStartTime       = "k8s.pod.startTime" // no semantic convention for this, but keeping a similar format for consistency
	deprecatedMetadataClusterName  = "clusterName"
)

// Option represents a configuration option that can be passes.
// to the k8s-tagger
type Option func(*kubernetesprocessor) error

// WithAPIConfig provides k8s API related configuration to the processor.
// It defaults the authentication method to in-cluster auth using service accounts.
func WithAPIConfig(cfg k8sconfig.APIConfig) Option {
	return func(p *kubernetesprocessor) error {
		p.apiConfig = cfg
		return p.apiConfig.Validate()
	}
}

// WithPassthrough enables passthrough mode. In passthrough mode, the processor
// only detects and tags the pod IP and does not invoke any k8s APIs.
func WithPassthrough() Option {
	return func(p *kubernetesprocessor) error {
		p.passthroughMode = true
		return nil
	}
}

// WithOwnerLookupEnabled makes the processor pull additional owner data from K8S API
func WithOwnerLookupEnabled() Option {
	return func(p *kubernetesprocessor) error {
		p.rules.OwnerLookupEnabled = true
		return nil
	}
}

// WithExtractMetadata allows specifying options to control extraction of pod metadata.
// If no fields explicitly provided, all metadata extracted by default.
func WithExtractMetadata(fields ...string) Option {
	return func(p *kubernetesprocessor) error {
		if len(fields) == 0 {
			fields = []string{
				metadataContainerID,
				metadataContainerImage,
				metadataContainerName,
				metadataDaemonSetName,
				metadataDeploymentName,
				metadataHostName,
				metadataNamespace,
				metadataNodeName,
				metadataPodName,
				metadataPodID,
				metadataReplicaSetName,
				metadataServiceName,
				metadataStartTime,
				metadataStatefulSetName,
			}
		}
		for _, field := range fields {
			switch field {
			case metadataContainerID, string(conventions.ContainerIDKey):
				p.rules.ContainerID = true
			case metadataContainerImage, string(conventions.ContainerImageNameKey):
				p.rules.ContainerImage = true
			case metadataContainerName, string(conventions.ContainerNameKey):
				p.rules.ContainerName = true
			case metadataCronJobName, string(conventions.K8SCronJobNameKey):
				p.rules.CronJobName = true
			case metadataDaemonSetName, string(conventions.K8SDaemonSetNameKey):
				p.rules.DaemonSetName = true
			case metadataDeploymentName, string(conventions.K8SDeploymentNameKey):
				p.rules.DeploymentName = true
			case metadataHostName, string(conventions.HostNameKey):
				p.rules.HostName = true
			case metadataJobName, string(conventions.K8SJobNameKey):
				p.rules.JobName = true
			case metadataNamespace, string(conventions.K8SNamespaceNameKey):
				p.rules.Namespace = true
			case metadataNodeName, string(conventions.K8SNodeNameKey):
				p.rules.NodeName = true
			case metadataPodID, string(conventions.K8SPodUIDKey):
				p.rules.PodUID = true
			case metadataPodName, string(conventions.K8SPodNameKey):
				p.rules.PodName = true
			case metadataReplicaSetName, string(conventions.K8SReplicaSetNameKey):
				p.rules.ReplicaSetName = true
			case metadataServiceName, metadataOtelSemconvServiceName:
				p.rules.ServiceName = true
			case metadataStartTime, metadataOtelPodStartTime:
				p.rules.StartTime = true
			case metadataStatefulSetName, string(conventions.K8SStatefulSetNameKey):
				p.rules.StatefulSetName = true
			case deprecatedMetadataClusterName, string(conventions.K8SClusterNameKey):
				p.logger.Warn("clusterName metadata field has been deprecated and will be removed soon")
			default:
				return fmt.Errorf("\"%s\" is not a supported metadata field", field)
			}
		}
		return nil
	}
}

// WithExtractTags allows specifying custom tag names
func WithExtractTags(tagsMap map[string]string) Option {
	return func(p *kubernetesprocessor) error {
		var tags = kube.NewExtractionFieldTags()
		for field, tag := range tagsMap {
			switch strings.ToLower(field) {
			case strings.ToLower(metadataContainerID):
				tags.ContainerID = tag
			case strings.ToLower(metadataContainerName):
				tags.ContainerName = tag
			case strings.ToLower(metadataContainerImage):
				tags.ContainerImage = tag
			case strings.ToLower(metadataDaemonSetName):
				tags.DaemonSetName = tag
			case strings.ToLower(metadataDeploymentName):
				tags.DeploymentName = tag
			case strings.ToLower(metadataHostName):
				tags.HostName = tag
			case strings.ToLower(metadataNamespace):
				tags.Namespace = tag
			case strings.ToLower(metadataNodeName):
				tags.NodeName = tag
			case strings.ToLower(metadataPodID):
				tags.PodUID = tag
			case strings.ToLower(metadataPodName):
				tags.PodName = tag
			case strings.ToLower(metadataReplicaSetName):
				tags.ReplicaSetName = tag
			case strings.ToLower(metadataServiceName):
				tags.ServiceName = tag
			case strings.ToLower(metadataStartTime):
				tags.StartTime = tag
			case strings.ToLower(metadataStatefulSetName):
				tags.StatefulSetName = tag
			case strings.ToLower(deprecatedMetadataClusterName):
				p.logger.Warn("clusterName metadata field has been deprecated and will be removed soon")
			default:
				return fmt.Errorf("\"%s\" is not a supported metadata field", field)
			}
		}
		p.rules.Tags = tags
		return nil
	}
}

// WithExtractLabels allows specifying options to control extraction of pod labels.
func WithExtractLabels(labels ...FieldExtractConfig) Option {
	return func(p *kubernetesprocessor) error {
		labels, err := extractFieldRules("labels", labels...)
		if err != nil {
			return err
		}
		p.rules.Labels = labels
		return nil
	}
}

// WithExtractNamespaceLabels allows specifying options to control extraction of namespace labels.
func WithExtractNamespaceLabels(labels ...FieldExtractConfig) Option {
	return func(p *kubernetesprocessor) error {
		labels, err := extractFieldRules("namespace_labels", labels...)
		if err != nil {
			return err
		}
		p.rules.NamespaceLabels = labels
		return nil
	}
}

// WithExtractAnnotations allows specifying options to control extraction of pod annotations tags.
func WithExtractAnnotations(annotations ...FieldExtractConfig) Option {
	return func(p *kubernetesprocessor) error {
		annotations, err := extractFieldRules("annotations", annotations...)
		if err != nil {
			return err
		}
		p.rules.Annotations = annotations
		return nil
	}
}

// WithExtractNamespaceAnnotations allows specifying options to control extraction of namespace annotations tags.
func WithExtractNamespaceAnnotations(annotations ...FieldExtractConfig) Option {
	return func(p *kubernetesprocessor) error {
		annotations, err := extractFieldRules("namespace_annotations", annotations...)
		if err != nil {
			return err
		}
		p.rules.NamespaceAnnotations = annotations
		return nil
	}
}

func extractFieldRules(fieldType string, fields ...FieldExtractConfig) ([]kube.FieldExtractionRule, error) {
	rules := []kube.FieldExtractionRule{}
	for _, a := range fields {
		name := a.TagName
		if name == "" {
			if a.Key == "*" {
				name = fmt.Sprintf("k8s.%s.%%s", fieldType)
			} else {
				name = fmt.Sprintf("k8s.%s.%s", fieldType, a.Key)
			}
		}

		var r *regexp.Regexp
		if a.Regex != "" {
			var err error
			r, err = regexp.Compile(a.Regex)
			if err != nil {
				return rules, err
			}
			names := r.SubexpNames()
			if len(names) != 2 || names[1] != "value" {
				return rules, fmt.Errorf("regex must contain exactly one named submatch (value)")
			}
		}

		rules = append(rules, kube.FieldExtractionRule{
			Name: name, Key: a.Key, Regex: r,
		})
	}
	return rules, nil
}

// WithFilterNode allows specifying options to control filtering pods by a node/host.
func WithFilterNode(node, nodeFromEnvVar string) Option {
	return func(p *kubernetesprocessor) error {
		if nodeFromEnvVar != "" {
			p.filters.Node = os.Getenv(nodeFromEnvVar)
			return nil
		}
		p.filters.Node = node
		return nil
	}
}

// WithFilterNamespace allows specifying options to control filtering pods by a namespace.
func WithFilterNamespace(ns string) Option {
	return func(p *kubernetesprocessor) error {
		p.filters.Namespace = ns
		return nil
	}
}

// WithFilterLabels allows specifying options to control filtering pods by pod labels.
func WithFilterLabels(filters ...FieldFilterConfig) Option {
	return func(p *kubernetesprocessor) error {
		labels := []kube.FieldFilter{}
		for _, f := range filters {
			if f.Op == "" {
				f.Op = filterOPEquals
			}

			var op selection.Operator
			switch f.Op {
			case filterOPEquals:
				op = selection.Equals
			case filterOPNotEquals:
				op = selection.NotEquals
			case filterOPExists:
				op = selection.Exists
			case filterOPDoesNotExist:
				op = selection.DoesNotExist
			default:
				return fmt.Errorf("'%s' is not a valid label filter operation for key=%s, value=%s", f.Op, f.Key, f.Value)
			}
			labels = append(labels, kube.FieldFilter{
				Key:   f.Key,
				Value: f.Value,
				Op:    op,
			})
		}
		p.filters.Labels = labels
		return nil
	}
}

// WithFilterFields allows specifying options to control filtering pods by pod fields.
func WithFilterFields(filters ...FieldFilterConfig) Option {
	return func(p *kubernetesprocessor) error {
		fields := []kube.FieldFilter{}
		for _, f := range filters {
			if f.Op == "" {
				f.Op = filterOPEquals
			}

			var op selection.Operator
			switch f.Op {
			case filterOPEquals:
				op = selection.Equals
			case filterOPNotEquals:
				op = selection.NotEquals
			default:
				return fmt.Errorf("'%s' is not a valid field filter operation for key=%s, value=%s", f.Op, f.Key, f.Value)
			}
			fields = append(fields, kube.FieldFilter{
				Key:   f.Key,
				Value: f.Value,
				Op:    op,
			})
		}
		p.filters.Fields = fields
		return nil
	}
}

// WithExtractPodAssociations allows specifying options to associate pod metadata with incoming resource
func WithExtractPodAssociations(podAssociations ...PodAssociationConfig) Option {
	return func(p *kubernetesprocessor) error {
		associations := make([]kube.Association, 0, len(podAssociations))
		for _, association := range podAssociations {
			associations = append(associations, kube.Association{
				From: association.From,
				Name: association.Name,
			})
		}
		p.podAssociations = associations
		return nil
	}
}

// WithDelimiter sets delimiter to use by kubernetesprocessor
func WithDelimiter(delimiter string) Option {
	return func(p *kubernetesprocessor) error {
		p.delimiter = delimiter
		return nil
	}
}

// WithLimit sets the limit to use by kubernetesprocessor
func WithLimit(limit int) Option {
	return func(p *kubernetesprocessor) error {
		p.limit = limit
		return nil
	}
}

// WithExcludes allows specifying pods to exclude
func WithExcludes(excludeConfig ExcludeConfig) Option {
	return func(p *kubernetesprocessor) error {
		excludes := kube.Excludes{}
		names := excludeConfig.Pods

		for _, name := range names {
			excludes.Pods = append(excludes.Pods, kube.ExcludePods{
				Name: regexp.MustCompile(name.Name)},
			)
		}

		p.podIgnore = excludes
		return nil
	}
}
