package opampprovider

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestPollWatcher(t *testing.T) {
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tf.Name())
	if _, err := fmt.Fprintln(tf, "hello"); err != nil {
		t.Fatal(err)
	}
	if err := tf.Sync(); err != nil {
		t.Fatal(err)
	}
	watcher := NewPollWatcher(100 * time.Millisecond)
	defer watcher.Close()
	if err := watcher.Add(tf.Name()); err != nil {
		t.Fatal(err)
	}
	if err := watcher.Add("asldkfj"); err == nil {
		t.Error("expected non-nil error")
	}
	events := watcher.Events()
	errors := watcher.Errors()
	go func() {
		// ensure enough time has passed that mtime will be different
		after := time.After(100 * time.Millisecond)
		watcher.WaitReady()
		<-after
		if _, err := fmt.Fprintln(tf, "hello, world"); err != nil {
			t.Error(err)
		}
		if err := tf.Close(); err != nil {
			t.Error(err)
		}
	}()
	select {
	case event := <-events:
		if got, want := tf.Name(), event.Name; got != want {
			t.Errorf("bad watch path: got %q, want %q", got, want)
		}
	case err := <-errors:
		t.Error(err)
	}
}
