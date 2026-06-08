package subkit

import "fmt"

func ExampleNew() {
	sub := New("Test_MyFunc/success")
	if sub.InMainProcess() {
		// sout, eout, err := sub.Run()
		return
	}
	// --- IN SUBPROCESS ---
	// The real test body runs here in the child process.

	// Output:
}

func ExampleNewPkg() {
	sub := NewPkg("github.com/ctx42/testkit/pkg/myservice")
	// sout, eout, err := sub.Run()
	_ = sub
	// Output:
}

func ExampleSubProcess_InSubProcess() {
	sub := New("Test_Example")
	fmt.Println(sub.InSubProcess())
	// Output:
	// false
}

func ExampleSubProcess_InMainProcess() {
	sub := New("Test_Example")
	fmt.Println(sub.InMainProcess())
	// Output:
	// true
}

func ExampleGetCovProfile() {
	args := []string{"-test.v", "-test.coverprofile", "/tmp/cover.out"}
	fmt.Println(GetCovProfile(args))
	// Output:
	// [-test.coverprofile /tmp/cover.out]
}

func ExampleGetCovProfile_notPresent() {
	args := []string{"-test.v", "-test.timeout", "30s"}
	fmt.Println(GetCovProfile(args))
	// Output:
	// []
}
