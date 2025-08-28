package sumologic_tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testContext struct {
	test            *testing.T
	commandErr      error
	commandOutput   []byte
	commandExitCode int
}

type checkFunc func(testContext) bool

func checkLogFileCreated(c testContext) bool {
	return assert.FileExists(c.test, logFilePath, "log file not found. "+string(c.commandOutput))
}

func checkValidateOutput(c testContext) bool {
	return assert.Equal(c.test, 0, c.commandExitCode, "Validate output code invalid. "+string(c.commandOutput))
}

func checkInvalidValidateOutput(c testContext) bool {
	ok1 := assert.Equal(c.test, 1, c.commandExitCode, "Invalid file successfully validated")
	ok2 := assert.Contains(c.test, string(c.commandOutput), "invalid keys: processorss", "Expected error message in output")
	return ok1 && ok2
}

func checkValidSumologicExporter(c testContext) bool {
	ok := assert.Nil(c.test, c.commandErr, "error occured while running collector: "+fmt.Sprint(c.commandErr))
	return ok
}

func preActionCreateCredentialsDir(c testContext) bool {
	err := os.MkdirAll(credentialsDir, 0755)
	return assert.NoError(c.test, err, "Failed to create credentials directory: "+credentialsDir)
}

func checkLogNumbersViaSumologicMock(c testContext) bool {
	resp, err := http.Get(sumlogicMockURL + sumologicMockLogCountPath)
	if !assert.NoError(c.test, err, "Failed to send GET request") {
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if !assert.NoError(c.test, err, "Failed to read response body") {
		return false
	}

	var result map[string]int
	err = json.Unmarshal(body, &result)
	if !assert.NoError(c.test, err, "Failed to unmarshal JSON response") {
		return false
	}

	return assert.Equal(c.test, 2, result["count"], "Expected count to be 2")
}
