//go:build !windows

package sumologic_scripts_tests

import (
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

func tearDown(t *testing.T) {
	t.Log("Cleaning up")
	ch := check{
		test: t,
		installOptions: installOptions{
			uninstall:   true,
			purge:       true,
			autoconfirm: true,
		},
	}

	_, _, err := runScript(ch)
	require.NoError(t, err)
}

func runTest(t *testing.T, spec *testSpec) {
	ch := check{
		test:                t,
		installOptions:      spec.options,
		expectedInstallCode: spec.installCode,
	}

	for _, a := range spec.conditionalChecks {
		if !a(ch) {
			t.SkipNow()
		}
	}

	defer tearDown(t)

	for _, a := range spec.preActions {
		a(ch)
	}

	for _, c := range spec.preChecks {
		c(ch)
	}

	ch.code, ch.output, ch.err = runScript(ch)
	checkRun(ch)

	for _, c := range spec.postChecks {
		c(ch)
	}
}
