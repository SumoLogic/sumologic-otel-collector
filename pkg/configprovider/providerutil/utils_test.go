package providerutil

import (
	"go.opentelemetry.io/collector/confmap"
	"reflect"
	"testing"
)

func TestPrepareForReplaceBehavior(t *testing.T) {
	tests := []struct {
		name        string
		srcMap      map[string]interface{}
		mergeMap    map[string]interface{}
		expectedMap map[string]interface{}
	}{
		{
			name: "Remove matching key path from source map",
			srcMap: map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"collector_fields": map[string]interface{}{
							"cluster": "cluster-1",
						},
						"childKey2": "value2",
					},
				},
				"anotherKey": "anotherValue",
			},
			mergeMap: map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"collector_fields": map[string]interface{}{
							"zone": "eu",
						},
						"childKey2": "value2",
					},
				},
				"anotherKey": "anotherValue",
			},
			expectedMap: map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"childKey2": "value2",
					},
				},
				"anotherKey": "anotherValue",
			},
		},
		{
			name: "No matching key paths to remove, source map remains unaffected",
			srcMap: map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"collector_fields1": map[string]interface{}{
							"cluster": "cluster-1",
						},
					},
				},
				"anotherKey": "anotherValue",
			},
			mergeMap: map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"collector_fields2": map[string]interface{}{
							"zone": "eu",
						},
					},
				},
				"anotherKey": "anotherValue",
			},
			expectedMap: map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"collector_fields1": map[string]interface{}{
							"cluster": "cluster-1",
						},
					},
				},
				"anotherKey": "anotherValue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcConf := confmap.NewFromStringMap(tt.srcMap)
			mergeConf := confmap.NewFromStringMap(tt.mergeMap)
			PrepareForReplaceBehavior(srcConf, mergeConf)
			if !reflect.DeepEqual(srcConf.ToStringMap(), tt.expectedMap) {
				t.Errorf("PrepareForReplaceBehavior() = %v, want %v", srcConf.ToStringMap(), tt.expectedMap)
			}
		})
	}
}
