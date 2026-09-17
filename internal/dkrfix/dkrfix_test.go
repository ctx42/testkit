// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrfix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
)

func Test_Simple(t *testing.T) {
	t.Run("single stage on the base image argument", func(t *testing.T) {
		// --- Given ---
		want := []string{"FROM $C42_BLD_IMG_BASE"}

		// --- When ---
		have := instructions(Simple.Content(), "FROM")

		// --- Then ---
		assert.Equal(t, want, have)
	})

	t.Run("stamps the build metadata labels", func(t *testing.T) {
		// --- When ---
		have := Simple.Content()

		// --- Then ---
		assert.Contain(t, "org.opencontainers.image.created", have)
		assert.Contain(t, "org.opencontainers.image.revision", have)
		assert.Contain(t, "org.opencontainers.image.version", have)
		assert.Contain(t, "org.opencontainers.image.source", have)
	})
}

func Test_Minimal(t *testing.T) {
	t.Run("single stage on the base image argument", func(t *testing.T) {
		// --- Given ---
		want := []string{"FROM $C42_BLD_IMG_BASE"}

		// --- When ---
		have := instructions(Minimal.Content(), "FROM")

		// --- Then ---
		assert.Equal(t, want, have)
	})

	t.Run("carries no build metadata labels", func(t *testing.T) {
		// --- When ---
		have := Minimal.Content()

		// --- Then ---
		assert.NotContain(t, "org.opencontainers.image.", have)
	})
}

func Test_Fixture_Content(t *testing.T) {
	t.Run("zero value is an empty Dockerfile", func(t *testing.T) {
		// --- Given ---
		fix := Fixture{}

		// --- When ---
		have := fix.Content()

		// --- Then ---
		assert.Equal(t, "", have)
	})
}

func Test_Fixture_Write(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		dir := t.TempDir()

		// --- When ---
		have := Minimal.Write(t, dir)

		// --- Then ---
		assert.Equal(t, filepath.Join(dir, FileName), have)
		assert.Equal(t, Minimal.Content(), readFile(have))
	})

	t.Run("overwrites an existing file", func(t *testing.T) {
		// --- Given ---
		dir := t.TempDir()
		Simple.Write(t, dir)

		// --- When ---
		have := Minimal.Write(t, dir)

		// --- Then ---
		assert.Equal(t, Minimal.Content(), readFile(have))
	})
}

func Test_Targets(t *testing.T) {
	t.Run("defines three chained targets", func(t *testing.T) {
		// --- Given ---
		want := []string{
			"FROM $C42_BLD_IMG_BASE AS first",
			"FROM first AS second",
			"FROM second AS third",
		}

		// --- When ---
		have := instructions(Targets.Content(), "FROM")

		// --- Then ---
		assert.Equal(t, want, have)
	})

	t.Run("every target stamps its own name", func(t *testing.T) {
		// --- Given ---
		want := []string{
			`ENV C42_BLD_IMG_TARGET="second"`,
			`ENV C42_BLD_IMG_TARGET="third"`,
		}

		// --- When ---
		have := instructions(Targets.Content(), "ENV")

		// --- Then ---
		assert.Equal(t, want, have[1:])

		want = []string{
			`LABEL com.ctx42.image.target="second"`,
			`LABEL com.ctx42.image.target="third"`,
		}
		have = instructions(Targets.Content(), "LABEL")
		assert.Equal(t, want, have[1:])

		lab := `      com.ctx42.image.target="first"`
		assert.Contain(t, lab, Targets.Content())
	})

	t.Run("every target echoes its own name", func(t *testing.T) {
		// --- Given ---
		want := []string{
			`ENTRYPOINT ["echo", "first"]`,
			`ENTRYPOINT ["echo", "second"]`,
			`ENTRYPOINT ["echo", "third"]`,
		}

		// --- When ---
		have := instructions(Targets.Content(), "ENTRYPOINT")

		// --- Then ---
		assert.Equal(t, want, have)
	})
}

func Test_TargetsNEP(t *testing.T) {
	t.Run("defines three chained targets", func(t *testing.T) {
		// --- Given ---
		want := []string{
			"FROM $C42_BLD_IMG_BASE AS first",
			"FROM first AS second",
			"FROM second AS third",
		}

		// --- When ---
		have := instructions(TargetsNEP.Content(), "FROM")

		// --- Then ---
		assert.Equal(t, want, have)
	})

	t.Run("every target stamps its own name", func(t *testing.T) {
		// --- Given ---
		want := []string{
			`ENV C42_BLD_IMG_TARGET="second"`,
			`ENV C42_BLD_IMG_TARGET="third"`,
		}

		// --- When ---
		have := instructions(TargetsNEP.Content(), "ENV")

		// --- Then ---
		assert.Equal(t, want, have[1:])

		want = []string{
			`LABEL com.ctx42.image.target="second"`,
			`LABEL com.ctx42.image.target="third"`,
		}
		have = instructions(TargetsNEP.Content(), "LABEL")
		assert.Equal(t, want, have[1:])

		lab := `      com.ctx42.image.target="first"`
		assert.Contain(t, lab, TargetsNEP.Content())
	})

	t.Run("no target sets an entrypoint or a command", func(t *testing.T) {
		// --- When ---
		have := instructions(TargetsNEP.Content(), "ENTRYPOINT", "CMD")

		// --- Then ---
		assert.Empty(t, have)
	})
}

// readFile returns the content of the file at the path.
func readFile(pth string) string {
	return string(must.Value(os.ReadFile(pth)))
}

// instructions returns the lines of the Dockerfile starting with any of the
// given instruction names. Comments mentioning a name are not matched.
func instructions(dockerfile string, names ...string) []string {
	var found []string
	for _, line := range strings.Split(dockerfile, "\n") {
		for _, name := range names {
			if strings.HasPrefix(line, name+" ") {
				found = append(found, line)
			}
		}
	}
	return found
}
