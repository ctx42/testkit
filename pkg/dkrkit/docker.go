// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ctx42/testing/pkg/check"
	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/xdef/pkg/xdef"

	"github.com/ctx42/testkit/pkg/randkit"
	"github.com/ctx42/testkit/pkg/testkit"
)

// DockerOption is a configuration option for [Docker].
type DockerOption func(dkr *Docker)

// WithEnv sets the environment variables passed to docker commands.
func WithEnv(env []string) DockerOption {
	return func(dkr *Docker) { dkr.env = env }
}

// Docker represents a docker command executor.
type Docker struct {
	env []string
}

// New returns a new [Docker] instance. By default, docker commands run with
// [os.Environ]. Use [WithEnv] to override it.
func New(opts ...DockerOption) *Docker {
	dkr := &Docker{}
	for _, opt := range opts {
		opt(dkr)
	}
	if dkr.env == nil {
		dkr.env = os.Environ()
	}
	return dkr
}

// ImgPull pulls docker image. Returns an error if the pull fails.
func (dkr *Docker) ImgPull(ref string) error {
	ctx := context.Background()
	args := []string{"pull", ref}
	_, _, err := dockerCmd(ctx, dkr.env, args)
	if err != nil {
		return notice.From(err, "pulling image")
	}
	return nil
}

// Build builds test Docker image. Returns image reference and image ID.
// It's up to the caller to remove the image when it's no longer needed.
//
// TODO: refactor to reduce cyclomatic complexity.
//
//nolint:cyclop
func (dkr *Docker) Build(opts ...BuildOption) (string, string, error) {
	def, err := DefaultBuildOptions(opts...)
	if err != nil {
		return "", "", err
	}

	if def.imgName == "" {
		def.imgName = RandName()
	}
	if def.imgTag == "" {
		def.imgTag = RandTag()
	}
	if !def.noCache && envGet(dkr.env, envBldNoCache) != "" {
		def.noCache = true
	}

	ref := fmt.Sprintf("%s:%s", def.imgName, def.imgTag)
	if def.iidPth == "" {
		def.iidPth = filepath.Join(os.TempDir(), randkit.Str()+".iid.log")
		defer func() { _ = os.Remove(def.iidPth) }()
	}

	args := []string{
		"build",
		"--rm",
		"-t", ref,
		"--iidfile", def.iidPth,
	}
	if envGet(dkr.env, "SSH_AUTH_SOCK") != "" {
		args = append(args, "--ssh=default")
	}
	for name, value := range def.labels {
		args = append(args, "--label", name+"="+value)
	}
	for name, value := range def.args {
		args = append(args, "--build-arg", name+"="+value)
	}
	if def.noCache {
		args = append(args, "--no-cache")
	}

	var dir string
	var sin io.Reader
	if def.bldRdr != nil {
		sin = def.bldRdr
		args = append(args, "-")
		if rw, ok := def.bldRdr.(io.ReadCloser); ok {
			defer func() { _ = rw.Close() }()
		}
	} else {
		dir = filepath.Dir(def.bldPth)
		dockerfile := filepath.Base(def.bldPth)
		if dockerfile == "." {
			dockerfile = "Dockerfile"
		}
		args = append(args, "--file", dockerfile, ".")
	}

	sinOpt := withCmdStdin(sin)
	wdOpt := withCmdWD(dir)
	envOpt := append(slices.Clone(dkr.env), "DOCKER_BUILDKIT=1")

	if def.dryRun != nil {
		out := "DOCKER_BUILDKIT=1 docker " + strings.Join(args, " ")
		_, _ = def.dryRun.Write([]byte(out))
		return ref, "", nil
	}

	ctx := context.Background()
	_, _, err = dockerCmd(ctx, envOpt, args, wdOpt, sinOpt)
	if err != nil {
		return "", "", notice.From(err, "building image")
	}

	iid, err := readIIDFile(ref, def.iidPth)
	if err != nil {
		return ref, "", err
	}
	return ref, iid, nil
}

