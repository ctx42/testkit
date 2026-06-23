package prjkit

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/check"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/testkit/pkg/exekit"
	"github.com/ctx42/testkit/pkg/oskit"
	"github.com/ctx42/testkit/pkg/randkit"
)

func Test_WithProjectCreate(t *testing.T) {
	t.Run("creates directory", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		root := filepath.Join(t.TempDir(), "project")
		prj := &Project{root: root, t: tspy}

		// --- When ---
		WithProjectCreate(prj)

		// --- Then ---
		assert.DirExist(t, root)
	})

	t.Run("error when the root directory is empty", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "test project root directory must be set"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		prj := &Project{root: "", t: tspy}

		// --- When ---
		assert.Panic(t, func() { WithProjectCreate(prj) })
	})

	t.Run("error when the project directory exists", func(t *testing.T) {
		// --- Given ---
		root := filepath.Join(t.TempDir(), "project")
		must.Nil(os.Mkdir(root, 0644))

		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "expected path to not be an existing directory:\n  path: %s"
		tspy.ExpectLogEqual(wMsg, root)
		tspy.Close()

		prj := &Project{root: root, t: tspy}

		// --- When ---
		WithProjectCreate(prj)
	})
}

func Test_WithProjectEnv(t *testing.T) {
	// --- Given ---
	env := []string{"A=1", "B=2"}
	prj := &Project{}

	// --- When ---
	WithProjectEnv(env)(prj)

	// --- Then ---
	assert.Equal(t, env, prj.env)
}

func Test_NewProject(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		// --- When ---
		prj := New(tspy, "/root/project")

		// --- Then ---
		prj.Close() // Must close to prevent error.
		tspy.Finish()

		assert.Equal(t, "/root/project", prj.root)
		assert.Equal(t, "project", prj.dirName)
		assert.Equal(t, GoModNameStem+"project", prj.modName)
		assert.Equal(t, must.Value(os.Getwd()), prj.testDir)
		assert.Equal(t, "", prj.tmpDir)
		assert.Equal(t, 0, prj.commits)
		assert.Same(t, tspy, prj.t)
	})

	t.Run("root must not be empty", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("test project root directory must be set")
		tspy.Close()

		// --- When ---
		assert.Panic(t, func() { New(tspy, "") })
	})

	t.Run("root must absolute path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFatal()
		wMsg := "test project root directory must be the absolute path"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		// --- When ---
		assert.Panic(t, func() { New(tspy, "dir") })
	})

	t.Run("must be closed at the test end", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogEqual("expected instance to be closed at the test end")
		tspy.Close()

		// --- When ---
		prj := New(tspy, "/dir")

		// --- Then ---
		assert.NotNil(t, prj)
		tspy.Finish()
	})
}

func Test_Project_Root(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, "/dir")
		prj.Close()

		// --- When ---
		have := prj.Root()

		// --- Then ---
		assert.Equal(t, "/dir", have)
	})

	t.Run("on open", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, "/dir")

		// --- When ---
		have := prj.Root()

		// --- Then ---
		assert.Equal(t, "/dir", have)

		prj.Close() // Must close to prevent error.
	})
}

func Test_Project_Path(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.ExpectCleanups(1)
	tspy.Close()

	prj := New(tspy, "/dir")
	prj.Close()

	// --- When ---
	have := prj.Path("a", "b")

	// --- Then ---
	assert.Equal(t, "/dir/a/b", have)
}

func Test_Project_Chdir(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		current := must.Value(os.Getwd())

		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.ExpectSetenv(oskit.ChdirChainEnvKey, current)
		tspy.Close()

		root := t.TempDir()
		prj := New(tspy, root)
		prj.Close()

		// --- When ---
		have := prj.Chdir()

		// --- Then ---
		assert.Equal(t, current, have)
		assert.Equal(t, root, oskit.Getwd(t))
	})

	t.Run("on open", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		before := oskit.Getwd(t)
		prj := New(tspy, t.TempDir())

		// --- When ---
		assert.Panic(t, func() { prj.Chdir() })

		// --- Then ---
		assert.Equal(t, before, oskit.Getwd(t))
	})
}

func Test_Project_ChdirBack(t *testing.T) {
	t.Run("on closed called after Chdir", func(t *testing.T) {
		// --- Given ---
		current := must.Value(os.Getwd())
		root := t.TempDir()

		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.ExpectSetenv(oskit.ChdirChainEnvKey, current)
		tspy.Close()

		prj := New(tspy, root)
		prj.Close()
		prj.Chdir()

		// --- When ---
		prj.ChdirBack()

		// --- Then ---
		assert.Equal(t, current, oskit.Getwd(t))
	})

	t.Run("on open", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		before := must.Value(os.Getwd())
		prj := New(tspy, t.TempDir())

		// --- When ---
		assert.Panic(t, func() { prj.ChdirBack() })

		// --- Then ---
		assert.Equal(t, before, must.Value(os.Getwd()))
	})

	t.Run("without previous call to Chdir", func(t *testing.T) {
		// --- Given ---
		current := must.Value(os.Getwd())
		root := t.TempDir()

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, root)
		prj.Close()

		// --- When ---
		prj.ChdirBack()

		// --- Then ---
		assert.Equal(t, current, must.Value(os.Getwd()))
	})

	t.Run("directory changed by some other code", func(t *testing.T) {
		// --- Given ---
		root := t.TempDir()
		other := t.TempDir()
		current := must.Value(os.Getwd())

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, root)
		prj.Close()

		oskit.Chdir(t, other)

		// --- When ---
		prj.ChdirBack()

		// --- Then ---
		assert.Equal(t, current, must.Value(os.Getwd()))
		tspy.Finish()
		assert.Equal(t, current, must.Value(os.Getwd()))
	})

	t.Run("error changing directory back", func(t *testing.T) {
		// --- Given ---
		current := must.Value(os.Getwd())
		other := t.TempDir()

		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.ExpectSetenv(oskit.ChdirChainEnvKey, current+":"+other)
		tspy.ExpectFatal()
		tspy.IgnoreLogs()
		tspy.Close()

		// Change to directory other than current one.
		oskit.Chdir(t, other)
		prj := New(tspy, t.TempDir())
		prj.Close()
		// Change to root directory saving the current one (other).
		prj.Chdir()
		// Remove other directory, so we cannot get back to it.
		assert.NoError(t, os.RemoveAll(other))

		// --- When ---
		assert.Panic(t, func() { prj.ChdirBack() })
	})
}

