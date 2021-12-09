// Copyright 2021, OpenTelemetry Authors
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

package servicemapreceiver

import (
	"fmt"
	"go.opentelemetry.io/collector/model/pdata"
	"math/rand"
	"time"
)

func buildTraces(input []*ebpfmessage) pdata.Traces {
	data := &flowsPackage{
		flows:      map[string]*flowData{},
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, msg := range input {
		data.addMessage(msg)
	}

	return data.buildTraces()
}

type flowData struct {
	ebpfmessage

	minTimestamp time.Time
	maxTimestamp time.Time
	count        int
}

func flowid(msg *ebpfmessage) string {
	// TODO: normalize on the port number?
	return fmt.Sprintf("%s:%d->%s:%d", msg.clientIp, msg.clientPort, msg.serverIp, msg.serverPort)
}

type flowsPackage struct {
	flows      map[string]*flowData
	randSource *rand.Rand
}

func (fp *flowsPackage) addMessage(msg *ebpfmessage) {
	id := flowid(msg)
	flow, found := fp.flows[id]
	if !found {
		flow = &flowData{
			ebpfmessage:  *msg,
			count:        1,
			minTimestamp: msg.ts,
			// Don't make it zero
			maxTimestamp: msg.ts.Add(time.Nanosecond),
		}
		fp.flows[id] = flow
	} else {
		flow.count++
		if msg.ts.After(flow.maxTimestamp) {
			flow.maxTimestamp = msg.ts
		}
		if msg.ts.Before(flow.minTimestamp) {
			flow.minTimestamp = msg.ts
		}
	}
}

func (fp *flowsPackage) buildTraces() pdata.Traces {
	traces := pdata.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	ils := rs.InstrumentationLibrarySpans().AppendEmpty()

	spans := ils.Spans()
	spans.EnsureCapacity(len(fp.flows) * 2)
	for _, v := range fp.flows {
		fp.createTraceInSlice(spans, v)
	}

	return traces
}

func (fp *flowsPackage) createTraceInSlice(spans pdata.SpanSlice, v *flowData) {
	tid := [16]byte{}
	fp.randSource.Read(tid[:])
	sid1 := [8]byte{}
	fp.randSource.Read(sid1[:])
	sid2 := [8]byte{}
	fp.randSource.Read(sid2[:])

	traceId := pdata.NewTraceID(tid)
	spanId1 := pdata.NewSpanID(sid1)
	spanId2 := pdata.NewSpanID(sid2)

	clientSpan := spans.AppendEmpty()
	serverSpan := spans.AppendEmpty()

	clientSpan.SetTraceID(traceId)
	serverSpan.SetTraceID(traceId)
	clientSpan.SetSpanID(spanId1)
	serverSpan.SetParentSpanID(spanId1)
	serverSpan.SetSpanID(spanId2)

	fp.fillSpans(clientSpan, serverSpan, v)
}

func (fp *flowsPackage) fillSpans(clientSpan pdata.Span, serverSpan pdata.Span, fd *flowData) {
	fillSpanBasics(clientSpan, fd)
	fillSpanBasics(serverSpan, fd)
	// This normally is at resource level, but it might work here as well
	clientSpan.Attributes().InsertString("service.name", fp.determineClientServiceName(fd))
	clientSpan.SetKind(pdata.SpanKindClient)
	serverSpan.Attributes().InsertString("service.name", fp.determineServerServiceName(fd))
	serverSpan.SetKind(pdata.SpanKindServer)
}

func fillSpanBasics(span pdata.Span, fd *flowData) {
	span.SetName(flowid(&fd.ebpfmessage))
	span.SetStartTimestamp(pdata.NewTimestampFromTime(fd.minTimestamp))
	span.SetEndTimestamp(pdata.NewTimestampFromTime(fd.maxTimestamp))
	span.Attributes().InsertInt("flows.count", int64(fd.count))
	if fd.statusCode != 0 {
		span.Attributes().InsertInt("http.status_code", int64(fd.statusCode))
		// TODO: set Span Status
	}
}

func (fp *flowsPackage) determineClientServiceName(fd *flowData) string {
	if fd.clientComm != "" {
		return fd.clientComm
	}
	// TODO: client IP+port should be used here
	if fd.clientPort < 5000 {
		return "foo"
	} else {
		return "foofoo"
	}
}

func (fp *flowsPackage) determineServerServiceName(fd *flowData) string {
	if fd.serverComm != "" {
		return fd.serverComm
	}
	// TODO: server IP+port should be used here
	if fd.serverPort < 5000 {
		return "bar"
	} else {
		return "snafu"
	}
}
