// Copyright 2021 Sumo Logic, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sumologicexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslateConfigValue(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic",
			input: "%{k8s.pod.name}-%{host.name}",
			want:  "%{pod}-%{host}",
		},
		{
			name:  "basic with sumo convention tags",
			input: "%{k8s.pod.name}-%{host.name}/%{pod}-%{host}",
			want:  "%{pod}-%{host}/%{pod}-%{host}",
		},
		{
			name:  "custom attributes",
			input: "%{_sourceCategory}-%{my_custom_vendor_attr}",
			want:  "%{_sourceCategory}-%{my_custom_vendor_attr}",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, translateConfigValue(tc.input))
		})
	}
}