func Test_Project_ReadFileStr(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := t.TempDir() // Project root.
		// Project root subdirectory.
		pth := oskit.MkdirAll(t, root, "a")
		fil := oskit.WriteStr(t, "abc", pth, "file*.txt") // File in subdirectory.
		fil = filepath.Base(fil)                          // Filename.

		prj := New(tspy, root)
		prj.Close()

		// --- When ---
		have := prj.ReadFileStr("a", fil)

		// --- Then ---
		assert.Equal(t, "abc", have)
	})

	t.Run("on open", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		pth := filepath.Join(t.TempDir(), "file.txt")

		// --- When ---
		assert.Panic(t, func() { prj.ReadFileStr(pth) })
	})
}

func Test_Project_Exe(t *testing.T) {
	t.Run("prints to stdout and stderr", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		args := []string{
			"--toStdout", "sout",
			"--toStderr", "eout",
		}
		haveSout, haveEout := prj.Exe(os.Args[0], args...)

		// --- Then ---
		assert.Equal(t, "|sout: sout|", haveSout)
		assert.Equal(t, "|eout: eout|", haveEout)
	})

	t.Run("command error", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogContain("exit status 10")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		args := []string{
			"--toStdout", "sout",
			"--toStderr", "eout",
			"--exitCode", "10",
		}
		haveSout, haveEout := prj.Exe(os.Args[0], args...)

		// --- Then ---
		assert.Equal(t, "|sout: sout|", haveSout)
		assert.Equal(t, "|eout: eout|", haveEout)
	})

	t.Run("env passed to exec", func(t *testing.T) {
		// --- Given ---
		kv := randkit.Str()
		env := append(os.Environ(), kv+"="+kv)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir(), WithProjectEnv(env))
		prj.Close()

		// --- When ---
		haveSout, haveEout := prj.Exe("env")

		// --- Then ---
		assert.Contain(t, kv+"="+kv, haveSout)
		assert.Equal(t, "", haveEout)
	})
}

func Test_Project_ExeStdout(t *testing.T) {
	t.Run("prints stdout", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		args := []string{"--toStdout", "sout"}
		have := prj.ExeStdout(os.Args[0], args...)

		// --- Then ---
		assert.Equal(t, "|sout: sout|", have)
	})

	t.Run("prints to stderr and stdout", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogContain("expected empty stderr got: %q", "|eout: eout|")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		args := []string{
			"--toStdout", "sout",
			"--toStderr", "eout",
		}
		have := prj.ExeStdout(os.Args[0], args...)

		// --- Then ---
		assert.Equal(t, "|sout: sout|", have)
	})

	t.Run("command error", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogContain("exit status 10")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		args := []string{
			"--toStdout", "sout",
			"--exitCode", "10",
		}
		have := prj.ExeStdout(os.Args[0], args...)

		// --- Then ---
		assert.Equal(t, "|sout: sout|", have)
	})

	t.Run("env passed to exec", func(t *testing.T) {
		// --- Given ---
		kv := randkit.Str()
		env := append(os.Environ(), kv+"="+kv)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir(), WithProjectEnv(env))
		prj.Close()

		// --- When ---
		args := []string{"--printEnv", kv}
		have := prj.ExeStdout(os.Args[0], args...)

		// --- Then ---
		assert.Equal(t, fmt.Sprintf("|env: %s|", kv), have)
	})
}

func Test_Project_ExeStderr(t *testing.T) {
	t.Run("prints stderr", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		args := []string{"--toStderr", "eout"}
		have := prj.ExeStderr(os.Args[0], args...)

		// --- Then ---
		assert.Equal(t, "|eout: eout|", have)
	})

	t.Run("prints to stderr and stdout", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogContain("expected empty stdout got: %q", "|sout: sout|")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		args := []string{
			"--toStdout", "sout",
			"--toStderr", "eout",
		}
		have := prj.ExeStderr(os.Args[0], args...)

		// --- Then ---
		assert.Equal(t, "|eout: eout|", have)
	})

	t.Run("command error", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogContain("exit status 10")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		args := []string{
			"--toStderr", "eout",
			"--exitCode", "10",
		}
		have := prj.ExeStderr(os.Args[0], args...)

		// --- Then ---
		assert.Equal(t, "|eout: eout|", have)
	})

	t.Run("env passed to exec", func(t *testing.T) {
		// --- Given ---
		kv := randkit.Str()
		env := append(os.Environ(), kv+"="+kv)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir(), WithProjectEnv(env))
		prj.Close()

		// --- When ---
		args := []string{"--printToStderr", "--printEnv", kv}
		have := prj.ExeStderr(os.Args[0], args...)

		// --- Then ---
		assert.Equal(t, fmt.Sprintf("|env: %s|", kv), have)
	})
}

