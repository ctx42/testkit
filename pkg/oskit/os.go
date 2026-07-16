// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package oskit provides test helpers for common [os] operations.
// Every exported function integrates with [tester.T]: on error it
// marks the test as failed, writes a diagnostic to the test log, and
// returns a safe zero value so the test can continue executing.
package oskit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/testing/pkg/tester"
)

// Getwd is a wrapper around [os.Getwd]. Returns the working directory on
// success, otherwise marks the test as failed, writes an error message to the
// test log and returns an empty string.
func Getwd(t tester.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		msg := notice.New("error calling os.Getwd").
			Append("error", "%s", err)
		t.Error(msg)
	}
	return wd
}

// Setenv is a wrapper around [os.Setenv]. On error marks the test as failed
// and writes an error message to test log.
func Setenv(t tester.T, key, val string) {
	t.Helper()
	if err := os.Setenv(key, val); err != nil {
		msg := notice.New("error calling os.Setenv").Append("error", "%s", err)
		t.Error(msg)
	}
}

// Stat is a wrapper around [os.Stat]. The path is constructed from pth and
// elems like in [filepath.Join] function. On error, marks the test as failed,
// writes an error message to test log and returns nil.
func Stat(t tester.T, pth string, elems ...string) os.FileInfo {
	t.Helper()
	pth = filepath.Join(append([]string{pth}, elems...)...)
	fi, err := os.Stat(pth)
	if err != nil {
		msg := notice.New("error calling os.Stat").Append("error", "%s", err)
		t.Error(msg)
		return nil
	}
	return fi
}

// ChdirChainEnvKey is the environment key set by [Chdir] each time it
// is called. The value is a [os.PathListSeparator]-separated list of
// the absolute paths that were the working directory before each call.
//
// Example:
//
//	/dir/sub:/dir/sub/sub_sub
const ChdirChainEnvKey = "__oskit_chdir_prev__"

// Chdir uses [os.Chdir] to change the current working directory to the path
// build from dir and elems (using [filepath.Join]). Returns the absolute path
// to the directory it changed to.
//
// It updates / creates an environment variable ([ChdirChainEnvKey]) with the
// list of absolute paths which were the current working directory (before it
// changed it) and adds cleanup call to change directory to the current working
// directory once the test finishes. On error, it marks the test as failed and
// returns an empty string.
func Chdir(t tester.T, dir string, elems ...string) string {
	t.Helper()
	curr, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return ""
	}

	// We always change directory using the absolute path,
	// we return absolute paths as well.
	if !filepath.IsAbs(dir) {
		if dir, err = filepath.Abs(dir); err != nil {
			t.Error(err)
			return ""
		}
	}

	pth := filepath.Join(append([]string{dir}, elems...)...)
	if err = os.Chdir(pth); err != nil {
		t.Error(err)
		return ""
	}
	t.Cleanup(func() { _ = os.Chdir(curr) })

	// The Setenv line is a safeguard! It will panic if we try to change
	// the directory while running parallel tests.
	history := curr
	if val := os.Getenv(ChdirChainEnvKey); val != "" {
		history = addToList(val, curr)
	}
	t.Setenv(ChdirChainEnvKey, history)
	return pth
}

// Unsetenv uses [os.Unsetenv] to unset environment variable. If the variable
// was set, it sets it to the same value after the test finishes.
func Unsetenv(t tester.T, key string) {
	t.Helper()
	previous, exists := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Cleanup(func() { _ = os.Setenv(key, previous) })
	}
}

// addToList adds the path to the list of paths. Where the list of paths is a
// string separated by [os.PathListSeparator].
func addToList(list, path string) string {
	if list == "" {
		return path
	}
	if path == "" {
		return list
	}
	return list + string(os.PathListSeparator) + path
}