// readIIDFile reads the image ID written by docker build's --iidfile flag.
// Returns an error if the file is missing or if its content contains no hex
// ID (e.g. a bare "sha256:" with no digest).
func readIIDFile(ref, pth string) (string, error) {
	content, err := os.ReadFile(pth) //nolint:gosec
	if err != nil {
		return "", err
	}
	iid := StripHashName(string(content))
	if iid == "" {
		msg := notice.New("building image: iid file has no image ID").
			Append("ref", "%s", ref).
			Append("path", "%s", pth).
			Append("content", "%q", string(content))
		return "", msg
	}
	return iid, nil
}

// testImageBuildOptions returns [BuildOption] slice for building the standard
// test image. It reads [xdef.EnvImgCreated] and [xdef.EnvImgRefName] from the
// environment; all other values are fixed constants.
func (dkr *Docker) testImageBuildOptions() []BuildOption {
	args := map[string]string{
		xdef.EnvImgCreated:  xdef.Created(dkr.env),
		xdef.EnvImgSrc:      "repo",
		xdef.EnvImgRev:      "12345678",
		xdef.EnvImgVer:      "v1.2.3",
		xdef.EnvImgRefName:  xdef.ImgRefName(dkr.env),
		xdef.EnvImgBaseName: TestImageBaseRef,
	}

	bldOpt := WithBuildRdr(bytes.NewReader(exBld))
	argOpt := WithBuildArgs(args)
	return []BuildOption{bldOpt, argOpt}
}

// BuildTestImg builds a test image based on an embedded Dockerfile. Returns
// image reference and image ID.
func (dkr *Docker) BuildTestImg() (string, string, error) {
	ops := dkr.testImageBuildOptions()
	return dkr.Build(ops...)
}

// ImgLs lists all docker images.
func (dkr *Docker) ImgLs(opts ...ImgListOption) (Images, error) {
	def := ImgListOptions{}
	for _, opt := range opts {
		opt(&def)
	}

	ctx := context.Background()
	args := []string{"image", "ls", "--no-trunc", "--format={{json .}}"}
	for _, filter := range def.filters {
		args = append(args, "--filter", filter)
	}
	sout, _, err := dockerCmd(ctx, dkr.env, args)
	if err != nil {
		return nil, notice.From(err, "getting images")
	}

	var ims []*Image
	dec := json.NewDecoder(strings.NewReader(sout))
	for dec.More() {
		var img *Image
		if err = dec.Decode(&img); err != nil {
			return nil, notice.From(err, "getting images")
		}
		ims = append(ims, img)
	}
	return ims, nil
}

// Labels returns the labels for the given ref as a map. The ref may be one of:
//   - Image ID
//   - Image reference
//   - Container ID
func (dkr *Docker) Labels(ref string) (map[string]string, error) {
	ctx := context.Background()
	return getLabels(ctx, dkr.env, ref)
}

// Label returns a label value for the given ref and label name. The ref may be
// one of the:
//   - Image ID
//   - Image reference
//   - Container ID
//
// Returns an error if the ref or label does not exist.
func (dkr *Docker) Label(ref, label string) (string, error) {
	lbs, err := dkr.Labels(ref)
	if err != nil {
		return "", notice.From(err, "getting label")
	}
	if val, exist := lbs[label]; exist {
		return val, nil
	}
	msg := notice.New("[getting label] expected label to exist").
		Append("ref", "%s", ShortID(ref)).
		Want("%q", label).
		Append("labels", "%s", formatMapKeys(lbs))
	return "", msg
}

// Envs returns the list of environment variables for the given ref. The
// ref may be one of the:
//   - Image ID
//   - Image reference
//   - Container ID
func (dkr *Docker) Envs(ref string) (map[string]string, error) {
	ctx := context.Background()
	return getEnvs(ctx, dkr.env, ref)
}

// Env returns the environment variable value for the given ref and name. The
// ref may be one of the:
//   - Image ID
//   - Image reference
//   - Container ID
//
// Returns an error if the ref or environment variable name does not exist.
func (dkr *Docker) Env(ref, name string) (string, error) {
	envs, err := dkr.Envs(ref)
	if err != nil {
		return "", notice.From(err, "getting environment variable")
	}
	if val, exist := envs[name]; exist {
		return val, nil
	}
	mHeader := "[getting environment variable] expected variable to exist"
	msg := notice.New(mHeader).
		Append("ref", "%s", ref).
		Want("%q", name).
		Append("env", "%s", formatMapKeys(envs))
	return "", msg
}