func Test_Project_TempDir(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.ExpectCleanups(1)
	tspy.ExpectTempDir(1)
	tspy.Close()

	prj := New(tspy, "/dir")

	// --- When ---
	have := prj.TempDir()

	// --- Then ---
	assert.Equal(t, prj.tmpDir, have)
	assert.DirExist(t, have)
	prj.Close() // Must close to prevent error.
	tspy.Finish()
	assert.NoDirExist(t, have)
}

func Test_Project_CreateDir(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		have := prj.CreateDir("a", "b")

		// --- Then ---
		assert.DirExist(t, filepath.Join(prj.Root(), "a"))
		assert.DirExist(t, filepath.Join(prj.Root(), "a", "b"))
		assert.Equal(t, filepath.Join(prj.Root(), "a", "b"), have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.CreateDir("a", "b") })
	})
}

func Test_Project_CreateFile(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		have := prj.CreateFile("file0.txt")

		// --- Then ---
		assert.Equal(t, filepath.Join(prj.root, "file0.txt"), have)
		assert.FileContain(t, "", have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("creates subdirectory", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		have := prj.CreateFile("dir", "file0.txt")

		// --- Then ---
		assert.Equal(t, filepath.Join(prj.root, "dir", "file0.txt"), have)
		assert.FileContain(t, "", have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("creates subdirectories", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		have := prj.CreateFile("dir", "sub", "file0.txt")

		// --- Then ---
		assert.Equal(t, filepath.Join(prj.root, "dir", "sub", "file0.txt"), have)
		assert.FileContain(t, "", have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.CreateFile("file0.txt") })
	})
}

func Test_Project_CreateFileWith(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		have := prj.CreateFileWith("abc", "file0.txt")

		// --- Then ---
		assert.Equal(t, filepath.Join(prj.root, "file0.txt"), have)
		assert.FileContain(t, "abc", have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("creates subdirectory", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		have := prj.CreateFileWith("abc", "dir", "file0.txt")

		// --- Then ---
		assert.Equal(t, filepath.Join(prj.root, "dir", "file0.txt"), have)
		assert.FileContain(t, "abc", have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("creates subdirectories", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		have := prj.CreateFileWith("abc", "dir", "sub", "file0.txt")

		// --- Then ---
		assert.Equal(t, filepath.Join(prj.root, "dir", "sub", "file0.txt"), have)
		assert.FileContain(t, "abc", have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.CreateFileWith("abc", "file0.txt") })
	})
}

func Test_Project_Rename(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		dir := t.TempDir()
		prj := New(tspy, dir)
		pth := filepath.Join(dir, "file0.txt")
		must.Nil(os.WriteFile(pth, []byte("abc"), 0644))

		// --- When ---
		have := prj.Rename("file0.txt", "file_abc.txt")

		// --- Then ---
		assert.Equal(t, filepath.Join(dir, "file_abc.txt"), have)
		assert.FileContain(t, "abc", have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("not existing file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFail()
		tspy.ExpectLogContain("no such file or directory")
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		assert.Panic(t, func() { prj.Rename("file0.txt", "file_abc.txt") })

		// --- Then ---
		prj.Close() // Must close to prevent error.
	})
}

func Test_Project_FileFrom(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		have := prj.FileFrom("testdata/file0.txt")

		// --- Then ---
		assert.Equal(t, "abc", oskit.ReadFileStr(t, have))
		assert.Equal(t, filepath.Join(prj.root, "file0.txt"), have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.FileFrom("testdata/file0.txt") })
	})
}

func Test_Project_FilesFrom(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		prj.FilesFrom("testdata")

		// --- Then ---
		assert.Equal(t, "abc", oskit.ReadFileStr(t, prj.root, "file0.txt"))
		assert.Equal(t, "def", oskit.ReadFileStr(t, prj.root, "file1.txt"))

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.FilesFrom("testdata") })
	})
}

func Test_Project_ProjectFrom(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		prj.ProjectFrom("testdata/dir")

		// --- Then ---
		assert.Equal(t, "ghi", oskit.ReadFileStr(t, prj.root, "file0_dir.txt"))
		assert.Equal(t, "jkl", oskit.ReadFileStr(t, prj.root, "file1_dir.txt"))
		assert.Equal(t, "dir", prj.dirName)
		assert.Equal(t, "example.com/comp/dir", prj.modName)

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.ProjectFrom("testdata/dir") })
	})
}

func Test_Project_ImpSpec(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, "/dir")
		prj.modName = "example.com/proj/my-project"

		// --- When ---
		have := prj.ImpSpec()

		// --- Then ---
		assert.Equal(t, "example.com/proj/my-project", have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, "/dir")
		prj.modName = "example.com/proj/my-project"
		prj.Close()

		// --- When ---
		have := prj.ImpSpec()

		// --- Then ---
		assert.Equal(t, "example.com/proj/my-project", have)
	})
}

func Test_Project_GoModInit(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		prj.GoModInit()

		// --- Then ---
		assert.FileExist(t, filepath.Join(prj.root, "go.mod"))

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.GoModInit() })
	})
}

func Test_Project_GoModTidy(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.GoModInit()
		content := "" +
			"package main\n" +
			"import \"github.com/ctx42/testing/pkg/notice\"\n" +
			"func main() { _ = notice.New(\"test\") }"
		prj.CreateFileWith(content, "main.go")
		prj.Exe(
			"go", "mod", "edit",
			"-require=github.com/ctx42/testing@v0.11.0",
		)

		// --- When ---
		prj.GoModTidy()

		// --- Then ---
		assert.FileExist(t, filepath.Join(prj.root, "go.mod"))
		assert.FileExist(t, filepath.Join(prj.root, "go.sum"))

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.GoModTidy() })
	})
}

