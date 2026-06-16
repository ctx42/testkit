// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xdef/pkg/xdef"

	"github.com/ctx42/testkit/pkg/exekit"
	"github.com/ctx42/testkit/pkg/netkit"
	"github.com/ctx42/testkit/pkg/oskit"
	"github.com/ctx42/testkit/pkg/randkit"
	"github.com/ctx42/testkit/pkg/testkit"
)

func Test_WithEnv(t *testing.T) {
	// --- Given ---
	env := []string{"KEY0=VAL0", "KEY1=VAL1"}

	// --- When ---
	dkr := New(WithEnv(env))

	// --- Then ---
	assert.Equal(t, env, dkr.env)
}

func Test_New(t *testing.T) {
	t.Run("default env", func(t *testing.T) {
		// --- When ---
		dkr := New()

		// --- Then ---
		assert.Equal(t, os.Environ(), dkr.env)
	})

	t.Run("WithEnv", func(t *testing.T) {
		// --- Given ---
		env := []string{"KEY=VALUE"}

		// --- When ---
		dkr := New(WithEnv(env))

		// --- Then ---
		assert.Equal(t, env, dkr.env)
	})
}

func Test_Docker_ImgPull(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		err := dkr.ImgPull(TestImageBaseRef)

		// --- Then ---
		assert.NoError(t, err)
		exekit.New(t).Exe("docker", "image", "inspect", TestImageBaseRef)
	})

	t.Run("error - non-existent ref", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		err := dkr.ImgPull("busybox:abc")

		// --- Then ---
		assert.ErrorContain(t, "failed to resolve reference", err)
		wMsg := "" +
			"[pulling image] docker command error:\n" +
			"   cmd: docker pull busybox:abc\n" +
			"   err: exit status 1\n" +
			"  eout: Error response from daemon: failed to resolve reference"
		assert.ErrorContain(t, wMsg, err)
	})

	t.Run("error - invalid ref", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		err := dkr.ImgPull("***")

		// --- Then ---
		wMsg := "" +
			"[pulling image] docker command error:\n" +
			"   cmd: docker pull ***\n" +
			"   err: exit status 1\n" +
			"  eout: invalid reference format"
		assert.ErrorEqual(t, wMsg, err)
	})
}

