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
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"
)

var payload2 = "\n75.2.16.224:80 -> 10.0.0.41:58358\nGET / HTTP/1.1\nHost: sumologic.com\nUser-Agent: curl/7.68.0\n"

var payload = "\n75.2.16.224:80 -> 10.0.0.41:58358\n" +
	"GET / HTTP/1.1\n" +
	"Host: onet.pl\n" +
	"User-Agent: curl/7.68.0\n" +
	"Accept: \n" +
	"99.83.207.202:80 -> 10.0.0.41:41044\n" +
	"HTTP/1.1 301 Moved Permanently\n" +
	"Server: Ring Publishing - Accelerator\n" +
	"Date: Fri, 10 Dec 2021 13:11:07 GMT\n" +
	"Content-Type: text/html\n" +
	"Content-Length: 162\n" +
	"Connection: keep-alive\n" +
	"Location: https://www.onet.pl/\n" +
	"set-cookie: acc_segment=15; Path=/; Max-Age=604800\n"

func TestPayloadExtractionShort(t *testing.T) {
	r := newServiceMapReceiver(zap.NewNop())

	msg := &ebpfmessage{}
	r.fillPayload(payload2, msg)

	assert.Equal(t, "GET", msg.op)
	assert.Equal(t, "/", msg.path)
	assert.Equal(t, 0, msg.statusCode)

	assert.Equal(t, "sumologic.com", msg.serverComm)
	assert.Equal(t, "curl/7.68.0", msg.clientComm)
}

func TestPayloadExtraction(t *testing.T) {
	r := newServiceMapReceiver(zap.NewNop())

	msg := &ebpfmessage{}
	r.fillPayload(payload, msg)

	assert.Equal(t, "GET", msg.op)
	assert.Equal(t, "/", msg.path)
	assert.Equal(t, 301, msg.statusCode)

	assert.Equal(t, "onet.pl", msg.serverComm)
	assert.Equal(t, "curl/7.68.0", msg.clientComm)
}

func TestBuildingFlow(t *testing.T) {
	msgs := []*ebpfmessage{
		{
			clientIp:   "1.1.1.1",
			serverIp:   "2.2.2.2",
			clientPort: 45000,
			serverPort: 80,
			clientComm: "",
			serverComm: "",
			statusCode: 0,
			op:         "",
			path:       "",
			ts:         time.Now(),
		},
		{
			clientIp:   "1.1.1.1",
			serverIp:   "2.2.2.2",
			clientPort: 45000,
			serverPort: 80,
			clientComm: "client",
			serverComm: "server",
			statusCode: 404,
			op:         "/GET",
			path:       "/home",
			ts:         time.Now(),
		},
	}
	traces := buildTraces(msgs)
	ss := traces.ResourceSpans().At(0).InstrumentationLibrarySpans().At(0).Spans()
	assert.Equal(t, 2, ss.Len())
	v, ok := ss.At(0).Attributes().Get("service.name")
	assert.True(t, ok)
	assert.Equal(t, "client", v.AsString())
	v, ok = ss.At(1).Attributes().Get("service.name")
	assert.True(t, ok)
	assert.Equal(t, "server", v.AsString())
	for i := 0; i < ss.Len(); i++ {
		span := ss.At(i)
		v, ok = span.Attributes().Get("http.status_code")
		assert.True(t, ok)
		assert.Equal(t, int64(404), v.IntVal())
	}
}
