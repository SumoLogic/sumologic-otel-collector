package commandtest

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// WrapTestMain can be used in TestMain to wrap the go test binary with basic
// command emulation, normalizing the test environment across platforms.
//
// Usage:
//
//	func TestMain(m *testing.M) {
//		commandtest.WrapTestMain(m)
//	}
func WrapTestMain(m *testing.M) {
	flag.Parse()

	pid := os.Getpid()
	if os.Getenv("GO_EXEC_TEST_PID") == "" {
		os.Setenv("GO_EXEC_TEST_PID", strconv.Itoa(pid))
		os.Exit(m.Run())
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	command, args := args[0], args[1:]
	switch command {
	case "echo":
		fmt.Fprintf(os.Stdout, "%s", strings.Join(args, " "))
	case "exit":
		if i, err := strconv.ParseInt(args[0], 10, 32); err == nil {
			os.Exit(int(i))
		}
		panic("unexpected exit argument")
	case "sleep":
		if d, err := time.ParseDuration(args[0]); err == nil {
			time.Sleep(d)
			return
		}
		if i, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			time.Sleep(time.Second * time.Duration(i))
			return
		}
	case "fork":
		childCommand := exec.CommandContext(context.Background(), os.Args[0], args...)
		if err := childCommand.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "fork error: %v", err)
			os.Exit(3)
		}
	}
}

// WrapCommand adjusts a command and arguments to run emuldated by a go test
// binary wrapped with WrapTestMain
func WrapCommand(cmd string, args []string) (string, []string) {
	return os.Args[0], append([]string{cmd}, args...)
}
