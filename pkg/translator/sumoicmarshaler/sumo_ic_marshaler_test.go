// Copyright 2022, OpenTelemetry Authors
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

package sumoicmarshaler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestMarshalerMissingAttributes(t *testing.T) {
	logs := plog.NewLogs()
	rls := logs.ResourceLogs().AppendEmpty()
	rls.ScopeLogs().AppendEmpty()
	marshaler := &SumoICMarshaler{"txt"}
	require.NotNil(t, marshaler)

	buf, err := marshaler.MarshalLogs(logs)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(buf))
}

func TestMarshalerMissingSourceHost(t *testing.T) {
	logs := plog.NewLogs()
	rls := logs.ResourceLogs().AppendEmpty()
	rls.Resource().Attributes().InsertString("_sourceCategory", "testcategory")

	marshaler := &SumoICMarshaler{"txt"}
	require.NotNil(t, marshaler)

	buf, err := marshaler.MarshalLogs(logs)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(buf))
}

func TestMarshalerMissingScopedLogs(t *testing.T) {
	logs := plog.NewLogs()
	rls := logs.ResourceLogs().AppendEmpty()
	rls.Resource().Attributes().InsertString("_sourceCategory", "testcategory")
	rls.Resource().Attributes().InsertString("_sourceHost", "testHost")

	marshaler := &SumoICMarshaler{"txt"}
	require.NotNil(t, marshaler)
	buf, err := marshaler.MarshalLogs(logs)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(buf))

}

func TestMarshalerMissingSourceName(t *testing.T) {
	logs := plog.NewLogs()
	rls := logs.ResourceLogs().AppendEmpty()
	rls.Resource().Attributes().InsertString("_sourceCategory", "testcategory")
	rls.Resource().Attributes().InsertString("_sourceHost", "testHost")

	sl := rls.ScopeLogs().AppendEmpty()
	const recordNum = 0

	ts := pcommon.Timestamp(int64(recordNum) * time.Millisecond.Nanoseconds())
	logRecord := sl.LogRecords().AppendEmpty()
	logRecord.Body().SetStringVal("entry1")
	logRecord.SetTimestamp(ts)

	marshaler := &SumoICMarshaler{"txt"}
	require.NotNil(t, marshaler)
	buf, err := marshaler.MarshalLogs(logs)
	assert.NoError(t, err)
	assert.Greater(t, len(buf), 0)
}

func TestMarshalerOkStructure(t *testing.T) {
	logs := plog.NewLogs()
	rls := logs.ResourceLogs().AppendEmpty()
	rls.Resource().Attributes().InsertString("_sourceCategory", "testcategory")
	rls.Resource().Attributes().InsertString("_sourceHost", "testHost")

	sl := rls.ScopeLogs().AppendEmpty()
	const recordNum = 0

	ts := pcommon.Timestamp(int64(recordNum) * time.Millisecond.Nanoseconds())
	logRecord := sl.LogRecords().AppendEmpty()
	logRecord.Body().SetStringVal("entry1")
	logRecord.Attributes().InsertString("log.file.path_resolved", "testSourceName")
	logRecord.SetTimestamp(ts)

	marshaler := &SumoICMarshaler{"txt"}
	require.NotNil(t, marshaler)
	buf, err := marshaler.MarshalLogs(logs)
	assert.NoError(t, err)
	expectedEntry := "{\"data\":\"1970-01-01 00:00:00 +0000 UTC\",\"sourceName\":\"testSourceName\",\"sourceHost\":\"testHost\""
	expectedEntry = expectedEntry + ",\"sourceCategory\":\"testcategory\",\"fields\":{\"_sourceCategory\":\"testcategory\",\"_sourceHost\":\"testHost\"},\"message\":\"entry1\"}"
	assert.Equal(t, string(buf), expectedEntry)
}
