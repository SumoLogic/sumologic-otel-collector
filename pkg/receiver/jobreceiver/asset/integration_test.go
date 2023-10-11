package asset

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestManagerIntegration(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.FileServer(http.Dir(fixturePath(""))))
	defer srv.Close()
	client := srv.Client()

	testStore, remove := tempDir(t)
	defer remove()
	nopLog := zap.NewNop().Sugar()
	manager := Manager{
		Fetcher:     NewFetcher(nopLog, client),
		Logger:      nopLog,
		StoragePath: testStore,
	}
	specs := []Spec{
		{
			Name:   "helloworld.tar.gz",
			SHA512: ExtractSHA(t, "helloworld.tar.gz.sha512"),
			URL:    srv.URL + "/helloworld.tar.gz",
		}, {
			Name:   "helloworld.tar",
			SHA512: ExtractSHA(t, "helloworld.tar.sha512"),
			URL:    srv.URL + "/helloworld.tar",
		},
	}
	err := manager.Validate(specs)
	assert.NoError(t, err)

	env := []string{}
	references, err := manager.InstallAll(context.Background(), specs)
	require.NoError(t, err)
	assert.NotEmpty(t, references)
	for _, ref := range references {
		info, err := os.Stat(ref.Path)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
		env = ref.MergeEnvironment(env)
	}
	var path string
	for _, v := range env {
		if strings.HasPrefix(v, "PATH") {
			path = v
			break
		}
	}
	assert.NotEmpty(t, path, env)
}
