// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package modkit provides test helpers for the Go module under test:
// locating the module root and reading module and Go versions from
// go.mod files.
//
// Root, Path, and Ver resolve paths against the module root and panic
// on failure, since a missing module root cannot be recovered from in
// a test. ModVer and GoVer read an explicit go.mod path and return an
// error instead. Tmp takes a [tester.T] and reports failures through
// it.
package modkit

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctx42/testing/pkg/tester"
)

// Root returns the module root directory path. Starting at the current working
// directory, it will walk directories up till it finds the "go.mod" file. If
// the "go.mod" is not found or an error occurred, it panics.
func Root() string {
	root, err := os.Getwd()
	if err != nil {
		panic("could not determine the working directory")
	}

	pth := root
	for {
		if pth == "/" {
			panic(fmt.Sprintf("could not find go.mod starting at %s", root))
		}
		_, err = os.Open(filepath.Join(pth, "go.mod")) //nolint:gosec
		if err == nil {
			return pth
		}
		pth = filepath.Dir(pth)
	}
}

// Path returns the absolute path relative to the current module root. If
// elem is empty, it returns the absolute path to the module root directory.
// Panics on error.
func Path(elem ...string) string {
	return filepath.Join(append([]string{Root()}, elem...)...)
}

// Tmp creates a module temporary directory with the given subdirectory
// structure. The module temporary directory is always named 'tmp' and is
// located in the same directory as the "go.mod" file. On error, it marks the
// test as failed and returns an empty string. Will panic if the module root
// directory cannot be determined, see [Root] function.
func Tmp(t tester.T, sub string, subs ...string) string {
	t.Helper()
	if sub == "" {
		t.Error("the subdirectory name must not be empty")
		return ""
	}

	sbs := []string{Root(), "tmp", sub}
	for _, dir := range subs {
		if dir == "" {
			t.Error("subdirectory name(s) must not be empty")
			return ""
		}
		sbs = append(sbs, dir)
	}

	pth := filepath.Join(sbs...)
	if err := os.MkdirAll(pth, 0750); err != nil {
		t.Error(err)
		return ""
	}
	t.Cleanup(func() { _ = os.RemoveAll(pth) })
	return pth
}

// Ver returns the version of a module used in go.mod which is in the current
// working directory or any of the parent directories.
func Ver(mod string) string {
	pth := Path("go.mod")
	ver, err := ModVer(pth, mod)
	if err != nil {
		panic(err)
	}
	return ver
}

// ModVer returns the version of a module used in the "go.mod" file pointed by
// the path.
func ModVer(pth, mod string) (string, error) {
	fil, err := os.Open(pth) //nolint:gosec
	if err != nil {
		return "", err
	}

	var cand []string
	scn := bufio.NewScanner(fil)
	for scn.Scan() {
		lin := scn.Text()
		if strings.Contains(lin, mod) {
			cand = append(cand, lin)
		}
	}

	switch len(cand) {
	case 0:
		return "", fmt.Errorf("no package \"%s\" is used", mod)
	case 1:
	default:
		return "", fmt.Errorf("too many package \"%s\" candidates", mod)
	}

	_, ver, _ := strings.Cut(cand[0], mod)
	ver = strings.TrimSpace(ver)
	if fls := strings.Fields(ver); len(fls) > 0 {
		ver = strings.TrimSpace(fls[0])
	}
	if ver == "" {
		return "", fmt.Errorf("no package %s is used", mod)
	}
	return ver, nil
}

// GoVer takes "go.mod" path and returns go version defined in it. The
// [filepath.Join] is used on pth slice to get the path.
func GoVer(pth ...string) (string, error) {
	fil, err := os.Open(filepath.Join(pth...)) //nolint:gosec
	if err != nil {
		return "", err
	}

	var cand []string
	scn := bufio.NewScanner(fil)
	for scn.Scan() {
		lin := scn.Text()
		if strings.HasPrefix(lin, "go ") {
			cand = append(cand, lin)
		}
	}

	errInv := errors.New("invalid go.mod file")
	switch len(cand) {
	case 0:
		return "", errInv
	case 1:
		cand[0], _ = strings.CutPrefix(cand[0], "go ")
	default:
		return "", errInv
	}

	return cand[0], nil
}
