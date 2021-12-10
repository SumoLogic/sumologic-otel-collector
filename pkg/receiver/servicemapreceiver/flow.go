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
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	opRegex     = regexp.MustCompile("(?P<op>\\w+)\\s+(?P<path>[a-z/].*) HTTP.*")
	statusRegex = regexp.MustCompile("HTTP/[0-9.]+\\s+(?P<op>\\d+).*")
	hostRegex   = regexp.MustCompile("Host:\\s+(?P<host>.*)")
	clientRegex = regexp.MustCompile("User-Agent:\\s+(?P<agent>.*)")
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

	minTimestamp    time.Time
	maxTimestamp    time.Time
	count           int
	combinedPayload string
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
			maxTimestamp:    msg.ts.Add(time.Nanosecond),
			combinedPayload: "",
		}
		fp.flows[id] = flow
	} else {
		flow.count++
		if flow.statusCode == 0 && msg.statusCode != 0 {
			flow.statusCode = msg.statusCode
		}
		if flow.serverComm == "" && msg.serverComm != "" {
			flow.serverComm = msg.serverComm
		}
		if flow.clientComm == "" && msg.clientComm != "" {
			flow.clientComm = msg.clientComm
		}
		if msg.ts.After(flow.maxTimestamp) {
			flow.maxTimestamp = msg.ts
		}
		if msg.ts.Before(flow.minTimestamp) {
			flow.minTimestamp = msg.ts
		}
	}
	flow.combinedPayload = flow.combinedPayload + msg.payload
}

func (fp *flowsPackage) buildTraces() pdata.Traces {
	traces := pdata.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	ils := rs.InstrumentationLibrarySpans().AppendEmpty()

	spans := ils.Spans()
	spans.EnsureCapacity(len(fp.flows) * 2)
	for _, v := range fp.flows {
		// THIS is not ideal, but we don't have the full payload at message level
		fillPayload(v.combinedPayload, v)
		fp.createTraceInSlice(spans, v)
	}

	return traces
}

func fillPayload(payload string, msg *flowData) {
	lines := strings.Split(payload, "\r\n")
	for i, line := range lines {
		if i < 3 {
			matches := opRegex.FindStringSubmatch(line)
			if len(matches) > 2 {
				msg.op = matches[1]
				msg.path = matches[2]
			}
		}
		var matches []string
		if msg.statusCode == 0 {
			matches = statusRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				val, err := strconv.Atoi(matches[1])
				if err == nil {
					msg.statusCode = val
				}
			}
		}
		if msg.serverComm == "" {
			matches = hostRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				msg.serverComm = matches[1]
			}
		}
		if msg.clientComm == "" {
			matches = clientRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				msg.clientComm = matches[1]
			}
		}
	}
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
	return fd.clientComm
}

func (fp *flowsPackage) determineServerServiceName(fd *flowData) string {
	return fd.serverComm
}
