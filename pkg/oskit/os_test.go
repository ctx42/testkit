// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package oskit

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/testkit/pkg/randkit"
	"github.com/ctx42/testkit/pkg/subkit"
)

func Test_Getwd(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := Getwd(tspy)

		// --- Then ---
		want, err := os.Getwd()
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("fail", func(t *testing.T) {
		if runtime.GOOS == "darwin" {
			t.Skip("skipping on darwin")
		}

		// --- SUBPROCESS SETUP ---
		sub := subkit.New(t.Name())
		if sub.InMainProcess() {
			// --- TEST SUBPROCESS OUTPUT AND ERROR ---
			sout, eout, err := sub.Run()
			if e, ok := errors.AsType[*exec.ExitError](err); ok && !e.Success() {
				assert.Contain(t, "getwd: no such file or directory", sout)
				assert.Empty(t, eout)
				return
			}
			t.Log(sout)
			t.Log(eout)
			t.Errorf("Process ran with err %v, want os.Exit(1)", err)
			return
		}
		// --- IN SUBPROCESS ---

		// --- Given ---
		dir := t.TempDir()
		assert.NoError(t, os.Chdir(dir))
		assert.NoError(t, os.Remove(dir))

		// --- Then ---
		Getwd(t)
	})
}

func Test_Setenv(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.Close()

	kv := randkit.Str()
	t.Cleanup(func() { _ = os.Unsetenv(kv) })

	// --- When ---
	Setenv(tspy, kv, kv)

	// --- Then ---
	assert.Equal(t, kv, os.Getenv(kv))
}

func Test_Stat(t *testing.T) {
	t.Run("file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		fi := Stat(tspy, "testdata/file.txt")

		// --- Then ---
		assert.Equal(t, int64(7), fi.Size())
		assert.Equal(t, "file.txt", fi.Name())
		assert.False(t, fi.Mode().IsDir())
		assert.True(t, fi.Mode().IsRegular())
		assert.False(t, fi.IsDir())
	})

	t.Run("construct file path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		fi := Stat(tspy, "testdata", "file.txt")

		// --- Then ---
		assert.Equal(t, int64(7), fi.Size())
		assert.Equal(t, "file.txt", fi.Name())
		assert.False(t, fi.Mode().IsDir())
		assert.True(t, fi.Mode().IsRegular())
		assert.False(t, fi.IsDir())
	})

	t.Run("directory", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		fi := Stat(tspy, "testdata")

		// --- Then ---
		assert.Equal(t, "testdata", fi.Name())
		assert.True(t, fi.Mode().IsDir())
		assert.False(t, fi.Mode().IsRegular())
		assert.True(t, fi.IsDir())
	})

	t.Run("not existing", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		// --- When ---
		Stat(tspy, "not_existing")

		// --- Then ---
		tspy.Finish().AssertExpectations()
	})
}

func Test_Chdir(t *testing.T) {
	t.Run("change directory", func(t *testing.T) {
		// --- Given ---
		wd := Mustwd()
		want := filepath.Join(wd, "testdata", "dir")

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectSetenv(ChdirChainEnvKey, wd)
		tspy.Close()

		// --- When ---
		have := Chdir(tspy, "testdata", "dir")

		// --- Then ---
		assert.Equal(t, want, have)
		assert.Equal(t, Mustwd(), have)
		assert.Equal(t, wd, os.Getenv(ChdirChainEnvKey))

		tspy.Finish()
		assert.Equal(t, wd, Mustwd())
	})

	t.Run("change directory twice", func(t *testing.T) {
		// --- Given ---
		wd := Mustwd()
		wantPth0 := filepath.Join(wd, "testdata", "dir")
		wantPth1 := filepath.Join(wd, "testdata", "dir", "sub")
		wantHist0 := wd
		wantHist1 := addToList(wd, wantPth0)

		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.ExpectSetenv(ChdirChainEnvKey, wantHist0)
		tspy.ExpectSetenv(ChdirChainEnvKey, wantHist1)
		tspy.Close()

		// --- When ---
		have0 := Chdir(tspy, "testdata", "dir")
		have1 := Chdir(tspy, "sub")

		// --- Then ---
		assert.Equal(t, wantPth0, have0)
		assert.Equal(t, wantPth1, have1)

		assert.Equal(t, Mustwd(), have1)
		assert.Equal(t, wantHist1, os.Getenv(ChdirChainEnvKey))

		tspy.Finish()
		assert.Equal(t, wd, Mustwd())
	})

	t.Run("change to not existing directory", func(t *testing.T) {
		// --- Given ---
		wd := Mustwd()

		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("no such file or directory")
		tspy.ExpectLogContain(filepath.Join(wd, "testdata", "not_existing"))
		tspy.Close()

		// --- When ---
		have := Chdir(tspy, "testdata", "not_existing")

		// --- Then ---
		assert.Empty(t, have)
		assert.Empty(t, os.Getenv(ChdirChainEnvKey))
		assert.Equal(t, wd, Mustwd())

		tspy.Finish()
		assert.Equal(t, wd, Mustwd())
	})
}

