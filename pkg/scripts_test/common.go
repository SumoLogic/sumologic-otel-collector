//go:build !windows

package sumologic_scripts_tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type testSpec struct {
	name              string
	options           installOptions
	preChecks         []checkFunc
	postChecks        []checkFunc
	preActions        []checkFunc
	conditionalChecks []condCheckFunc
	installCode       int
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
		expectedInstallCode: spec.installCode,
	}

	t.Log("Running conditional checks")
	for _, a := range spec.conditionalChecks {
		if !a(ch) {
			t.SkipNow()
		}
	}

	defer tearDown(t)

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
