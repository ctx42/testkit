// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/xdef/pkg/xdef"

	"github.com/ctx42/testkit/pkg/exekit"
	"github.com/ctx42/testkit/pkg/netkit"
	"github.com/ctx42/testkit/pkg/randkit"
)

func Test_NewT(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		dkr := NewT(tspy)

		// --- Then ---
		assert.Equal(t, tspy, dkr.t)
		assert.Equal(t, os.Environ(), dkr.dkr.env)
	})

	t.Run("WithEnv option", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()
		env := []string{"KEY=VAL"}

		// --- When ---
		dkr := NewT(tspy, WithEnv(env))

		// --- Then ---
		assert.Equal(t, env, dkr.dkr.env)
	})
}

func Test_DockerT_ImgPull(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		dkr.ImgPull(TestImageBaseRef)
		exekit.New(t).Exe("docker", "image", "inspect", TestImageBaseRef)
	})

	t.Run("error - invalid ref", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[pulling image] docker command error:\n" +
			"   cmd: docker pull \n" +
			"   err: exit status 1\n" +
			"  eout: invalid reference format"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		dkr.ImgPull("")
	})
}

func Test_DockerT_Build(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectNames(1)
		tspy.Close()

		dkr := NewT(tspy)
		bldOpt := WithBuildPth("testdata/simple/Dockerfile")
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)

		// --- When ---
		ref, iid := dkr.Build(bldOpt, argOpt)

		// --- Then ---
		exekit.New(t).Exe("docker", "history", ref)
		assert.NotEmpty(t, iid)

		hLabels := must.Value(getLabels(t.Context(), os.Environ(), ref))
		wLabels := map[string]string{xdef.LabImgAuthors: t.Name()}
		assert.MapSubset(t, wLabels, hLabels)

		hEnv := must.Value(getEnvs(t.Context(), os.Environ(), ref))
		wEnv := map[string]string{xdef.EnvImgAuthors: t.Name()}
		assert.MapSubset(t, wEnv, hEnv)

		tspy.Finish()
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "history", ref)
	})

	t.Run("error", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectNames(1)
		tspy.ExpectError()
		wMsg := "WithBuildPth and WithBuildRdr are mutually exclusive"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		dkr := NewT(tspy)
		bldPthOpt := WithBuildPth("testdata/invalid/Dockerfile")
		bldRdrOpt := WithBuildRdr(strings.NewReader("abc"))

		// --- When ---
		ref, iid := dkr.Build(bldPthOpt, bldRdrOpt)

		// --- Then ---
		assert.Empty(t, ref)
		assert.Empty(t, iid)
	})
}

func Test_DockerT_BuildTestImg(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectNames(1)
		tspy.ExpectCleanups(1)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		ref, iid := dkr.BuildTestImg()

		// --- Then ---
		exekit.New(t).Exe("docker", "history", ref)
		assert.NotEmpty(t, iid)

		hLabels := must.Value(getLabels(t.Context(), os.Environ(), ref))
		wLabels := map[string]string{xdef.LabImgAuthors: t.Name()}
		assert.MapSubset(t, wLabels, hLabels)

		hEnv := must.Value(getEnvs(t.Context(), os.Environ(), ref))
		wEnv := map[string]string{xdef.EnvImgAuthors: t.Name()}
		assert.MapSubset(t, wEnv, hEnv)

		assert.Equal(t, t.Name(), NewT(t).Label(iid, xdef.LabImgAuthors))
		tspy.Finish()
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "history", ref)
	})
}

func Test_DockerT_ImgLs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		ims := dkr.ImgLs()

		// --- Then ---
		img := ims.FindByRef(TestImg0.ref)
		assert.NotNil(t, img)
		assert.Equal(t, TestImg0.iid, img.ID)
		assert.Equal(t, TestImg0.name, img.Repository)
		assert.Equal(t, TestImg0.tag, img.Tag)
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)

		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[getting images] docker command error:")
		tspy.Close()

		dkr := NewT(tspy, WithEnv(env))

		// --- When ---
		ims := dkr.ImgLs()

		// --- Then ---
		assert.Nil(t, ims)
	})
}