func Test_Unsetenv(t *testing.T) {
	t.Run("unset existing", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		kv := randkit.Str()
		assert.NoError(t, os.Setenv(kv, kv))
		t.Cleanup(func() { _ = os.Unsetenv(kv) })
		assert.Equal(t, kv, os.Getenv(kv))

		// --- When ---
		Unsetenv(tspy, kv)

		// --- Then ---
		_, exist := os.LookupEnv(kv)
		assert.False(t, exist)

		tspy.Finish()
		have, exist := os.LookupEnv(kv)
		assert.True(t, exist)
		assert.Equal(t, kv, have)
	})

	t.Run("unset not existing", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		kv := randkit.Str()

		// --- When ---
		Unsetenv(tspy, kv)

		// --- Then ---
		_, exist := os.LookupEnv(kv)
		assert.False(t, exist)

		tspy.Finish()
		_, exist = os.LookupEnv(kv)
		assert.False(t, exist)
	})
}

func Test_addToList_tabular(t *testing.T) {
	tt := []struct {
		testN string

		curr string
		path string
		want []string
	}{
		{"1", "", "", []string{""}},
		{"2", "", "/a", []string{"/a"}},
		{"3", "/a", "b", []string{"/a", "b"}},
		{"4", "/a", "", []string{"/a"}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := addToList(tc.curr, tc.path)

			// --- Then ---
			parts := strings.Split(have, string(os.PathListSeparator))
			assert.Equal(t, tc.want, parts)
		})
	}
}

func Test_MkdirTemp(t *testing.T) {
	t.Run("dir and prefix", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectTempDir(0)
		tspy.ExpectCleanups(1)
		tspy.Close()
		dir := t.TempDir()

		// --- When ---
		pth := MkdirTemp(tspy, dir, "dir-*-sfx")

		// --- Then ---
		fi, err := os.Stat(pth)
		assert.NoError(t, err)
		assert.True(t, fi.Mode().IsDir())

		name := filepath.Base(pth)
		assert.True(t, strings.HasPrefix(name, "dir-"))
		assert.True(t, strings.HasSuffix(name, "-sfx"))

		tspy.Finish().AssertExpectations()
		_, err = os.Stat(pth)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("empty dir and set prefix", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectTempDir(0)
		tspy.ExpectCleanups(1)
		tspy.Close()

		// --- When ---
		pth := MkdirTemp(tspy, "", "dir-*-sfx")

		// --- Then ---
		fi, err := os.Stat(pth)
		assert.NoError(t, err)
		assert.True(t, fi.Mode().IsDir())

		name := filepath.Base(pth)
		assert.True(t, strings.HasPrefix(name, "dir-"))
		assert.True(t, strings.HasSuffix(name, "-sfx"))
		assert.Equal(t, filepath.Clean(os.TempDir()), filepath.Dir(pth))

		tspy.Finish().AssertExpectations()
		_, err = os.Stat(pth)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("empty dir and prefix", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectTempDir(0)
		tspy.ExpectCleanups(1)
		tspy.Close()

		// --- When ---
		pth := MkdirTemp(tspy, "", "")

		// --- Then ---
		fi, err := os.Stat(pth)
		assert.NoError(t, err)
		assert.True(t, fi.Mode().IsDir())

		name := filepath.Base(pth)
		assert.NotEmpty(t, name)
		assert.Equal(t, filepath.Clean(os.TempDir()), filepath.Dir(pth))

		tspy.Finish().AssertExpectations()
		_, err = os.Stat(pth)
		assert.True(t, os.IsNotExist(err))
	})
}

func Test_Open(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t).ExpectCleanups(1).Close()

		// --- When ---
		fil := Open(tspy, "testdata/file.txt")

		// --- Then ---
		have, err := io.ReadAll(fil)
		assert.NoError(t, err)
		assert.Equal(t, "content", string(have))
		tspy.Finish().AssertExpectations()
		assert.ErrorIs(t, os.ErrClosed, fil.Close())
	})

	t.Run("join the path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t).ExpectCleanups(1).Close()

		// --- When ---
		fil := Open(tspy, "testdata", "file.txt")

		// --- Then ---
		have, err := io.ReadAll(fil)
		assert.NoError(t, err)
		assert.Equal(t, "content", string(have))
		tspy.Finish().AssertExpectations()
		assert.ErrorIs(t, os.ErrClosed, fil.Close())
	})

	t.Run("fail", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "open testdata/not_existing.txt: no such file or directory"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		// --- When ---
		fil := Open(tspy, "testdata/not_existing.txt")

		// --- Then ---
		assert.Nil(t, fil)
	})
}

