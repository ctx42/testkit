// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/ctx42/testing/pkg/notice"
)

// cmdOption is a functional option for [dockerCmd].
type cmdOption func(cmd *cmdOptions)

// cmdOptions represents [dockerCmd] options.
type cmdOptions struct {
	wd  string    // Working directory, by default current working directory.
	env []string  // Environment, by default current environment.
	sin io.Reader // Standard input, by default [os.Stdin].

	// When non-nil, stdout is written here and the returned stdout string is
	// always "".
	sout io.Writer
}

// withCmdWD is dockerCmd option setting the current working directory.
func withCmdWD(pth string) cmdOption {
	return func(opts *cmdOptions) { opts.wd = pth }
}

// withCmdStdin is dockerCmd option setting custom standard input.
func withCmdStdin(sin io.Reader) cmdOption {
	return func(opts *cmdOptions) { opts.sin = sin }
}

// withCmdStdout is dockerCmd option redirecting stdout to w instead of
// buffering it as a string. Use for binary output (e.g. docker cp tar stream)
// where TrimSpace must not touch the data.
func withCmdStdout(w io.Writer) cmdOption {
	return func(opts *cmdOptions) { opts.sout = w }
}

// dockerCmd runs docker command with arguments. Returns messages written by
// docker to standard output, standard error.
func dockerCmd(
	ctx context.Context,
	env []string,
	args []string,
	opts ...cmdOption,
) (string, string, error) {

	def := &cmdOptions{
		env: env,
		sin: os.Stdin,
	}
	for _, opt := range opts {
		opt(def)
	}

	soutBuf, eout := &bytes.Buffer{}, &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdin = def.sin
	cmd.Stdout = soutBuf
	if def.sout != nil {
		cmd.Stdout = def.sout
	}
	cmd.Stderr = eout
	cmd.Dir = def.wd
	cmd.Env = def.env

	err := cmd.Run()
	soutS := ""
	if def.sout == nil {
		soutS = strings.TrimSpace(soutBuf.String())
	}
	eoutS := strings.TrimSpace(eout.String())
	if err != nil {
		msg := notice.New("docker command error").
			Append("cmd", "%s", "docker "+strings.Join(args, " ")).
			Append("err", "%s", err).
			Wrap(err)

		if soutS != "" {
			msg = msg.Append("sout", "%s", soutS)
		}
		if eoutS != "" {
			msg = msg.Append("eout", "%s", eoutS)
		}

		return soutS, eoutS, msg
	}
	return soutS, eoutS, nil
}
