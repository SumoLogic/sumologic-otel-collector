//go:build linux || darwin

package sumologic_scripts_tests

import (
	"io/fs"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func checkAbortedDueToNoToken(c check) {
	require.Greater(c.test, len(c.output), 1)
	require.Contains(c.test, c.output[len(c.output)-2], "Installation token has not been provided. Please set the 'SUMOLOGIC_INSTALLATION_TOKEN' environment variable.")
	require.Contains(c.test, c.output[len(c.output)-1], "You can ignore this requirement by adding '--skip-installation-token argument.")
}

func preActionMockConfig(c check) {
	err := os.MkdirAll(etcPath, fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)

	f, err := os.Create(configPath)
	require.NoError(c.test, err)

	err = f.Chmod(fs.FileMode(configPathFilePermissions))
	require.NoError(c.test, err)
}

func preActionMockUserConfig(c check) {
	err := os.MkdirAll(etcPath, fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)

	err = os.MkdirAll(confDPath, fs.FileMode(configPathDirPermissions))
	require.NoError(c.test, err)

	f, err := os.Create(userConfigPath)
	require.NoError(c.test, err)

	err = f.Chmod(fs.FileMode(commonConfigPathFilePermissions))
	require.NoError(c.test, err)
}

func PathHasOwner(t *testing.T, path string, ownerName string, groupName string) {
	info, err := os.Stat(path)
	require.NoError(t, err)

	// get the owning user and group
	stat := info.Sys().(*syscall.Stat_t)
	uid := strconv.FormatUint(uint64(stat.Uid), 10)
	gid := strconv.FormatUint(uint64(stat.Gid), 10)

	usr, err := user.LookupId(uid)
	require.NoError(t, err)

	group, err := user.LookupGroupId(gid)
	require.NoError(t, err)

	require.Equal(t, ownerName, usr.Username, "%s should be owned by user '%s'", path, ownerName)
	require.Equal(t, groupName, group.Name, "%s should be owned by group '%s'", path, groupName)
}
