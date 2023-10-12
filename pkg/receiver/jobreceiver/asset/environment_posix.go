//go:build !windows
// +build !windows

package asset

// POSIX compliant platforms use case-sensitive variables, no coercion
// required.
func coerceKey(k string) string {
	return k
}