func Test_Project_GitInitAddAll(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())
		oskit.CopyFile(t, prj.root, "testdata/file0.txt")

		// --- When ---
		cm := prj.GitInitAddAll()

		// --- Then ---
		assert.DirExist(t, filepath.Join(prj.root, ".git"))
		assert.FileExist(t, filepath.Join(prj.root, "file0.txt"))

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		want := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")
		assert.Equal(t, want, cm.Hash)

		want = "On branch master\nnothing to commit, working tree clean"
		have := exe.ExeStdout("git", "status")
		assert.Equal(t, want, have)

		want = fmt.Sprintf("%s (HEAD -> master) Initial commit.", cm.Hash)
		have = exe.ExeStdout("git", "log", "--pretty=format:%h%d %s")
		assert.Equal(t, want, have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("tag", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())
		oskit.CopyFile(t, prj.root, "testdata/file0.txt")

		// --- When ---
		cm := prj.GitInitAddAll("v1.1.1")

		// --- Then ---
		assert.DirExist(t, filepath.Join(prj.root, ".git"))
		assert.FileExist(t, filepath.Join(prj.root, "file0.txt"))

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		want := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")
		assert.Equal(t, want, cm.Hash)

		want = "On branch master\nnothing to commit, working tree clean"
		have := exe.ExeStdout("git", "status")
		assert.Equal(t, want, have)

		want = fmt.Sprintf(
			"%s (HEAD -> master, tag: v1.1.1) Initial commit.", cm.Hash)
		have = exe.ExeStdout("git", "log", "--pretty=format:%h%d %s")
		assert.Equal(t, want, have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.GitInitAddAll() })
	})
}

func Test_Project_GitCommit(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		oskit.CopyFile(t, prj.root, "testdata/file0.txt")
		exe.Exe("git", "init")

		// --- When ---
		cm := prj.GitCommit("")

		// --- Then ---
		want := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")
		assert.Equal(t, want, cm.Hash)

		want = "On branch master\nnothing to commit, working tree clean"
		have := exe.ExeStdout("git", "status")
		assert.Equal(t, want, have)

		want = fmt.Sprintf("%s (HEAD -> master) commit 0", cm.Hash)
		have = exe.ExeStdout("git", "log", "--pretty=format:%h%d %s")
		assert.Equal(t, want, have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("tag", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		oskit.CopyFile(t, prj.root, "testdata/file0.txt")
		exe.Exe("git", "init")

		// --- When ---
		cm := prj.GitCommit("v1.1.1")

		// --- Then ---
		want := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")
		assert.Equal(t, want, cm.Hash)

		want = "On branch master\nnothing to commit, working tree clean"
		have := exe.ExeStdout("git", "status")
		assert.Equal(t, want, have)

		want = fmt.Sprintf("%s (HEAD -> master, tag: v1.1.1) commit 0", cm.Hash)
		have = exe.ExeStdout("git", "log", "--pretty=format:%h%d %s")
		assert.Equal(t, want, have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("tag and custom message", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		oskit.CopyFile(t, prj.root, "testdata/file0.txt")
		exe.Exe("git", "init")

		// --- When ---
		cm := prj.GitCommit("v1.1.1", "my message")

		// --- Then ---
		want := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")
		assert.Equal(t, want, cm.Hash)

		want = "On branch master\nnothing to commit, working tree clean"
		have := exe.ExeStdout("git", "status")
		assert.Equal(t, want, have)

		want = fmt.Sprintf("%s (HEAD -> master, tag: v1.1.1) my message", cm.Hash)
		have = exe.ExeStdout("git", "log", "--pretty=format:%h%d %s")
		assert.Equal(t, want, have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("on not initialized", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogContain("fatal: not a git repository")
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		cm := prj.GitCommit("v1.1.1")

		// --- Then ---
		assert.Nil(t, cm)
	})

	t.Run("sets git identity", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())
		oskit.CopyFile(t, prj.root, "testdata/file0.txt")

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		exe.Exe("git", "init")

		// --- When ---
		prj.GitCommit("")
		oskit.CopyFile(t, prj.root, "testdata/file1.txt")
		prj.GitCommit("") // second commit must not reset the flag

		// --- Then ---
		assert.True(t, prj.gitID)

		have := exe.ExeStdout("git", "config", "user.email")
		assert.Equal(t, "test@example.com", have)

		have = exe.ExeStdout("git", "config", "user.name")
		assert.Equal(t, "Test User", have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.GitCommit("v1.1.1") })
	})
}

func Test_Project_GitSetRemote(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		oskit.CopyFile(t, prj.root, "testdata/file0.txt")
		exe.Exe("git", "init")

		// --- When ---
		prj.GitSetRemote()

		// --- Then ---
		want := "" +
			"origin\tgit@example.com:comp/project.git (fetch)\n" +
			"origin\tgit@example.com:comp/project.git (push)"
		have := exe.ExeStdout("git", "remote", "-v")
		assert.Equal(t, want, have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("with custom repo", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		oskit.CopyFile(t, prj.root, "testdata/file0.txt")
		exe.Exe("git", "init")

		// --- When ---
		prj.GitSetRemote(GitSSHOrigin)

		// --- Then ---
		want := "" +
			"origin\tssh://git@example.com:comp/project.git (fetch)\n" +
			"origin\tssh://git@example.com:comp/project.git (push)"
		have := exe.ExeStdout("git", "remote", "-v")
		assert.Equal(t, want, have)

		prj.Close() // Must close to prevent error.
	})

	t.Run("on not initialized", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogContain("fatal: not a git repository")
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		prj.GitSetRemote()

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.GitSetRemote() })
	})
}

