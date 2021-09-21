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

const (
	// AttributeK8SProcessorStartTime Will be removed when new fields get merged
	// to https://github.com/open-telemetry/opentelemetry-collector/blob/main/model/semconv/opentelemetry.go
	AttributeK8SProcessorStartTime = "k8s.pod.start_time"

	// AttributeK8SContainerID and others are additional tags used in Sumo Logic version
	AttributeK8SContainerID    = "k8s.container.id"
	AttributeK8SContainerImage = "k8s.container.image"
	AttributeK8SHostName       = "k8s.pod.hostname"
	AttributeK8SServiceName    = "k8s.service.name"
)
