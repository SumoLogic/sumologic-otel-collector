package asset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandValid(t *testing.T) {

	testCases := []string{
		"helloworld.tar.gz",
		"helloworld.tar",
	}
	for i := range testCases {
		fileName := testCases[i]
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			assetPath := fixturePath(fileName)
			f, err := os.Open(assetPath)
			require.NoError(t, err)
			defer f.Close()
			tmpdir, remove := tempDir(t)
			defer remove()

			expander := archiveExpander{}
			targetDirectory := filepath.Join(tmpdir, strings.ReplaceAll(fileName, ".", "-"))
			require.NoError(t, os.Mkdir(targetDirectory, 0755))
			err = expander.Expand(f, targetDirectory)
			assert.NoError(t, err)
			info, err := os.Stat(filepath.Join(targetDirectory, "bin"))
			assert.NoError(t, err)
			assert.True(t, info.IsDir())
		})
	}
}

func TestExpandInvalid(t *testing.T) {

	testCases := []string{
		"invalid.tar",
		"invalid.tar.gz",
		"unsupported.zip",
	}
	for i := range testCases {
		fileName := testCases[i]
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			assetPath := fixturePath(fileName)
			f, err := os.Open(assetPath)
			require.NoError(t, err)
			defer f.Close()
			tmpdir, remove := tempDir(t)
			defer remove()

			expander := archiveExpander{}
			targetDirectory := filepath.Join(tmpdir, strings.ReplaceAll(fileName, ".", "-"))
			require.NoError(t, os.Mkdir(targetDirectory, 0755))
			err = expander.Expand(f, targetDirectory)
			assert.Error(t, err)
		})
	}
}
func fixturePath(f string) string {
	abs, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	return filepath.Join(abs, "fixtures", f)
}

func tempDir(t testing.TB) (tmpDir string, remove func()) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "jobreceiver-test")
	if err != nil {
		t.FailNow()
	}

	return tmpDir, func() { _ = os.RemoveAll(tmpDir) }
}
