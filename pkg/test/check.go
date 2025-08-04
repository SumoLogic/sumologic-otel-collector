package sumologic_tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type check struct {
	test            *testing.T
	commandErr      error
	commandOutput   []byte
	commandExitCode int
}

type checkFunc func(check) bool

func checkLogFileCreated(c check) bool {
	return assert.FileExists(c.test, logFilePath, "log file not found. "+string(c.commandOutput))
}

func checkValidateOutput(c check) bool {
	return assert.Equal(c.test, 0, c.commandExitCode, "Validate output code invalid. "+string(c.commandOutput))
}

func checkInvalidValidateOutput(c check) bool {
	ok1 := assert.Equal(c.test, 1, c.commandExitCode, "Invalid file successfully validated")
	ok2 := assert.Contains(c.test, string(c.commandOutput), "invalid keys: processorss", "Expected error message in output")
	return ok1 && ok2
}

func checkValidSumologicExporter(c check) bool {
	ok := assert.Nil(c.test, c.commandErr, "error occured while running collector: "+fmt.Sprint(c.commandErr))
	return ok
}

func preActionCreateCredentialsDir(c check) bool {
	err := os.MkdirAll(credentialsDir, 0755)
	return assert.NoError(c.test, err, "Failed to create credentials directory: "+credentialsDir)
}
