package metricfrequencyprocessor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestAccumulate(t *testing.T) {
	sieve := newMetricSieve(createDefaultConfig().(*Config))
	var timestamp = time.Unix(0, 0)
	setupHistory(sieve, map[time.Time]float64{timestamp: 0.0})

	result := sieve.Sift(dataPointsToMetric(map[time.Time]float64{
		timestamp.Add(1 * time.Minute): 0.0,
	}))

	// metric is not filtered, because sieve is still accumulating points
	assert.False(t, result)
}

func TestIsConstant(t *testing.T) {
	type testCase struct {
		dataPoint     pmetric.NumberDataPoint
		values        map[int64]float64
		expectedValue bool
	}

	testCases := []*testCase{
		{
			dataPoint: createDataPoint(time.Unix(0.0, 0.0), 0.0),
			values: map[int64]float64{
				1.0: 0.0,
				2.0: 0.0,
			},
			expectedValue: true,
		},
		{
			dataPoint: createDataPoint(time.Unix(0.0, 0.0), 0.0),
			values: map[int64]float64{
				1.0: 0.0,
				2.0: 1.0,
			},
			expectedValue: false,
		},
		{
			dataPoint: createDataPoint(time.Unix(0.0, 0.0), 1.0),
			values: map[int64]float64{
				1.0: 0.0,
				2.0: 1.0,
			},
			expectedValue: false,
		},
		{
			dataPoint: createDataPoint(time.Unix(0.0, 0.0), 0.0),
			values: map[int64]float64{
				1.0: 0.0 + float64EqualityThreshold/10,
			},
			expectedValue: true,
		},
	}

	for _, test := range testCases {
		result := isConstant(test.dataPoint, unixPointsToPdata(test.values))
		assert.Equal(t, result, test.expectedValue)
	}
}

func TestIsLowInfo(t *testing.T) {
	type testCase struct {
		values        map[int64]float64
		expectedValue bool
	}

	testCases := []*testCase{
		{
			values: map[int64]float64{
				0: 0.0,
				1: 1.0,
				2: 2.0,
				3: 3.0,
				4: 4.0,
			},
			expectedValue: true,
		},
		{
			values: map[int64]float64{
				0: 0.0,
				1: 4.0,
				2: 1.0,
				3: 3.0,
				4: 2.0,
			},
			expectedValue: false,
		},
		{
			values: map[int64]float64{
				0: 0.0,
				1: 1.0,
				2: 2.0,
				3: 3.0,
				4: 20.0,
			},
			expectedValue: false,
		},
	}

	sieve := newMetricSieve(createDefaultConfig().(*Config))

	for _, test := range testCases {
		result := sieve.isLowInformation(unixPointsToPdata(test.values))
		assert.Equal(t, result, test.expectedValue)
	}
}

func TestQuantiles(t *testing.T) {
	type testCase struct {
		values          map[int64]float64
		expectedQ1Value float64
		expectedQ3Value float64
	}

	testCases := []*testCase{
		{
			values: map[int64]float64{
				0: 0.0,
				1: 1.0,
				2: 2.0,
				3: 3.0,
				4: 4.0,
			},
			expectedQ1Value: 1.0,
			expectedQ3Value: 3.0,
		},
		{
			values: map[int64]float64{
				0: 1.0,
				1: 4.0,
				2: 0.0,
				3: 2.0,
				4: 3.0,
			},
			expectedQ1Value: 1.0,
			expectedQ3Value: 3.0,
		},
	}

	for _, test := range testCases {
		resultQ1, resultQ3 := calculateQ1Q3(unixPointsToPdata(test.values))
		assert.True(t, almostEqual(resultQ1, test.expectedQ1Value))
		assert.True(t, almostEqual(resultQ3, test.expectedQ3Value))
	}
}

func TestVariation(t *testing.T) {
	type testCase struct {
		values        map[int64]float64
		expectedValue float64
	}

	testCases := []*testCase{
		{
			values: map[int64]float64{
				0: 0.0,
				1: 0.0,
			},
			expectedValue: 0.0,
		},
		{
			values: map[int64]float64{
				0: 0.0,
				1: 1.0,
			},
			expectedValue: 1.0,
		},
		{
			values: map[int64]float64{
				0: 0.0,
				1: 0.5,
				2: 1,
			},
			expectedValue: 1.0,
		},
		{
			values: map[int64]float64{
				0: 0.0,
				1: 1.0,
				2: 0.0,
			},
			expectedValue: 2.0,
		},
		{
			values: map[int64]float64{
				0: 1.0,
				1: 0.0,
				2: 1.0,
			},
			expectedValue: 2.0,
		},
	}

	for _, test := range testCases {
		result := calculateVariation(unixPointsToPdata(test.values))
		assert.True(t, almostEqual(result, test.expectedValue))
	}
}

func unixPointsToPdata(points map[int64]float64) map[pcommon.Timestamp]float64 {
	out := make(map[pcommon.Timestamp]float64)
	for unix, value := range points {
		timestamp := pcommon.NewTimestampFromTime(time.Unix(unix, 0))
		out[timestamp] = value
	}

	return out
}

func createDataPoint(timestamp time.Time, value float64) pmetric.NumberDataPoint {
	pdataTimestamp := pcommon.NewTimestampFromTime(timestamp)
	out := pmetric.NewNumberDataPoint()
	out.SetTimestamp(pdataTimestamp)
	out.SetDoubleValue(value)
	return out
}

func setupHistory(sieve metricSieve, dataPoints map[time.Time]float64) {
	sieve.Sift(dataPointsToMetric(dataPoints))
}

func dataPointsToMetric(dataPoints map[time.Time]float64) pmetric.Metric {
	out := pmetric.NewMetric()
	out.SetName("test")
	out.SetEmptyGauge()
	target := out.Gauge().DataPoints()
	for timestamp, value := range dataPoints {
		createDataPoint(timestamp, value).CopyTo(target.AppendEmpty())
	}
	return out
}
