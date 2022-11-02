//go:build !windows

package sumologic_scripts_tests

import (
	"fmt"
	"runtime"
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
	fakeTTY           bool
}

// These checks always have to be true after a script execution
var commonPostChecks = []checkFunc{checkNoBakFilesPresent}

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

	for _, c := range commonPostChecks {
		c(ch)
	}

	for _, c := range spec.postChecks {
		c(ch)
	}
}

func getRootGroupName() string {
	if runtime.GOOS == "darwin" {
		return "wheel"
	} else if runtime.GOOS == "linux" {
		return "root"
	}

	panic(fmt.Sprintf("Encountered unsupported OS: %s", runtime.GOOS))
}
