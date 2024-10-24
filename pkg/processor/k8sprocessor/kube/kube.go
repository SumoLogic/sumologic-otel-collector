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

package kube

import (
	"regexp"
	"time"

	conventions "go.opentelemetry.io/collector/semconv/v1.18.0"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
)

const (
	podNodeField            = "spec.nodeName"
	ignoreAnnotation string = "opentelemetry.io/k8s-processor/ignore"

	defaultTagContainerID     = "k8s.container.id"
	defaultTagContainerImage  = "k8s.container.image"
	defaultTagContainerName   = "k8s.container.name"
	defaultTagDaemonSetName   = "k8s.daemonset.name"
	defaultTagHostName        = "k8s.pod.hostname"
	defaultTagCronJobName     = "k8s.cronjob.name"
	defaultTagJobName         = "k8s.job.name"
	defaultTagNodeName        = "k8s.node.name"
	defaultTagPodUID          = "k8s.pod.uid"
	defaultTagReplicaSetName  = "k8s.replicaset.name"
	defaultTagServiceName     = "k8s.service.name"
	defaultTagStatefulSetName = "k8s.statefulset.name"
	defaultTagStartTime       = "k8s.pod.startTime"
)

// PodIdentifier is a custom type to represent IP Address or Pod UID
type PodIdentifier string

const (
	DefaultPodDeleteGracePeriod = time.Second * 120
	watchSyncPeriod             = time.Minute * 5
)

// Client defines the main interface that allows querying pods by metadata.
type Client interface {
	GetPodAttributes(PodIdentifier) (map[string]string, bool)
	Start()
	Stop()
}

// ClientProvider defines a func type that returns a new Client.
type ClientProvider func(
	*zap.Logger,
	k8sconfig.APIConfig,
	ExtractionRules,
	Filters,
	[]Association,
	Excludes,
	APIClientsetProvider,
	InformerProvider,
	OwnerProvider,
	string,
	int,
	time.Duration,
	time.Duration,
) (Client, error)

// APIClientsetProvider defines a func type that initializes and return a new kubernetes
// Clientset object.
type APIClientsetProvider func(config k8sconfig.APIConfig) (kubernetes.Interface, error)

// Pod represents a kubernetes pod.
type Pod struct {
	Attributes      map[string]string
	StartTime       *metav1.Time
	Name            string
	Namespace       string
	Address         string
	PodUID          string
	Ignore          bool
	OwnerReferences *[]metav1.OwnerReference
}

func (p Pod) GetName() string {
	return p.Name
}

func (p Pod) GetNamespace() string {
	return p.Namespace
}

type deleteRequest struct {
	ts time.Time
	// id is identifier (IP address or Pod UID) of pod to remove from pods map
	id PodIdentifier
	// name contains name of pod to remove from pods map
	podName string
}

// Filters is used to instruct the client on how to filter out k8s pods.
// Right now only filters supported are the ones supported by k8s API itself
// for performance reasons. We can support adding additional custom filters
// in future if there is a real need.
type Filters struct {
	Node            string
	Namespace       string
	Fields          []FieldFilter
	Labels          []FieldFilter
	NamespaceLabels []FieldFilter
}

// FieldFilter represents exactly one filter by field rule.
type FieldFilter struct {
	// Key matches the field name.
	Key string
	// Value matches the field value.
	Value string
	// Op determines the matching operation.
	// Currently only two operations are supported,
	//  - Equals
	//  - NotEquals
	Op selection.Operator
}

// ExtractionRules is used to specify the information that needs to be extracted
// from pods and added to the spans as tags.
type ExtractionRules struct {
	ContainerID     bool
	ContainerImage  bool
	ContainerName   bool
	DaemonSetName   bool
	DeploymentName  bool
	HostName        bool
	JobName         bool
	CronJobName     bool
	PodUID          bool
	PodName         bool
	ReplicaSetName  bool
	ServiceName     bool
	StatefulSetName bool
	StartTime       bool
	Namespace       bool
	NodeName        bool

	OwnerLookupEnabled bool

	Tags                 ExtractionFieldTags
	Annotations          []FieldExtractionRule
	NamespaceAnnotations []FieldExtractionRule
	Labels               []FieldExtractionRule
	NamespaceLabels      []FieldExtractionRule
}

// ExtractionFieldTags is used to describe selected exported key names for the extracted data
type ExtractionFieldTags struct {
	ContainerID     string
	ContainerImage  string
	ContainerName   string
	DaemonSetName   string
	DeploymentName  string
	HostName        string
	CronJobName     string
	JobName         string
	PodUID          string
	PodName         string
	Namespace       string
	NodeName        string
	ReplicaSetName  string
	ServiceName     string
	StartTime       string
	StatefulSetName string
}

// NewExtractionFieldTags builds a new instance of tags with default values
func NewExtractionFieldTags() ExtractionFieldTags {
	tags := ExtractionFieldTags{}
	tags.ContainerID = defaultTagContainerID
	tags.ContainerImage = defaultTagContainerImage
	tags.ContainerName = defaultTagContainerName
	tags.DaemonSetName = defaultTagDaemonSetName
	tags.DeploymentName = conventions.AttributeK8SDeploymentName
	tags.HostName = defaultTagHostName
	tags.CronJobName = defaultTagCronJobName
	tags.JobName = defaultTagJobName
	tags.PodUID = defaultTagPodUID
	tags.PodName = conventions.AttributeK8SPodName
	tags.Namespace = conventions.AttributeK8SNamespaceName
	tags.NodeName = defaultTagNodeName
	tags.ReplicaSetName = defaultTagReplicaSetName
	tags.ServiceName = defaultTagServiceName
	tags.StartTime = defaultTagStartTime
	tags.StatefulSetName = defaultTagStatefulSetName
	return tags
}

// FieldExtractionRule is used to specify which fields to extract from pod fields
// and inject into spans as attributes.
type FieldExtractionRule struct {
	// Regex is a regular expression used to extract a sub-part of a field value.
	// Full value is extracted when no regexp is provided.
	Regex *regexp.Regexp
	// Name is used to as the Span tag name.
	Name string
	// Key is used to lookup k8s pod fields.
	Key string
}

// Associations represent a list of rules for Pod metadata associations with resources
type Associations struct {
	Associations []Association
}

// Association represents one association rule
type Association struct {
	From string
	Name string
}

// Excludes represent a list of Pods to ignore
type Excludes struct {
	Pods []ExcludePods
}

// ExcludePods represent a Pod name to ignore
type ExcludePods struct {
	Name *regexp.Regexp
}