func Test_Docker_Build(t *testing.T) {
	t.Run("ref and iid format", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		dkfOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)

		// --- When ---
		ref, iid, err := dkr.Build(dkfOpt, argOpt)

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })

		assert.NotContain(t, ":", iid)
		assert.Contain(t, "ctx42-tst-img-", ref)
		assert.Contain(t, ":ctx42-tst-tag-", ref)
		exekit.New(t).Exe("docker", "history", iid)
	})

	t.Run("build from the current working directory", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)
		oskit.Chdir(t, "testdata/simple")

		// --- When ---
		ref, iid, err := dkr.Build(argOpt)

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })

		assert.NotContain(t, ":", iid)
		assert.Contain(t, "ctx42-tst-img-", ref)
		assert.Contain(t, ":ctx42-tst-tag-", ref)
		exekit.New(t).Exe("docker", "history", iid)
	})

	t.Run("labels and env values - default build args", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		dkfOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)

		// --- When ---
		ref, iid, err := dkr.Build(dkfOpt, argOpt)

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })

		args := []string{"image", "inspect", "--format", "{{slice .Id 7}}", ref}
		sout := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		assert.Equal(t, iid, sout)

		hLabels := must.Value(getLabels(t.Context(), os.Environ(), ref))
		wLabels := map[string]string{
			xdef.LabImgCreated: "0001-01-01T00:00:00Z",
			xdef.LabImgAuthors: "unknown",
			xdef.LabImgRefName: "unknown",
			xdef.LabImgSrc:     "unknown",
			xdef.LabImgRev:     "0000000",
			xdef.LabImgVer:     "v0.0.0",
			labTestEmpty:       "",
			labTestValue:       "value",
		}
		assert.MapSubset(t, wLabels, hLabels)

		hEnv := must.Value(getEnvs(t.Context(), os.Environ(), ref))
		wEnv := map[string]string{
			xdef.EnvImgCreated: "0001-01-01T00:00:00Z",
			xdef.EnvImgAuthors: "unknown",
			xdef.EnvImgRefName: "unknown",
			xdef.EnvImgSrc:     "unknown",
			xdef.EnvImgRev:     "0000000",
			xdef.EnvImgVer:     "v0.0.0",
			envTestEmpty:       "",
			envTestValue:       "value",
		}
		assert.MapSubset(t, wEnv, hEnv)
	})

	t.Run("labels and env values - custom build args", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		dkfOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		argOpt := WithBuildArgs(map[string]string{
			xdef.EnvImgBaseName: TestImageBaseRef,
			xdef.EnvImgCreated:  "2000-01-02T03:04:05Z",
			xdef.EnvImgAuthors:  "author",
			xdef.EnvImgRefName:  "ccid",
			xdef.EnvImgSrc:      "repo",
			xdef.EnvImgRev:      "abc",
			xdef.EnvImgVer:      "v1.2.3",
		})

		// --- When ---
		ref, iid, err := dkr.Build(dkfOpt, argOpt)

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })

		args := []string{"image", "inspect", "--format", "{{slice .Id 7}}", ref}
		sout := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		assert.Equal(t, iid, sout)

		hLabels := must.Value(getLabels(t.Context(), os.Environ(), ref))
		wLabels := map[string]string{
			xdef.LabImgCreated: "2000-01-02T03:04:05Z",
			xdef.LabImgAuthors: "author",
			xdef.LabImgRefName: "ccid",
			xdef.LabImgSrc:     "repo",
			xdef.LabImgRev:     "abc",
			xdef.LabImgVer:     "v1.2.3",
			labTestEmpty:       "",
			labTestValue:       "value",
		}
		assert.MapSubset(t, wLabels, hLabels)

		hEnv := must.Value(getEnvs(t.Context(), os.Environ(), ref))
		wEnv := map[string]string{
			xdef.EnvImgCreated: "2000-01-02T03:04:05Z",
			xdef.EnvImgAuthors: "author",
			xdef.EnvImgRefName: "ccid",
			xdef.EnvImgSrc:     "repo",
			xdef.EnvImgRev:     "abc",
			xdef.EnvImgVer:     "v1.2.3",
			envTestEmpty:       "",
			envTestValue:       "value",
		}
		assert.MapSubset(t, wEnv, hEnv)
	})

	t.Run("provide a Dockerfile as a reader", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		rdr := must.Value(os.Open("testdata/simple/Dockerfile"))
		dkfOpt := WithBuildDkfRdr(rdr)
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)

		// --- When ---
		ref, iid, err := dkr.Build(dkfOpt, argOpt)

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })

		args := []string{"image", "inspect", "--format", "{{slice .Id 7}}", ref}
		sout := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		assert.Equal(t, iid, sout)
	})

	t.Run("error - empty Dockerfile", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		dkfOpt := WithBuildDkfPth("testdata/invalid/Dockerfile")

		// --- When ---
		ref, iid, err := dkr.Build(dkfOpt)

		// --- Then ---
		assert.ErrorContain(t, "[building image] docker command error", err)
		assert.ErrorContain(t, "the Dockerfile cannot be empty", err)
		assert.Empty(t, ref)
		assert.Empty(t, iid)
	})

	t.Run("error - WithBuildDkfPth with WithBuildDkfRdr", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		dkfPthOpt := WithBuildDkfPth("testdata/invalid/Dockerfile")
		dkfRdrOpt := WithBuildDkfRdr(strings.NewReader("abc"))

		// --- When ---
		ref, iid, err := dkr.Build(dkfPthOpt, dkfRdrOpt)

		// --- Then ---
		wMsg := "WithBuildDkfPth and WithBuildDkfRdr are mutually exclusive"
		assert.ErrorContain(t, wMsg, err)
		assert.Empty(t, ref)
		assert.Empty(t, iid)
	})

	t.Run("WithBuildName", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		dkfOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)
		name := RandName() + "-my-name"
		nameOpt := WithBuildName(name)

		// --- When ---
		ref, _, err := dkr.Build(dkfOpt, argOpt, nameOpt)

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })

		assert.Contain(t, name+":", ref)
	})

	t.Run("WithBuildTag", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		dkfOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)
		tag := RandName() + "-my-tag"
		tagOpt := WithBuildTag(tag)

		// --- When ---
		ref, _, err := dkr.Build(dkfOpt, argOpt, tagOpt)

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })

		assert.Contain(t, ":"+tag, ref)
	})

	t.Run("dry run - SSH_AUTH_SOCK set", func(t *testing.T) {
		// --- Given ---
		env := WithEnv(append(os.Environ(), "SSH_AUTH_SOCK=/tmp/ssh.sock"))
		dkr := New(env)
		idfPth := filepath.Join(t.TempDir(), "iid.log")
		have := &bytes.Buffer{}
		dryOpt := WithBuildDryRun(have)
		dkfOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		idfOpt := withBuildIIDFile(idfPth)
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)

		// --- When ---
		ref, iid, err := dkr.Build(dryOpt, dkfOpt, argOpt, idfOpt)

		// --- Then ---
		assert.NoError(t, err)
		assert.Empty(t, iid)
		assert.Contain(t, ":", ref)
		assert.Contain(t, "ctx42-tst-img-", ref)
		assert.Contain(t, "ctx42-tst-tag-", ref)

		want := "" +
			"DOCKER_BUILDKIT=1 docker build --rm " +
			"-t %s " +
			"--iidfile %s " +
			"--ssh=default " +
			"--build-arg OCI_IMAGE_BASE_NAME=busybox:1.38-uclibc " +
			"--file Dockerfile ."
		assert.Equal(t, fmt.Sprintf(want, ref, idfPth), have.String())
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "image", "rm", ref)
	})

	t.Run("dry run - no SSH_AUTH_SOCK", func(t *testing.T) {
		// --- Given ---
		envMap := oskit.EnvSplit(os.Environ())
		delete(envMap, "SSH_AUTH_SOCK")
		dkr := New(WithEnv(oskit.EnvJoin(envMap)))
		idfPth := filepath.Join(t.TempDir(), "iid.log")
		have := &bytes.Buffer{}
		dryOpt := WithBuildDryRun(have)
		dkfOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		idfOpt := withBuildIIDFile(idfPth)
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)

		// --- When ---
		ref, iid, err := dkr.Build(dryOpt, dkfOpt, argOpt, idfOpt)

		// --- Then ---
		assert.NoError(t, err)
		assert.Empty(t, iid)

		want := "" +
			"DOCKER_BUILDKIT=1 docker build --rm " +
			"-t %s " +
			"--iidfile %s " +
			"--build-arg OCI_IMAGE_BASE_NAME=busybox:1.38-uclibc " +
			"--file Dockerfile ."
		assert.Equal(t, fmt.Sprintf(want, ref, idfPth), have.String())
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "image", "rm", ref)
	})

	t.Run("dry run - dockerfile as reader", func(t *testing.T) {
		// --- Given ---
		env := WithEnv(append(os.Environ(), "SSH_AUTH_SOCK=/tmp/ssh.sock"))
		dkr := New(env)
		idfPth := filepath.Join(t.TempDir(), "iid.log")
		have := &bytes.Buffer{}
		rdr := must.Value(os.Open("testdata/simple/Dockerfile"))
		dryOpt := WithBuildDryRun(have)
		dkfOpt := WithBuildDkfRdr(rdr)
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)
		idfOpt := withBuildIIDFile(idfPth)

		// --- When ---
		ref, iid, err := dkr.Build(dryOpt, dkfOpt, argOpt, idfOpt)

		// --- Then ---
		assert.NoError(t, err)
		assert.Empty(t, iid)
		assert.Contain(t, ":", ref)

		want := "" +
			"DOCKER_BUILDKIT=1 docker build --rm " +
			"-t %s " +
			"--iidfile %s " +
			"--ssh=default " +
			"--build-arg OCI_IMAGE_BASE_NAME=busybox:1.38-uclibc " +
			"-"
		assert.Equal(t, fmt.Sprintf(want, ref, idfPth), have.String())
	})

	t.Run("iidfile is kept when its path is set manually", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		idfPth := filepath.Join(t.TempDir(), "iid.log")
		dkfOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)
		idfOpt := withBuildIIDFile(idfPth)

		// --- When ---
		ref, iid, err := dkr.Build(dkfOpt, argOpt, idfOpt)

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })

		content := must.Value(os.ReadFile(idfPth))
		assert.Equal(t, iid, StripHashName(strings.TrimSpace(string(content))))
	})

	t.Run("WithBuildLabel", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		dryBuf := &bytes.Buffer{}
		dryOpt := WithBuildDryRun(dryBuf)
		dkfOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		idfPth := filepath.Join(t.TempDir(), "iid.log")
		idfOpt := withBuildIIDFile(idfPth)
		labOpt := WithBuildLabel("my.label", "val")

		// --- When ---
		_, _, err := dkr.Build(dryOpt, dkfOpt, idfOpt, labOpt)

		// --- Then ---
		assert.NoError(t, err)
		assert.Contain(t, "--label my.label=val", dryBuf.String())
	})

	t.Run("build with no cache - set by env variable", func(t *testing.T) {
		// --- Given ---
		env := append(os.Environ(), envBuildNoCache+"=1")
		dkr := New(WithEnv(env))
		idPth := filepath.Join(t.TempDir(), "iid.log")
		have := &bytes.Buffer{}
		dryOpt := WithBuildDryRun(have)
		dkfOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)
		idfOpt := withBuildIIDFile(idPth)

		// --- When ---
		_, _, err := dkr.Build(dryOpt, dkfOpt, argOpt, idfOpt)

		// --- Then ---
		assert.NoError(t, err)
		assert.Contain(t, "--no-cache", have.String())
	})

	t.Run("build with no cache - set by option", func(t *testing.T) {
		// --- Given ---
		env := WithEnv(append(os.Environ(), "SSH_AUTH_SOCK=default"))
		dkr := New(env)
		idfPth := filepath.Join(t.TempDir(), "iid.log")
		buf := &bytes.Buffer{}
		dryOpt := WithBuildDryRun(buf)
		dkfOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		idfOpt := withBuildIIDFile(idfPth)
		argOpt := WithBuildArg(xdef.EnvImgBaseName, TestImageBaseRef)
		nocOpt := WithBuildNoCache()

		// --- When ---
		ref, iid, err := dkr.Build(dryOpt, dkfOpt, argOpt, idfOpt, nocOpt)

		// --- Then ---
		assert.NoError(t, err)
		assert.Empty(t, iid)
		assert.Contain(t, ":", ref)
		assert.Contain(t, "ctx42-tst-img-", ref)
		assert.Contain(t, "ctx42-tst-tag-", ref)

		have := buf.String()
		want := "" +
			"DOCKER_BUILDKIT=1 docker build " +
			"--rm -t %s " +
			"--iidfile %s " +
			"--ssh=default " +
			"--build-arg OCI_IMAGE_BASE_NAME=busybox:1.38-uclibc " +
			"--no-cache " +
			"--file Dockerfile ."
		assert.Equal(t, fmt.Sprintf(want, ref, idfPth), have)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "image", "rm", ref)
	})
}

