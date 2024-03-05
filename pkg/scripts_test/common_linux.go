// go:

package sumologic_scripts_tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func tearDown(t *testing.T) {
	ch := check{
		test: t,
		installOptions: installOptions{
			uninstall:   true,
			purge:       true,
			autoconfirm: true,
		},
	}

	_, _, _, err := runScript(ch)
	require.NoError(t, err)
}
