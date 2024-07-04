package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetSortedActions(t *testing.T) {
	// test that actions are carried out in a specific order
	flagValues := newFlagValues()
	fs := makeFlagSet(flagValues)
	if err := fs.Parse([]string{"--read-kv", ".sumologic", "--add-tag", "foo=bar", "--set-installation-token", "foo"}); err != nil {
		t.Fatal(err)
	}
	sortedActions := getSortedActions(fs)
	exp := []string{flagAddTag, flagSetInstallationToken, flagReadKV}
	if got, want := sortedActions, exp; !cmp.Equal(got, want) {
		t.Errorf("actions not correctly sorted: got %v, want %v", got, want)
	}
}
