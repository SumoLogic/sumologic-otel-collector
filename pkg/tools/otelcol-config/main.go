package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/pflag"
)

const (
	hostmetricsLinux  = "hostmetrics-linux.yaml"
	hostmetricsDarwin = "hostmetrics-darwin.yaml"
)

// errorCoder is here to give actions a way to set the exit status of the program
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
	sortedActions := getSortedActions(fs)
	for _, name := range sortedActions {
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

func getSortedActions(fs *pflag.FlagSet) []string {
	flags := []*pflag.Flag{}

	fs.Visit(func(flag *pflag.Flag) {
		flags = append(flags, flag)
	})

	actions := map[string]struct{}{}

	for _, flag := range flags {
		actions[flag.Name] = struct{}{}
	}

	sortedActions := []string{}
	for _, action := range actionOrder {
		if _, ok := actions[action]; ok {
			sortedActions = append(sortedActions, action)
		}
	}

	return sortedActions
}

func getConfDWriter(values *flagValues, fileName string) func(doc []byte) (int, error) {
	return func(doc []byte) (int, error) {
		return len(doc), os.WriteFile(filepath.Join(values.ConfigDir, ConfDotD, fileName), doc, 0600)
	}
}

func getSumologicRemoteWriter(values *flagValues) func([]byte) (int, error) {
	docPath := filepath.Join(values.ConfigDir, SumologicRemoteDotYaml)
	return func(doc []byte) (int, error) {
		if doc == nil {
			// Special case: when doc is nil, we delete the file. This tells
			// the packaging that it should not use --remote-config for
			// otelcol-sumo.
			return 0, os.Remove(docPath)
		}
		return len(doc), os.WriteFile(docPath, doc, 0600)
	}
}

func getHostMetricsFilename() string {
	switch runtime.GOOS {
	case "linux":
		return hostmetricsLinux
	case "darwin":
		return hostmetricsDarwin
	default:
		panic("unsupported os: " + runtime.GOOS)
	}
}

func isLinkError(err error) bool {
	_, linkError := err.(*os.LinkError)
	return linkError
}

func getHostMetricsLinker(values *flagValues) func() error {
	filename := getHostMetricsFilename()
	hostmetricsAvailPath := filepath.Join(values.ConfigDir, ConfDotDAvailable, filename)
	hostmetricsPath := filepath.Join(values.ConfigDir, ConfDotD, filename)
	return func() error {
		if err := os.Symlink(hostmetricsAvailPath, hostmetricsPath); isLinkError(err) {
			// if the link already exists, hostmetrics are already enabled
			return nil
		} else {
			return err
		}
	}
}

func getHostMetricsUnlinker(values *flagValues) func() error {
	filename := getHostMetricsFilename()
	hostmetricsPath := filepath.Join(values.ConfigDir, ConfDotD, filename)
	return func() error {
		if err := os.Remove(hostmetricsPath); err == os.ErrNotExist {
			// if the link does not exist, hostmetrics are already disabled
			return nil
		} else {
			return nil
		}
	}
}

// main. here is what it does:
//
// 1. Check basic OS compatibility
// 1. Parse flags, or exit 2 on failure
// 2. Visit flags alphabetically according to flagActions, or exit on failure
func main() {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		stderrOrBust(fmt.Errorf("unsupported OS: %s", runtime.GOOS))
		os.Exit(1)
	}

	flagValues := newFlagValues()
	fs := makeFlagSet(flagValues)

	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fs.PrintDefaults()
		os.Exit(2)
	}

	if err := fs.Parse(os.Args); err != nil {
		stderrOrBust(err)
		os.Exit(2)
	}

	ctx := &actionContext{
		ConfigDir:            os.DirFS(flagValues.ConfigDir),
		Flags:                flagValues,
		Stdout:               os.Stdout,
		Stderr:               os.Stderr,
		WriteConfD:           getConfDWriter(flagValues, ConfDSettings),
		WriteConfDOverrides:  getConfDWriter(flagValues, ConfDOverrides),
		WriteSumologicRemote: getSumologicRemoteWriter(flagValues),
		LinkHostMetrics:      getHostMetricsLinker(flagValues),
		UnlinkHostMetrics:    getHostMetricsUnlinker(flagValues),
	}

	if err := visitFlags(fs, ctx); err != nil {
		stderrOrBust(err)
		exit(err)
	}

	os.Exit(0)
}
