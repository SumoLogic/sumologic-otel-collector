package asset

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

type Fetcher interface {
	// Fetch a payload from the URL given as the argument, and save it in a
	// temporary file.
	Fetch(ctx context.Context, URL string) (*os.File, error)
}

// An HTTPFetcher fetches the contents of files at a given URL.
type httpFetcher struct {
	client      *http.Client
	logger      *zap.SugaredLogger
	makeBackoff func() backoff.BackOff
}

// NewFetcher creates a new HTTP based Fetcher.
//
// Uses an exponential backoff to retry failed requests.
func NewFetcher(log *zap.SugaredLogger, client *http.Client) Fetcher {
	return &httpFetcher{
		client:      client,
		logger:      log,
		makeBackoff: func() backoff.BackOff { return backoff.NewExponentialBackOff() },
	}
}

// Fetch the file found at the specified url, and return the file or an
// error indicating why the fetch failed.
func (h *httpFetcher) Fetch(ctx context.Context, url string) (*os.File, error) {
	var fetchErr error
	var attempts int
	b := h.makeBackoff()
	for {
		if attempts > 0 {
			h.logger.Errorf("retrying failed asset fetch for %s: %s", url, fetchErr)
		}

		out, err := h.tryFetch(ctx, url)
		if err == nil {
			return out, nil
		}
		attempts += 1
		fetchErr = err
		duration := b.NextBackOff()
		if duration == backoff.Stop {
			break
		}
		select {
		case <-time.After(duration):
		case <-ctx.Done():
			return nil, ctx.Err()
		}

	}
	return nil, fmt.Errorf("failed to fetch resource after %d attempts: %s", attempts, fetchErr)
}

func (h *httpFetcher) tryFetch(ctx context.Context, url string) (*os.File, error) {
	resp, err := h.get(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	// Write response to tmp
	tmpFile, err := os.CreateTemp(os.TempDir(), "runtime-asset")
	if err != nil {
		return nil, fmt.Errorf("can't open tmp file for asset: %s", err)
	}

	cleanup := func() {
		tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}

	buffered := bufio.NewWriter(tmpFile)
	if _, err = io.Copy(buffered, resp); err != nil {
		cleanup()
		return nil, fmt.Errorf("error downloading asset: %s", err)
	}
	if err := buffered.Flush(); err != nil {
		cleanup()
		return nil, fmt.Errorf("error downloading asset: %s", err)
	}

	if err := tmpFile.Sync(); err != nil {
		cleanup()
		return nil, err
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		cleanup()
		return nil, err
	}

	return tmpFile, nil
}

// Get the target URL and return an io.ReadCloser
func (h *httpFetcher) get(ctx context.Context, path string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching asset: %s", err)
	}
	req = req.WithContext(ctx)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching asset: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching asset: Response Code %d", resp.StatusCode)
	}

	return resp.Body, nil
}