func Test_readIIDFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		ref := "ctx42-tst-img:ctx42-tst-tag"
		pth := filepath.Join(t.TempDir(), "iid.log")
		must.Nil(os.WriteFile(pth, []byte("sha256:abc123"), 0600))

		// --- When ---
		have, err := readIIDFile(ref, pth)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "abc123", have)
	})

	t.Run("error - file not found", func(t *testing.T) {
		// --- Given ---
		ref := "ctx42-tst-img:ctx42-tst-tag"
		pth := filepath.Join(t.TempDir(), "not-there.log")

		// --- When ---
		have, err := readIIDFile(ref, pth)

		// --- Then ---
		assert.Error(t, err)
		assert.Equal(t, "", have)
	})

	t.Run("error - empty iid after strip", func(t *testing.T) {
		// --- Given ---
		ref := "ctx42-tst-img:ctx42-tst-tag"
		pth := filepath.Join(t.TempDir(), "iid.log")
		must.Nil(os.WriteFile(pth, []byte("sha256:"), 0600))

		// --- When ---
		have, err := readIIDFile(ref, pth)

		// --- Then ---
		wMsg := "" +
			"building image: iid file has no image ID:\n" +
			"      ref: ctx42-tst-img:ctx42-tst-tag\n" +
			"     path: %s\n" +
			"  content: \"sha256:\""
		wMsg = fmt.Sprintf(wMsg, pth)
		assert.ErrorEqual(t, wMsg, err)
		assert.Equal(t, "", have)
	})
}

