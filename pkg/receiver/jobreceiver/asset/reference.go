package asset

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Reference to an installed runtime asset
type Reference struct {
	Name string
	Path string
}

// MergeEnvironment merges the environment variables associated with a runtime
// asset into an existing set of environment variables.
//
// Includes executables in the PATH, shared libraries in LD_LIBRARY_PATH and
// include source in CPATH. Also introduces an {{ asset name }}_PATH variable
// pointing at the installation directory.
func (r Reference) MergeEnvironment(env []string) []string {
	assetEnv := []string{
		fmt.Sprintf("PATH=%s", joinPaths(filepath.Join(r.Path, binDir), "${PATH}")),
		fmt.Sprintf("LD_LIBRARY_PATH=%s", joinPaths(filepath.Join(r.Path, libDir), "${LD_LIBRARY_PATH}")),
		fmt.Sprintf("CPATH=%s", joinPaths(filepath.Join(r.Path, includeDir), "${CPATH}")),
	}
	for i, envVar := range assetEnv {
		// ExpandEnv replaces ${var} with the contents of var from the current
		// environment, or an empty string if var doesn't exist.
		assetEnv[i] = os.ExpandEnv(envVar)
	}

	assetEnv = append(assetEnv, fmt.Sprintf("%s=%s",
		fmt.Sprintf("%s_PATH", key(r.Name)),
		r.Path,
	))

	return mergeEnvironments(env, assetEnv)
}

// joinPath writes a path with its platform specific separator. For example:
// PATH={{ joinPath("/bin/sh") }}${PATH} would yield "PATH=/bin/sh:${PATH}"
func joinPaths(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}
	if len(paths) == 1 {
		return paths[0]
	}
	var sb strings.Builder
	for _, path := range paths[:len(paths)-1] {
		sb.WriteString(path)
		sb.WriteRune(os.PathListSeparator)
	}
	sb.WriteString(paths[len(paths)-1])
	return sb.String()
}