func Test_ReadFile(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := ReadFile(tspy, "testdata/file.txt")

		// --- Then ---
		assert.Equal(t, []byte("content"), have)
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := ReadFile(tspy, "testdata", "file.txt")

		// --- Then ---
		assert.Equal(t, []byte("content"), have)
	})

	t.Run("not existing file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "open testdata/not_existing.txt: no such file or directory"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		// --- When ---
		have := ReadFile(tspy, "testdata/not_existing.txt")

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_ReadFileStr(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := ReadFileStr(tspy, "testdata/file.txt")

		// --- Then ---
		assert.Equal(t, "content", have)
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := ReadFileStr(tspy, "testdata", "file.txt")

		// --- Then ---
		assert.Equal(t, "content", have)
	})

	t.Run("not existing file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "open testdata/not_existing.txt: no such file or directory"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		// --- When ---
		have := ReadFileStr(tspy, "testdata/not_existing.txt")

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_Readdirnames(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		lst := Readdirnames(tspy, "testdata/tree")

		// --- Then ---
		sort.Strings(lst)
		assert.Equal(t, []string{"dir0", "dir1", "file.txt"}, lst)
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		lst := Readdirnames(tspy, "testdata", "tree")

		// --- Then ---
		sort.Strings(lst)
		assert.Equal(t, []string{"dir0", "dir1", "file.txt"}, lst)
	})

	t.Run("not existing directory", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		// --- When ---
		lst := Readdirnames(tspy, "testdata", "not_existing")

		// --- Then ---
		assert.Nil(t, lst)
		tspy.Finish().AssertExpectations()
	})
}

func Test_ModTime(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dir := t.TempDir()
		pth := filepath.Join(dir, "test_file.txt")
		must.Value(os.Create(pth))

		// --- When ---
		mt := ModTime(tspy, pth)

		// --- Then ---
		fi := must.Value(os.Stat(pth))
		assert.Equal(t, fi.ModTime().In(time.UTC), mt)
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()
		dir := t.TempDir()
		pth := filepath.Join(dir, "file.txt")
		_, err := os.Create(pth)
		assert.NoError(t, err)

		// --- When ---
		mt := ModTime(tspy, dir, "file.txt")

		// --- Then ---
		fi := must.Value(os.Stat(pth))
		assert.Equal(t, fi.ModTime().In(time.UTC), mt)
	})

	t.Run("not existing file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "stat not_existing_file.txt: no such file or directory"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		pth := "not_existing_file.txt"

		// --- When ---
		mt := ModTime(tspy, pth)

		// --- Then ---
		assert.Zero(t, mt)
	})
}