func Test_Docker_testImageBuildOptions(t *testing.T) {
	t.Run("static args", func(t *testing.T) {
		// --- Given ---
		env := append(os.Environ(), xdef.EnvImgCreated+"=2000-01-02T03:04:05Z")
		env = append(env, xdef.EnvImgRefName+"=ccid")
		dkr := New(WithEnv(env))

		// --- When ---
		have := dkr.testImageBuildOptions()

		// --- Then ---
		opts := &BuildOptions{}
		for _, opt := range have {
			opt(opts)
		}
		assert.NotNil(t, opts.dkfRdr)
		assert.Equal(t, "", opts.dkfPth)
		assert.Equal(t, "", opts.imgName)
		assert.Equal(t, "", opts.imgTag)
		assert.Nil(t, opts.labels)
		assert.Equal(t, "", opts.iidPth)
		assert.False(t, opts.noCache)
		assert.Nil(t, opts.dryRun)
		want := map[string]string{
			xdef.EnvImgCreated:  "2000-01-02T03:04:05Z",
			xdef.EnvImgRefName:  "ccid",
			xdef.EnvImgBaseName: TestImageBaseRef,
			xdef.EnvImgSrc:      "repo",
			xdef.EnvImgRev:      "12345678",
			xdef.EnvImgVer:      "v1.2.3",
		}
		assert.Equal(t, want, opts.args)
		assert.Fields(t, 9, BuildOptions{})
	})

	t.Run("OCI_IMAGE_CREATED from env", func(t *testing.T) {
		// --- Given ---
		env := append(os.Environ(), xdef.EnvImgCreated+"=2000-01-02T03:04:05Z")
		dkr := New(WithEnv(env))

		// --- When ---
		have := dkr.testImageBuildOptions()

		// --- Then ---
		opts := &BuildOptions{}
		for _, opt := range have {
			opt(opts)
		}
		assert.Equal(t, "2000-01-02T03:04:05Z", opts.args[xdef.EnvImgCreated])
	})

	t.Run("OCI_IMAGE_REF_NAME from env", func(t *testing.T) {
		// --- Given ---
		env := append(os.Environ(), xdef.EnvImgRefName+"=ccid")
		dkr := New(WithEnv(env))

		// --- When ---
		have := dkr.testImageBuildOptions()

		// --- Then ---
		opts := &BuildOptions{}
		for _, opt := range have {
			opt(opts)
		}
		assert.Equal(t, "ccid", opts.args[xdef.EnvImgRefName])
	})
}

func Test_Docker_BuildTestImg(t *testing.T) {
	t.Run("created and ccid set from env variables", func(t *testing.T) {
		// --- Given ---
		env := append(os.Environ(), xdef.EnvImgCreated+"=2000-01-02T03:04:05Z")
		env = append(env, xdef.EnvImgRefName+"=ccid")
		dkr := New(WithEnv(env))

		// --- When ---
		ref, iid, err := dkr.BuildTestImg()

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })
		assert.NotEmpty(t, iid)

		hLabels := must.Value(getLabels(t.Context(), os.Environ(), ref))
		wLabels := map[string]string{
			xdef.LabImgCreated:  "2000-01-02T03:04:05Z",
			xdef.LabImgSrc:      "repo",
			xdef.LabImgRev:      "12345678",
			xdef.LabImgVer:      "v1.2.3",
			xdef.LabImgRefName:  "ccid",
			xdef.LabImgBaseName: TestImageBaseRef,
		}
		assert.MapSubset(t, wLabels, hLabels)

		hEnv := must.Value(getEnvs(t.Context(), os.Environ(), ref))
		wEnv := map[string]string{
			xdef.EnvImgCreated:  "2000-01-02T03:04:05Z",
			xdef.EnvImgSrc:      "repo",
			xdef.EnvImgRev:      "12345678",
			xdef.EnvImgVer:      "v1.2.3",
			xdef.EnvImgRefName:  "ccid",
			xdef.EnvImgBaseName: TestImageBaseRef,
		}
		assert.MapSubset(t, wEnv, hEnv)
	})

	t.Run("empty environment build date is ignored", func(t *testing.T) {
		// --- Given ---
		env := append(os.Environ(), xdef.EnvImgCreated+"=")
		env = append(env, xdef.EnvImgRefName+"=ccid")
		dkr := New(WithEnv(env))

		// --- When ---
		ref, iid, err := dkr.BuildTestImg()

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })
		assert.NotEmpty(t, iid)

		hLabels := must.Value(getLabels(t.Context(), os.Environ(), ref))
		valBDate, _ := assert.HasKey(t, xdef.LabImgCreated, hLabels)
		assert.Within(t, time.Now(), "3s", valBDate)
		assert.True(t, strings.HasSuffix(valBDate, "Z"))

		hEnv := must.Value(getEnvs(t.Context(), os.Environ(), ref))
		valBDate, _ = assert.HasKey(t, xdef.EnvImgCreated, hEnv)
		assert.Within(t, time.Now(), "3s", valBDate)
	})

	t.Run("empty environment CCID is ignored", func(t *testing.T) {
		// --- Given ---
		env := append(os.Environ(), xdef.EnvImgRefName+"=")
		dkr := New(WithEnv(env))

		// --- When ---
		ref, iid, err := dkr.BuildTestImg()

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })
		assert.NotEmpty(t, iid)

		hLabels := must.Value(getLabels(t.Context(), os.Environ(), ref))
		val, _ := assert.HasKey(t, xdef.LabImgRefName, hLabels)
		assert.True(t, strings.HasPrefix(val, "no-ccid-"))

		hEnv := must.Value(getEnvs(t.Context(), os.Environ(), ref))
		val, _ = assert.HasKey(t, xdef.EnvImgRefName, hEnv)
		assert.True(t, strings.HasPrefix(val, "no-ccid-"))
	})

	t.Run("entrypoint without args", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		ref, iid, err := dkr.BuildTestImg()

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })

		sout := exekit.New(t).ExeStdout("docker", "run", "--rm", iid)
		assert.Equal(t, "hello\n", sout)
	})

	t.Run("entrypoint with args", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		ref, iid, err := dkr.BuildTestImg()

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "image", "rm", ref) })

		args := []string{"run", "--rm", iid, "echo", "hello"}
		sout := exekit.New(t).ExeStdout("docker", args...)
		assert.Equal(t, "hello\n", sout)
	})
}