func Test_Project_GitHash(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		oskit.CopyFile(t, prj.root, "testdata/file0.txt")
		exe.Exe("git", "init")
		exe.Exe("git", "add", "-A")
		exe.Exe("git", "commit", "-m", "Initial commit.")
		want := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")

		// --- When ---
		have := prj.GitHash()

		// --- Then ---
		assert.Equal(t, want, have)
	})

	t.Run("on not initialized", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogContain("fatal: not a git repository")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		have := prj.GitHash()

		// --- Then ---
		assert.Empty(t, have)
	})

	t.Run("on open", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		assert.Panic(t, func() { prj.GitHash() })
	})
}

func Test_Project_GitLog(t *testing.T) {
	t.Run("success on open", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		exe.Exe("git", "init")

		oskit.WriteStr(t, "0", prj.root, "file0.txt")
		exe.Exe("git", "add", "-A")
		exe.Exe("git", "commit", "-m", "commit 0.")
		hash0 := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")

		time.Sleep(1100 * time.Millisecond)
		oskit.WriteStr(t, "1", prj.root, "file1.txt")
		exe.Exe("git", "add", "-A")
		exe.Exe("git", "commit", "-m", "commit 1.")
		hash1 := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")

		// --- When ---
		cms := prj.GitCommitLog()

		// --- Then ---
		assert.Len(t, 2, cms)

		cm := cms[0]
		wantDate := exe.ExeStdout("git", "log", "-1", "--pretty=format:%at", hash1)
		assert.Equal(t, hash1, cm.Hash)
		assert.Time(t, wantDate, cm.Date, check.WithTimeFormat("str-ts"))
		assert.Equal(t, "commit 1.", cm.Summary)

		cm = cms[1]
		wantDate = exe.ExeStdout("git", "log", "-1", "--pretty=format:%at", hash0)
		assert.Equal(t, hash0, cm.Hash)
		assert.Time(t, wantDate, cm.Date, check.WithTimeFormat("str-ts"))
		assert.Equal(t, "commit 0.", cm.Summary)

		prj.Close() // Must close to prevent error.
	})

	t.Run("success on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		exe.Exe("git", "init")

		oskit.WriteStr(t, "0", prj.root, "file0.txt")
		exe.Exe("git", "add", "-A")
		exe.Exe("git", "commit", "-m", "commit 0.")
		hash0 := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")

		time.Sleep(1100 * time.Millisecond)
		oskit.WriteStr(t, "1", prj.root, "file1.txt")
		exe.Exe("git", "add", "-A")
		exe.Exe("git", "commit", "-m", "commit 1.")
		hash1 := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")

		prj.Close()

		// --- When ---
		cms := prj.GitCommitLog()

		// --- Then ---
		assert.Len(t, 2, cms)

		cm := cms[0]
		wantDate := exe.ExeStdout("git", "log", "-1", "--pretty=format:%at", hash1)
		assert.Equal(t, hash1, cm.Hash)
		assert.Equal(t, wantDate, strconv.FormatInt(cm.Date.Unix(), 10))
		assert.Equal(t, "commit 1.", cm.Summary)
		assert.Equal(t, "commit 1.", cm.Summary)

		cm = cms[1]
		wantDate = exe.ExeStdout("git", "log", "-1", "--pretty=format:%at", hash0)
		assert.Equal(t, hash0, cm.Hash)
		assert.Equal(t, wantDate, strconv.FormatInt(cm.Date.Unix(), 10))
		assert.Equal(t, "commit 0.", cm.Summary)
	})

	t.Run("commit with tag", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())

		exe := exekit.New(t, exekit.WithWd(prj.root), exekit.WithTrim)
		exe.Exe("git", "init")

		oskit.WriteStr(t, "0", prj.root, "file0.txt")
		exe.Exe("git", "add", "-A")
		exe.Exe("git", "commit", "-m", "commit 0.")
		hash0 := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")

		time.Sleep(1100 * time.Millisecond)
		oskit.WriteStr(t, "1", prj.root, "file1.txt")
		exe.Exe("git", "add", "-A")
		exe.Exe("git", "commit", "-m", "commit 1.")
		exe.Exe("git", "tag", "-a", "-m", "test tag v1.2.3", "v1.2.3")
		hash1 := exe.ExeStdout("git", "rev-parse", "--short", "HEAD")

		prj.Close()

		// --- When ---
		cms := prj.GitCommitLog()

		// --- Then ---
		assert.Len(t, 2, cms)

		cm := cms[0]
		wantDate := exe.ExeStdout("git", "log", "-1", "--pretty=format:%at", hash1)
		assert.Equal(t, hash1, cm.Hash)
		assert.Equal(t, "v1.2.3", cm.Rev)
		assert.Equal(t, wantDate, strconv.FormatInt(cm.Date.Unix(), 10))
		assert.Equal(t, "commit 1.", cm.Summary)

		cm = cms[1]
		wantDate = exe.ExeStdout("git", "log", "-1", "--pretty=format:%at", hash0)
		assert.Equal(t, hash0, cm.Hash)
		assert.Equal(t, "", cm.Rev)
		assert.Equal(t, wantDate, strconv.FormatInt(cm.Date.Unix(), 10))
		assert.Equal(t, "commit 0.", cm.Summary)
	})

	t.Run("on not initialized", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogContain("fatal: not a git repository")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		have := prj.GitCommitLog()

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_Project_Compile(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectTempDir(1)
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		exe := exekit.New(t, exekit.WithWd(prj.root))
		oskit.CopyDir(t, prj.root, "testdata/compile/simple")
		exe.Exe("go", "mod", "init", "project")

		// --- When ---
		have := prj.Compile()

		// --- Then ---
		exe = exekit.New(t, exekit.WithWd(prj.root))
		assert.Equal(t, "hello\n", exe.ExeStdout(have))
		assert.NotEqual(t, prj.root, filepath.Dir(have))
	})

	t.Run("compilation failed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectTempDir(1)
		tspy.ExpectFatal()
		tspy.ExpectLogContain("main redeclared in this block")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		exe := exekit.New(t, exekit.WithWd(prj.root))
		oskit.CopyDir(t, prj.root, "testdata/compile/error")
		oldpath := filepath.Join(prj.root, "main")
		newpath := filepath.Join(prj.root, "main.go")
		assert.NoError(t, os.Rename(oldpath, newpath))
		exe.Exe("go", "mod", "init", "project")

		// --- When ---
		assert.Panic(t, func() { prj.Compile() })
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		prj := New(tspy, t.TempDir())

		// --- When ---
		assert.Panic(t, func() { prj.Compile() })
	})
}

