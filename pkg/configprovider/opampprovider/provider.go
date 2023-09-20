package opampprovider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	"go.opentelemetry.io/collector/confmap"
)

const (
	schemeName = "opamp"
)

var ErrProviderAlreadyRunning = errors.New("only one URI can be retrieved with the op-amp provider")

type Provider struct {
	watcher FileWatcher
	called  bool
}

func New() confmap.Provider {
	return &Provider{}
}

func (p *Provider) Retrieve(ctx context.Context, uri string, fn confmap.WatcherFunc) (*confmap.Retrieved, error) {
	if p.called {
		// Retrieve starts up a goroutine that can call fn. Given the existing
		// use of Retrieve is a once-per-URI sort of affair, and given that the
		// op-amp provider is only supposed to work with a single URI, we can
		// assume that more than a single call is a programmer error.
		return nil, ErrProviderAlreadyRunning
	}
	p.called = true
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("can't load %q: %s", uri, err)
	}

	if got, want := u.Scheme, schemeName; got != want {
		return nil, fmt.Errorf("opamp provider called with scheme %q (want %q)", got, want)
	}

	content, err := os.ReadFile(u.Path)
	if err != nil {
		return nil, fmt.Errorf("can't read %q: %s", uri, err)
	}

	var config map[string]any

	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	// test p.watcher == nil so that we can patch in our own watcher in test
	// TODO: see if it's necessary to maintain consistency with other providers
	// or if New() can be used to inject the watcher instead.
	if p.watcher == nil {
		if err := p.initWatcher(); err != nil {
			return nil, err
		}
	}

	go p.watchForEvents(ctx, fn)

	if err := p.watcher.Add(u.Path); err != nil {
		return nil, fmt.Errorf("can't watch %q: %s", u.Path, err)
	}

	return confmap.NewRetrieved(config)
}

func (p *Provider) initWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err == nil {
		p.watcher = wrappedWatcher{watcher: watcher}
	} else {
		p.watcher = NewPollWatcher(time.Minute)
	}
	return nil
}

func (p *Provider) watchForEvents(ctx context.Context, fn confmap.WatcherFunc) {
	ch := p.watcher.Events()
	errs := p.watcher.Errors()
	for {
		select {
		case <-ctx.Done():
			if err := p.watcher.Close(); err != nil {
				fn(&confmap.ChangeEvent{Error: err})
			}
		case _, ok := <-ch:
			if !ok {
				return
			}
			fn(&confmap.ChangeEvent{})
		case err, ok := <-errs:
			if !ok {
				return
			}
			fn(&confmap.ChangeEvent{Error: err})
		}
	}
}

type FileWatcher interface {
	Add(name string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

type wrappedWatcher struct {
	watcher *fsnotify.Watcher
}

func (w wrappedWatcher) Add(name string) error {
	return w.watcher.Add(name)
}

func (w wrappedWatcher) Close() error {
	return w.watcher.Close()
}

func (w wrappedWatcher) Remove(name string) error {
	return w.watcher.Remove(name)
}

func (w wrappedWatcher) Events() <-chan fsnotify.Event {
	return w.watcher.Events
}

func (w wrappedWatcher) Errors() <-chan error {
	return w.watcher.Errors
}

func (*Provider) Scheme() string {
	return "opamp"
}

func (*Provider) Shutdown(context.Context) error {
	return nil
}