func Test_Docker_ImgLs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		have, err := dkr.ImgLs()

		// --- Then ---
		assert.NoError(t, err)

		img := have.FindByRef(TestImg0.ref)
		assert.NotNil(t, img)
		assert.Equal(t, TestImg0.iid, img.ID)
		assert.Equal(t, TestImg0.name, img.Repository)
		assert.Equal(t, TestImg0.tag, img.Tag)
	})

	t.Run("found with filters", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		filOpt := WithImgLsFilter("reference=" + TestImg1.ref)

		// --- When ---
		have, err := dkr.ImgLs(filOpt)

		// --- Then ---
		assert.NoError(t, err)
		assert.Len(t, 1, have)

		img := have.FindByRef(TestImg1.ref)
		assert.NotNil(t, img)
		assert.Equal(t, TestImg1.iid, img.ID)
		assert.Equal(t, TestImg1.name, img.Repository)
		assert.Equal(t, TestImg1.tag, img.Tag)
	})

	t.Run("not found with filters", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		ref := fmt.Sprintf("%s:%s", RandName(), RandTag())
		filOpt := WithImgLsFilter("reference=" + ref)

		// --- When ---
		ims, err := dkr.ImgLs(filOpt)

		// --- Then ---
		assert.NoError(t, err)
		assert.Len(t, 0, ims)
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := New(WithEnv(env))

		// --- When ---
		ims, err := dkr.ImgLs()

		// --- Then ---
		assert.ErrorContain(t, "[getting images] docker command error", err)
		assert.ErrorContain(t, host, err)
		assert.Nil(t, ims)
	})
}

func Test_Docker_Labels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		have, err := dkr.Labels(TestImg0.ref)

		// --- Then ---
		assert.NoError(t, err)
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
		dkr := New()
		ref := RandRef()

		// --- When ---
		have, err := dkr.Labels(ref)

		// --- Then ---
		assert.ErrorContain(t, "[getting labels]", err)
		assert.Nil(t, have)
	})
}

func Test_Docker_Label(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		have, err := dkr.Label(TestImg0.ref, xdef.LabImgTitle)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "Image0", have)
	})

	t.Run("empty existing label", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		have, err := dkr.Label(TestImg0.ref, labTestEmpty)

		// --- Then ---
		assert.NoError(t, err)
		assert.Empty(t, have)
	})

	t.Run("error - non-existent label", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		have, err := dkr.Label(TestImg0.ref, "not.existing.label")

		// --- Then ---
		wMsg := "" +
			"[getting label] expected label to exist:\n" +
			"     ref: %s\n" +
			"    want: \"not.existing.label\"\n" +
			"  labels:\n" +
			"          \"com.ctx42.test.empty\"\n" +
			"          \"org.opencontainers.image.base.name\"\n" +
			"          \"org.opencontainers.image.created\"\n" +
			"          \"org.opencontainers.image.title\""
		wMsg = fmt.Sprintf(wMsg, TestImg0.ref)
		assert.ErrorEqual(t, wMsg, err)
		assert.Empty(t, have)
	})

	t.Run("error - non-existent reference", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		ref := RandRef()

		// --- When ---
		have, err := dkr.Label(ref, xdef.LabImgAuthors)

		// --- Then ---
		wMsg := "" +
			"[getting label] docker command error:\n" +
			"   cmd: docker inspect --format {{ json .Config.Labels }} %s\n" +
			"   err: exit status 1\n" +
			"  eout: error: no such object: %s"
		wMsg = fmt.Sprintf(wMsg, ref, ref)
		assert.ErrorEqual(t, wMsg, err)
		assert.Empty(t, have)
	})
}

func Test_Docker_Envs(t *testing.T) {
	t.Run("success by iid", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		have, err := dkr.Envs(TestImg0.iid)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			xdef.EnvImgCreated: "2000-01-02T03:04:05Z",
			xdef.EnvImgTitle:   "Image0",
			envTestEmpty:       "",
		}
		assert.MapSubset(t, want, have)
	})

	t.Run("error - non-existent ref", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		ref := RandRef()

		// --- When ---
		have, err := dkr.Envs(ref)

		// --- Then ---
		assert.ErrorContain(t, "[getting environment variables]", err)
		assert.Nil(t, have)
	})
}

func Test_Docker_Env(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		have, err := dkr.Env(TestImg0.iid, xdef.EnvImgTitle)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "Image0", have)
	})

	t.Run("empty existing environment variable", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		have, err := dkr.Env(TestImg0.ref, envTestEmpty)

		// --- Then ---
		assert.NoError(t, err)
		assert.Empty(t, have)
	})

	t.Run("error - non-existent environment variable", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		have, err := dkr.Env(TestImg0.iid, "NOT_EXISTING")

		// --- Then ---
		wMsg := "[getting environment variable] expected variable to exist:\n" +
			"   ref: %s\n" +
			"  want: \"NOT_EXISTING\"\n" +
			"   env:\n" +
			"        \"CTX42_TEST_EMPTY\"\n" +
			"        \"OCI_IMAGE_BASE_NAME\"\n" +
			"        \"OCI_IMAGE_CREATED\"\n" +
			"        \"OCI_IMAGE_TITLE\"\n" +
			"        \"PATH\""
		wMsg = fmt.Sprintf(wMsg, TestImg0.iid)
		assert.ErrorEqual(t, wMsg, err)
		assert.Empty(t, have)
	})

	t.Run("error - non-existent reference", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		ref := RandRef()

		// --- When ---
		have, err := dkr.Env(ref, "CTX42_TEST_VALUE0")

		// --- Then ---
		wMsg := "[getting environment variable] docker command error"
		assert.ErrorContain(t, wMsg, err)
		assert.Empty(t, have)
	})
}

