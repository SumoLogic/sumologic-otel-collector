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

	shaFileContent, err := os.ReadFile(fixturePath("helloworld.tar.gz.sha512"))
	require.NoError(t, err)
	expectedSha := strings.TrimSpace(string(shaFileContent))
	err = new(sha512Verifier).Verify(f, expectedSha)
	assert.NoError(t, err)

	badSha := string(append([]byte(expectedSha)[1:], []byte(expectedSha)[0]))
	// Check that a muddled sha512 does not match
	err = new(sha512Verifier).Verify(f, badSha)
	assert.Error(t, err)
}
