//go:build linux || darwin

package sumologic_scripts_tests

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// These checks always have to be true after a script execution
var commonPostChecks = []checkFunc{checkNoBakFilesPresent}

func cleanCache(t *testing.T) {
	err := os.RemoveAll(cacheDirectory)
	require.NoError(t, err)
}

func runTest(t *testing.T, spec *testSpec) {
	ch := check{
		test:                t,
		installOptions:      spec.options,
		expectedInstallCode: spec.installCode,
	}

	t.Log("Running conditional checks")
	for _, a := range spec.conditionalChecks {
		if !a(ch) {
			t.SkipNow()
		}
	}

	defer tearDown(t)

	t.Log("Starting HTTP server")
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, "200 OK\n")
		require.NoError(t, err)
	})

	listener, err := net.Listen("tcp", ":3333")
	require.NoError(t, err)

	httpServer := &http.Server{
		Handler: mux,
	}
	go func() {
		err := httpServer.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			require.NoError(t, err)
		}
	}()
	defer func() {
		require.NoError(t, httpServer.Shutdown(context.Background()))
	}()

	t.Log("Running pre actions")
	for _, a := range spec.preActions {
		a(ch)
	}

	t.Log("Running pre checks")
	for _, c := range spec.preChecks {
		c(ch)
	}

	ch.code, ch.output, ch.errorOutput, ch.err = runScript(ch)

	// Remove cache in case of curl issue
	if ch.code == curlTimeoutErrorCode {
		cleanCache(t)
	}

	checkRun(ch)

	t.Log("Running common post checks")
	for _, c := range commonPostChecks {
		c(ch)
	}

	t.Log("Running post checks")
	for _, c := range spec.postChecks {
		c(ch)
	}
}