func Test_ModTimeSet(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dir := t.TempDir()
		pth := filepath.Join(dir, "test_file.txt")
		must.Value(os.Create(pth))
		tim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)

		// --- When ---
		have := ModTimeSet(tspy, tim, pth)

		// --- Then ---
		assert.Equal(t, pth, have)
		assert.Exact(t, tim, must.Value(os.Stat(pth)).ModTime().UTC())
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dir := t.TempDir()
		pth := filepath.Join(dir, "file.txt")
		must.Value(os.Create(pth))
		tim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)

		// --- When ---
		have := ModTimeSet(tspy, tim, dir, "file.txt")

		// --- Then ---
		assert.Equal(t, pth, have)
		assert.Exact(t, tim, must.Value(os.Stat(pth)).ModTime().UTC())
	})

	t.Run("not existing file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "chtimes not_existing_file.txt: no such file or directory"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		pth := "not_existing_file.txt"

		// --- When ---
		have := ModTimeSet(tspy, time.Now(), pth)

		// --- Then ---
		assert.Equal(t, pth, have)
	})
}

func Test_FileSize(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		size := FileSize(tspy, "testdata/file.txt")

		// --- Then ---
		assert.Equal(t, int64(7), size)
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		size := FileSize(tspy, "testdata", "file.txt")

		// --- Then ---
		assert.Equal(t, int64(7), size)
	})

	t.Run("not existing file", func(t *testing.T) {
		// --- Given ---
		pth := "not_existing_file.txt"

		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "stat not_existing_file.txt: no such file or directory"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		// --- When ---
		size := FileSize(tspy, pth)

		// --- Then ---
		assert.Zero(t, size)
	})
}

func Test_Create(t *testing.T) {
	t.Run("create new file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dir := t.TempDir()
		pth := filepath.Join(dir, "file.txt")

		// --- When ---
		have := Create(tspy, []byte("def"), pth)

		// --- Then ---
		assert.Equal(t, pth, have)

		content, err := os.ReadFile(pth)
		assert.NoError(t, err)
		assert.Equal(t, "def", string(content))
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dir := t.TempDir()

		// --- When ---
		have := Create(tspy, []byte("def"), dir, "file.txt")

		// --- Then ---

		pth := filepath.Join(dir, "file.txt")
		assert.Equal(t, pth, have)

		content, err := os.ReadFile(pth)
		assert.NoError(t, err)
		assert.Equal(t, "def", string(content))
	})

	t.Run("does not append to existing file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		pth := filepath.Join(t.TempDir(), "file.txt")
		err := os.WriteFile(pth, []byte("abc"), 0600)
		assert.NoError(t, err)

		// --- When ---
		have := Create(tspy, []byte("def"), pth)

		// --- Then ---
		assert.Equal(t, pth, have)

		content, err := os.ReadFile(pth)
		assert.NoError(t, err)
		assert.Equal(t, "def", string(content))
	})

	t.Run("truncates before override", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		pth := filepath.Join(t.TempDir(), "file.txt")
		err := os.WriteFile(pth, []byte("abc"), 0600)
		assert.NoError(t, err)

		// --- When ---
		have := Create(tspy, []byte("de"), pth)

		// --- Then ---
		assert.Equal(t, pth, have)

		content, err := os.ReadFile(pth)
		assert.NoError(t, err)
		assert.Equal(t, "de", string(content))
	})

	t.Run("string content", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		pth := filepath.Join(t.TempDir(), "file.txt")

		// --- When ---
		have := Create(tspy, "def", pth)

		// --- Then ---
		assert.Equal(t, pth, have)

		content, err := os.ReadFile(pth)
		assert.NoError(t, err)
		assert.Equal(t, "def", string(content))
	})

	t.Run("error - write into not existing directory", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		pth := filepath.Join(t.TempDir(), "not_existing", "file.txt")

		// --- When ---
		have := Create(tspy, "def", pth)

		// --- Then ---
		assert.Equal(t, pth, have)
		assert.NoFileExist(t, pth)
		tspy.Finish().AssertExpectations()
	})
}

