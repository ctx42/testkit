// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"errors"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/ctx42/testkit/pkg/randkit"
)

// BuildOption represents option for [Docker.Build] and [DockerT.Build].
type BuildOption func(*BuildOptions)

// BuildOptions represents docker image build options.
type BuildOptions struct {
	imgName string            // Image name (random when not set).
	imgTag  string            // Image tag (random when not set).
	labels  map[string]string // Container labels.
	args    map[string]string // Build arguments.
	bldRdr  io.Reader         // Reader for Dockerfile.
	bldPth  string            // Path to Dockerfile.
	iidPth  string            // Path to an image ID file written by --iidfile.
	noCache bool              // Do not use cache.
	dryRun  io.Writer         // Dry run a command and print it to the writer.
}

// DefaultBuildOptions returns customizable default build options.
func DefaultBuildOptions(ops ...BuildOption) (*BuildOptions, error) {
	opts := &BuildOptions{}
	for _, op := range ops {
		op(opts)
	}
	if opts.bldPth != "" && opts.bldRdr != nil {
		msg := "WithBuildPth and WithBuildRdr are mutually exclusive"
		return nil, errors.New(msg)
	}
	return opts, nil
}

// normalize fills in the options which depend on the environment env and
// have not been set explicitly. The image ID file is created in the tmpDir
// directory unless its path was set explicitly. Returns the image reference
// and a function removing the image ID file when normalize generated its
// path. The caller must always call the returned function.
func (def *BuildOptions) normalize(
	env []string,
	tmpDir string,
) (string, func()) {

	if def.imgName == "" {
		def.imgName = RandName()
	}
	if def.imgTag == "" {
		def.imgTag = RandTag()
	}
	if !def.noCache && envGet(env, envBldNoCache) != "" {
		def.noCache = true
	}

	cleanup := func() {}
	if def.iidPth == "" {
		def.iidPth = filepath.Join(tmpDir, randkit.Str()+".iid.log")
		cleanup = func() { _ = os.Remove(def.iidPth) }
	}
	return def.imgName + ":" + def.imgTag, cleanup
}

// cmdArgs returns the docker build command arguments building the ref image.
// The env is examined for the SSH agent socket. The arguments pointing docker
// at the Dockerfile are not included, see source.
func (def *BuildOptions) cmdArgs(ref string, env []string) []string {
	args := []string{
		"build",
		"--rm",
		"-t", ref,
		"--iidfile", def.iidPth,
	}
	if envGet(env, "SSH_AUTH_SOCK") != "" {
		args = append(args, "--ssh=default")
	}
	for _, name := range slices.Sorted(maps.Keys(def.labels)) {
		args = append(args, "--label", name+"="+def.labels[name])
	}
	for _, name := range slices.Sorted(maps.Keys(def.args)) {
		args = append(args, "--build-arg", name+"="+def.args[name])
	}
	if def.noCache {
		args = append(args, "--no-cache")
	}
	return args
}

// source returns the build context directory, the standard input for the
// docker build command, and the command arguments pointing docker at the
// Dockerfile.
func (def *BuildOptions) source() (string, io.Reader, []string) {
	if def.bldRdr != nil {
		return "", def.bldRdr, []string{"-"}
	}
	dockerfile := filepath.Base(def.bldPth)
	if dockerfile == "." {
		dockerfile = "Dockerfile"
	}
	dir := filepath.Dir(def.bldPth)
	return dir, nil, []string{"--file", dockerfile, "."}
}

// WithBuildName is an option for setting image name.
func WithBuildName(name string) BuildOption {
	return func(opts *BuildOptions) { opts.imgName = name }
}

// WithBuildTag is an option for setting image tag.
func WithBuildTag(tag string) BuildOption {
	return func(opts *BuildOptions) { opts.imgTag = tag }
}

// WithBuildLabel is an option for setting image label.
func WithBuildLabel(name, value string) BuildOption {
	return func(opts *BuildOptions) {
		if opts.labels == nil {
			opts.labels = make(map[string]string)
		}
		opts.labels[name] = value
	}
}

// WithBuildLabels is an option for setting image labels.
func WithBuildLabels(labels map[string]string) BuildOption {
	return func(opts *BuildOptions) {
		if opts.labels == nil {
			opts.labels = make(map[string]string)
		}
		maps.Copy(opts.labels, labels)
	}
}

// WithBuildArg is an option for setting build argument.
func WithBuildArg(name, value string) BuildOption {
	return func(opts *BuildOptions) {
		if opts.args == nil {
			opts.args = make(map[string]string)
		}
		opts.args[name] = value
	}
}

// WithBuildArgs is an option for setting build arguments.
func WithBuildArgs(args map[string]string) BuildOption {
	return func(opts *BuildOptions) {
		if opts.args == nil {
			opts.args = make(map[string]string)
		}
		maps.Copy(opts.args, args)
	}
}

// WithBuildRdr is an option for setting the reader to the Dockerfile. If the
// reader can be cast to [io.ReadCloser], its Close method is called
// automatically, but the error is ignored.
func WithBuildRdr(r io.Reader) BuildOption {
	return func(opts *BuildOptions) { opts.bldRdr = r }
}

// WithBuildPth is an option for setting the path to Dockerfile.
func WithBuildPth(pth string) BuildOption {
	return func(opts *BuildOptions) { opts.bldPth = pth }
}

// withBuildIIDFile is an internal option for setting the path to
// the file where docker writes the built image ID (--iidfile). When not set,
// BuildImg generates a temporary path and removes the file after the build.
func withBuildIIDFile(pth string) BuildOption {
	return func(opts *BuildOptions) { opts.iidPth = pth }
}

// WithBuildNoCache is an option for disabling the build cache.
func WithBuildNoCache() BuildOption {
	return func(opts *BuildOptions) { opts.noCache = true }
}

// WithBuildDryRun is an option for setting dry run mode.
func WithBuildDryRun(w io.Writer) BuildOption {
	return func(opts *BuildOptions) { opts.dryRun = w }
}
