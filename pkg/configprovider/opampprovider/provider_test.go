package opampprovider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/collector/confmap"
)

func noopCB(*confmap.ChangeEvent) {}

func tempContent(t testing.TB, content string) *os.File {
	t.Helper()
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprint(f, content); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	return f
}

func opampURI(f *os.File) string {
	return "opamp:" + f.Name()
}

type fakeWatcher struct {
	AddError    error
	RemoveError error
	CloseError  error
	EventsC     chan fsnotify.Event
	ErrsC       chan error
}

func (f fakeWatcher) Add(name string) error {
	return f.AddError
}

func (f fakeWatcher) Remove(name string) error {
	return f.RemoveError
}

func (f fakeWatcher) Close() error {
	return f.CloseError
}

func (f fakeWatcher) Events() <-chan fsnotify.Event {
	return f.EventsC
}

func (f fakeWatcher) Errors() <-chan error {
	return f.ErrsC
}

func newFakeWatcherWithEvent() fakeWatcher {
	events := make(chan fsnotify.Event, 1)
	events <- fsnotify.Event{}
	return fakeWatcher{
		EventsC: events,
	}
}

func newFakeWatcherWithError() fakeWatcher {
	errs := make(chan error, 1)
	errs <- errors.New("boom")
	return fakeWatcher{
		ErrsC: errs,
	}
}

type expectedToBeCalled struct {
	called    bool
	lastEvent *confmap.ChangeEvent
	wg        sync.WaitGroup
}

func (e *expectedToBeCalled) Call(event *confmap.ChangeEvent) {
	if e != nil {
		e.called = true
		e.lastEvent = event
		e.wg.Done()
	}
}

func (e *expectedToBeCalled) Called(t testing.TB) {
	if e == nil {
		return
	}
	e.wg.Wait()
	if !e.called {
		t.Error("callback not called")
	}
	if e.lastEvent == nil {
		t.Error("nil last event")
	}
}

func (e *expectedToBeCalled) CalledWithError(t testing.TB, err error) {
	if e == nil {
		return
	}
	e.Called(t)
	if e.lastEvent == nil {
		t.Error("nil last event")
	}
	if got, want := e.lastEvent.Error, err; got != want {
		t.Errorf("errors not equal: got %q, want %q", got, want)
	}
}

func newExpectCall(numcalls int) *expectedToBeCalled {
	c := &expectedToBeCalled{}
	c.wg.Add(1)
	return c
}

func TestProviderRetrieve(t *testing.T) {
	bg := context.Background()
	tests := []struct {
		Name           string
		Ctx            context.Context
		URI            string
		CB             *expectedToBeCalled
		Content        *os.File
		ExpRetrieved   any
		ExpErr         bool
		PatchedWatcher FileWatcher
	}{
		{
			Name:   "invalid scheme",
			Ctx:    bg,
			URI:    "file:///mulder.yaml",
			ExpErr: true,
		},
		{
			Name:   "file missing",
			Ctx:    bg,
			URI:    "opamp:/scully.yaml",
			ExpErr: true,
		},
		{
			Name:    "invalid yaml",
			Ctx:     bg,
			Content: tempContent(t, "bad yaml document"),
			ExpErr:  true,
		},
		{
			Name:         "valid yaml",
			Ctx:          bg,
			Content:      tempContent(t, `{"this is my yaml": true}`),
			ExpRetrieved: map[string]any{"this is my yaml": true},
		},
		{
			Name:           "watcher event results in callback call",
			Ctx:            bg,
			CB:             newExpectCall(1),
			Content:        tempContent(t, `{"this is my yaml": true}`),
			ExpRetrieved:   map[string]any{"this is my yaml": true},
			PatchedWatcher: newFakeWatcherWithEvent(),
		},
		{
			Name:           "watcher error results in callback call with error",
			Ctx:            bg,
			CB:             newExpectCall(1),
			Content:        tempContent(t, `{"this is my yaml": true}`),
			ExpRetrieved:   map[string]any{"this is my yaml": true},
			PatchedWatcher: newFakeWatcherWithError(),
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.Content != nil {
				defer test.Content.Close()
				defer os.Remove(test.Content.Name())
			}
			if test.URI == "" {
				test.URI = opampURI(test.Content)
			}
			p := &Provider{watcher: test.PatchedWatcher}
			retrieved, err := p.Retrieve(test.Ctx, test.URI, test.CB.Call)
			if test.ExpErr && err == nil {
				t.Error("expected non-nil error")
				return
			} else if !test.ExpErr && err != nil {
				t.Error(err)
				return
			}
			var conf any
			if retrieved != nil {
				conf, err = retrieved.AsRaw()
				if err != nil {
					t.Error(err)
					return
				}
			}
			if !cmp.Equal(conf, test.ExpRetrieved) {
				t.Errorf("retrieved not as expected: %s", cmp.Diff(retrieved, test.ExpRetrieved))
			}
			test.CB.Called(t)
		})
	}
}
