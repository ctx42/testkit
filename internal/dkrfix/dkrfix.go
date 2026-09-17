// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package dkrfix owns the example Dockerfiles shared by the testkit packages.
// Each fixture is embedded once here and handed out as a [Fixture], so the
// packages writing or building them cannot drift apart.
//
// The fixtures are whole, self-contained Dockerfiles: each one builds on its
// own with "docker build -f", and each one takes its base image from the
// C42_BLD_IMG_BASE build argument so a test can point it at an image it
// already has.
package dkrfix

import (
	_ "embed"

	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/testkit/pkg/oskit"
)

// FileName is the name [Fixture.Write] gives every fixture it writes, and the
// name docker looks for when it is given a build context without an explicit
// file.
const FileName = "Dockerfile"

// Each fixture is embedded by its own directive, so renaming a directory under
// data breaks the build instead of the tests reading it.
var (
	//go:embed data/simple/Dockerfile
	simple string

	//go:embed data/minimal/Dockerfile
	minimal string

	//go:embed data/targets/Dockerfile
	targets string

	//go:embed data/targets-nep/Dockerfile
	targetsNEP string
)

// The example Dockerfiles the package ships.
var (
	// Simple is a single stage Dockerfile stamping the OCI image metadata
	// labels and the C42_* environment variables, and adding the /file0.txt
	// and /entrypoint.sh files the tests inspect.
	Simple = Fixture{content: simple}

	// Minimal is a single stage Dockerfile adding the same files and test
	// labels as [Simple] but no build metadata, which keeps its label and
	// environment sets small enough for a test to assert on in full.
	Minimal = Fixture{content: minimal}

	// Targets is a Dockerfile defining three chained targets named "first",
	// "second" and "third", each with an entrypoint echoing its own target
	// name.
	Targets = Fixture{content: targets}

	// TargetsNEP is a Dockerfile (NEP: no entrypoint) defining the same
	// three chained targets as [Targets], except none of them sets an
	// ENTRYPOINT or a CMD.
	TargetsNEP = Fixture{content: targetsNEP}
)

// Fixture is an example Dockerfile embedded in the package. The zero value is
// an empty Dockerfile; use one of the package level fixtures.
type Fixture struct {
	content string // The embedded Dockerfile.
}

// Content returns the Dockerfile. Multiple calls to this method return the
// same value.
func (fix Fixture) Content() string {
	return fix.content
}

// Write writes the Dockerfile to a file named [FileName] in the dir directory,
// so the directory can be used as a docker build context. On error, it marks
// the test as failed. Returns the path to the written file.
func (fix Fixture) Write(t tester.T, dir string) string {
	t.Helper()
	return oskit.Create(t, fix.content, dir, FileName)
}