// ImgRm removes the image with the given ref. If the image cannot be removed
// because it's used by a running container, it will retry. Returns nil on
// success, error on failure.
//
// Special cases when the method returns nil:
//   - The ref is an empty string,
//   - The ref is not found.
//
// TODO: refactor to reduce cognitive complexity.
//
//nolint:gocognit,cyclop
func (dkr *Docker) ImgRm(ref string, opts ...ImgRmOption) error {
	if ref == "" {
		return nil
	}
	def := DefaultImgRmOptions(opts...)

	tries := 0
	iid := ref
	force := false
	ctx := context.Background()
	for {
		tries++
		args := []string{"image", "rm"}
		if force {
			args = append(args, "--force")
		}
		args = append(args, ref)

		_, eout, err := dockerCmd(ctx, dkr.env, args)
		if err == nil {
			if force && iid != ref {
				args = []string{"image", "rm", iid}
				_, eout, err = dockerCmd(ctx, dkr.env, args)
				if err != nil && !strings.Contains(eout, "No such image") {
					s := classifyImgRmErr(eout)
					if s.action == imgRmForce {
						args = []string{"image", "rm", "--force", iid}
						_, _, _ = dockerCmd(ctx, dkr.env, args)
					}
				}
			}
			return nil
		}
		if strings.Contains(eout, "No such image") {
			return nil
		}
		if tries >= def.tries {
			return notice.From(err, "removing image")
		}

		s := classifyImgRmErr(eout)
		switch s.action {
		case imgRmWait:
			time.Sleep(def.sleep)
		case imgRmForce:
			if s.iid != "" {
				iid = s.iid
			}
			if s.force {
				force = true
			}
		default:
			return notice.From(err, "removing image")
		}
	}
}

// CtrRun runs docker container based on reference. Returns stdout and an error.
//
// When [WithCtrRunCID] is passed, a goroutine polls for the cidfile every
// 50 ms (configurable) for up to 1 s and sends the container ID on the channel
// once the file appears. On timeout or any other error the channel is closed
// without sending; the caller can detect this with:
//
//	cid, ok := <-cidCh
//
// The channel is always closed exactly once. When the container runs detached,
// it is the caller's responsibility to remove it after it stops.
func (dkr *Docker) CtrRun(ref string, opts ...CtrRunOption) (string, error) {
	def := DefaultCtrRunOptions(opts...)

	if def.cidCh != nil {
		tmpDir := ""
		if def.cidPth == "" {
			dir, err := os.MkdirTemp("", "dkrkit-")
			if err != nil {
				return "", err
			}
			def.cidPth = filepath.Join(dir, "cid.log")
			tmpDir = dir
		}
		// Start a goroutine waiting (with configured timeout) for the CID file
		// to be created and send it on the channel. On error the channel is
		// closed without sending; the caller can distinguish via ok in
		// cid, ok := <-cidCh.
		cidPth := def.cidPth
		go func() {
			if tmpDir != "" {
				defer func() { _ = os.RemoveAll(tmpDir) }()
			}
			thr := check.WithWaitThrottle(def.cidPoll)
			cid, err := testkit.Wait4File("1s", cidPth, thr)
			if err == nil {
				def.cidCh <- cid
			}
			def.cidClose()
		}()
	}

	args := []string{"run"}
	if def.remove {
		args = append(args, "--rm")
	}
	if def.cidPth != "" {
		args = append(args, "--cidfile", def.cidPth)
	}
	if def.detach {
		args = append(args, "--detach")
	}
	for name, value := range def.labels {
		label := name
		if value != "" {
			label += "=" + value
		}
		args = append(args, "--label", label)
	}
	args = append(args, ref)
	args = append(args, def.args...)

	ctx := context.Background()
	sout, _, err := dockerCmd(ctx, dkr.env, args)
	if err != nil {
		return "", notice.From(err, "running image")
	}
	return sout, nil
}

// CtrRm removes the container with the given ID.
func (dkr *Docker) CtrRm(cid string) error {
	ctx := context.Background()
	args := []string{"rm", "--force", cid}
	_, _, err := dockerCmd(ctx, dkr.env, args)
	if err != nil {
		return notice.From(err, "removing container")
	}
	return nil
}