func Test_Write(t *testing.T) {
	t.Run("create new file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		pth := filepath.Join(t.TempDir(), "file.txt")

		// --- When ---
		have := Write(tspy, []byte("def"), pth)

		// --- Then ---
		assert.Equal(t, pth, have)

		content, err := os.ReadFile(pth)
		assert.NoError(t, err)
		assert.Equal(t, "def", string(content))
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dir := t.TempDir()

		// --- When ---
		have := Write(tspy, []byte("def"), dir, "file.txt")

		// --- Then ---
		exp := filepath.Join(dir, "file.txt")
		assert.Equal(t, exp, have)

		content, err := os.ReadFile(exp)
		assert.NoError(t, err)
		assert.Equal(t, "def", string(content))
	})

	t.Run("append to existing file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dir := t.TempDir()
		pth := filepath.Join(dir, "file.txt")
		err := os.WriteFile(pth, []byte("abc\n"), 0600)
		assert.NoError(t, err)

		// --- When ---
		have := Write(tspy, []byte("def"), dir, "file.txt")

		// --- Then ---
		assert.Equal(t, pth, have)

		content, err := os.ReadFile(pth)
		assert.NoError(t, err)
		assert.Equal(t, "abc\ndef", string(content))
	})

	t.Run("string content", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		pth := filepath.Join(t.TempDir(), "file.txt")

		// --- When ---
		have := Write(tspy, "def", pth)

		// --- Then ---
		assert.Equal(t, pth, have)

		content, err := os.ReadFile(pth)
		assert.NoError(t, err)
		assert.Equal(t, "def", string(content))
	})

	t.Run("error - write into not existing directory", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		pth := filepath.Join(t.TempDir(), "not_existing", "file.txt")

		// --- When ---
		have := Write(tspy, "def", pth)

		// --- Then ---
		assert.Equal(t, pth, have)
		assert.NoFileExist(t, pth)
		tspy.Finish().AssertExpectations()
	})
}

func Test_MkdirAll(t *testing.T) {
	t.Run("create not existing", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dir := t.TempDir()

		// --- When ---
		have := MkdirAll(tspy, dir, "dir")

		// --- Then ---
		want := filepath.Join(dir, "dir")
		assert.Equal(t, want, have)
		assert.DirExist(t, want)
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dir := t.TempDir()

		// --- When ---
		have := MkdirAll(tspy, dir, "dir", "sub")

		// --- Then ---
		want := filepath.Join(dir, "dir", "sub")
		assert.Equal(t, want, have)
		assert.DirExist(t, want)
	})

	t.Run("create existing", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dir := t.TempDir()

		// --- When ---
		have := MkdirAll(tspy, dir)

		// --- Then ---
		assert.Equal(t, dir, have)
		assert.DirExist(t, dir)
	})

	t.Run("error - creating directory", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("permission denied")
		tspy.Close()

		dir := filepath.Join(t.TempDir(), "read_only")
		must.Nil(os.Mkdir(dir, 0500))
		t.Cleanup(func() { _ = os.Chmod(dir, 0700) })

		// --- When ---
		have := MkdirAll(tspy, dir, "sub")

		// --- Then ---
		want := filepath.Join(dir, "sub")
		assert.Equal(t, want, have)
		assert.NoDirExist(t, want)
		tspy.Finish().AssertExpectations()
	})
}

func Test_PathExists(t *testing.T) {
	t.Run("file exists", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		exists := PathExists(tspy, "testdata/file.txt")

		// --- Then ---
		assert.True(t, exists)
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		exists := PathExists(tspy, "testdata", "file.txt")

		// --- Then ---
		assert.True(t, exists)
	})

	t.Run("directory exists", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		exists := PathExists(tspy, "testdata")

		// --- Then ---
		assert.True(t, exists)
	})

	t.Run("file does not exist", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		exists := PathExists(tspy, "testdata/not_existing.txt")

		// --- Then ---
		assert.False(t, exists)
	})

	t.Run("directory does not exist", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		exists := PathExists(tspy, "not_existing")

		// --- Then ---
		assert.False(t, exists)
	})

	t.Run("symlink exists", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		exists := PathExists(tspy, "testdata/file_sym_link.txt")

		// --- Then ---
		assert.True(t, exists)
	})
}

