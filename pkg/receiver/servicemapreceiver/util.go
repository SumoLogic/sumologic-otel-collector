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
	"encoding/binary"
	"fmt"
	"unsafe"

	linux "golang.org/x/sys/unix"
)

var NativeEndian binary.ByteOrder

func initEndian() {
	if isBigEndian() {
		NativeEndian = binary.BigEndian
	} else {
		NativeEndian = binary.LittleEndian
	}
}

func isBigEndian() (ret bool) {
	i := int(0x1)
	bs := (*[int(unsafe.Sizeof(i))]byte)(unsafe.Pointer(&i))
	return bs[0] == 0
}

// UnmarshalBinary unmarshals a ringbuf record into an Event.
func (e *Event) unmarshalBinary(b []byte) error {
	if len(b) == eventHeadLength {
		e.hasPayload = false
		e.SPort = NativeEndian.Uint16(b[:2])
		e.DPort = NativeEndian.Uint16(b[2:4])

		e.SAddr = binary.BigEndian.Uint32(b[4:8])
		e.DAddr = binary.BigEndian.Uint32(b[8:12])

		return nil
	} else if len(b) == eventPayloadLength {
		e.hasPayload = true
		e.Payload = linux.ByteSliceToString(b[:32])
		return nil
	} else {
		return fmt.Errorf("unexpected event length %d", len(b))
	}
}
