// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package randkit

import (
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_WithChars(t *testing.T) {
	// --- Given ---
	def := &options{}

	// --- When ---
	WithChars("ab", "cd", "ef")(def)

	// --- Then ---
	want := &options{
		chars:  "abcdef",
		n:      0,
		prefix: "",
		suffix: "",
	}
	assert.Equal(t, want, def)
}

func Test_WithPrefix(t *testing.T) {
	// --- Given ---
	def := &options{}

	// --- When ---
	WithPrefix("abc")(def)

	// --- Then ---
	want := &options{
		chars:  "",
		n:      0,
		prefix: "abc",
		suffix: "",
	}
	assert.Equal(t, want, def)
}

func Test_WithSuffix(t *testing.T) {
	// --- Given ---
	def := &options{}

	// --- When ---
	WithSuffix("abc")(def)

	// --- Then ---
	want := &options{
		chars:  "",
		n:      0,
		prefix: "",
		suffix: "abc",
	}
	assert.Equal(t, want, def)
}

func Test_WithExt(t *testing.T) {
	// --- Given ---
	def := &options{}

	// --- When ---
	WithExt(".exe")(def)

	// --- Then ---
	want := &options{
		chars:  "",
		n:      0,
		prefix: "",
		suffix: ".exe",
	}
	assert.Equal(t, want, def)
}

func Test_WithLen(t *testing.T) {
	// --- Given ---
	def := &options{}

	// --- When ---
	WithLen(42)(def)

	// --- Then ---
	want := &options{
		chars:  "",
		n:      42,
		prefix: "",
		suffix: "",
	}
	assert.Equal(t, want, def)
}

func Test_seededRand(t *testing.T) {
	t.Run("is deterministic", func(t *testing.T) {
		// --- When ---
		a := seededRand(42)
		b := seededRand(42)

		// --- Then ---
		for i := 0; i < 100; i++ {
			assert.Equal(t, a(52), b(52))
		}
	})

	t.Run("different seeds differ", func(t *testing.T) {
		// --- When ---
		a := seededRand(1)
		b := seededRand(2)

		// --- Then ---
		differ := false
		for i := 0; i < 100; i++ {
			if a(1000) != b(1000) {
				differ = true
				break
			}
		}
		assert.True(t, differ)
	})

	t.Run("panics on zero", func(t *testing.T) {
		assert.PanicContain(t, "n must be > 0", func() { seededRand(1)(0) })
	})

	t.Run("panics on negative", func(t *testing.T) {
		assert.PanicContain(t, "n must be > 0", func() { seededRand(1)(-1) })
	})
}

func Test_seedToBytes(t *testing.T) {
	t.Run("is deterministic", func(t *testing.T) {
		assert.Equal(t, seedToBytes(42), seedToBytes(42))
	})

	t.Run("different seeds differ", func(t *testing.T) {
		assert.NotEqual(t, seedToBytes(1), seedToBytes(2))
	})

	t.Run("encodes seed as little-endian in first 8 bytes", func(t *testing.T) {
		// --- When ---
		b := seedToBytes(0x0102030405060708)

		// --- Then ---
		assert.Equal(t, byte(0x08), b[0])
		assert.Equal(t, byte(0x07), b[1])
		assert.Equal(t, byte(0x06), b[2])
		assert.Equal(t, byte(0x05), b[3])
		assert.Equal(t, byte(0x04), b[4])
		assert.Equal(t, byte(0x03), b[5])
		assert.Equal(t, byte(0x02), b[6])
		assert.Equal(t, byte(0x01), b[7])
	})
}

func Test_WithSeed(t *testing.T) {
	// --- Given ---
	def := &options{}

	// --- When ---
	WithSeed(1)(def)

	// --- Then ---
	assert.NotNil(t, def.rng)
}

func Test_Str(t *testing.T) {
	t.Run("does not repeat", func(t *testing.T) {
		// --- Given ---
		history := make(map[string]struct{})

		// --- Then ---
		for i := 0; i < 1000; i++ {
			str := Str()
			if _, ok := history[str]; ok {
				t.Error("did not expect string to repeat")
			}
			history[str] = struct{}{}
			want := 10
			have := len(str)
			if want != have {
				t.Errorf("expected len(s)=%d got %d", want, have)
			}
		}
	})

	t.Run("set prefix", func(t *testing.T) {
		// --- When ---
		str := Str(WithPrefix("pref-"))

		// --- Then ---
		assert.True(t, strings.HasPrefix(str, "pref-"))
	})

	t.Run("set suffix", func(t *testing.T) {
		// --- When ---
		str := Str(WithSuffix("-suf"))

		// --- Then ---
		assert.True(t, strings.HasSuffix(str, "-suf"))
	})

	t.Run("set prefix and suffix", func(t *testing.T) {
		// --- When ---
		str := Str(WithPrefix("pref-"), WithSuffix("-suf"))

		// --- Then ---
		assert.True(t, strings.HasPrefix(str, "pref-"))
		assert.True(t, strings.HasSuffix(str, "-suf"))
	})

	t.Run("has only lowercase letters", func(t *testing.T) {
		// --- Then ---
		for i := 0; i < 500; i++ {
			str := Str(WithChars(Lowercase))
			want := strings.ToLower(str)
			have := str
			if want != have {
				t.Errorf("expected only lowercase letters got %s", have)
			}
		}
	})

	t.Run("has only uppercase letters", func(t *testing.T) {
		// --- Then ---
		for i := 0; i < 500; i++ {
			str := Str(WithChars(Uppercase))
			want := strings.ToUpper(str)
			have := str
			if want != have {
				t.Errorf("expected only uppercase letters got %s", have)
			}
		}
	})

	t.Run("seeded is deterministic", func(t *testing.T) {
		assert.Equal(t, "qLKZasgepC", Str(WithSeed(1)))
		assert.Equal(t, "qLKZas", Str(WithSeed(1), WithLen(6)))
		assert.Equal(t, "37790310", Str(WithChars(Digits), WithSeed(1), WithLen(8)))
		assert.Equal(t, "test-qLKZas-end",
			Str(WithPrefix("test-"), WithSuffix("-end"), WithLen(6), WithSeed(1)))
	})
}

func Test_FileName(t *testing.T) {
	t.Run("no dir prefix ext", func(t *testing.T) {
		// --- When ---
		name := FileName("")

		// --- Then ---
		assert.True(t, strings.HasPrefix(name, "file-"))
		assert.True(t, strings.HasSuffix(name, ".txt"))
		assert.Len(t, len("file-")+7+len(".txt"), name)
	})

	t.Run("no dir prefix with ext", func(t *testing.T) {
		// --- When ---
		name := FileName("", WithExt(".my"))

		// --- Then ---
		assert.True(t, strings.HasPrefix(name, "file-"))
		assert.True(t, strings.HasSuffix(name, ".my"))
		assert.Len(t, len("file-")+7+len(".my"), name)
	})

	t.Run("no dir with prefix ext", func(t *testing.T) {
		// --- When ---
		name := FileName("", WithPrefix("prefix-"), WithExt(".my"))

		// --- Then ---
		assert.True(t, strings.HasPrefix(name, "prefix-"))
		assert.True(t, strings.HasSuffix(name, ".my"))
		assert.Len(t, len("prefix-")+7+len(".my"), name)
	})

	t.Run("with dir prefix ext", func(t *testing.T) {
		// --- When ---
		name := FileName("/dir", WithPrefix("prefix-"), WithExt(".my"))

		// --- Then ---
		assert.True(t, strings.HasPrefix(name, "/dir/prefix-"))
		assert.True(t, strings.HasSuffix(name, ".my"))
		assert.Len(t, len("/dir/prefix-")+7+len(".my"), name)
	})

	t.Run("has only lowercase letters", func(t *testing.T) {
		// --- Then ---
		for i := 0; i < 500; i++ {
			name := FileName(
				"/dir",
				WithPrefix(""),
				WithSuffix(""),
				WithChars(Lowercase),
			)
			want := strings.ToLower(name)
			have := name
			if want != have {
				t.Errorf("expected only lowercase letters got %s", have)
			}
		}
	})

	t.Run("has only uppercase letters", func(t *testing.T) {
		// --- Then ---
		for i := 0; i < 500; i++ {
			name := FileName(
				"/DIR",
				WithPrefix(""),
				WithSuffix(""),
				WithChars(Uppercase),
			)
			want := strings.ToUpper(name)
			have := name
			if want != have {
				t.Errorf("expected only uppercase letters got %s", have)
			}
		}
	})

	t.Run("seeded is deterministic", func(t *testing.T) {
		assert.Equal(t, "file-qLKZasg.txt", FileName("", WithSeed(1)))
		assert.Equal(t, "/tmp/file-qLKZasg.txt", FileName("/tmp", WithSeed(1)))
		assert.Equal(t, "/tmp/file-qLKZasg.json",
			FileName("/tmp", WithExt(".json"), WithSeed(1)))
	})
}

func Test_Int(t *testing.T) {
	t.Run("range", func(t *testing.T) {
		const count = 100_000
		const maxVal = 999
		for i := 0; i < count; i++ {
			num := Int(maxVal)
			if num < 1 || num > maxVal {
				t.Errorf("expected num [1, %d]: %d", maxVal, num)
			}
		}
	})

	t.Run("seeded is deterministic", func(t *testing.T) {
		assert.Equal(t, 32, Int(100, WithSeed(1)))
	})
}

func Test_Password(t *testing.T) {
	t.Run("length", func(t *testing.T) {
		const count = 100_000
		for i := 0; i < count; i++ {
			pass := Password(16)
			if len(pass) != 16 {
				t.Errorf("expected password 16 characters long got: %s", pass)
			}
		}
	})

	t.Run("seeded is deterministic", func(t *testing.T) {
		assert.Equal(t, "tSR9avhesITXkYun", Password(16, WithSeed(1)))
	})
}