func Test_List(t *testing.T) {
	t.Run("dir", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := List(tspy, "testdata/list")

		// --- Then ---
		want := []string{
			"d|dir",
			"file0.txt",
			"file1.txt",
		}
		sort.Strings(have)
		assert.Equal(t, want, have)
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := List(tspy, "testdata", "list")

		// --- Then ---
		want := []string{
			"d|dir",
			"file0.txt",
			"file1.txt",
		}
		sort.Strings(have)
		assert.Equal(t, want, have)
	})

	t.Run("not a dir", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogContain("list/file0.txt: not a directory")
		tspy.Close()

		// --- When ---
		have := List(tspy, "testdata/list/file0.txt")

		// --- Then ---
		assert.Empty(t, have)
	})

	t.Run("not existing directory", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogContain("list/not_existing: no such file or directory")
		tspy.Close()

		// --- When ---
		have := List(tspy, "testdata/list/not_existing")

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_ListAbs(t *testing.T) {
	t.Run("dir", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		src := filepath.Join("testdata", "list")
		srcAbs, err := filepath.Abs(src)
		assert.NoError(t, err)

		// --- When ---
		have := ListAbs(tspy, "testdata/list")

		// --- Then ---
		want := []string{
			filepath.Join(srcAbs, "file0.txt"),
			filepath.Join(srcAbs, "file1.txt"),
			"d|" + filepath.Join(srcAbs, "dir"),
		}
		sort.Strings(have)
		assert.Equal(t, want, have)
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		src := filepath.Join("testdata", "list")
		srcAbs, err := filepath.Abs(src)
		assert.NoError(t, err)

		// --- When ---
		have := ListAbs(tspy, "testdata", "list")

		// --- Then ---
		want := []string{
			filepath.Join(srcAbs, "file0.txt"),
			filepath.Join(srcAbs, "file1.txt"),
			"d|" + filepath.Join(srcAbs, "dir"),
		}
		sort.Strings(have)
		assert.Equal(t, want, have)
	})

	t.Run("not a dir", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogContain("list/file0.txt: not a directory")
		tspy.Close()

		// --- When ---
		have := ListAbs(tspy, "testdata/list/file0.txt")

		// --- Then ---
		assert.Empty(t, have)
	})

	t.Run("not existing directory", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogContain("list/not_existing: no such file or directory")
		tspy.Close()

		// --- When ---
		have := ListAbs(tspy, "testdata/list/not_existing")

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_CopyFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dst := t.TempDir()

		// --- When ---
		have := CopyFile(tspy, dst, "testdata/file.txt")

		// --- Then ---
		wPth := filepath.Join(dst, "file.txt")
		assert.FileExist(t, wPth)
		assert.Equal(t, wPth, have)

		haveCont := must.Value(os.ReadFile(have))
		assert.Equal(t, "content", string(haveCont))
	})

	t.Run("not existing source", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("testdata/not_existing.txt")
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		dst := t.TempDir()

		// --- When ---
		have := CopyFile(tspy, dst, "testdata/not_existing.txt")

		// --- Then ---
		wPth := filepath.Join(dst, "not_existing.txt")
		assert.NoFileExist(t, wPth)
		assert.Empty(t, have)
	})

	t.Run("not existing destination", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("destination/file.txt")
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		dst := filepath.Join(t.TempDir(), "destination")

		// --- When ---
		have := CopyFile(tspy, dst, "testdata/file.txt")

		// --- Then ---
		assert.NoDirExist(t, filepath.Join(dst, "destination.txt"))
		assert.Empty(t, have)
	})

	t.Run("empty source path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("open : no such file or directory")
		tspy.Close()

		dst := t.TempDir()

		// --- When ---
		have := CopyFile(tspy, dst, "")

		// --- Then ---
		assert.Empty(t, have)
	})

	t.Run("dot source path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain(". is not a regular file")
		tspy.Close()

		dst := t.TempDir()

		// --- When ---
		have := CopyFile(tspy, dst, ".")

		// --- Then ---
		assert.Empty(t, have)
	})

	t.Run("empty destination path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("empty destination directory")
		tspy.Close()

		// --- When ---
		have := CopyFile(tspy, "", "testdata/file.txt")

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_CopyDir(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dst := t.TempDir()
		src := must.Value(filepath.Abs("testdata/dir"))

		// --- When ---
		have := CopyDir(tspy, dst, src)

		// --- Then ---
		assert.Equal(t, dst, have)

		var files []string
		fn := func(path string, info os.FileInfo, err error) error {
			files = append(files, path)
			return nil
		}
		assert.NoError(t, filepath.Walk(dst, fn))

		want := append(
			[]string{},
			dst,
			filepath.Join(dst, "sub"),
			filepath.Join(dst, "sub", "file.txt"),
			filepath.Join(dst, "sub", "sub_sub"),
			filepath.Join(dst, "sub", "sub_sub", "file.txt"),
		)
		assert.Equal(t, want, files)

		pth := filepath.Join(dst, "sub", "file.txt")
		data := must.Value(os.ReadFile(pth))
		assert.Equal(t, "dir/sub/file.txt", string(data))

		pth = filepath.Join(dst, "sub", "sub_sub", "file.txt")
		data = must.Value(os.ReadFile(pth))
		assert.Equal(t, "dir/sub/sub_sub/file.txt", string(data))
	})

	t.Run("symlinks to files", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dst := t.TempDir()
		src := must.Value(filepath.Abs("testdata/with_sym_link"))

		// --- When ---
		have := CopyDir(tspy, dst, src)

		// --- Then ---
		assert.Equal(t, dst, have)
		data := must.Value(os.ReadFile(filepath.Join(dst, "file.txt")))
		assert.Equal(t, "abc", string(data))
		data = must.Value(os.ReadFile(filepath.Join(dst, "file_sym_link.txt")))
		assert.Equal(t, "abc", string(data))
	})

	t.Run("source is a file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("not a directory")
		tspy.ExpectLogContain("testdata/file.txt")
		tspy.Close()

		dst := t.TempDir()
		src := must.Value(filepath.Abs("testdata/file.txt"))

		// --- When ---
		have := CopyDir(tspy, dst, src)

		// --- Then ---
		assert.Empty(t, have)
	})

	t.Run("not existing source", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("testdata/not_existing")
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		dst := t.TempDir()

		// --- When ---
		have := CopyDir(tspy, dst, "testdata/not_existing")

		// --- Then ---
		wPth := filepath.Join(dst, "not_existing.txt")
		assert.NoFileExist(t, wPth)
		assert.Empty(t, have)
	})

	t.Run("not existing destination", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("destination/sub")
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		dst := filepath.Join(t.TempDir(), "destination")

		// --- When ---
		have := CopyDir(tspy, dst, "testdata/dir")

		// --- Then ---
		wPth := filepath.Join(dst, "destination.txt")
		assert.NoDirExist(t, wPth)
		assert.Empty(t, have)
	})

	t.Run("empty source path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("open : no such file or directory")
		tspy.Close()

		dst := t.TempDir()

		// --- When ---
		have := CopyDir(tspy, dst, "")

		// --- Then ---
		assert.Empty(t, have)
	})

	t.Run("dot source path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dst := t.TempDir()

		current := must.Value(os.Getwd())
		assert.NoError(t, os.Chdir("testdata/dir"))
		defer func() { _ = os.Chdir(current) }()

		// --- When ---
		have := CopyDir(tspy, dst, ".")

		// --- Then ---
		assert.Equal(t, dst, have)

		var files []string
		fn := func(path string, info os.FileInfo, err error) error {
			files = append(files, path)
			return nil
		}
		assert.NoError(t, filepath.Walk(dst, fn))

		want := append(
			[]string{},
			dst,
			filepath.Join(dst, "sub"),
			filepath.Join(dst, "sub", "file.txt"),
			filepath.Join(dst, "sub", "sub_sub"),
			filepath.Join(dst, "sub", "sub_sub", "file.txt"),
		)
		assert.Equal(t, want, files)
		wPth := filepath.Join(dst, "sub", "file.txt")
		assert.FileContain(t, "dir/sub/file.txt", wPth)
		wPth = filepath.Join(dst, "sub", "sub_sub", "file.txt")
		assert.FileContain(t, "dir/sub/sub_sub/file.txt", wPth)
	})

	t.Run("empty destination path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("empty destination directory")
		tspy.Close()

		// --- When ---
		have := CopyDir(tspy, "", "testdata/file.txt")

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_EnvSplit_tabular(t *testing.T) {
	tt := []struct {
		testN string

		env  []string
		want map[string]string
	}{
		{"1", []string{}, map[string]string{}},
		{"1a", []string{""}, map[string]string{}},
		{"2", []string{"A=B"}, map[string]string{"A": "B"}},
		{"3", []string{"A=B=C"}, map[string]string{"A": "B=C"}},
		{"4", []string{"A="}, map[string]string{"A": ""}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := EnvSplit(tc.env)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_EnvSplitOrdered(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		env := []string{
			"key0=val0",
			"key1=val1",
			"key2=val2",
		}

		// --- When ---
		haveMap, haveOrder := EnvSplitOrdered(env)

		// --- Then ---
		wantMap := map[string]string{
			"key0": "val0",
			"key1": "val1",
			"key2": "val2",
		}
		assert.Equal(t, wantMap, haveMap)
		assert.Equal(t, []string{"key0", "key1", "key2"}, haveOrder)
	})

	t.Run("environment variable with empty value", func(t *testing.T) {
		// --- Given ---
		env := []string{
			"key0=val0",
			"key1=",
			"key2=val2",
		}

		// --- When ---
		haveMap, haveOrder := EnvSplitOrdered(env)

		// --- Then ---
		wantMap := map[string]string{
			"key0": "val0",
			"key1": "",
			"key2": "val2",
		}
		assert.Equal(t, wantMap, haveMap)
		assert.Equal(t, []string{"key0", "key1", "key2"}, haveOrder)
	})

	t.Run("environment variable with empty name", func(t *testing.T) {
		// --- Given ---
		env := []string{
			"key0=val0",
			"=val1",
			"key2=val2",
		}

		// --- When ---
		haveMap, haveOrder := EnvSplitOrdered(env)

		// --- Then ---
		wantMap := map[string]string{
			"key0": "val0",
			"key2": "val2",
		}
		assert.Equal(t, wantMap, haveMap)
		assert.Equal(t, []string{"key0", "key2"}, haveOrder)
	})

	t.Run("entry without equal sign", func(t *testing.T) {
		// --- Given ---
		env := []string{
			"key0=val0",
			"key1val1",
			"key2=val2",
		}

		// --- When ---
		haveMap, haveOrder := EnvSplitOrdered(env)

		// --- Then ---
		wantMap := map[string]string{
			"key0": "val0",
			"key2": "val2",
		}
		assert.Equal(t, wantMap, haveMap)
		assert.Equal(t, []string{"key0", "key2"}, haveOrder)
	})

	t.Run("empty entry is skipped", func(t *testing.T) {
		// --- Given ---
		env := []string{
			"key0=val0",
			"",
			"key2=val2",
		}

		// --- When ---
		haveMap, haveOrder := EnvSplitOrdered(env)

		// --- Then ---
		wantMap := map[string]string{
			"key0": "val0",
			"key2": "val2",
		}
		assert.Equal(t, wantMap, haveMap)
		assert.Equal(t, []string{"key0", "key2"}, haveOrder)
	})
}

func Test_EnvJoin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		m := map[string]string{
			"key0": "val0",
			"key1": "val1",
		}

		// --- When ---
		have := EnvJoin(m)

		// --- Then ---
		sort.Strings(have)
		want := []string{"key0=val0", "key1=val1"}
		sort.Strings(want)
		assert.Equal(t, want, have)
	})

	t.Run("empty key names are skipped", func(t *testing.T) {
		// --- Given ---
		m := map[string]string{
			"key0": "val0",
			"key1": "val1",
			"":     "val2",
		}

		// --- When ---
		have := EnvJoin(m)

		// --- Then ---
		want := []string{"key0=val0", "key1=val1"}
		sort.Strings(have)
		assert.Equal(t, want, have)
	})
}
