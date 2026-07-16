// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package exekit provides test helpers for running external commands and
// asserting their exit code, standard output, and standard error.
//
// The central type is [Exe], which wraps [os/exec] with a configurable
// timeout, environment, and exit-code expectation. Construct one with [New]
// and run commands with [Exe.Exe], [Exe.ExeStdout], or [Exe.ExeStderr].
//
// Coverage helpers ([WithDetCov], [WithDevOsCov], [MaybeAddGoCovDir],
// [IsWithCoverage]) assist with propagating Go coverage instrumentation
// into sub-processes launched during tests.
package exekit
