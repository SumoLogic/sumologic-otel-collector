// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package bigendianconverter

import (
	"encoding/binary"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

// NOTE:
// This code was copied over from:
// https://github.com/open-telemetry/opentelemetry-collector/blob/v0.28.0/internal/idutils/big_endian_converter.go
// to allow processor tests to still run as they used to.

// UInt64ToTraceID converts the pair of uint64 representation of a TraceID to pcommon.TraceID.
func UInt64ToTraceID(high, low uint64) pcommon.TraceID {
	traceID := [16]byte{}
	binary.BigEndian.PutUint64(traceID[:8], high)
	binary.BigEndian.PutUint64(traceID[8:], low)
	return pcommon.TraceID(traceID)
}

// SpanIDToUInt64 converts the pcommon.SpanID to uint64 representation.
func SpanIDToUInt64(spanID pcommon.SpanID) uint64 {
	bytes := spanID
	return binary.BigEndian.Uint64(bytes[:])
}

// UInt64ToSpanID converts the uint64 representation of a SpanID to pcommon.SpanID.
func UInt64ToSpanID(id uint64) pcommon.SpanID {
	spanID := [8]byte{}
	binary.BigEndian.PutUint64(spanID[:], id)
	return pcommon.SpanID(spanID)
}