// MkdirTemp is a wrapper around [os.MkdirTemp]. The difference between this
// function and t.TempDir is that you can specify the parent directory of the
// directory to create. The created directory will be automatically deleted
// when the test ends.
//
// Example:
//
//	dir := MkdirTemp(t, t.TempDir(), "")
//
// On error, it marks the test as failed and returns an empty string.
func MkdirTemp(t tester.T, dir, prefix string) string {
	t.Helper()
	pth, err := os.MkdirTemp(dir, prefix)
	if err != nil {
		t.Error(err)
		return ""
	}
	t.Cleanup(func() { _ = os.RemoveAll(pth) })
	return pth
}

// Open is a wrapper around [os.Open]. The path is constructed from pth
// and elems like in [filepath.Join] function. The returned file descriptor is
// automatically closed after the test finishes. On error, it marks the test as
// failed and returns nil.
func Open(t tester.T, pth string, elems ...string) *os.File {
	t.Helper()
	pth = filepath.Join(append([]string{pth}, elems...)...)
	fil, err := os.Open(pth) //nolint:gosec
	if err != nil {
		t.Error(err)
		return nil
	}
	t.Cleanup(func() { _ = fil.Close() })
	return fil
}

// ReadFile is a wrapper around [os.ReadFile]. The path is constructed from pth
// and elems like in [filepath.Join] function. On error, it marks the test as
// failed and returns nil.
func ReadFile(t tester.T, pth string, elems ...string) []byte {
	t.Helper()
	pth = filepath.Join(append([]string{pth}, elems...)...)
	buf, err := os.ReadFile(pth) //nolint:gosec
	if err != nil {
		t.Error(err)
		return nil
	}
	return buf
}

// ReadFileStr is a wrapper function for [ReadFile] returning string instead of
// byte slice. The path is constructed from pth and elems like in
// [filepath.Join] function. On error, it marks the test as failed and returns
// an empty string.
func ReadFileStr(t tester.T, pth string, elems ...string) string {
	t.Helper()
	pth = filepath.Join(append([]string{pth}, elems...)...)
	return string(ReadFile(t, pth))
}

// Readdirnames returns a slice of files in the directory dir. The path is
// constructed from dir and elems like in [filepath.Join] function. On error,
// it marks the test as failed and returns nil.
func Readdirnames(t tester.T, dir string, elems ...string) []string {
	t.Helper()
	dir = filepath.Join(append([]string{dir}, elems...)...)
	fh := Open(t, dir)
	if fh == nil {
		return nil
	}
	names, err := fh.Readdirnames(0)
	if err != nil {
		t.Error(err)
		return nil
	}
	sort.Strings(names)
	return names
}

// ModTime calls [os.Stat] and returns modification time in UTC. The path is
// constructed from pth and elems like in [filepath.Join] function. On error,
// it marks the test as failed and returns zero value time.
func ModTime(t tester.T, pth string, elems ...string) time.Time {
	t.Helper()
	pth = filepath.Join(append([]string{pth}, elems...)...)
	fi, err := os.Stat(pth)
	if err != nil {
		t.Error(err)
		return time.Time{}
	}
	return fi.ModTime().In(time.UTC)
}

// ModTimeSet calls [os.Chtimes] and sets access and modification time on a
// file. The path is constructed from pth and elems like in [filepath.Join]
// function. On error, it marks the test as failed. Always returns the same
// path it constructed from pth and elems arguments.
func ModTimeSet(t tester.T, tim time.Time, pth string, elems ...string) string {
	t.Helper()
	pth = filepath.Join(append([]string{pth}, elems...)...)
	err := os.Chtimes(pth, tim, tim)
	if err != nil {
		t.Error(err)
	}
	return pth
}

// FileSize calls [os.Stat] and returns the file size. The path is constructed
// from pth and elems like in [filepath.Join] function. On error, it marks the
// test as failed and returns zero.
func FileSize(t tester.T, pth string, elems ...string) int64 {
	t.Helper()
	pth = filepath.Join(append([]string{pth}, elems...)...)
	fi, err := os.Stat(pth)
	if err != nil {
		t.Error(err)
		return 0
	}
	return fi.Size()
}

// stringOrBytes constrains content arguments to string or []byte.
type stringOrBytes interface{ ~string | ~[]byte }