func Test_Project_WithConfig(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		have := prj.WithConfig()

		// --- Then ---
		assert.Equal(t, "project", filepath.Base(prj.Root()))
		assert.DirExist(t, filepath.Join(prj.Root(), "configs"))
		pth := filepath.Join(prj.Root(), "configs", "project.conf")
		assert.Equal(t, pth, have)
		assert.Equal(t, "", oskit.ReadFileStr(t, pth))

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		prj := New(tspy, t.TempDir())
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.WithConfig() })
	})
}

func Test_Project_CfgAdd(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		have := prj.CfgAdd("key", "val")

		// --- Then ---
		pth := filepath.Join(prj.Root(), "configs", "project.conf")
		assert.Equal(t, pth, have)
		assert.Equal(t, "key=val\n", oskit.ReadFileStr(t, pth))

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.CfgAdd("key", "val") })
	})
}

func Test_Project_CfgDkrRepoDef(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		prj.CfgDkrRepoDef()

		// --- Then ---
		pth := filepath.Join(prj.Root(), "configs", "project.conf")
		want := "" +
			"VRIT_DKR_REG_HOST=my.nexus.dev\n" +
			"VRIT_DKR_REPO=my.nexus.dev/repo\n"
		assert.Equal(t, want, oskit.ReadFileStr(t, pth))

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.CfgDkrRepoDef() })
	})
}

func Test_Project_CfgDkrRepo(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		prj.CfgDkrRepo("host", "repo")

		// --- Then ---
		pth := filepath.Join(prj.Root(), "configs", "project.conf")
		want := "" +
			"VRIT_DKR_REG_HOST=host\n" +
			"VRIT_DKR_REPO=host/repo\n"
		assert.Equal(t, want, oskit.ReadFileStr(t, pth))

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.CfgDkrRepo("host", "repo") })
	})

	t.Run("set multiple times", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		prj.CfgDkrRepo("host0", "repo0")
		prj.CfgDkrRepo("host1", "repo1")

		// --- Then ---
		pth := filepath.Join(prj.Root(), "configs", "project.conf")
		want := "" +
			"VRIT_DKR_REG_HOST=host0\n" +
			"VRIT_DKR_REPO=host0/repo0\n" +
			"VRIT_DKR_REG_HOST=host1\n" +
			"VRIT_DKR_REPO=host1/repo1\n"
		assert.Equal(t, want, oskit.ReadFileStr(t, pth))

		prj.Close() // Must close to prevent error.
	})
}

func Test_Project_DkrRepo(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.Close()

		// --- When ---
		have := prj.DkrRepo()

		// --- Then ---
		assert.Equal(t, "", have)
	})

	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		assert.Panic(t, func() { prj.DkrRepo() })
	})

	t.Run("returns set value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.CfgDkrRepo("host", "repo")
		prj.Close()

		// --- When ---
		have := prj.DkrRepo()

		// --- Then ---
		assert.Equal(t, "host/repo", have)
	})

	t.Run("returns latest value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.CfgDkrRepo("host0", "repo0")
		prj.CfgDkrRepo("host1", "repo1")
		prj.Close()

		// --- When ---
		have := prj.DkrRepo()

		// --- Then ---
		assert.Equal(t, "host1/repo1", have)
	})
}

func Test_Project_CfgDkrTargets(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		prj.CfgDkrTargets("first,second")

		// --- Then ---
		pth := filepath.Join(prj.Root(), "configs", "project.conf")
		want := "VRIT_DKF_TARGETS=first,second\n"
		assert.Equal(t, want, oskit.ReadFileStr(t, pth))

		prj.Close() // Must close to prevent error.
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.CfgDkrTargets("first,second") })
	})

	t.Run("set multiple times", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		prj.CfgDkrTargets("first,second")
		prj.CfgDkrTargets("second,third")

		// --- Then ---
		pth := filepath.Join(prj.Root(), "configs", "project.conf")
		want := "" +
			"VRIT_DKF_TARGETS=first,second\n" +
			"VRIT_DKF_TARGETS=second,third\n"
		assert.Equal(t, want, oskit.ReadFileStr(t, pth))

		prj.Close() // Must close to prevent error.
	})
}

func Test_Project_WithDockerfile(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		have := prj.WithDockerfile()

		// --- Then ---
		pth := filepath.Join(prj.Root(), "Dockerfile")
		assert.Equal(t, pth, have)
		assert.Equal(t, string(dockerfile), oskit.ReadFileStr(t, pth))
		assert.NotEmpty(t, prj.imgName)
		assert.NotEmpty(t, prj.imgTag)

		prj.Close() // Must close to prevent error.
		tspy.Finish()
		wantImgRem := []string{
			prj.dkrImgRef(),
			prj.dkrImgRefLatest(),
		}
		assert.Equal(t, wantImgRem, prj.removed)
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.WithDockerfile() })
	})
}

