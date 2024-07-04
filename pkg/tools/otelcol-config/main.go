package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

type errorCoder interface {
	ErrorCode() int
}

func stderrOrBust(args ...interface{}) {
	if _, err := fmt.Fprintln(os.Stderr, args...); err != nil {
		panic(err)
	}
}

func exit(err error) {
	code := 1
	if e, ok := err.(errorCoder); ok {
		code = e.ErrorCode()
	}
	os.Exit(code)
}

// visitFlags visits all of the flags and runs their associated action. Only
// flags that have been explicitly set will be acted on. This means that there
// are no actions taken by default, when flags are omitted.
func visitFlags(fs *pflag.FlagSet, ctx *actionContext) error {
	flags := []*pflag.Flag{}

	fs.Visit(func(flag *pflag.Flag) {
		flags = append(flags, flag)
	})

	for _, flag := range flags {
		name := flag.Name
		action := flagActions[name]
		if action == nil {
			return fmt.Errorf("developer error: action undefined: %s", name)
		}
		if err := action(ctx); err != nil {
			return err
		}
	}

	return nil
}

func getConfDWriter(values *flagValues, fileName string) func(doc []byte) error {
	return func(doc []byte) (err error) {
		return os.WriteFile(filepath.Join(values.ConfigDir, ConfDotD, fileName), doc, 0600)
	}
}

// main. here is what it does:
//
// 1. parse flags, or exit 2 on failure
// 2. visit flags alphabetically according to flagActions, or exit on failure
func main() {
	flagValues := newFlagValues()
	fs := makeFlagSet(flagValues)

	if err := fs.Parse(os.Args); err != nil {
		stderrOrBust(err)
		os.Exit(2)
	}

	ctx := makeActionContext(flagValues)
	if err := visitFlags(fs, ctx); err != nil {
		stderrOrBust(err)
		exit(err)
	}

	os.Exit(0)
}
