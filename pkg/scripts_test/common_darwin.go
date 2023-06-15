package sumologic_scripts_tests

import (
	"os"
	"path"
)

func getPackagePath() string {
	tmpDir := os.Getenv("TMPDIR")
	return path.Join(tmpDir, darwinPackageName)
}
