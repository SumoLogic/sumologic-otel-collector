package command

import (
	"bytes"
	"sync"
)

// SyncBuffer can be used to buffer both output streams to
// in a monitoring plugin spec compliant way.
type SyncBuffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (s *SyncBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *SyncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}
