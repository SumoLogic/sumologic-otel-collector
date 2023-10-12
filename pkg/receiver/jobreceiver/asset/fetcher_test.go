package asset

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFetcher(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.FileServer(http.Dir(fixturePath(""))))
	defer srv.Close()
	client := srv.Client()
	nl := zap.NewNop().Sugar()
	fetcher := NewFetcher(nl, client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("can fetch tar gz file", func(t *testing.T) {
		helloTGZURL, err := url.JoinPath(srv.URL, "helloworld.tar.gz")
		require.NoError(t, err)
		f, err := fetcher.Fetch(ctx, helloTGZURL)
		require.NoError(t, err)

		tarGZSHA := ExtractSHA(t, "helloworld.tar.gz.sha512")
		err = new(sha512Verifier).Verify(f, tarGZSHA)
		assert.NoError(t, err)
	})

	t.Run("can fetch tar file", func(t *testing.T) {
		helloTURL, err := url.JoinPath(srv.URL, "helloworld.tar")
		require.NoError(t, err)
		f, err := fetcher.Fetch(ctx, helloTURL)
		require.NoError(t, err)

		tarSHA := ExtractSHA(t, "helloworld.tar.sha512")
		err = new(sha512Verifier).Verify(f, tarSHA)
		assert.NoError(t, err)
	})

}

func TestFetcherBackOff(t *testing.T) {
	t.Parallel()

	observedIntervals := make(chan time.Duration, 16)
	respondOK := make(chan []byte, 1)
	srv := httptest.NewTLSServer(&intervalObserver{
		Response: respondOK,
		Out:      observedIntervals,
	})
	defer srv.Close()
	fetcher := NewFetcher(zap.NewNop().Sugar(), srv.Client())
	fetcher.(*httpFetcher).makeBackoff = func() backoff.BackOff {
		return backoff.NewConstantBackOff(time.Millisecond * 150)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetchDone := make(chan struct{})
	var fetchResponse *os.File
	var fetchErr error
	go func() {
		fetchResponse, fetchErr = fetcher.Fetch(ctx, srv.URL)
		close(fetchDone)
	}()

	intervals := make([]time.Duration, 0, 3)
	assert.Eventually(t, func() bool {
		select {
		case interval := <-observedIntervals:
			intervals = append(intervals, interval)
		default:
		}
		return len(intervals) >= 3
	}, time.Second*5, time.Millisecond*150, "expected fetcher to retry repeatedly")

	// retried according to backoff schedule with .3 to 10x tolerance
	for _, interval := range intervals {
		assert.Less(t, time.Millisecond*50, interval)
		assert.Greater(t, time.Millisecond*1500, interval)
	}

	// respond to fetcher successfully with payload
	respondOK <- []byte("sensu")

	assert.Eventually(t, func() bool {
		select {
		case <-fetchDone:
			return true
		default:
			return false
		}
	}, time.Second*5, time.Millisecond*150, "expected fetcher to return")

	require.NoError(t, fetchErr)
	defer func() { assert.NoError(t, fetchResponse.Close()) }()
	actualText, err := io.ReadAll(fetchResponse)
	require.NoError(t, err)
	assert.Equal(t, "sensu", string(actualText))
}

func TestFetcherBackOffExpires(t *testing.T) {
	t.Parallel()

	observedIntervals := make(chan time.Duration, 16)
	srv := httptest.NewTLSServer(&intervalObserver{
		Response: make(chan []byte, 1),
		Out:      observedIntervals,
	})
	defer srv.Close()
	fetcher := NewFetcher(zap.NewNop().Sugar(), srv.Client())
	// set backoff max elapsed time to .5s
	fetcher.(*httpFetcher).makeBackoff = func() backoff.BackOff {
		b := backoff.NewExponentialBackOff()
		b.InitialInterval = 50 * time.Millisecond
		b.Multiplier = 1
		b.RandomizationFactor = 0.1
		b.MaxElapsedTime = 2 * time.Second
		return b
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetchDone := make(chan struct{})
	var fetchResponse *os.File
	var fetchErr error
	go func() {
		fetchResponse, fetchErr = fetcher.Fetch(ctx, srv.URL)
		close(fetchDone)
	}()

	atLeastOneRetry := make(chan struct{})
	go func() {
		<-observedIntervals
		close(atLeastOneRetry)
		for {
			select {
			case <-observedIntervals:
			case <-ctx.Done():
				return
			}
		}
	}()
	assert.Eventually(t, func() bool {
		select {
		case <-atLeastOneRetry:
			return true
		default:
			return false
		}
	}, time.Second*5, time.Millisecond*150, "expected at least one retry")

	assert.Eventually(t, func() bool {
		select {
		case <-fetchDone:
			return true
		default:
			return false
		}
	}, time.Second*2, time.Millisecond*50, "expected fetcher to return")

	assert.Error(t, fetchErr)
	if fetchResponse != nil {
		assert.NoError(t, fetchResponse.Close())
	}
}

func TestFetcherInvalidURL(t *testing.T) {
	t.Parallel()

	failFastClient := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: time.Millisecond * 500,
			}).DialContext,
		},
	}
	fetcher := NewFetcher(zap.NewNop().Sugar(), failFastClient)
	// set backoff max elapsed time to 1s
	fetcher.(*httpFetcher).makeBackoff = func() backoff.BackOff {
		b := backoff.NewExponentialBackOff()
		b.InitialInterval = 50 * time.Millisecond
		b.Multiplier = 1
		b.RandomizationFactor = 0.1
		b.MaxElapsedTime = 1 * time.Second
		return b
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetchDone := make(chan struct{})
	var fetchResponse *os.File
	var fetchErr error
	go func() {
		fetchResponse, fetchErr = fetcher.Fetch(ctx, "https://sumo-otel-collector.invalid/myasset.tar.gz")
		close(fetchDone)
	}()

	assert.Eventually(t, func() bool {
		select {
		case <-fetchDone:
			return true
		default:
			return false
		}
	}, time.Second*5, time.Millisecond*50, "expected fetcher to return")

	assert.Error(t, fetchErr)
	if fetchResponse != nil {
		assert.NoError(t, fetchResponse.Close())
	}
}

type intervalObserver struct {
	Last time.Time

	Response <-chan []byte
	Out      chan<- time.Duration
}

var _ http.Handler = (*intervalObserver)(nil)

func (i *intervalObserver) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	prev := i.Last
	i.Last = time.Now()

	select {
	case resp := <-i.Response:
		_, err := rw.Write(resp)
		if err != nil {
			panic(err)
		}
	default:
		if !prev.IsZero() {
			i.Out <- i.Last.Sub(prev)
		}
		rw.WriteHeader(http.StatusServiceUnavailable)
	}
}