// Create writes content to a file at the path. The path is constructed from
// pth and elems like in [filepath.Join] function. Content may be a string or
// []byte. If the file exists, it will be truncated, and the content will be
// written at the beginning, otherwise a new file will be created with given
// content. On error, it marks the test as failed. Always returns the same path
// it constructed from pth and elems arguments.
func Create[T stringOrBytes](
	t tester.T,
	content T,
	pth string,
	elems ...string,
) string {

	t.Helper()
	b := []byte(content)
	pth = filepath.Join(append([]string{pth}, elems...)...)
	write(t, b, pth, os.O_CREATE|os.O_WRONLY)
	if err := os.Truncate(pth, int64(len(b))); err != nil {
		t.Error(err)
	}
	return pth
}

// Write writes content to a file at the path. The path is constructed from pth
// and elems like in [filepath.Join] function. Content may be a string or
// []byte. If the file exists, the content will be appended, otherwise the new
// file will be created with the given content. On error, it marks the test as
// failed. Always returns the same path it constructed from pth and elems
// arguments.
func Write[T stringOrBytes](
	t tester.T,
	content T,
	pth string,
	elems ...string,
) string {

	t.Helper()
	pth = filepath.Join(append([]string{pth}, elems...)...)
	return write(t, []byte(content), pth, os.O_CREATE|os.O_APPEND|os.O_WRONLY)
}

// write uses os.OpenFile with given flags to write content to a file at pth.
// On error, it marks the test as failed. Always returns the same path it
// received in the pth argument.
func write(t tester.T, content []byte, pth string, flag int) string {
	t.Helper()
	f, err := os.OpenFile(pth, flag, 0644) //nolint:gosec
	if err != nil {
		t.Error(err.Error())
		return pth
	}
	defer func() { _ = f.Close() }()

	if _, err = f.Write(content); err != nil {
		t.Error(err.Error())
	}
	return pth
}

// MkdirAll creates a directory. The path is constructed from dir and elems
// like in [filepath.Join] function. If the directory exists, it will do
// nothing. The directory is created with 0755 permissions. On error, it marks
// the test as failed but continues execution. Always returns the constructed
// path.
func MkdirAll(t tester.T, dir string, elems ...string) string {
	t.Helper()
	pth := filepath.Join(append([]string{dir}, elems...)...)
	if PathExists(t, pth) {
		return pth
	}
	if err := os.MkdirAll(pth, 0750); err != nil {
		err = fmt.Errorf("failed to create directory: %s, error: %w", pth, err)
		t.Error(err)
	}
	return pth
}

// PathExists returns true if pth exists on the filesystem. The path is
// constructed from pth and elems like in [filepath.Join] function. It doesn't
// matter what kind of file it is. On error, it marks the test as failed but
// continues execution.
func PathExists(t tester.T, pth string, elems ...string) bool {
	t.Helper()
	pth = filepath.Join(append([]string{pth}, elems...)...)
	_, err := os.Stat(pth)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		t.Error(err)
		return false
	}
	return true
}

// List returns a list of files in the given directory. The path to the
// directory is constructed from dir and elems like in [filepath.Join] function.
// All directories are prefixed with the "d|" string. On error, it marks the
// test as failed, returns nil, but continues execution.
func List(t tester.T, dir string, elems ...string) []string {
	t.Helper()
	dir = filepath.Join(append([]string{dir}, elems...)...)
	fh, err := os.Open(dir) //nolint:gosec
	if err != nil {
		t.Error(err)
		return nil
	}
	defer func() { _ = fh.Close() }()
	fls, err := fh.Readdir(0)
	if err != nil {
		t.Error(err)
		return nil
	}

	ret := make([]string, len(fls))
	for i, v := range fls {
		var name string
		if v.IsDir() {
			name = "d|"
		}
		ret[i] = name + v.Name()
	}
	sort.Strings(ret)
	return ret
}

