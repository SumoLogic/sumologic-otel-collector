package sumologic_scripts_tests

import (
	"os"
)

func getSystemdConfig(path string) (string, error) {
	systemdFile, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(systemdFile), nil
}
