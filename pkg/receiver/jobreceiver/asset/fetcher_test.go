package asset

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

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

	helloTGZURL, err := url.JoinPath(srv.URL, "helloworld.tar.gz")
	require.NoError(t, err)
	f, err := fetcher.Fetch(ctx, helloTGZURL)
	require.NoError(t, err)

	tarGZShaBytes, err := os.ReadFile(fixturePath("helloworld.tar.gz.sha512"))
	require.NoError(t, err)
	tarGZSHA := strings.TrimSpace(string(tarGZShaBytes))
	err = new(sha512Verifier).Verify(f, tarGZSHA)
	assert.NoError(t, err)

	helloTURL, err := url.JoinPath(srv.URL, "helloworld.tar")
	require.NoError(t, err)
	f, err = fetcher.Fetch(ctx, helloTURL)
	require.NoError(t, err)

	tarShaBytes, err := os.ReadFile(fixturePath("helloworld.tar.sha512"))
	require.NoError(t, err)
	tarSHA := strings.TrimSpace(string(tarShaBytes))
	err = new(sha512Verifier).Verify(f, tarSHA)
	assert.NoError(t, err)
}