// ListAbs returns a list of absolute paths to files in the given directory.
// The path to the directory is constructed from dir and elems like in
// [filepath.Join] function. All directories are prefixed with the "d|" string.
// On error, it marks the test as failed, returns nil, but continues execution.
func ListAbs(t tester.T, dir string, elems ...string) []string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join(append([]string{dir}, elems...)...))
	if err != nil {
		t.Error(err)
		return nil
	}
	fh, err := os.Open(dir) //nolint:gosec
	if err != nil {
		t.Error(err)
		return nil
	}
	defer func() { _ = fh.Close() }()
	fls, err := fh.Readdir(0)
	if err != nil {
		t.Error(err)
		return nil
	}

	ret := make([]string, len(fls))
	for i, v := range fls {
		var name string
		if v.IsDir() {
			name = "d|"
		}
		ret[i] = name + filepath.Join(dir, v.Name())
	}
	sort.Strings(ret)
	return ret
}

// CopyFile copies a src file to a dst file where src and dst are regular
// files. On success returns value of dst. On error, it marks the test as
// failed, returns an empty string, but continues execution.
func CopyFile(t tester.T, dst, src string) string {
	t.Helper()
	if dst == "" {
		t.Error("empty destination directory")
		return ""
	}
	dst = filepath.Join(dst, filepath.Base(src))

	srcF, err := os.Open(src) //nolint:gosec
	if err != nil {
		t.Error(err)
		return ""
	}
	defer func() { _ = srcF.Close() }()

	fi, err := srcF.Stat()
	if err != nil {
		t.Error(err)
		return ""
	}

	if !fi.Mode().IsRegular() {
		t.Errorf("%s is not a regular file", src)
		return ""
	}

	dstF, err := os.Create(dst) //nolint:gosec
	if err != nil {
		t.Error(err)
		return ""
	}
	defer func() { _ = dstF.Close() }()

	if _, err = io.Copy(dstF, srcF); err != nil {
		t.Error(err)
	}
	return dst
}

// CopyDir recursively copies a src directory to a destination directory. On
// success returns value of dst. On error, it marks the test as failed, returns
// an empty string, but continues execution.
//
// NOTE: It will copy symlinks as regular files.
func CopyDir(t tester.T, dst, src string) string {
	t.Helper()
	if dst == "" {
		t.Error("empty destination directory")
		return ""
	}

	des, err := os.ReadDir(src)
	if err != nil {
		t.Error(err)
		return ""
	}
	for _, ent := range des {
		srcPth := filepath.Join(src, ent.Name())
		fi, err := os.Stat(srcPth)
		if err != nil {
			t.Error(err)
			return ""
		}

		mode := fi.Mode() & os.ModeType
		switch mode {
		case os.ModeDir:
			dstPth := filepath.Join(dst, ent.Name())
			if err := os.Mkdir(dstPth, 0750); err != nil {
				t.Error(err)
				return ""
			}
			CopyDir(t, dstPth, srcPth)

		case 0:
			CopyFile(t, dst, srcPth)

		default:
			format := "only regular files and directories are supported: %s"
			t.Errorf(format, srcPth)
		}
	}
	return dst
}

// EnvSplit parses [os.Environ] results and returns it as the key value map.
func EnvSplit(env []string) map[string]string {
	m, _ := EnvSplitOrdered(env)
	return m
}

// EnvSplitOrdered parses os.Environ results and returns it as a key value map
// and a slice with the order of keys returned by [os.Environ].
func EnvSplitOrdered(env []string) (map[string]string, []string) {
	k := make([]string, 0, 10)
	m := make(map[string]string, 10)
	for _, s := range env {
		if s == "" {
			continue
		}
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 2 {
			if parts[0] == "" {
				continue
			}
			k = append(k, parts[0])
			m[parts[0]] = parts[1]
		}
	}
	return m, k
}

// EnvJoin is reversing [EnvSplit] and joining keys values so they match
// whatever [os.Environ] returns.
func EnvJoin(env map[string]string) []string {
	var ret []string
	for key, val := range env {
		if key == "" {
			continue
		}
		ret = append(ret, key+"="+val)
	}
	return ret
}
