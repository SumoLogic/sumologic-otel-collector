package sumologic_scripts_tests

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runScript(t *testing.T) {
	cmd := exec.Command("bash", "-c", "/sumologic/scripts/install.sh")

	in, err := cmd.StdinPipe()
	if err != nil {
		require.NoError(t, err)
	}

	defer in.Close()

	out, err := cmd.StdoutPipe()
	if err != nil {
		require.NoError(t, err)
	}

	defer out.Close()

	// We want to read line by line
	bufOut := bufio.NewReader(out)

	// Start the process
	if err = cmd.Start(); err != nil {
		require.NoError(t, err)
	}

	// Read the results from the process
	for {
		line, _, err := bufOut.ReadLine()
		strLine := string(line)
		t.Log(strLine)

		// exit if script finished
		if err == io.EOF {
			break
		}

		// otherwise ensure there is no error
		require.NoError(t, err)

		if strings.Contains(strLine, "Showing full changelog") {
			// show changelog
			_, err = in.Write([]byte("\n"))
			require.NoError(t, err)

			// accept changes and proceed with the installation
			_, err = in.Write([]byte("y\n"))
			require.NoError(t, err)
		}
	}

	code, err := exitCode(cmd)
	assert.NoError(t, err)

	t.Logf("%v", code)
}

func tearDown(t *testing.T) {
	t.Log("Cleaning up")

	err := os.RemoveAll("/var/lib/sumologic/file_storage")
	assert.NoError(t, err, "no permissions to remove storage directory")

	err = os.RemoveAll("/etc/otelcol-sumo")
	assert.NoError(t, err, "no permissions to remove configuration")

	err = os.RemoveAll("/etc/systemd/system/otelcol-sumo.service")
	assert.NoError(t, err, "no permissions to remove systemd configuration")

	err = os.RemoveAll("/usr/local/bin/otelcol-sumo")
	assert.NoError(t, err, "removing binary")
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

func TestInstallation(t *testing.T) {
	root, err := os.Open("/")
	require.NoError(t, err)

	defer removeEnvironment(t, root)

	defer tearDown(t)
	prepareEnvironment(t)

	_, err := os.Stat("/usr/local/bin/otelcol-sumo")
	require.ErrorIs(t, err, os.ErrNotExist, "/usr/local/bin/otelcol-sumo is already created")
	runScript(t)

	_, err = os.Stat("/usr/local/bin/otelcol-sumo")
	require.NoError(t, err, "/usr/local/bin/otelcol-sumo has not been created")

	cmd := exec.Command("/usr/local/bin/otelcol-sumo", "--version")
	err = cmd.Start()
	require.NoError(t, err)

	code, err := exitCode(cmd)
	require.NoError(t, err)
	require.Equal(t, 0, code)
}

func prepareEnvironment(t *testing.T) {

	dir := os.TempDir()
	path := filepath.Join(dir, "bin")
	fmt.Printf("Local dir: %v", dir)

	os.MkdirAll(path, os.ModePerm)
	for _, command := range []string{"sudo", "sed", "curl", "head", "bash", "sort", "mv", "chmod", "envsubst", "hostname", "tac", "tail", "users", "whoami"} {
		cmd := exec.Command("bash", "-c", fmt.Sprintf("cp $(which %s) %s", command, path))

		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}

	path = filepath.Join(dir, "dev")
	os.MkdirAll(path, os.ModePerm)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("mount --rbind /dev %s", path))
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	path = filepath.Join(dir, "proc")
	os.MkdirAll(path, os.ModePerm)
	cmd = exec.Command("bash", "-c", fmt.Sprintf("mount -t proc /proc %s", path))
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	path = filepath.Join(dir, "sys")
	os.MkdirAll(path, os.ModePerm)
	cmd = exec.Command("bash", "-c", fmt.Sprintf("mount -t sysfs /sys %s", path))
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	path = filepath.Join(dir, "usr")
	os.MkdirAll(path, os.ModePerm)
	cmd = exec.Command("bash", "-c", fmt.Sprintf("mount -o bind /usr %s", path))
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	path = filepath.Join(dir, "lib")
	os.MkdirAll(path, os.ModePerm)
	cmd = exec.Command("bash", "-c", fmt.Sprintf("mount -o bind /lib %s", path))
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	path = filepath.Join(dir, "lib64")
	os.MkdirAll(path, os.ModePerm)
	cmd = exec.Command("bash", "-c", fmt.Sprintf("mount -o bind /lib64 %s", path))
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	path = filepath.Join(dir, "run")
	os.MkdirAll(path, os.ModePerm)
	cmd = exec.Command("bash", "-c", fmt.Sprintf("mount -o bind /run %s", path))
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	path = filepath.Join(dir, "sumologic")
	os.MkdirAll(path, os.ModePerm)
	cmd = exec.Command("bash", "-c", fmt.Sprintf("cp -r ../.. %s", path))
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	path = filepath.Join(dir, "etc/ssl")
	os.MkdirAll(path, os.ModePerm)
	cmd = exec.Command("bash", "-c", fmt.Sprintf("cp -r /etc/ssl/certs %s", path))
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	path = filepath.Join(dir, "etc")
	os.MkdirAll(path, os.ModePerm)
	cmd = exec.Command("bash", "-c", fmt.Sprintf("cp -r /etc/resolv.conf %s", path))
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	cmd = exec.Command("bash", "-c", fmt.Sprintf("cp -r /etc/sudoers %s", path))
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	cmd = exec.Command("bash", "-c", fmt.Sprintf("cp -r /etc/passwd %s", path))
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	path = filepath.Join(dir, "tmp")
	os.MkdirAll(path, os.ModePerm)

	err = syscall.Chroot(dir)
	require.NoError(t, err)
}

func removeEnvironment(t *testing.T, root *os.File) {
	t.Log("Cleaning up env")
	defer root.Close()

	dir := os.TempDir()

	err := root.Chdir()
	require.NoError(t, err)

	err = syscall.Chroot(".")
	require.NoError(t, err)

	path := filepath.Join(dir, "dev")
	cmd := exec.Command("bash", "-c", fmt.Sprintf("mount --make-rslave %s", path))
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	for _, subdir := range []string{
		"dev",
		"proc",
		"sys",
		"usr",
		"lib",
		"lib64",
		"run",
	} {
		path := filepath.Join(dir, subdir)
		cmd := exec.Command("bash", "-c", fmt.Sprintf("umount -R %s", path))
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
}