// CtrKill kills a running container with the given ID.
func (dkr *Docker) CtrKill(cid string) error {
	ctx := context.Background()
	args := []string{"kill", "-s", "9", cid}
	_, _, err := dockerCmd(ctx, dkr.env, args)
	if err != nil {
		return notice.From(err, "killing container")
	}
	return nil
}

// CtrExec executes a command in the running container. Returns stdout and
// error.
func (dkr *Docker) CtrExec(cid string, cmd ...string) (string, error) {
	ctx := context.Background()
	args := []string{"exec", cid}
	args = append(args, cmd...)
	sout, _, err := dockerCmd(ctx, dkr.env, args)
	if err != nil {
		return "", notice.From(err, "container exec")
	}
	return sout, nil
}

// CtrPs returns a list of all containers. Container details (Env, Labels,
// Networks, State) are populated lazily when FindByImage or FindByID is
// called.
func (dkr *Docker) CtrPs() (Containers, error) {
	ctx := context.Background()
	args := []string{
		"ps",
		"--all",
		"--no-trunc",
		"--format", "{{.ID}} {{.Image}}",
	}
	sout, _, err := dockerCmd(ctx, dkr.env, args)
	if err != nil {
		return Containers{}, notice.From(err, "listing containers")
	}
	var ctrs []*Container
	for line := range strings.SplitSeq(sout, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		ctrs = append(ctrs, &Container{ID: parts[0], Image: parts[1]})
	}
	return Containers{env: dkr.env, ctrs: ctrs}, nil
}

// CtrFile returns the content of a file from a running container.
func (dkr *Docker) CtrFile(cid, pth string) (string, error) {
	ctx := context.Background()
	tarBuf := &bytes.Buffer{}
	args := []string{"cp", cid + ":" + pth, "-"}
	_, _, err := dockerCmd(
		ctx, dkr.env, args, withCmdStdout(tarBuf))
	if err != nil {
		return "", notice.From(err, "getting file")
	}
	name := filepath.Base(pth)

	rdr := tar.NewReader(tarBuf)
	for {
		var header *tar.Header
		if header, err = rdr.Next(); errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			msg := notice.New("getting file tar error").
				Append("file", "%s", pth).
				Append("tar error", "%s", err.Error())
			return "", msg
		}

		if header.Typeflag == tar.TypeReg && header.Name == name {
			buf := &bytes.Buffer{}
			if _, err = io.Copy(buf, rdr); err != nil { //nolint:gosec
				msg := notice.New("getting file reader error").
					Append("file", "%s", pth).
					Append("tar error", "%s", err.Error())
				return "", msg
			}
			return buf.String(), nil
		}
	}

	msg := notice.New("expected file to exist in the container").
		Append("ref", "%q", cid).
		Append("file", "%s", pth)
	return "", msg
}

// NetLs lists all docker networks.
func (dkr *Docker) NetLs() (Networks, error) {
	ctx := context.Background()
	args := []string{"network", "ls", "--no-trunc", "--format={{.ID}}"}
	sout, _, err := dockerCmd(ctx, dkr.env, args)
	if err != nil {
		return nil, notice.From(err, "getting networks")
	}
	ids := strings.Split(sout, "\n")
	if len(ids) == 1 && ids[0] == "" {
		return Networks{}, nil
	}

	args = []string{"network", "inspect", "--format=json"}
	args = append(args, ids...)
	sout, _, err = dockerCmd(ctx, dkr.env, args)
	if err != nil {
		return nil, notice.From(err, "getting networks")
	}

	nts := make([]*Network, 0, len(ids))
	if err = json.Unmarshal([]byte(sout), &nts); err != nil {
		return nil, notice.From(err, "getting networks")
	}
	return nts, nil
}

// NetRm removes Docker network by ID or name.
func (dkr *Docker) NetRm(ref string) error {
	ctx := context.Background()
	args := []string{"network", "rm", "--force", ref}
	_, eout, err := dockerCmd(ctx, dkr.env, args)
	if err != nil {
		if strings.Contains(eout, "not found") {
			return nil
		}
		return notice.From(err, "removing network")
	}
	return nil
}
