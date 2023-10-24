package asset

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerify(t *testing.T) {
	t.Parallel()

	assetPath := fixturePath("helloworld.tar.gz")
	f, err := os.Open(assetPath)
	require.NoError(t, err)

	expectedSha := ExtractSHA(t, "helloworld.tar.gz.sha512")
	err = new(sha512Verifier).Verify(f, expectedSha)
	assert.NoError(t, err)

	badSha := string(append([]byte(expectedSha)[1:], []byte(expectedSha)[0]))
	// Check that a muddled sha512 does not match
	err = new(sha512Verifier).Verify(f, badSha)
	assert.Error(t, err)
}

func ExtractSHA(t *testing.T, fileName string) string {
	t.Helper()
	shaFileContent, err := os.ReadFile(fixturePath(fileName))
	require.NoError(t, err)
	return strings.TrimSpace(string(shaFileContent))
}