func Test_Project_WithDockerfileNEP(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		have := prj.WithDockerfileNEP()

		// --- Then ---
		pth := filepath.Join(prj.Root(), "Dockerfile")
		assert.Equal(t, pth, have)
		assert.Equal(t, string(dockerfileNEP), oskit.ReadFileStr(t, pth))
		assert.NotEmpty(t, prj.imgName)
		assert.NotEmpty(t, prj.imgTag)

		prj.Close() // Must close to prevent error.
		tspy.Finish()
		wantImgRem := []string{
			prj.dkrImgRef(),
			prj.dkrImgRefLatest(),
		}
		assert.Equal(t, wantImgRem, prj.removed)
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.WithDockerfileNEP() })
	})
}

func Test_Project_DkrImgNameTagSet(t *testing.T) {
	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		prj.DkrImgNameTagSet("name", "tag")

		// --- Then ---
		assert.Equal(t, "name", prj.imgName)
		assert.Equal(t, "tag", prj.imgTag)

		prj.Close() // Must close to prevent error.
		tspy.Finish()
		wantImgRem := []string{"name:tag", "name:latest"}
		assert.Equal(t, wantImgRem, prj.removed)
	})

	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be open")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.Close()

		// --- When ---
		assert.Panic(t, func() { prj.DkrImgNameTagSet("name", "tag") })
	})
}

func Test_Project_DkrImgName(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.imgName = "name"
		prj.Close()

		// --- When ---
		have := prj.DkrImgName()

		// --- Then ---
		assert.Equal(t, "name", have)
	})

	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		assert.Panic(t, func() { prj.DkrImgName() })
	})
}

func Test_Project_DkrImgTag(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.imgTag = "tag"
		prj.Close()

		// --- When ---
		have := prj.DkrImgTag()

		// --- Then ---
		assert.Equal(t, "tag", have)
	})

	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		assert.Panic(t, func() { prj.DkrImgTag() })
	})
}

func Test_Project_DkrImgRef(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.imgName = "name"
		prj.imgTag = "tag"
		prj.Close()

		// --- When ---
		have := prj.DkrImgRef()

		// --- Then ---
		assert.Equal(t, "name:tag", have)
	})

	t.Run("with repo", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.dkrRepo = "repo"
		prj.imgName = "name"
		prj.imgTag = "tag"
		prj.Close()

		// --- When ---
		have := prj.DkrImgRef()

		// --- Then ---
		assert.Equal(t, "repo/name:tag", have)
	})

	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		assert.Panic(t, func() { prj.DkrImgRef() })
	})
}

func Test_Project_DkrImgRefLatest(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.imgName = "name"
		prj.imgTag = "tag"
		prj.Close()

		// --- When ---
		have := prj.DkrImgRefLatest()

		// --- Then ---
		assert.Equal(t, "name:latest", have)
	})

	t.Run("with repo", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.dkrRepo = "repo"
		prj.imgName = "name"
		prj.imgTag = "tag"
		prj.Close()

		// --- When ---
		have := prj.DkrImgRefLatest()

		// --- Then ---
		assert.Equal(t, "repo/name:latest", have)
	})

	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		assert.Panic(t, func() { prj.DkrImgRefLatest() })
	})
}

func Test_Project_DkrTgtName(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.imgName = "name"
		prj.imgTag = "tag"
		prj.Close()

		// --- When ---
		have := prj.DkrTgtName("target")

		// --- Then ---
		assert.Equal(t, "name-target", have)

		tspy.Finish()
		assert.Empty(t, prj.removed)
	})

	t.Run("with repo", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.dkrRepo = "repo"
		prj.imgName = "name"
		prj.imgTag = "tag"
		prj.Close()

		// --- When ---
		have := prj.DkrTgtName("target")

		// --- Then ---
		assert.Equal(t, "repo/name-target", have)
	})

	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		assert.Panic(t, func() { prj.DkrTgtName("target") })
	})
}

func Test_Project_DkrTgtRef(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.imgName = "name"
		prj.imgTag = "tag"
		prj.Close()

		// --- When ---
		have := prj.DkrTgtRef("target")

		// --- Then ---
		assert.Equal(t, "name-target:tag", have)

		tspy.Finish()
		wantImgRem := []string{"name-target:tag", "name-target:latest"}
		assert.Equal(t, wantImgRem, prj.removed)
	})

	t.Run("with repo", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.dkrRepo = "repo"
		prj.imgName = "name"
		prj.imgTag = "tag"
		prj.Close()

		// --- When ---
		have := prj.DkrTgtRef("target")

		// --- Then ---
		assert.Equal(t, "repo/name-target:tag", have)
	})

	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		assert.Panic(t, func() { prj.DkrTgtRef("target") })
	})
}

func Test_Project_DkrTgtRefLatest(t *testing.T) {
	t.Run("on closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.imgName = "name"
		prj.imgTag = "tag"
		prj.Close()

		// --- When ---
		have := prj.DkrTgtRefLatest("target")

		// --- Then ---
		tspy.Finish()
		assert.Equal(t, "name-target:latest", have)
		wantImgRem := []string{"name-target:tag", "name-target:latest"}
		assert.Equal(t, wantImgRem, prj.removed)
	})

	t.Run("with repo", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)
		prj.dkrRepo = "repo"
		prj.imgName = "name"
		prj.imgTag = "tag"
		prj.Close()

		// --- When ---
		have := prj.DkrTgtRefLatest("target")

		// --- Then ---
		assert.Equal(t, "repo/name-target:latest", have)
	})

	t.Run("on opened", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectFatal()
		tspy.ExpectLogEqual("expected test project instance to be closed")
		tspy.Close()

		root := oskit.MkdirAll(t, t.TempDir(), "project")
		prj := New(tspy, root)

		// --- When ---
		assert.Panic(t, func() { prj.DkrTgtRefLatest("target") })
	})
}

