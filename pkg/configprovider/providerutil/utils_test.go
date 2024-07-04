package providerutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/confmap"
)

func TestPrepareForReplaceBehavior(t *testing.T) {
	tests := []struct {
		name         string
		srcConf      *confmap.Conf
		mergeConf    *confmap.Conf
		expectedConf *confmap.Conf
	}{
		{
			name: "Remove matching key path from source map",
			srcConf: confmap.NewFromStringMap(map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"childKey2": "value2",
						"collector_fields": map[string]interface{}{
							"cluster": "cluster-1",
						},
					},
				},
				"anotherKey": "anotherValue",
			}),
			mergeConf: confmap.NewFromStringMap(map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"collector_fields": map[string]interface{}{
							"zone": "eu",
						},
						"childKey2": "value2",
					},
				},
				"anotherKey": "anotherValue",
			}),
			expectedConf: confmap.NewFromStringMap(map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"childKey2": "value2",
					},
				},
				"anotherKey": "anotherValue",
			}),
		},
		{
			name: "No matching key paths to remove, source map remains unaffected",
			srcConf: confmap.NewFromStringMap(map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"collector_fields1": map[string]interface{}{
							"cluster": "cluster-1",
						},
					},
				},
				"anotherKey": "anotherValue",
			}),
			mergeConf: confmap.NewFromStringMap(map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"collector_fields2": map[string]interface{}{
							"zone": "eu",
						},
					},
				},
				"anotherKey": "anotherValue",
			}),
			expectedConf: confmap.NewFromStringMap(map[string]interface{}{
				"extensions": map[string]interface{}{
					"sumologic": map[string]interface{}{
						"collector_fields1": map[string]interface{}{
							"cluster": "cluster-1",
						},
					},
				},
				"anotherKey": "anotherValue",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PrepareForReplaceBehavior(tt.srcConf, tt.mergeConf)
			assert.Equal(t, tt.expectedConf, tt.srcConf)
		})
	}
}
