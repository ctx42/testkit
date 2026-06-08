// Package subkit runs Go tests in a child "go test" process. The primary use
// case is code that terminates the process (os.Exit, log.Fatal) — calls that
// kill the entire test binary and cannot be tested in-process. A sentinel
// environment variable lets the same test function detect whether it is
// running in the parent or the child via [SubProcess.InMainProcess] and
// [SubProcess.InSubProcess].
package subkit

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

// SubProcess simplifies running 'go test' in subprocesses.
type SubProcess struct {
	name string // Test or package name to run tests on.
	pkg  bool   // True when name represents a package path.
}

// New returns a subprocess helper scoped to a single test by name.
//
// Both the parent and the child call New with the same name. Run sets
// a sentinel environment variable before launching the child, so
// InMainProcess and InSubProcess can branch the same test function
// into two paths: the parent asserts; the child runs the code under
// test.
//
//	func TestFatal(t *testing.T) {
//		sub := subkit.New(t.Name())
//		if sub.InMainProcess() {
//			sout, eout, err := sub.Run()
//			// assert sout, eout, err …
//			return
//		}
//		// In subprocess: run the code under test.
//		mypackage.FunctionThatCallsOsExit()
//	}
func New(name string) *SubProcess {
	return &SubProcess{
		name: name,
		pkg:  false,
	}
}

// NewPkg returns a 'go test' subprocess testing helper for a package.
// pkg must be an import path accepted by "go test", for example
// "./pkg/foo" or "github.com/ctx42/testkit/pkg/foo". Use [New] to
// target a specific test by name instead.
func NewPkg(pkg string) *SubProcess {
	return &SubProcess{
		name: pkg,
		pkg:  true,
	}
}

// InSubProcess returns true if we are in a subprocess.
func (sp *SubProcess) InSubProcess() bool {
	return os.Getenv(sp.envName()) == "1"
}

// InMainProcess returns true if we are in the main process.
func (sp *SubProcess) InMainProcess() bool {
	return !sp.InSubProcess()
}

// Run runs 'go test' in a subprocess and returns stdout and stderr.
// The returned error is the exit error from "go test"; a non-zero exit
// code indicates test failures, not an execution error.
func (sp *SubProcess) Run() (string, string, error) {
	args := []string{"test", "-v"}
	arg := sp.name
	if !sp.pkg {
		arg = "-test.run=" + arg
	}
	args = append(args, arg)

	sout := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), sp.envName()+"=1")
	cmd.Stdout = sout
	cmd.Stderr = eout
	err := cmd.Run()
	return sout.String(), eout.String(), err
}

// envNamePrefix is the prefix for the environment variable name used to
// indicate that a process is a subprocess.
const envNamePrefix = "GO_TEST_SUBPROCESS_"

// envName returns the environment variable name used to detect whether
// we are in a 'go test' subprocess.
func (sp *SubProcess) envName() string {
	name := sp.name
	if sp.pkg {
		name = strings.ReplaceAll(name, "/", "_")
	}
	return envNamePrefix + name
}

// GetCovProfile returns the "-test.coverprofile" argument and its value
// (path to the coverage file) if present in args, nil otherwise. It is
// typically called with [os.Args].
func GetCovProfile(args []string) []string {
	for i, arg := range args {
		if arg == "-test.coverprofile" && i+1 < len(args) {
			return []string{"-test.coverprofile", args[i+1]}
		}
	}
	return nil
}