func Test_Project_Close(t *testing.T) {
	t.Run("sets as closed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		prj := New(tspy, "/dir")

		// --- When ---
		prj.Close()

		// --- Then ---
		assert.True(t, prj.closed)
	})

	t.Run("error when closed twice", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogEqual("instance already closed")
		tspy.Close()

		prj := New(tspy, "/dir")
		prj.Close()

		// --- When ---
		prj.Close()

		// --- Then ---
		assert.True(t, prj.closed)
	})
}

func Test_NewGitCommit(t *testing.T) {
	tt := []struct {
		testN string

		lin        string
		expHash    string
		expTag     string
		expDate    time.Time
		expSummary string
	}{
		{
			"simple",
			"16c806f () 946782245 commit 1",
			"16c806f",
			"",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			"commit 1",
		},
		{
			"with tag",
			"6907f52 (HEAD -> master, tag: v0.5.11, " +
				"origin/master) 946782245 Bump version to 0.5.11.",
			"6907f52",
			"v0.5.11",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			"Bump version to 0.5.11.",
		},
		{
			"with tag only",
			"2728491 (HEAD -> master, tag: v1.2.3) 946782245 commit 2.",
			"2728491",
			"v1.2.3",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			"commit 2.",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			gc, err := NewGitCommit(tc.lin)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.expHash, gc.Hash)
			assert.Equal(t, tc.expDate, gc.Date)
			assert.Equal(t, tc.expSummary, gc.Summary)
		})
	}
}

func Test_GitCommits_Find(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		// --- Given ---
		cm0 := GitCommit{Hash: "hash0", Summary: "summary 0"}
		cm1 := GitCommit{Hash: "hash1", Summary: "summary 1"}
		cm2 := GitCommit{Hash: "hash2", Summary: "summary 2"}

		gcs := GitCommits([]GitCommit{cm0, cm1, cm2})

		// --- When ---
		have := gcs.Find("hash1")

		// --- Then ---
		assert.Equal(t, cm1, *have)
	})

	t.Run("not existing", func(t *testing.T) {
		// --- Given ---
		cm0 := GitCommit{Hash: "hash0", Summary: "summary 0"}
		cm1 := GitCommit{Hash: "hash1", Summary: "summary 1"}

		gcs := GitCommits([]GitCommit{cm0, cm1})

		// --- When ---
		have := gcs.Find("hash3")

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("find on empty", func(t *testing.T) {
		// --- Given ---
		gcs := GitCommits([]GitCommit{})

		// --- When ---
		have := gcs.Find("hash3")

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("find on nil", func(t *testing.T) {
		// --- Given ---
		var gcs GitCommits

		// --- When ---
		have := gcs.Find("hash3")

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_GitCommits_Latest(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		// --- Given ---
		cm0 := GitCommit{Hash: "hash0", Summary: "summary 0"}
		cm1 := GitCommit{Hash: "hash1", Summary: "summary 1"}
		cm2 := GitCommit{Hash: "hash2", Summary: "summary 2"}

		gcs := GitCommits([]GitCommit{cm0, cm1, cm2})

		// --- When ---
		have := gcs.Latest()

		// --- Then ---
		assert.Equal(t, cm0, *have)
	})

	t.Run("empty collection", func(t *testing.T) {
		// --- Given ---
		gcs := GitCommits([]GitCommit{})

		// --- When ---
		have := gcs.Latest()

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("nil collection", func(t *testing.T) {
		// --- Given ---
		var gcs GitCommits

		// --- When ---
		have := gcs.Latest()

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_GitCommits_First(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		// --- Given ---
		cm0 := GitCommit{Hash: "hash0", Summary: "summary 0"}
		cm1 := GitCommit{Hash: "hash1", Summary: "summary 1"}
		cm2 := GitCommit{Hash: "hash2", Summary: "summary 2"}

		gcs := GitCommits([]GitCommit{cm0, cm1, cm2})

		// --- When ---
		have := gcs.First()

		// --- Then ---
		assert.Equal(t, cm2, *have)
	})

	t.Run("empty collection", func(t *testing.T) {
		// --- Given ---
		gcs := GitCommits([]GitCommit{})

		// --- When ---
		have := gcs.First()

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("nil collection", func(t *testing.T) {
		// --- Given ---
		var gcs GitCommits

		// --- When ---
		have := gcs.First()

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_GitCommits_N(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		// --- Given ---
		cm0 := GitCommit{Hash: "hash0", Summary: "summary 0"}
		cm1 := GitCommit{Hash: "hash1", Summary: "summary 1"}
		cm2 := GitCommit{Hash: "hash2", Summary: "summary 2"}

		gcs := GitCommits([]GitCommit{cm0, cm1, cm2})

		// --- When ---
		have := gcs.N(1)

		// --- Then ---
		assert.Equal(t, cm1, *have)
	})

	t.Run("not existing index", func(t *testing.T) {
		// --- Given ---
		cm0 := GitCommit{Hash: "hash0", Summary: "summary 0"}
		cm1 := GitCommit{Hash: "hash1", Summary: "summary 1"}
		cm2 := GitCommit{Hash: "hash2", Summary: "summary 2"}

		gcs := GitCommits([]GitCommit{cm0, cm1, cm2})

		// --- When ---
		have := gcs.N(10)

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("empty collection", func(t *testing.T) {
		// --- Given ---
		gcs := GitCommits([]GitCommit{})

		// --- When ---
		have := gcs.N(1)

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("nil collection", func(t *testing.T) {
		// --- Given ---
		var gcs GitCommits

		// --- When ---
		have := gcs.N(1)

		// --- Then ---
		assert.Nil(t, have)
	})
}
