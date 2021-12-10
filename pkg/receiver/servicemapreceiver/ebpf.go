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

//go:build linux
// +build linux

package servicemapreceiver

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

// $BPF_CLANG and $BPF_CFLAGS are set by the Makefile.
//go-:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS bpf cgroup_skb.c -- -I../headers

// Length of struct event_t sent from kernelspace.
var eventHeadLength = 12
var eventPayloadLength = 32

type Event struct {
	hasPayload bool
	SPort      uint16
	DPort      uint16
	SAddr      uint32
	DAddr      uint32
	Payload    string
}

// UnmarshalBinary unmarshals a ringbuf record into an Event.
func (e *Event) UnmarshalBinary(b []byte) error {
	if len(b) == eventHeadLength {
		e.hasPayload = false
		e.SPort = NativeEndian.Uint16(b[:2])
		e.DPort = NativeEndian.Uint16(b[2:4])

		e.SAddr = binary.BigEndian.Uint32(b[4:8])
		e.DAddr = binary.BigEndian.Uint32(b[8:12])

		return nil
	} else if len(b) == eventPayloadLength {
		e.hasPayload = true
		e.Payload = unix.ByteSliceToString(b[:32])
		return nil
	} else {
		return fmt.Errorf("unexpected event length %d", len(b))
	}
}

func run(r *servicemapreceiver) {
	initEndian()

	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}

	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()

	// Get the first-mounted cgroupv2 path.
	cgroupPath, err := detectCgroupPath()
	if err != nil {
		log.Fatal(err)
	}

	// Link the count_egress_packets program to the cgroup.
	l, err := link.AttachCgroup(link.CgroupOptions{
		Path:    cgroupPath,
		Attach:  ebpf.AttachCGroupInetEgress,
		Program: objs.DumpEgressPackets,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	// Link the count_egress_packets program to the cgroup.
	l2, err2 := link.AttachCgroup(link.CgroupOptions{
		Path:    cgroupPath,
		Attach:  ebpf.AttachCGroupInetIngress,
		Program: objs.DumpIngressPackets,
	})
	if err2 != nil {
		log.Fatal(err2)
	}
	defer l2.Close()

	log.Println("Reading event entries...")

	rd, err := ringbuf.NewReader(objs.bpfMaps.Events)
	if err != nil {
		log.Fatalf("opening ringbuf reader: %s", err)
	}
	defer rd.Close()

	go func() {
		<-stopper

		if err := rd.Close(); err != nil {
			log.Fatalf("closing ringbuf reader: %s", err)
		}
	}()

	var event Event
	for {
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) {
				log.Println("received signal, exiting..")
				return
			}
			log.Printf("reading from reader: %s", err)
			continue
		}

		// Parse the ringbuf event entry into an Event structure.
		if err := event.UnmarshalBinary(record.RawSample); err != nil {
			log.Fatal("unmarshaling event failed", zap.Error(err))
		}

		//r.addMessage(r.buildMessage(
		//	intToIP(event.SAddr).String(),
		//	event.SPort,
		//	intToIP(event.DAddr).String(),
		//	event.DPort,
		//	event.Payload,
		//))

		if !event.hasPayload {
			//fmt.Printf("\n%s:%d -> %s:%d\n",
			//	intToIP(event.SAddr),
			//	event.SPort,
			//	intToIP(event.DAddr),
			//	event.DPort,
			//)
			r.addMessage(r.buildMessage(
				intToIP(event.SAddr).String(),
				event.SPort,
				intToIP(event.DAddr).String(),
				event.DPort,
				event.Payload,
			))
		} else {
			r.addMessage(r.buildMessage(
				intToIP(event.SAddr).String(),
				event.SPort,
				intToIP(event.DAddr).String(),
				event.DPort,
				event.Payload,
			))
			//fmt.Printf("%s",
			//	event.Payload)
		}
	}
}

// detectCgroupPath returns the first-found mount point of type cgroup2
// and stores it in the cgroupPath global variable.
func detectCgroupPath() (string, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// example fields: cgroup2 /sys/fs/cgroup/unified cgroup2 rw,nosuid,nodev,noexec,relatime 0 0
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) >= 3 && fields[2] == "cgroup2" {
			return fields[1], nil
		}
	}

	return "", errors.New("cgroup2 not mounted")
}

// intToIP converts IPv4 number to net.IP
func intToIP(ipNum uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, ipNum)
	return ip
}
