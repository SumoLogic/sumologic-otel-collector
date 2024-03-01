package sumologic_scripts_tests

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/stretchr/testify/require"
)

type installOptions struct {
	installToken       string
	tags               map[string]string
	fips               bool
	envs               map[string]string
	apiBaseURL         string
	installHostmetrics bool
	remotelyManaged    bool
	ephemeral          bool
}

func (io *installOptions) string() []string {
	opts := []string{
		"-Command",
		scriptPath,
	}

	if io.fips {
		opts = append(opts, "-Fips", "1")
	}

	if io.installHostmetrics {
		opts = append(opts, "-InstallHostMetrics", "1")
	}

	if io.remotelyManaged {
		opts = append(opts, "-RemotelyManaged", "1")
	}

	if io.ephemeral {
		opts = append(opts, "-Ephemeral", "1")
	}

	if len(io.tags) > 0 {
		opts = append(opts, "-Tags", getTagOptValue(io.tags))
	}

	if io.apiBaseURL != "" {
		opts = append(opts, "-Api", io.apiBaseURL)
	}

	return opts
}

func (io *installOptions) buildEnvs() []string {
	e := os.Environ()

	for k, v := range io.envs {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}

	if io.installToken != "" {
		e = append(e, fmt.Sprintf("%s=%s", installTokenEnv, io.installToken))
	}

	return e
}

func exitCode(cmd *exec.Cmd) (int, error) {
	err := cmd.Wait()

	if err == nil {
		return cmd.ProcessState.ExitCode(), nil
	}

	if exiterr, ok := err.(*exec.ExitError); ok {
		return exiterr.ExitCode(), nil
	}

	return 0, fmt.Errorf("cannot obtain exit code: %v", err)
}

func runScript(ch check) (int, []string, []string, error) {
	cmd := exec.Command("powershell", ch.installOptions.string()...)
	cmd.Env = ch.installOptions.buildEnvs()
	output := []string{}

	in, err := cmd.StdinPipe()
	if err != nil {
		require.NoError(ch.test, err)
	}

	defer in.Close()

	out, err := cmd.StdoutPipe()
	if err != nil {
		require.NoError(ch.test, err)
	}
	defer out.Close()

	errOut, err := cmd.StderrPipe()
	if err != nil {
		require.NoError(ch.test, err)
	}
	defer errOut.Close()

	// We want to read line by line
	bufOut := bufio.NewReader(out)

	// Start the process
	if err = cmd.Start(); err != nil {
		require.NoError(ch.test, err)
	}

	// Read the results from the process
	for {
		line, _, err := bufOut.ReadLine()
		strLine := strings.TrimSpace(string(line))

		if len(strLine) > 0 {
			output = append(output, strLine)
		}
		ch.test.Log(strLine)

		// exit if script finished
		if err == io.EOF {
			break
		}

		// otherwise ensure there is no error
		require.NoError(ch.test, err)

	}

	// Handle stderr separately
	bufErrOut := bufio.NewReader(errOut)
	errorOutput := []string{}
	for {
		line, _, err := bufErrOut.ReadLine()
		strLine := strings.TrimSpace(string(line))

		if len(strLine) > 0 {
			errorOutput = append(errorOutput, strLine)
		}
		ch.test.Log(strLine)

		// exit if script finished
		if err == io.EOF {
			break
		}
	}

	code, err := exitCode(cmd)
	return code, output, errorOutput, err
}

func getTagOptValue(tags map[string]string) string {
	tagOpts := []string{}
	for k, v := range tags {
		tagOpts = append(tagOpts, fmt.Sprintf("%s = \"%s\"", k, v))
	}
	tagOptString := strings.Join(tagOpts, " ; ")
	return fmt.Sprintf("@{ %s }", tagOptString)
}