func Test_Docker_ImgRm(t *testing.T) {
	t.Run("remove existing iid", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		img := must.Value(createMinImage(t.Name()))
		t.Cleanup(func() { exekit.New(t).Exe("docker", "rmi", "-f", img.iid) })

		// --- When ---
		err := dkr.ImgRm(img.iid)

		// --- Then ---
		assert.NoError(t, err)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "history", img.ref)
	})

	t.Run("remove existing image by ref", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		img := must.Value(createMinImage(t.Name()))
		t.Cleanup(func() { exekit.New(t).Exe("docker", "rmi", "-f", img.iid) })

		// --- When ---
		err := dkr.ImgRm(img.ref)

		// --- Then ---
		assert.NoError(t, err)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "history", img.ref)
	})

	t.Run("error - used by running container", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		img := must.Value(createMinImage(t.Name()))

		// Run the container in the background long
		// enough for ImgRm to fail at least once.
		args := []string{"run", "-d", "--rm", img.iid, "sleep", "3"}
		cid := exekit.New(t).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t).Exe("docker", "kill", cid)
			exekit.New(t).Exe("docker", "rmi", "-f", img.iid)
		})

		tryOpt := WithImgRmTries(1)
		slpOpt := WithImgRmSleep(100 * time.Millisecond)

		// --- When ---
		err := dkr.ImgRm(img.iid, tryOpt, slpOpt)

		// --- Then ---
		assert.Error(t, err)
		wMsg := "[removing image] docker command error:\n" +
			"   cmd: docker image rm %s\n" +
			"   err: exit status 1\n" +
			"  eout: Error response from daemon: conflict: unable to delete"
		wMsg = fmt.Sprintf(wMsg, img.iid)
		assert.ErrorContain(t, wMsg, err)
		wMsg = "(cannot be forced) - image is being used by running container"
		assert.ErrorContain(t, wMsg, err)
	})

	t.Run("remove image used by running container by ref", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		img := must.Value(createMinImage(t.Name()))

		// Run the container in the background long
		// enough for ImgRm to fail at least once.
		args := []string{"run", "-d", "--rm", img.iid, "sleep", "3"}
		cid := exekit.New(t).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t).Exe("docker", "kill", cid)
			exekit.New(t).Exe("docker", "rmi", "-f", img.iid)
		})

		tryOpt := WithImgRmTries(2)
		slpOpt := WithImgRmSleep(100 * time.Millisecond)

		// --- When ---
		err := dkr.ImgRm(img.ref, tryOpt, slpOpt)

		// --- Then ---
		assert.NoError(t, err)
		exekit.New(t).Exe("docker", "history", img.iid)
	})

	t.Run("remove image used by stopped container by iid", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		img := must.Value(createMinImage(t.Name()))

		// Run container in the background but do not remove it after it exits.
		args := []string{"run", "-d", img.iid, "echo", "ok"}
		cid := exekit.New(t).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t).Exe("docker", "rm", cid)
			exekit.New(t).Exe("docker", "rmi", "-f", img.iid)
		})

		// --- When ---
		err := dkr.ImgRm(img.iid)

		// --- Then ---
		assert.NoError(t, err)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "history", img.iid)
	})

	t.Run("remove image used by stopped container by ref", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		img := must.Value(createMinImage(t.Name()))

		// Run container in the background but do not remove it after it exits.
		args := []string{"run", "-d", img.ref, "echo", "ok"}
		cid := exekit.New(t).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t).Exe("docker", "rm", cid)
			exekit.New(t).Exe("docker", "rmi", "-f", img.iid)
		})

		// --- When ---
		err := dkr.ImgRm(img.ref)

		// --- Then ---
		assert.NoError(t, err)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "history", img.iid)
	})

	t.Run("non-existent image", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		err := dkr.ImgRm(RandRef())

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty ref", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		err := dkr.ImgRm("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := New(WithEnv(env))
		ref := RandRef()

		// --- When ---
		err := dkr.ImgRm(ref)

		// --- Then ---
		wMsg := "" +
			"[removing image] docker command error:\n" +
			"   cmd: docker image rm %s\n" +
			"   err: exit status 1\n" +
			"  eout: Cannot connect to the Docker daemon at %s. " +
			"Is the docker daemon running?"
		wMsg = fmt.Sprintf(wMsg, ref, host)
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("WithImgRmIgnoreErrors does not suppress error return", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		img := must.Value(createMinImage(t.Name()))

		// Run the container in the background long
		// enough for ImgRm to fail at least once.
		args := []string{"run", "-d", "--rm", img.iid, "sleep", "3"}
		cid := exekit.New(t).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
			exekit.New(t).Exe("docker", "rmi", "-f", img.iid)
		})

		tryOpt := WithImgRmTries(1)
		slpOpt := WithImgRmSleep(100 * time.Millisecond)
		ignOpt := WithImgRmIgnoreErrors()

		// --- When ---
		err := dkr.ImgRm(img.iid, tryOpt, slpOpt, ignOpt)

		// --- Then ---
		assert.Error(t, err)
		exekit.New(t).Exe("docker", "history", img.iid)
	})
}

