// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package dkrkit wraps the docker CLI in Go helper functions for use in
// integration tests. It provides two APIs:
//
//   - [Docker]: returns errors; suitable for production-facing test helpers.
//   - [DockerT]: wraps [Docker] and calls t.Error on failure; designed for
//     use directly inside test functions.
//
// Package-level assertion helpers ([HasLabel], [HasEnv], etc.) inspect images
// and containers and report failures via a [tester.T].
package dkrkit

import (
	_ "embed"

	"github.com/ctx42/testing/pkg/check"
	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/testing/pkg/tester"
)

// exDkf represents minimal dockerfile contents.
//
//go:embed testdata/simple/Dockerfile
var exDkf []byte

// Test image constants used in integration tests.
const (
	TestImageName    = "busybox"
	TestImageTag     = "1.38-uclibc"
	TestImageBaseRef = TestImageName + ":" + TestImageTag
)

// Test-only DockerT image labels set by testdata Dockerfiles to verify
// that helpers read both empty and non-empty label values correctly.
const (
	// labTestEmpty is a test-only label with an intentionally empty
	// value. Used in tests to verify that helpers handle labels with
	// empty values correctly.
	labTestEmpty = "com.ctx42.test.empty"

	// labTestValue is a test-only label set to "value". Used in tests to
	// verify that helpers correctly detect and read non-empty label
	// values.
	labTestValue = "com.ctx42.test.value"
)

// Test-only and build-control environment variables. CTX42_TEST_* vars
// are baked into the image; CTX42_BUILD_* vars control build behavior.
const (
	// envTestEmpty is a test-only environment variable with an empty value.
	envTestEmpty = "CTX42_TEST_EMPTY"

	// envTestValue is a test-only environment variable with a non-empty value.
	envTestValue = "CTX42_TEST_VALUE"

	// envBuildNoCache, when set to any non-empty value in the test
	// runner environment, forces DockerT builds to skip the build cache.
	// Equivalent to passing --no-cache to docker build.
	envBuildNoCache = "CTX42_BUILD_NO_CACHE"
)

// HasBuildArg asserts a not nil set has the key with the "want" value.
// Returns true if it does, otherwise marks the test as failed, writes the
// error message to the test log, and returns false.
func HasBuildArg(
	t tester.T,
	set map[string]*string,
	want, key string,
	opts ...any,
) bool {
	t.Helper()

	ops := check.DefaultOptions(opts...)
	have, err := check.HasKey(key, set, opts...)
	if err != nil {
		t.Error(notice.From(err, "build args"))
		return false
	}

	if have != nil && want == *have {
		return true
	}
	msg := notice.New("[build args] expected the map to have a key with value").
		Append("key", "%s", ops.Dumper.Any(key)).
		Want("%s", ops.Dumper.Any(want)).
		Have("%s", ops.Dumper.Any(have))
	t.Error(msg)
	return false
}

// HasLabel returns true if the ref has the given label set to the expected
// value. The ref may be one of: image ID, image reference, or container ID.
// Fails the test and returns false if the ref or label does not exist, if the
// value differs, or if the call fails.
func HasLabel(t tester.T, ref, label, want string) bool {
	t.Helper()
	lbs, err := New().Labels(ref)
	if err != nil {
		t.Error(notice.From(err, "getting labels"))
		return false
	}
	have, exist := lbs[label]
	if !exist {
		msg := notice.New("[getting labels] expected label to exist").
			Append("ref", "%s", ShortID(ref)).
			Want("%q", label).
			Append("labels", "%s", formatMapKeys(lbs))
		t.Error(msg)
		return false
	}
	if want == have {
		return true
	}
	msg := notice.New("[getting labels] expected label value").
		Append("ref", "%s", ShortID(ref)).
		Append("label", "%q", label).
		Want("%q", want).
		Append("have", "%q", have)
	t.Error(msg)
	return false
}

// HasNoLabel returns true if the ref has no label with the given name. The ref
// may be one of: image ID, image reference, or container ID. Fails the test
// and returns false if the label exists or the call fails.
func HasNoLabel(t tester.T, ref, label string) bool {
	t.Helper()
	lbs, err := New().Labels(ref)
	if err != nil {
		t.Error(notice.From(err, "getting labels"))
		return false
	}
	if _, exist := lbs[label]; !exist {
		return true
	}
	msg := notice.New("[getting labels] expected label not to exist").
		Append("ref", "%s", ShortID(ref)).
		Append("label", "%q", label)
	t.Error(msg)
	return false
}

// HasLabels returns true if the ref has all labels in want. The ref may
// have more labels than listed in want. The ref may be one of: image ID, image
// reference, or container ID. Fails the test and returns false on error.
func HasLabels(t tester.T, ref string, want map[string]string) bool {
	t.Helper()
	lbs, err := New().Labels(ref)
	if err != nil {
		t.Error(notice.From(err, "getting labels"))
		return false
	}
	if err = check.MapSubset(want, lbs); err != nil {
		err = notice.From(err, "getting labels").
			Prepend("ref", "%s", ShortID(ref))
		t.Error(err)
		return false
	}
	return true
}

// HasEnv returns true if the ref has the given environment variable set to
// the expected value. The ref may be one of: image ID, image reference, or
// container ID. Fails the test and returns false if the ref or variable does
// not exist, if the value differs, or if the call fails.
func HasEnv(t tester.T, ref, name, want string) bool {
	t.Helper()
	envs, err := New().Envs(ref)
	if err != nil {
		t.Error(err)
		return false
	}
	have, exist := envs[name]
	if !exist {
		mHeader := "[getting environment variable] expected variable to exist"
		msg := notice.New(mHeader).
			Append("ref", "%s", ShortID(ref)).
			Want("%q", name).
			Append("env", "%s", formatMapKeys(envs))
		t.Error(msg)
		return false
	}
	if want == have {
		return true
	}
	mHeader := "[getting environment variable] expected variable value"
	msg := notice.New(mHeader).
		Append("ref", "%s", ShortID(ref)).
		Append("name", "%q", name).
		Want("%q", want).
		Have("%q", have)
	t.Error(msg)
	return false
}

// HasNoEnv returns true if the ref has no environment variable with the given
// name. The ref may be one of: image ID, image reference, or container ID.
// Fails the test and returns false if the variable exists or the call fails.
func HasNoEnv(t tester.T, ref, name string) bool {
	t.Helper()
	envs, err := New().Envs(ref)
	if err != nil {
		t.Error(err)
		return false
	}
	val, exist := envs[name]
	if !exist {
		return true
	}
	mHeader := "[getting environment variable] expected variable not to exist"
	msg := notice.New(mHeader).
		Append("ref", "%s", ShortID(ref)).
		Append("name", "%q", name).
		Append("has value", "%q", val)
	t.Error(msg)
	return false
}

// HasEnvs returns true if the ref has all environment variables in want.
// The ref may have more variables than listed in want. The ref may be one of:
// image ID, image reference, or container ID. Fails the test and returns false
// on error.
func HasEnvs(t tester.T, ref string, want map[string]string) bool {
	t.Helper()
	lbs, err := New().Envs(ref)
	if err != nil {
		t.Error(notice.From(err, "getting environment variables"))
		return false
	}
	if err = check.MapSubset(want, lbs); err != nil {
		err = notice.From(err, "getting environment variables").
			Prepend("ref", "%s", ShortID(ref))
		t.Error(err)
		return false
	}
	return true
}
