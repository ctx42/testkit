// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xdef/pkg/xdef"

	"github.com/ctx42/testkit/pkg/exekit"
)

func Test_Ref_tabular(t *testing.T) {
	tt := []struct {
		testN string

		repo string
		name string
		tag  string
		want string
	}{
		{
			"repo, name, tag",
			"example.com/repo",
			"name",
			"tag",
			"example.com/repo/name:tag",
		},
		{
			"repo ending with slash",
			"example.com/repo/",
			"name",
			"tag",
			"example.com/repo/name:tag",
		},
		{"name, tag", "", "name", "tag", "name:tag"},
		{"name", "", "name", "", "name"},
		{"no name, tag", "", "", "tag", ""},
		{"all empty", "", "", "", ""},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			assert.Equal(t, tc.want, Ref(tc.repo, tc.name, tc.tag))
		})
	}
}

func Test_stripHashName(t *testing.T) {
	t.Run("trailing colon returns empty string", func(t *testing.T) {
		assert.Equal(t, "", StripHashName("sha256:"))
		assert.Equal(t, "", StripHashName(":"))
	})
}

func Test_StripHashName_tabular(t *testing.T) {
	tt := []struct {
		testN string

		have string
		want string
	}{
		{
			"with hash",
			"sha256:b3aab1576e98b7f41f01fa",
			"b3aab1576e98b7f41f01fa",
		},
		{"no hash", "b3aab1576e98b7f41f01fa", "b3aab1576e98b7f41f01fa"},
		{"empty", "", ""},
		{"starts with colon", ":abc", "abc"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := StripHashName(tc.have)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_ShortID_tabular(t *testing.T) {
	tt := []struct {
		testN string

		id   string
		want string
	}{
		{
			"id",
			"62321923585b",
			"62321923585b",
		},
		{
			"long id",
			"785e9f61d4598b65c6c86c5f122830f72a57f42c5ed1b0e40b59000a4ecf0925",
			"785e9f61d459",
		},
		{
			"reference",
			"ctx42-tst-img-cc93bcf1a44d:tst-tag-d6986dbe0ec6",
			"ctx42-tst-img-cc93bcf1a44d:tst-tag-d6986dbe0ec6",
		},
		{
			"shorter than 12",
			"5e9f61d4",
			"5e9f61d4",
		},
		{
			"empty",
			"",
			"",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := ShortID(tc.id)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_isHex_tabular(t *testing.T) {
	tt := []struct {
		testN string

		id   string
		want bool
	}{
		{
			"id",
			"785e9f61d459",
			true,
		},
		{
			"reference",
			"ctx42-tst-img-cc93bcf1a44d:tst-tag-d6986dbe0ec6",
			false,
		},
		{
			"empty",
			"",
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := isHex(tc.id)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_randString(t *testing.T) {
	const count = 1000
	seen := make(map[string]struct{}, count)
	for range count {
		s := randString()
		if len(s) != 12 {
			t.Errorf("expected length 12, got %d: %q", len(s), s)
			return
		}
		if !isHex(s) {
			t.Errorf("expected hex string, got: %q", s)
			return
		}
		seen[s] = struct{}{}
	}
	if len(seen) != count {
		t.Errorf("expected %d unique values, got %d (collision)", count, len(seen))
	}
}

func Test_RandName(t *testing.T) {
	const count = 1000
	const prefix = "ctx42-tst-img-"
	for range count {
		name := RandName()
		wantLen := 12 + len(prefix)
		if len(name) != wantLen {
			format := "expected name to be %d characters long got %d: %s"
			t.Errorf(format, wantLen, len(name), name)
			return
		}
		if !strings.HasPrefix(name, prefix) {
			format := "expected name to be prefixed with %q got: %q"
			t.Errorf(format, prefix, name)
			return
		}
	}
}

func Test_RandTag(t *testing.T) {
	const count = 1000
	const prefix = "ctx42-tst-tag-"
	for range count {
		tag := RandTag()
		wantLen := 12 + len(prefix)
		if len(tag) != wantLen {
			format := "expected tag to be %d characters long got %d: %q"
			t.Errorf(format, wantLen, len(tag), tag)
			return
		}
		if !strings.HasPrefix(tag, prefix) {
			format := "expected tag to be prefixed with %q got: %q"
			t.Errorf(format, prefix, tag)
			return
		}
	}
}

func Test_RandRef(t *testing.T) {
	const count = 1000
	const tagPrefix = "ctx42-tst-tag-"
	const imgPrefix = "ctx42-tst-img-"
	for range count {
		ref := RandRef()
		wantLen := len(imgPrefix) + 12 + 1 + len(tagPrefix) + 12
		if len(ref) != wantLen {
			format := "expected ref to be %d characters long got %d: %q"
			t.Errorf(format, wantLen, len(ref), ref)
			return
		}

		elems := strings.Split(ref, ":")
		if len(elems) != 2 {
			format := "expected ref to have %d elements got: %d"
			t.Errorf(format, 2, len(elems))
			return
		}

		if !strings.HasPrefix(elems[0], imgPrefix) {
			format := "expected image name to be prefixed with %q got: %q"
			t.Errorf(format, imgPrefix, elems[0])
			return
		}
		if !strings.HasPrefix(elems[1], tagPrefix) {
			format := "expected image tag to be prefixed with %q got: %q"
			t.Errorf(format, tagPrefix, elems[1])
			return
		}
	}
}

func Test_RandNet(t *testing.T) {
	const count = 1000
	const prefix = "ctx42-tst-net-"
	for range count {
		id := RandNet()
		wantLen := 12 + len(prefix)
		if len(id) != wantLen {
			format := "expected name to be %d characters long got %d: %s"
			t.Errorf(format, wantLen, len(id), id)
			return
		}
		if !strings.HasPrefix(id, prefix) {
			format := "expected name to be prefixed with %q got: %q"
			t.Errorf(format, prefix, id)
			return
		}
	}
}

func Test_docker_ToBuildArgs(t *testing.T) {
	// --- Given ---
	ags := map[string]string{
		"ARG1": "VALUE1",
		"ARG2": "VALUE2",
	}

	// --- When ---
	args := ToBuildArgs(ags)

	// --- Then ---
	assert.Len(t, 2, ags)
	assert.HasKey(t, "ARG1", args)
	assert.HasKey(t, "ARG2", args)
	assert.Equal(t, "VALUE1", *args["ARG1"])
	assert.Equal(t, "VALUE2", *args["ARG2"])
}

func Test_getLabels(t *testing.T) {
	t.Run("success by iid", func(t *testing.T) {
		// --- When ---
		have, err := getLabels(t.Context(), os.Environ(), TestImg0.iid)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			xdef.LabImgCreated:  "2000-01-02T03:04:05Z",
			xdef.LabImgTitle:    "Image0",
			xdef.LabImgBaseName: TestImageBaseRef,
			labTestEmpty:        "",
		}
		assert.Equal(t, want, have)
	})

	t.Run("success by ref", func(t *testing.T) {
		// --- When ---
		have, err := getLabels(t.Context(), os.Environ(), TestImg1.ref)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			xdef.LabImgCreated:  "2000-01-02T03:04:05Z",
			xdef.LabImgTitle:    "Image1",
			xdef.LabImgBaseName: TestImageBaseRef,
			labTestEmpty:        "",
		}
		assert.Equal(t, want, have)
	})

	t.Run("error - non-existent reference", func(t *testing.T) {
		// --- Given ---
		ref := RandRef()

		// --- When ---
		have, err := getLabels(t.Context(), os.Environ(), ref)

		// --- Then ---
		wMsg := "" +
			"[getting labels] docker command error:\n" +
			"   cmd: docker inspect --format {{ json .Config.Labels }} %s\n" +
			"   err: exit status 1\n" +
			"  eout: error: no such object: %s"
		wMsg = fmt.Sprintf(wMsg, ref, ref)
		assert.ErrorEqual(t, wMsg, err)
		assert.ExitCode(t, 1, err)
		assert.Nil(t, have)
	})

	t.Run("error - non-existent ID", func(t *testing.T) {
		// --- Given ---
		iid := "000000000000"

		// --- When ---
		have, err := getLabels(t.Context(), os.Environ(), iid)

		// --- Then ---
		wMsg := "" +
			"[getting labels] docker command error:\n" +
			"   cmd: docker inspect --format {{ json .Config.Labels }} %s\n" +
			"   err: exit status 1\n" +
			"  eout: error: no such object: %s"
		wMsg = fmt.Sprintf(wMsg, iid, iid)
		assert.ErrorEqual(t, wMsg, err)
		assert.ExitCode(t, 1, err)
		assert.Nil(t, have)
	})
}

func Test_getEnvs(t *testing.T) {
	t.Run("success by iid", func(t *testing.T) {
		// --- When ---
		have, err := getEnvs(t.Context(), os.Environ(), TestImg0.iid)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			xdef.EnvImgCreated: "2000-01-02T03:04:05Z",
			xdef.EnvImgTitle:   "Image0",
			envTestEmpty:       "",
		}
		assert.MapSubset(t, want, have)
	})

	t.Run("success by ref", func(t *testing.T) {
		// --- When ---
		have, err := getEnvs(t.Context(), os.Environ(), TestImg1.ref)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			xdef.EnvImgCreated: "2000-01-02T03:04:05Z",
			xdef.EnvImgTitle:   "Image1",
			envTestEmpty:       "",
		}
		assert.MapSubset(t, want, have)
	})

	t.Run("success by cid", func(t *testing.T) {
		// --- Given ---
		args := []string{"create", TestImg1.ref, "true"}
		cid := exekit.New(t, exekit.WithTrim).ExeStdout("docker", args...)
		t.Cleanup(func() { exekit.New(t).Exe("docker", "rm", cid) })

		// --- When ---
		have, err := getEnvs(t.Context(), os.Environ(), cid)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			xdef.EnvImgCreated: "2000-01-02T03:04:05Z",
			xdef.EnvImgTitle:   "Image1",
			envTestEmpty:       "",
		}
		assert.MapSubset(t, want, have)
	})

	t.Run("error - non-existent ref", func(t *testing.T) {
		// --- Given ---
		ref := RandRef()

		// --- When ---
		have, err := getEnvs(t.Context(), os.Environ(), ref)

		// --- Then ---
		wMsg := "" +
			"[getting environment variables] docker command error:\n" +
			"   cmd: docker inspect --format {{ json .Config.Env }} %s\n" +
			"   err: exit status 1\n" +
			"  eout: error: no such object: %s"
		wMsg = fmt.Sprintf(wMsg, ref, ref)
		assert.ErrorEqual(t, wMsg, err)
		assert.ExitCode(t, 1, err)
		assert.Nil(t, have)
	})

	t.Run("error - non-existent id", func(t *testing.T) {
		// --- Given ---
		iid := "000000000000"

		// --- When ---
		have, err := getEnvs(t.Context(), os.Environ(), iid)

		// --- Then ---
		wMsg := "" +
			"[getting environment variables] docker command error:\n" +
			"   cmd: docker inspect --format {{ json .Config.Env }} %s\n" +
			"   err: exit status 1\n" +
			"  eout: error: no such object: %s"
		wMsg = fmt.Sprintf(wMsg, iid, iid)
		assert.ErrorEqual(t, wMsg, err)
		assert.ExitCode(t, 1, err)
		assert.Nil(t, have)
	})
}

func Test_formatMapKeys(t *testing.T) {
	t.Run("sorted", func(t *testing.T) {
		// --- Given ---
		m := map[string]string{
			"KEY0": "VAl0",
			"KEY1": "VAl1",
			"KEY2": "VAl2",
		}

		// --- When ---
		have := formatMapKeys(m)

		// --- Then ---
		assert.Equal(t, "\"KEY0\"\n\"KEY1\"\n\"KEY2\"", have)
	})

	t.Run("empty map", func(t *testing.T) {
		// --- Given ---
		m := map[string]string{}

		// --- When ---
		have := formatMapKeys(m)

		// --- Then ---
		assert.Equal(t, "", have)
	})

	t.Run("nil map", func(t *testing.T) {
		// --- When ---
		have := formatMapKeys(nil)

		// --- Then ---
		assert.Equal(t, "", have)
	})
}

func Test_envGet_tabular(t *testing.T) {
	tt := []struct {
		testN string

		env       []string
		findKey   string
		wantValue string
	}{
		{"found", []string{"key0=val0", "key1=val1"}, "key1", "val1"},
		{"not found", []string{"key0=val0", "key1=val1"}, "key9", ""},
		{"partial", []string{"key0=val0", "key1=val1"}, "key", ""},
		{"empty env", []string{}, "key", ""},
		{"empty key", []string{"key0=val0", "key1=val1"}, "", ""},
		{
			"last value counts",
			[]string{"key0=val0", "key1=val1", "key0=abc"},
			"key0",
			"abc",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			haveValue := envGet(tc.env, tc.findKey)

			// --- Then ---
			assert.Equal(t, tc.wantValue, haveValue)
		})
	}
}