func Test_Docker_CtrRun(t *testing.T) {
	t.Run("success by iid", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		have, err := dkr.CtrRun(TestImg0.iid)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "hello", have)
	})

	t.Run("success by image ref", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		have, err := dkr.CtrRun(TestImg0.ref)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "hello", have)
	})

	t.Run("error - unknown ref", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		ref := "img-not-existing:tag-not-existing"

		// --- When ---
		have, err := dkr.CtrRun(ref)

		// --- Then ---
		wMsg := "" +
			"[running image] docker command error:\n" +
			"   cmd: docker run --rm %s\n" +
			"   err: exit status 125\n" +
			"  eout:\n" +
			"        Unable to find image '%s' locally"
		wMsg = fmt.Sprintf(wMsg, ref, ref)
		assert.ErrorContain(t, wMsg, err)
		assert.Equal(t, "", have)
	})

	t.Run("get container ID via a channel", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		noRmOpt := WithCtrRunNoRemove()
		kv := randkit.Str()
		labOpt := WithCtrRunLabel(kv, kv)
		argOpt := WithCtrRunArgs("echo", "abc")
		cidCh, cidOpt := WithCtrRunCID()

		// --- When ---
		have, err := dkr.CtrRun(TestImg0.iid, cidOpt, noRmOpt, labOpt, argOpt)

		// --- Then ---
		assert.NoError(t, err)
		cid, ok := <-cidCh
		assert.True(t, ok)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "rm", cid) })

		assert.Equal(t, "abc", have)
		format := fmt.Sprintf(`{{ index .Config.Labels %q}}`, kv)
		args := []string{"inspect", cid, "--format", format}
		sout := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		assert.Equal(t, kv, sout)
	})

	t.Run("WithCtrRunNoRemove - container persists after run", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		noRmOpt := WithCtrRunNoRemove()
		cidChn, cidOpt := WithCtrRunCID()

		// --- When ---
		have, err := dkr.CtrRun(TestImg0.iid, noRmOpt, cidOpt)

		// --- Then ---
		assert.NoError(t, err)
		cid, ok := <-cidChn
		assert.True(t, ok)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "rm", cid) })

		assert.Equal(t, "hello", have)
		exekit.New(t).Exe("docker", "inspect", cid)
	})

	t.Run("detach", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		noRmOpt := WithCtrRunNoRemove()
		detOpt := WithCtrRunDetach()

		// --- When ---
		cid, err := dkr.CtrRun(TestImg0.iid, noRmOpt, detOpt)

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "rm", cid) })

		exekit.New(t).Exe("docker", "inspect", cid)
	})

	t.Run("with label", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		noRmOpt := WithCtrRunNoRemove()
		detOpt := WithCtrRunDetach()
		kv := randkit.Str()
		labOpt := WithCtrRunLabel(kv, kv)

		// --- When ---
		cid, err := dkr.CtrRun(TestImg0.iid, noRmOpt, detOpt, labOpt)

		// --- Then ---
		assert.NoError(t, err)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "rm", cid) })

		format := fmt.Sprintf(`{{ index .Config.Labels %q}}`, kv)
		args := []string{"inspect", cid, "--format", format}
		sout := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		assert.Equal(t, kv, sout)
	})

	t.Run("get container ID via CID file", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		noRmOpt := WithCtrRunNoRemove()
		cifPth := filepath.Join(t.TempDir(), "cid.log")
		cifOpt := WithCtrRunCIDPth(cifPth)
		kv := randkit.Str()
		labOpt := WithCtrRunLabel(kv, kv)
		argOpt := WithCtrRunArgs("echo", "abc")

		// --- When ---
		have, err := dkr.CtrRun(TestImg0.iid, cifOpt, noRmOpt, labOpt, argOpt)

		// --- Then ---
		assert.NoError(t, err)
		cid := must.Value(testkit.Wait4File("1s", cifPth))
		t.Cleanup(func() { exekit.New(t).Exe("docker", "rm", cid) })

		assert.Equal(t, "abc", have)
		format := fmt.Sprintf(`{{ index .Config.Labels %q}}`, kv)
		args := []string{"inspect", cid, "--format", format}
		sout := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		assert.Equal(t, kv, sout)
	})

	t.Run("cidfile is kept when its path is set manually", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		noRmOpt := WithCtrRunNoRemove()
		cidPth := filepath.Join(t.TempDir(), "cid.log")
		cidCh, cidOpt := WithCtrRunCID()
		cidPthOpt := WithCtrRunCIDPth(cidPth)

		// --- When ---
		_, err := dkr.CtrRun(TestImg0.iid, noRmOpt, cidOpt, cidPthOpt)

		// --- Then ---
		assert.NoError(t, err)
		cid, ok := <-cidCh
		assert.True(t, ok)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "rm", cid) })

		content := must.Value(os.ReadFile(cidPth))
		assert.Equal(t, cid, strings.TrimSpace(string(content)))
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := New(WithEnv(env))

		// --- When ---
		have, err := dkr.CtrRun(TestImg0.iid)

		// --- Then ---
		wMsg := "" +
			"[running image] docker command error:\n" +
			"   cmd: docker run --rm %s\n" +
			"   err: exit status 1\n" +
			"  eout: Cannot connect to the Docker daemon at %s. " +
			"Is the docker daemon running?"
		wMsg = fmt.Sprintf(wMsg, TestImg0.iid, host)
		assert.ErrorEqual(t, wMsg, err)
		assert.Empty(t, have)
	})
}

func Test_Docker_CtrRm(t *testing.T) {
	t.Run("remove the stopped container", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		args := []string{"run", "-d", TestImg0.iid}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		// --- When ---
		err := dkr.CtrRm(cid)

		// --- Then ---
		assert.NoError(t, err)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "inspect", cid)
	})

	t.Run("remove the running container", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		args := []string{"run", "-d", TestImg0.iid, "sleep", "5"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		// --- When ---
		err := dkr.CtrRm(cid)

		// --- Then ---
		assert.NoError(t, err)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "inspect", cid)
	})

	t.Run("remove a non-existent container", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		cid := "000000000000"

		// --- When ---
		err := dkr.CtrRm(cid)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := New(WithEnv(env))

		// --- When ---
		err := dkr.CtrRm(TestImg0.iid)

		// --- Then ---
		wMsg := "" +
			"[removing container] docker command error:\n" +
			"   cmd: docker rm --force %s\n" +
			"   err: exit status 1\n" +
			"  eout: Cannot connect to the Docker daemon at %s. " +
			"Is the docker daemon running?"
		wMsg = fmt.Sprintf(wMsg, TestImg0.iid, host)
		assert.ErrorEqual(t, wMsg, err)
	})
}

