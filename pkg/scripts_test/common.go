//go:build !windows

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

type installType int

const (
	BINARY_INSTALL installType = 1 << iota
	PACKAGE_INSTALL
)

type testSpec struct {
	name              string
	installType       installType
	options           installOptions
	preChecks         []checkFunc
	postChecks        []checkFunc
	preActions        []checkFunc
	conditionalChecks []condCheckFunc
	installCode       int
}

func NewTestSpecFromTestSpec(old testSpec) testSpec {
	options := installOptions{
		installToken:           old.options.installToken,
		deprecatedInstallToken: old.options.deprecatedInstallToken,
		autoconfirm:            old.options.autoconfirm,
		skipSystemd:            old.options.skipSystemd,
		tags:                   make(map[string]string),
		skipConfig:             old.options.skipConfig,
		skipInstallToken:       old.options.skipInstallToken,
		fips:                   old.options.fips,
		envs:                   make(map[string]string),
		uninstall:              old.options.uninstall,
		purge:                  old.options.purge,
		apiBaseURL:             old.options.apiBaseURL,
		configBranch:           old.options.configBranch,
		downloadOnly:           old.options.downloadOnly,
		dontKeepDownloads:      old.options.dontKeepDownloads,
		installHostmetrics:     old.options.installHostmetrics,
		timeout:                old.options.timeout,
		useNativePackaging:     old.options.useNativePackaging,
		version:                old.options.version,
	}
	for k, v := range old.options.tags {
		options.tags[k] = v
	}
	for k, v := range old.options.envs {
		options.envs[k] = v
	}

	new := testSpec{
		name:              old.name,
		installType:       old.installType,
		options:           options,
		preChecks:         []checkFunc{},
		postChecks:        []checkFunc{},
		preActions:        []checkFunc{},
		conditionalChecks: []condCheckFunc{},
		installCode:       old.installCode,
	}
	new.preChecks = append(new.preChecks, old.preChecks...)
	new.postChecks = append(new.postChecks, old.postChecks...)
	new.preActions = append(new.preActions, old.preActions...)
	new.conditionalChecks = append(new.conditionalChecks, old.conditionalChecks...)

	return new
}

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
		installType:         spec.installType,
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

	t.Log("Running install script")
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
