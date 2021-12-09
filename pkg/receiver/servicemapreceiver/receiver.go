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
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/collector/consumer"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

var (
	ErrAlreadyStarted = errors.New("component already started")
	ErrAlreadyStopped = errors.New("component already stopped")
)

type servicemapreceiver struct {
	mu             sync.Mutex
	cancel         context.CancelFunc
	logger         *zap.Logger
	messageCache   []*ebpfmessage
	tracesConsumer consumer.Traces
}

type ebpfmessage struct {
	clientIp   string
	serverIp   string
	clientPort uint16
	serverPort uint16
	// empty is not present
	clientComm string
	// empty is not present
	serverComm string
	// 0  is not set
	statusCode int

	ts time.Time
}

// Ensure this receiver adheres to required interface.
var _ component.MetricsReceiver = (*servicemapreceiver)(nil)
var _ component.TracesReceiver = (*servicemapreceiver)(nil)

// Start tells the receiver to start.
func (r *servicemapreceiver) Start(ctx context.Context, host component.Host) error {
	r.logger.Info("Starting Service Map receiver")

	ctx, r.cancel = context.WithCancel(ctx)

	go r.loop(ctx)

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.dumpData(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (r *servicemapreceiver) dumpData(ctx context.Context) {
	r.mu.Lock()
	r.logger.Info("Going to dump messages", zap.Int("count", len(r.messageCache)))
	copiedData := make([]*ebpfmessage, len(r.messageCache))
	copy(copiedData, r.messageCache)
	r.messageCache = nil
	r.mu.Unlock()

	traces := buildTraces(copiedData)

	// Now we could aggregate and such
	if r.tracesConsumer != nil {
		r.tracesConsumer.ConsumeTraces(ctx, traces)
	}
}

// Shutdown is invoked during service shutdown.
func (r *servicemapreceiver) Shutdown(context.Context) error {
	r.logger.Info("Shutting down Service Map receiver")
	r.cancel()
	r.logger.Info("Shutting down Service Map receiver finalized")
	return nil
}

func (r *servicemapreceiver) addMessage(msg *ebpfmessage) {
	r.logger.Info(
		fmt.Sprintf("Received a message %s:%d -> %s:%d",
			msg.clientIp, msg.clientPort,
			msg.serverIp, msg.serverPort))

	r.mu.Lock()
	defer r.mu.Unlock()

	// FIXME: This is far from optimal

	r.messageCache = append(r.messageCache, msg)
}

func buildMessage(srcIp string, srcPort uint16, destIp string, destPort uint16, srcComm string, destComm string, statusCode int) *ebpfmessage {
	if srcPort < destPort {
		return &ebpfmessage{
			clientIp:   destIp,
			serverIp:   srcIp,
			clientPort: destPort,
			serverPort: srcPort,
			clientComm: destComm,
			serverComm: srcComm,
			statusCode: statusCode,
			ts:         time.Now(),
		}
	} else {
		return &ebpfmessage{
			clientIp:   srcIp,
			serverIp:   destIp,
			clientPort: srcPort,
			serverPort: destPort,
			clientComm: srcComm,
			serverComm: destComm,
			statusCode: statusCode,
			ts:         time.Now(),
		}
	}
}

func (r *servicemapreceiver) loop(ctx context.Context) {
	for {
		r.addMessage(buildMessage(
			"10.0.0.1",
			1000,
			"11.1.1.1",
			2000,
			"",
			"",
			200,
		))
		time.Sleep(2 * time.Second)
	}
}