func Test_DockerT_Labels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		env := append(os.Environ(), xdef.EnvImgCreated+"=2000-01-02T03:04:05Z")
		env = append(env, xdef.EnvImgRefName+"=ccid")
		dkr := NewT(tspy, WithEnv(env))

		// --- When ---
		have := dkr.Labels(TestImg0.ref)

		// --- Then ---
		want := map[string]string{
			xdef.LabImgCreated:  "2000-01-02T03:04:05Z",
			xdef.LabImgBaseName: TestImageBaseRef,
			xdef.LabImgTitle:    "Image0",
			labTestEmpty:        "",
		}
		assert.Equal(t, want, have)
	})

	t.Run("error - non-existent reference", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[getting labels] docker command error:")
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.Labels(RandRef())

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_DockerT_Label(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.Label(TestImg0.ref, xdef.LabImgTitle)

		// --- Then ---
		assert.Equal(t, "Image0", have)
	})

	t.Run("error - non-existent label", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[getting label] expected label to exist:")
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.Label(TestImg0.ref, "not.existing.label")

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_DockerT_Envs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.Envs(TestImg0.ref)

		// --- Then ---
		want := map[string]string{
			xdef.EnvImgCreated: "2000-01-02T03:04:05Z",
			xdef.EnvImgTitle:   "Image0",
			envTestEmpty:       "",
		}
		assert.MapSubset(t, want, have)
	})

	t.Run("error - non-existent reference", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "[getting environment variables] docker command error:"
		tspy.ExpectLogContain(wMsg)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.Envs(RandRef())

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_DockerT_Env(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.Env(TestImg0.ref, xdef.EnvImgTitle)

		// --- Then ---
		assert.Equal(t, "Image0", have)
	})

	t.Run("error - non-existent environment variable", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "[getting environment variable] expected variable to exist:"
		tspy.ExpectLogContain(wMsg)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.Env(TestImg0.ref, "NOT_EXISTING")

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_DockerT_ImgRm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		img := must.Value(createMinImage(t.Name()))
		dkr := NewT(tspy)

		// --- When ---
		have := dkr.ImgRm(img.ref)

		// --- Then ---
		assert.True(t, have)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "history", img.ref)
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[removing image] docker command error:")
		tspy.Close()

		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := NewT(tspy, WithEnv(env))

		// --- When ---
		have := dkr.ImgRm(RandRef())

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("WithImgRmIgnoreErrors uses log not error", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectLogContain("[removing image] docker command error:")
		tspy.Close()

		img := must.Value(createMinImage(t.Name()))
		args := []string{"run", "-d", "--rm", img.iid, "sleep", "3"}
		cid := exekit.New(t).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
			exekit.New(t).Exe("docker", "rmi", "-f", img.iid)
		})

		dkr := NewT(tspy)
		tryOpt := WithImgRmTries(1)
		slpOpt := WithImgRmSleep(100 * time.Millisecond)
		ignOpt := WithImgRmIgnoreErrors()

		// --- When ---
		have := dkr.ImgRm(img.iid, tryOpt, slpOpt, ignOpt)

		// --- Then ---
		assert.False(t, have)
		exekit.New(t).Exe("docker", "history", img.iid)
	})
}

func Test_DockerT_CtrRun(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.CtrRun(TestImg0.iid)

		// --- Then ---
		assert.Equal(t, "hello", have)
	})

	t.Run("get container ID via a channel", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectTempDir(1)
		tspy.Close()

		dkr := NewT(tspy)
		noRmOpt := WithCtrRunNoRemove()
		kv := randkit.Str()
		labOpt := WithCtrRunLabel(kv, kv)
		cidCh, cidOpt := WithCtrRunCID()
		argOpt := WithCtrRunArgs("echo", "abc")

		// --- When ---
		have := dkr.CtrRun(TestImg0.iid, cidOpt, noRmOpt, labOpt, argOpt)

		// --- Then ---
		assert.Equal(t, "abc", have)

		cid, ok := <-cidCh
		assert.True(t, ok)
		format := fmt.Sprintf(`{{ index .Config.Labels %q}}`, kv)
		args := []string{"inspect", cid, "--format", format}
		sout := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		assert.Equal(t, kv, sout)
		exekit.New(t).Exe("docker", "rm", cid)
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[running image] docker command error:")
		tspy.Close()

		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := NewT(tspy, WithEnv(env))

		// --- When ---
		ims := dkr.CtrRun(TestImg0.iid)

		// --- Then ---
		assert.Empty(t, ims)
	})
}

func Test_DockerT_CtrRm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		args := []string{"run", "-d", TestImg0.iid}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		tspy := tester.New(t)
		tspy.ExpectLogEqual("CtrRm: successfully removed container: %s", cid)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.CtrRm(cid)

		// --- Then ---
		assert.True(t, have)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "inspect", cid)
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[removing container] docker command error:")
		tspy.Close()

		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := NewT(tspy, WithEnv(env))

		// --- When ---
		have := dkr.CtrRm(TestImg0.iid)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_DockerT_CtrKill(t *testing.T) {
	t.Run("kill running container", func(t *testing.T) {
		// --- Given ---
		args := []string{"run", "-d", TestImg0.iid, "sleep", "5"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		tspy := tester.New(t)
		tspy.ExpectLogEqual("CtrKill: successfully killed container: %s", cid)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.CtrKill(cid)

		// --- Then ---
		assert.True(t, have)
		exekit.New(t).Exe("docker", "inspect", cid)
	})

	t.Run("error - kill unknown container", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[killing container] docker command error:")
		tspy.Close()

		dkr := NewT(tspy)
		cid := "000000000000"

		// --- When ---
		have := dkr.CtrKill(cid)

		// --- Then ---
		assert.False(t, have)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "inspect", cid)
	})
}

