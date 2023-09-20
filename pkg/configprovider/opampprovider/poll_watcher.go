package opampprovider

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

func NewPollWatcher(every time.Duration) *PollWatcher {
	return &PollWatcher{
		events:       make(chan fsnotify.Event, 1),
		errors:       make(chan error, 1),
		readyRequest: make(chan struct{}),
		every:        every,
	}
}

type PollWatcher struct {
	path         string
	cancel       context.CancelFunc
	every        time.Duration
	events       chan fsnotify.Event
	errors       chan error
	mtime        time.Time
	readyRequest chan struct{}
}

func (p *PollWatcher) Add(path string) error {
	if p.path != "" {
		return errors.New("can only poll a single path")
	}
	p.path = path
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	go p.start(ctx)
	return nil
}

func (p *PollWatcher) Close() error {
	p.cancel()
	return nil
}

func (p *PollWatcher) Events() <-chan fsnotify.Event {
	return p.events
}

func (p *PollWatcher) Errors() <-chan error {
	return p.errors
}

func (p *PollWatcher) WaitReady() {
	p.readyRequest <- struct{}{}
}

func (p *PollWatcher) start(ctx context.Context) {
	p.doPoll()
	ticker := time.NewTicker(p.every)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.doPoll()
		case <-ctx.Done():
			return
		case <-p.readyRequest:
		}
	}
}

func (p *PollWatcher) doPoll() {
	stat, err := os.Stat(p.path)
	if err != nil {
		p.errors <- err
		return
	}
	mtime := stat.ModTime()
	if !mtime.Equal(p.mtime) && !p.mtime.IsZero() {
		p.events <- fsnotify.Event{
			Name: p.path,
			Op:   fsnotify.Write,
		}
	}
	p.mtime = mtime
}
