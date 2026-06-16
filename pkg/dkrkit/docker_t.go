// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"path/filepath"

	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/xdef/pkg/xdef"
)

// DockerT represents a docker command with test helper methods.
type DockerT struct {
	dkr *Docker
	t   tester.T
}

// NewT returns a new [DockerT] instance.
func NewT(t tester.T, opts ...DockerOption) *DockerT {
	t.Helper()
	return &DockerT{t: t, dkr: New(opts...)}
}

// ImgPull pulls the image for the given ref. Fails the test on error.
func (dt *DockerT) ImgPull(ref string) {
	dt.t.Helper()
	if err := dt.dkr.ImgPull(ref); err != nil {
		dt.t.Error(err)
	}
}

// Build builds a test Docker image and returns the image reference and ID.
// Fails the test on error.
//
// If you are going to run the built image, you must stop or kill it before
// trying to remove it, otherwise the cleanup function will only untag the
// image and will not be able to remove it.
func (dt *DockerT) Build(opts ...BuildOption) (string, string) {
	dt.t.Helper()
	name := dt.t.Name()
	def := []BuildOption{
		WithBuildArg(xdef.EnvImgAuthors, name),
	}
	opts = append(def, opts...)
	ref, iid, err := dt.dkr.Build(opts...)
	dt.t.Cleanup(func() { dt.t.Helper(); dt.ImgRm(iid) })
	if err != nil {
		dt.t.Error(err)
	}
	return ref, iid
}

// BuildTestImg builds a test image from an embedded Dockerfile. Fails the
// test on error.
func (dt *DockerT) BuildTestImg() (string, string) {
	dt.t.Helper()
	return dt.Build(dt.dkr.testImageBuildOptions()...)
}

// ImgLs lists all docker images. Fails the test and returns nil on error.
func (dt *DockerT) ImgLs(opts ...ImgListOption) Images {
	dt.t.Helper()
	ims, err := dt.dkr.ImgLs(opts...)
	if err != nil {
		dt.t.Error(err)
	}
	return ims
}

// Labels returns the labels for the given ref. The ref may be one of:
// image ID, image reference, or container ID. Fails the test and returns nil
// on error.
func (dt *DockerT) Labels(ref string) map[string]string {
	dt.t.Helper()
	lbs, err := dt.dkr.Labels(ref)
	if err != nil {
		dt.t.Error(err)
	}
	return lbs
}

// Label returns the value of the given label for ref. The ref may be one of:
// image ID, image reference, or container ID. Fails the test and returns an
// empty string if the ref or label does not exist.
func (dt *DockerT) Label(ref, label string) string {
	dt.t.Helper()
	val, err := dt.dkr.Label(ref, label)
	if err != nil {
		dt.t.Error(err)
	}
	return val
}

// Envs uses "docker inspect" to list environment variables for the given
// ref. The ref may be one of: image ID, image reference, or container ID.
// Fails the test and returns nil on error.
func (dt *DockerT) Envs(ref string) map[string]string {
	dt.t.Helper()
	envs, err := dt.dkr.Envs(ref)
	if err != nil {
		dt.t.Error(notice.From(err, "getting environment variables"))
	}
	return envs
}

// Env uses "docker inspect" to get the value of an environment variable for
// the given ref. The ref may be one of: image ID, image reference, or
// container ID. Fails the test and returns an empty string if the ref or
// variable does not exist.
func (dt *DockerT) Env(ref, name string) string {
	dt.t.Helper()
	val, err := dt.dkr.Env(ref, name)
	if err != nil {
		dt.t.Error(notice.From(err, "getting environment variables"))
	}
	return val
}

// ImgRm removes the image with the given ref, returning true on success.
// Fails the test and returns false on error.
//
// By default, errors are reported via t.Error. Pass
// [WithImgRmIgnoreErrors] to log them via t.Log instead, which lets the
// test continue without a failure while still returning false.
//
// Special cases when the method returns true and does not trigger an error:
//   - ref is an empty string,
//   - ref is not found.
func (dt *DockerT) ImgRm(ref string, opts ...ImgRmOption) bool {
	dt.t.Helper()
	def := DefaultImgRmOptions(opts...)
	if err := dt.dkr.ImgRm(ref, opts...); err != nil {
		if !def.ignoreErrors {
			dt.t.Error(err)
		} else {
			dt.t.Log(err)
		}
		return false
	}
	return true
}

// CtrRun runs a container for the given ref. Fails the test and returns an
// empty string on error.
func (dt *DockerT) CtrRun(ref string, opts ...CtrRunOption) string {
	dt.t.Helper()
	def := DefaultCtrRunOptions(opts...)
	if def.cidCh != nil {
		if def.cidPth == "" {
			opt := WithCtrRunCIDPth(filepath.Join(dt.t.TempDir(), "cid.log"))
			opts = append(opts, opt)
		}
		dt.t.Cleanup(def.cidClose)
	}
	sout, err := dt.dkr.CtrRun(ref, opts...)
	if err != nil {
		dt.t.Error(err)
	}
	return sout
}

// CtrRm force-removes the container with the given ID, returning true on
// success. Fails the test and returns false on error. Does not report an
// error if cid does not exist.
func (dt *DockerT) CtrRm(cid string) bool {
	dt.t.Helper()
	if err := dt.dkr.CtrRm(cid); err != nil {
		dt.t.Error(err)
		return false
	}
	dt.t.Logf("CtrRm: successfully removed container: %s", cid)
	return true
}

// CtrKill kills the running container with the given ID, returning true on
// success. Fails the test and returns false on error. Does not remove the
// stopped container.
func (dt *DockerT) CtrKill(cid string) bool {
	dt.t.Helper()
	if err := dt.dkr.CtrKill(cid); err != nil {
		dt.t.Error(err)
		return false
	}
	dt.t.Logf("CtrKill: successfully killed container: %s", cid)
	return true
}

// CtrExec executes a command in the running container. Fails the test and
// returns an empty string on error.
func (dt *DockerT) CtrExec(cid string, cmd ...string) string {
	dt.t.Helper()
	sout, err := dt.dkr.CtrExec(cid, cmd...)
	if err != nil {
		dt.t.Error(err)
	}
	return sout
}

// CtrPs returns a list of all containers. Fails the test and returns an
// empty collection on error.
func (dt *DockerT) CtrPs() Containers {
	dt.t.Helper()
	cts, err := dt.dkr.CtrPs()
	if err != nil {
		dt.t.Error(err)
	}
	return cts
}

// CtrFile returns the content of file pth from the running container cid.
// Fails the test and returns an empty string on error.
func (dt *DockerT) CtrFile(cid, pth string) string {
	dt.t.Helper()
	sout, err := dt.dkr.CtrFile(cid, pth)
	if err != nil {
		dt.t.Error(err)
	}
	return sout
}

// NetLs lists all docker networks. Fails the test and returns nil on error.
func (dt *DockerT) NetLs() Networks {
	dt.t.Helper()
	nts, err := dt.dkr.NetLs()
	if err != nil {
		dt.t.Error(err)
	}
	return nts
}

// NetRm removes the Docker network by ID or name, returning true on success.
// Fails the test and returns false on error. Does not report an error if the
// network does not exist.
func (dt *DockerT) NetRm(ref string) bool {
	dt.t.Helper()
	if err := dt.dkr.NetRm(ref); err != nil {
		dt.t.Error(err)
		return false
	}
	return true
}
