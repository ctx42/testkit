// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/xdef/pkg/xdef"
)

// TestImage holds identifiers for a minimal Docker image built by
// createMinImage function and shared across all tests in this package.
type TestImage struct {
	iid   string // Docker image ID.
	name  string // Docker image name.
	tag   string // Docker image tag.
	ref   string // Docker image reference.
	title string // The value of [xdef.LabImgTitle] label.
}

// TestImg0, TestImg1, TestImg2 are minimal Docker images built once before all
// tests run and removed afterward. Tests that need a pre-existing image
// without building one themselves should reference these.
var (
	TestImg0 = TestImage{}
	TestImg1 = TestImage{}
)

func TestMain(m *testing.M) { os.Exit(runTests(m)) }

// runTests builds shared test images, runs all tests, then removes the images.
func runTests(m *testing.M) int {
	var err error
	if TestImg0, err = createMinImage("Image0"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	defer func() { _ = removeMinImage(TestImg0) }()

	if TestImg1, err = createMinImage("Image1"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}
	defer func() { _ = removeMinImage(TestImg1) }()

	return m.Run()
}

// createMinImage builds a minimal Docker image based on Dockerfile in
// testdata/minimal directory. The image is tagged with a random name and tag.
// The caller is responsible for removing the image.
func createMinImage(title string) (TestImage, error) {
	img := TestImage{
		name:  "ctx42-tst-min-img-" + randString(),
		tag:   "ctx42-tst-min-tag-" + randString(),
		title: title,
	}
	img.ref = img.name + ":" + img.tag
	idf := filepath.Join(os.TempDir(), img.name)

	args := []string{
		"build",
		"-t", img.ref,
		"-f", "testdata/minimal/Dockerfile",
		"--iidfile", idf,
		"--build-arg", xdef.EnvImgBaseName + "=" + TestImageBaseRef,
		"--build-arg", xdef.EnvImgTitle + "=" + title,
		"testdata/minimal",
	}
	cmd := exec.Command("docker", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		msg := notice.New("failed to build a Docker test image").
			Append("ref", "%s", img.ref).
			Append("out", "%s", string(out))
		return TestImage{}, msg
	}

	id, err := os.ReadFile(idf)
	if err != nil {
		msg := notice.New("failed to read the Docker test image ID").
			Append("ref", "%s", img.ref).
			Append("err", "%s", err)
		_ = removeMinImage(img)
		return TestImage{}, msg
	}
	_ = os.Remove(idf)
	img.iid = strings.TrimSpace(string(id))
	return img, nil
}

// removeMinImage force-removes the Docker image identified by img.ref.
func removeMinImage(img TestImage) error {
	args := []string{"rmi", "-f", img.ref}
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := notice.New("failed to remove the Docker test image").
			Append("ref", "%s", img.ref).
			Append("out", "%s", string(out))
		return msg
	}
	return nil
}
