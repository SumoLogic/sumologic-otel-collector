//go:build windows

package sumologic_scripts_tests

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

var (
	modAdvapi32                   = syscall.NewLazyDLL("advapi32.dll")
	procGetExplicitEntriesFromACL = modAdvapi32.NewProc("GetExplicitEntriesFromAclW")
)

// A Windows ACL record. Windows represents these as windows.EXPLICIT_ACCESS, which comes with an impractical
// representation of trustees. Instead, we just use a string representation of SIDs.
type ACLRecord struct {
	SID               string
	AccessPermissions windows.ACCESS_MASK
	AccessMode        windows.ACCESS_MODE
}

func checkAbortedDueToNoToken(c check) {
	require.Greater(c.test, len(c.output), 1)
	require.Greater(c.test, len(c.errorOutput), 1)
	// The exact formatting of the error message can be different depending on Powershell version
	errorOutput := strings.Join(c.errorOutput, " ")
	require.Contains(c.test, errorOutput, "Installation token has not been provided.")
	require.Contains(c.test, errorOutput, "Please set the SUMOLOGIC_INSTALLATION_TOKEN environment variable.")
}

func checkEphemeralNotInConfig(p string) func(c check) {
	return func(c check) {
		assert.False(c.test, c.installOptions.ephemeral, "ephemeral was specified")

		conf, err := getConfig(p)
		require.NoError(c.test, err, "error while reading configuration")

		assert.False(c.test, conf.Extensions.Sumologic.Ephemeral, "ephemeral is true")
	}
}

func checkEphemeralInConfig(p string) func(c check) {
	return func(c check) {
		assert.True(c.test, c.installOptions.ephemeral, "ephemeral was not specified")

		conf, err := getConfig(p)
		require.NoError(c.test, err, "error while reading configuration")

		assert.True(c.test, conf.Extensions.Sumologic.Ephemeral, "ephemeral is not true")
	}
}

func checkTokenInConfig(c check) {
	require.NotEmpty(c.test, c.installOptions.installToken, "installation token has not been provided")

	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, c.installOptions.installToken, conf.Extensions.Sumologic.InstallationToken, "installation token is different than expected")
}

func checkTokenInSumoConfig(c check) {
	require.NotEmpty(c.test, c.installOptions.installToken, "installation token has not been provided")

	conf, err := getConfig(configPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, c.installOptions.installToken, conf.Extensions.Sumologic.InstallationToken, "installation token is different than expected")
}

func checkConfigFilesOwnershipAndPermissions(ownerSid string) func(c check) {
	return func(c check) {
		etcPathGlob := filepath.Join(etcPath, "*")
		etcPathNestedGlob := filepath.Join(etcPath, "*", "*")

		for _, glob := range []string{etcPathGlob, etcPathNestedGlob} {
			paths, err := filepath.Glob(glob)
			require.NoError(c.test, err)
			for _, path := range paths {
				var aclRecords []ACLRecord
				info, err := os.Stat(path)
				require.NoError(c.test, err)
				if info.IsDir() {
					if path == opampDPath {
						aclRecords = opampDPermissions
					} else {
						aclRecords = configPathDirPermissions
					}
				} else {
					aclRecords = configPathFilePermissions
				}
				PathHasWindowsACLs(c.test, path, aclRecords)
				PathHasOwner(c.test, path, ownerSid)
			}
		}
	}
}

func PathHasOwner(t *testing.T, path string, ownerSID string) {
	securityDescriptor, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION,
	)
	require.NoError(t, err)

	// get the owning user
	owner, _, err := securityDescriptor.Owner()
	require.NoError(t, err)

	require.Equal(t, ownerSID, owner.String(), "%s should be owned by user '%s'", path, ownerSID)
}

func PathHasWindowsACLs(t *testing.T, path string, expectedACLs []ACLRecord) {
	securityDescriptor, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION,
	)
	require.NoError(t, err)

	// get the ACL entries
	acl, _, err := securityDescriptor.DACL()
	require.NoError(t, err)
	require.NotNil(t, acl)
	entries, err := GetExplicitEntriesFromACL(acl)
	require.NoError(t, err)
	aclRecords := []ACLRecord{}
	for _, entry := range entries {
		aclRecord := ExplicitEntryToACLRecord(entry)
		if aclRecord != nil {
			aclRecords = append(aclRecords, *aclRecord)
		}
	}
	assert.Equal(t, expectedACLs, aclRecords, "invalid ACLs for %s", path)
}

// GetExplicitEntriesFromACL gets a list of explicit entries from an ACL
// This doesn't exist in golang.org/x/sys/windows so we need to define it ourselves.
func GetExplicitEntriesFromACL(acl *windows.ACL) ([]windows.EXPLICIT_ACCESS, error) {
	var pExplicitEntries *windows.EXPLICIT_ACCESS
	var explicitEntriesSize uint64
	// Get dacl
	r1, _, err := procGetExplicitEntriesFromACL.Call(
		uintptr(unsafe.Pointer(acl)),
		uintptr(unsafe.Pointer(&explicitEntriesSize)),
		uintptr(unsafe.Pointer(&pExplicitEntries)),
	)
	if r1 != 0 {
		return nil, err
	}
	if pExplicitEntries == nil {
		return []windows.EXPLICIT_ACCESS{}, nil
	}

	// convert the pointer we got from Windows to a Go slice by doing some gnarly looking pointer arithmetic
	explicitEntries := make([]windows.EXPLICIT_ACCESS, explicitEntriesSize)
	for i := 0; i < int(explicitEntriesSize); i++ {
		elementPtr := unsafe.Pointer(
			uintptr(unsafe.Pointer(pExplicitEntries)) +
				uintptr(i)*unsafe.Sizeof(pExplicitEntries),
		)
		explicitEntries[i] = *(*windows.EXPLICIT_ACCESS)(elementPtr)
	}
	return explicitEntries, nil
}

// ExplicitEntryToACLRecord converts a windows.EXPLICIT_ACCESS to a ACLRecord. If the trustee type is not SID,
// we return nil.
func ExplicitEntryToACLRecord(entry windows.EXPLICIT_ACCESS) *ACLRecord {
	trustee := entry.Trustee
	if trustee.TrusteeType != windows.TRUSTEE_IS_SID {
		return nil
	}
	trusteeSid := (*windows.SID)(unsafe.Pointer(entry.Trustee.TrusteeValue))
	return &ACLRecord{
		SID:               trusteeSid.String(),
		AccessMode:        entry.AccessMode,
		AccessPermissions: entry.AccessPermissions,
	}
}