func Test_DockerT_CtrExec(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		args := []string{"run", "-d", TestImg0.iid, "sleep", "2"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		tspy := tester.New(t)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.CtrExec(cid, "ls", "/")

		// --- Then ---
		want := "" +
			"bin\n" +
			"dev\n" +
			"entrypoint.sh\n" +
			"etc\n" +
			"file0.txt\n" +
			"home\n" +
			"proc\n" +
			"root\n" +
			"sys\n" +
			"tmp\n" +
			"usr\n" +
			"var"
		assert.Equal(t, want, have)
	})

	t.Run("error - unknown command", func(t *testing.T) {
		// --- Given ---
		args := []string{"run", "-d", TestImg0.iid, "sleep", "2"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[container exec] docker command error:")
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.CtrExec(cid, "not-existing")

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_DockerT_CtrPs(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		// --- Given ---
		ctr0args := []string{"run", "-d", "--label", "ctr=ctr0", TestImg0.iid}
		ctr1args := []string{"run", "-d", "--label", "ctr=ctr1", TestImg0.iid}
		cid0 := exekit.New(t, exekit.WithTrim).ExeStdout("docker", ctr0args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid0)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid0)
		})
		cid1 := exekit.New(t, exekit.WithTrim).ExeStdout("docker", ctr1args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid1)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid1)
		})

		tspy := tester.New(t)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.CtrPs()

		// --- Then ---
		ctr0 := must.Value(have.FindByID(cid0))
		assert.NotNil(t, ctr0)
		wLabels := map[string]string{"ctr": "ctr0"}
		assert.MapSubset(t, wLabels, ctr0.Labels)
		assert.Equal(t, "exited", ctr0.State.Status)
		assert.HasKey(t, "bridge", ctr0.Networks)

		ctr1 := must.Value(have.FindByID(cid1))
		assert.NotNil(t, ctr1)
		wLabels = map[string]string{"ctr": "ctr1"}
		assert.MapSubset(t, wLabels, ctr1.Labels)
		assert.Equal(t, "exited", ctr1.State.Status)
		assert.HasKey(t, "bridge", ctr1.Networks)
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[listing containers] docker command error:")
		tspy.Close()

		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := NewT(tspy, WithEnv(env))

		// --- When ---
		_ = dkr.CtrPs()
	})
}

func Test_DockerT_CtrFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		args := []string{"run", "-d", TestImg0.ref, "sleep", "2"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})

		tspy := tester.New(t)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.CtrFile(cid, "/file0.txt")

		// --- Then ---
		assert.Equal(t, "file0\n", have)
	})

	t.Run("error - asking for a non-existent file", func(t *testing.T) {
		// --- Given ---
		args := []string{"run", "-d", TestImg0.ref, "sleep", "2"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[getting file] docker command error:")
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.CtrFile(cid, "/not-existing")

		// --- Then ---
		assert.Equal(t, "", have)
	})
}

func Test_DockerT_NetLs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		name := RandNet()
		args := []string{
			"network",
			"create",
			"--attachable",
			"--label", xdef.LabImgAuthors + "=" + t.Name(),
			"--label", "com.ctx42.meta.abc=abc",
			name,
		}
		nid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "network", "rm", nid) })
		exekit.New(t).Exe("docker", "inspect", nid)

		tspy := tester.New(t)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.NetLs()

		// --- Then ---
		network := have.FindByName(name)
		assert.NotEmpty(t, network.ID)
		assert.Equal(t, "bridge", network.Driver)
		assert.Equal(t, name, network.Name)
		assert.True(t, network.Attachable)
		wLabels := map[string]string{
			xdef.LabImgAuthors:   t.Name(),
			"com.ctx42.meta.abc": "abc",
		}
		assert.Equal(t, wLabels, network.Labels)
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[getting networks] docker command error:")
		tspy.Close()

		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := NewT(tspy, WithEnv(env))

		// --- When ---
		ims := dkr.NetLs()

		// --- Then ---
		assert.Nil(t, ims)
	})
}

func Test_DockerT_NetRm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		name := RandNet()
		args := []string{"network", "create", name}
		nid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "network", "rm", nid)
		})
		exekit.New(t).Exe("docker", "inspect", nid)

		tspy := tester.New(t)
		tspy.Close()

		dkr := NewT(tspy)

		// --- When ---
		have := dkr.NetRm(name)

		// --- Then ---
		assert.True(t, have)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "inspect", nid)
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("[removing network] docker command error:")
		tspy.Close()

		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := NewT(tspy, WithEnv(env))

		// --- When ---
		have := dkr.NetRm("00000000-0000-0000-0000-000000000000")

		// --- Then ---
		assert.False(t, have)
	})
}