func Test_Docker_CtrKill(t *testing.T) {
	t.Run("kill running container", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		args := []string{"run", "-d", TestImg0.iid, "sleep", "5"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		// --- When ---
		err := dkr.CtrKill(cid)

		// --- Then ---
		assert.NoError(t, err)
		exekit.New(t).Exe("docker", "inspect", cid)
	})

	t.Run("error - kill unknown container", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		cid := "000000000000"

		// --- When ---
		err := dkr.CtrKill(cid)

		// --- Then ---
		wMsg := "" +
			"[killing container] docker command error:\n" +
			"   cmd: docker kill -s 9 000000000000\n" +
			"   err: exit status 1\n" +
			"  eout: Error response from daemon: " +
			"cannot kill container: 000000000000: " +
			"No such container: 000000000000"
		assert.ErrorEqual(t, wMsg, err)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "inspect", cid)
	})
}

func Test_Docker_CtrExec(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		args := []string{"run", "-d", TestImg0.iid, "sleep", "2"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		// --- When ---
		have, err := dkr.CtrExec(cid, "ls", "/")

		// --- Then ---
		assert.NoError(t, err)
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
		dkr := New()
		args := []string{"run", "-d", TestImg0.iid, "sleep", "2"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		// --- When ---
		have, err := dkr.CtrExec(cid, "not-existing")

		// --- Then ---
		wMsg := "" +
			"[container exec] docker command error:\n" +
			"   cmd: docker exec %s not-existing\n" +
			"   err: exit status 127\n" +
			"  sout: OCI runtime exec failed: exec failed: " +
			"unable to start container process: exec: \"not-existing\": " +
			"executable file not found in $PATH"
		wMsg = fmt.Sprintf(wMsg, cid)
		assert.ErrorEqual(t, wMsg, err)
		assert.Empty(t, have)
	})
}

func Test_Docker_CtrPs(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		// --- Given ---
		dkr := New()
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

		// --- When ---
		have, err := dkr.CtrPs()

		// --- Then ---
		assert.NoError(t, err)

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
		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := New(WithEnv(env))

		// --- When ---
		_, err := dkr.CtrPs()

		// --- Then ---
		wMsg := "" +
			"[listing containers] docker command error:\n" +
			"   cmd: docker ps --all --no-trunc --format {{.ID}} {{.Image}}\n" +
			"   err: exit status 1\n" +
			"  eout: Cannot connect to the Docker daemon at %s. " +
			"Is the docker daemon running?"
		wMsg = fmt.Sprintf(wMsg, host)
		assert.ErrorEqual(t, wMsg, err)
	})
}

func Test_Docker_CtrFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		args := []string{"run", "-d", TestImg0.ref, "sleep", "2"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})

		// --- When ---
		have, err := dkr.CtrFile(cid, "/file0.txt")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "file0\n", have)
	})

	t.Run("error - asking for directory", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		args := []string{"run", "-d", TestImg0.ref, "sleep", "2"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		// --- When ---
		have, err := dkr.CtrFile(cid, "/etc")

		// --- Then ---
		assert.ErrorContain(t, "expected file to exist in the container", err)
		assert.Equal(t, "", have)
	})

	t.Run("error - asking for a non-existent file", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		args := []string{"run", "-d", TestImg0.ref, "sleep", "2"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "kill", cid)
			exekit.New(t, exekit.WithLax).Exe("docker", "rm", cid)
		})
		exekit.New(t).Exe("docker", "inspect", cid)

		// --- When ---
		have, err := dkr.CtrFile(cid, "/not-existing")

		// --- Then ---
		wMsg := "[getting file] docker command error:\n" +
			"   cmd: docker cp %s:/not-existing -\n   " +
			"err: exit status 1\n" +
			"  eout: Error response from daemon: " +
			"Could not find the file /not-existing in container %s"
		wMsg = fmt.Sprintf(wMsg, cid, cid)
		assert.ErrorEqual(t, wMsg, err)
		assert.Equal(t, "", have)
	})
}

func Test_Docker_NetLs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		dkr := New()
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

		// --- When ---
		have, err := dkr.NetLs()

		// --- Then ---
		assert.NoError(t, err)

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
		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		dkr := New(WithEnv(env))

		// --- When ---
		nts, err := dkr.NetLs()

		// --- Then ---
		wMsg := "" +
			"[getting networks] docker command error:\n" +
			"   cmd: docker network ls --no-trunc --format={{.ID}}\n" +
			"   err: exit status 1\n" +
			"  eout: Cannot connect to the Docker daemon at %s. " +
			"Is the docker daemon running?"
		wMsg = fmt.Sprintf(wMsg, host)
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, nts)
	})
}

func Test_Docker_NetRm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		dkr := New()
		name := RandNet()
		args := []string{"network", "create", name}
		nid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() {
			exekit.New(t, exekit.WithLax).Exe("docker", "network", "rm", nid)
		})
		exekit.New(t).Exe("docker", "inspect", nid)

		// --- When ---
		err := dkr.NetRm(name)

		// --- Then ---
		assert.NoError(t, err)
		exekit.New(t, exekit.WithExitCode(1)).Exe("docker", "inspect", nid)
	})

	t.Run("non-existent", func(t *testing.T) {
		// --- Given ---
		dkr := New()

		// --- When ---
		err := dkr.NetRm("00000000-0000-0000-0000-000000000000")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - cannot connect to docker host", func(t *testing.T) {
		// --- Given ---
		port := must.Value(netkit.GetFreePort())
		host := fmt.Sprintf("tcp://127.0.0.1:%d", port)
		env := append(os.Environ(), "DOCKER_HOST="+host)
		nid := "00000000-0000-0000-0000-000000000000"
		dkr := New(WithEnv(env))

		// --- When ---
		err := dkr.NetRm(nid)

		// --- Then ---
		wMsg := "" +
			"[removing network] docker command error:\n" +
			"   cmd: docker network rm --force %s\n" +
			"   err: exit status 1\n" +
			"  eout:\n" +
			"        Cannot connect to the Docker daemon at %s. " +
			"Is the docker daemon running?\n        exit status 1"
		wMsg = fmt.Sprintf(wMsg, nid, host)
		assert.ErrorEqual(t, wMsg, err)
	})
}
