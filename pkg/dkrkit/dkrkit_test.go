// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/xdef/pkg/xdef"
)

func Test_HasBuildArg(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		set := map[string]*string{
			"KEY0": new("VAL0"),
			"KEY1": new("VAL1"),
		}

		// --- When ---
		have := HasBuildArg(tspy, set, "VAL0", "KEY0")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - value not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[build args] expected the map to have a key with value:\n" +
			"   key: \"KEY0\"\n" +
			"  want: \"VALX\"\n" +
			"  have: \"VAL0\""
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		set := map[string]*string{
			"KEY0": new("VAL0"),
			"KEY1": new("VAL1"),
		}

		// --- When ---
		have := HasBuildArg(tspy, set, "VALX", "KEY0")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - nil value in map", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[build args] expected the map to have a key with value:\n" +
			"   key: \"KEY1\"\n" +
			"  want: \"VAL1\"\n" +
			"  have: nil"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		set := map[string]*string{
			"KEY0": new("VAL0"),
			"KEY1": nil,
		}

		// --- When ---
		have := HasBuildArg(tspy, set, "VAL1", "KEY1")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - non-existent key", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[build args] expected map to have a key:\n" +
			"  key: \"KEYX\"\n" +
			"  map:\n" +
			"       map[string]*string{\n" +
			"         \"KEY0\": \"VAL0\",\n" +
			"         \"KEY1\": \"VAL1\",\n" +
			"       }"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		set := map[string]*string{
			"KEY0": new("VAL0"),
			"KEY1": new("VAL1"),
		}

		// --- When ---
		have := HasBuildArg(tspy, set, "VALX", "KEYX")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_HasLabel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := HasLabel(tspy, TestImg0.ref, xdef.LabImgTitle, "Image0")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("empty existing label", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := HasLabel(tspy, TestImg0.ref, labTestEmpty, "")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - non-existent label", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting labels] expected label to exist:\n" +
			"     ref: %s\n" +
			"    want: \"not.existing.label\"\n" +
			"  labels:\n" +
			"          \"" + labTestEmpty + "\"\n" +
			"          \"" + xdef.LabImgBaseName + "\"\n" +
			"          \"" + xdef.LabImgCreated + "\"\n" +
			"          \"" + xdef.LabImgTitle + "\""
		tspy.ExpectLogEqual(wMsg, TestImg0.ref)
		tspy.Close()

		// --- When ---
		have := HasLabel(tspy, TestImg0.ref, "not.existing.label", "abc")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - not matching value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting labels] expected label value:\n" +
			"    ref: %s\n" +
			"  label: %q\n" +
			"   want: \"abc\"\n" +
			"   have: \"Image0\""
		tspy.ExpectLogEqual(wMsg, TestImg0.ref, xdef.LabImgTitle)
		tspy.Close()

		// --- When ---
		have := HasLabel(tspy, TestImg0.ref, xdef.LabImgTitle, "abc")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - non-existent reference", func(t *testing.T) {
		// --- Given ---
		ref := RandRef()

		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting labels] docker command error:\n" +
			"   cmd: docker inspect --format {{ json .Config.Labels }} %s\n" +
			"   err: exit status 1\n" +
			"  eout: error: no such object: %s"
		tspy.ExpectLogEqual(wMsg, ref, ref)
		tspy.Close()

		// --- When ---
		have := HasLabel(tspy, ref, xdef.LabImgAuthors, "abc")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_HasNoLabel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := HasNoLabel(tspy, TestImg0.ref, "not.existing.label")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - has label", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting labels] expected label not to exist:\n" +
			"    ref: %s\n" +
			"  label: %q"
		tspy.ExpectLogEqual(wMsg, TestImg0.ref, xdef.LabImgCreated)
		tspy.Close()

		// --- When ---
		have := HasNoLabel(tspy, TestImg0.ref, xdef.LabImgCreated)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - non-existent reference", func(t *testing.T) {
		// --- Given ---
		ref := RandRef()

		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting labels] docker command error:\n" +
			"   cmd: docker inspect --format {{ json .Config.Labels }} %s\n" +
			"   err: exit status 1\n" +
			"  eout: error: no such object: %s"
		tspy.ExpectLogEqual(wMsg, ref, ref)
		tspy.Close()

		// --- When ---
		have := HasNoLabel(tspy, ref, xdef.LabImgAuthors)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_HasLabels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		want := map[string]string{
			xdef.LabImgCreated: "2000-01-02T03:04:05Z",
			xdef.LabImgTitle:   "Image0",
			labTestEmpty:       "",
		}
		have := HasLabels(tspy, TestImg0.ref, want)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - missing", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting labels] expected the map to have keys:\n" +
			"   ref: %s\n" +
			"  keys: \"com.ctx42.meta.missing0\", \"com.ctx42.meta.missing1\""
		tspy.ExpectLogEqual(wMsg, TestImg0.ref)
		tspy.Close()

		// --- When ---
		want := map[string]string{
			xdef.LabImgCreated:        "2000-01-02T03:04:05Z",
			xdef.LabImgTitle:          "Image0",
			"com.ctx42.meta.missing0": "no",
			"com.ctx42.meta.missing1": "no",
		}
		have := HasLabels(tspy, TestImg0.ref, want)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - missing and wrong values", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"multiple expectations violated:\n" +
			"  error: [getting labels] expected values to be equal\n" +
			"  trail: map[\"com.ctx42.test.empty\"]\n" +
			"   want: \"wrong0\"\n" +
			"   have: \"\"\n" +
			"      ---\n" +
			"  error: [getting labels] expected values to be equal\n" +
			"  trail: map[\"org.opencontainers.image.created\"]\n" +
			"   want: \"2001-01-01T01:01:01Z\"\n" +
			"   have: \"2000-01-02T03:04:05Z\"\n" +
			"      ---\n" +
			"  error: [getting labels] expected the map to have keys\n" +
			"    ref: %s\n" +
			"   keys: \"com.ctx42.meta.missing0\", \"com.ctx42.meta.missing1\""
		tspy.ExpectLogEqual(wMsg, TestImg0.ref)
		tspy.Close()

		// --- When ---
		want := map[string]string{
			xdef.LabImgCreated:        "2001-01-01T01:01:01Z",
			xdef.LabImgTitle:          "Image0",
			labTestEmpty:              "wrong0",
			"com.ctx42.meta.missing0": "no",
			"com.ctx42.meta.missing1": "no",
		}
		have := HasLabels(tspy, TestImg0.ref, want)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - non-existent reference", func(t *testing.T) {
		// --- Given ---
		ref := RandRef()

		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting labels] docker command error:\n" +
			"   cmd: docker inspect --format {{ json .Config.Labels }} %s\n" +
			"   err: exit status 1\n" +
			"  eout: error: no such object: %s"
		tspy.ExpectLogEqual(wMsg, ref, ref)
		tspy.Close()

		// --- When ---
		have := HasLabels(tspy, ref, nil)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_HasEnv(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := HasEnv(tspy, TestImg0.ref, xdef.EnvImgTitle, "Image0")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("empty existing environment variable", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := HasEnv(tspy, TestImg0.ref, envTestEmpty, "")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - non-existent environment variable", func(t *testing.T) {
		// --- Given ---
		find := "NOT_EXISTING"

		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting environment variable] expected variable to exist:\n" +
			"   ref: %s\n" +
			"  want: \"NOT_EXISTING\"\n" +
			"   env:\n" +
			"        \"C42_TEST_EMPTY\"\n" +
			"        \"OCI_IMAGE_BASE_NAME\"\n" +
			"        \"OCI_IMAGE_CREATED\"\n" +
			"        \"OCI_IMAGE_TITLE\"\n" +
			"        \"PATH\""
		tspy.ExpectLogEqual(wMsg, TestImg0.ref)
		tspy.Close()

		// --- When ---
		have := HasEnv(tspy, TestImg0.ref, find, "abc")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - not matching value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting environment variable] expected variable value:\n" +
			"   ref: %s\n" +
			"  name: \"OCI_IMAGE_TITLE\"\n" +
			"  want: \"other\"\n" +
			"  have: \"Image0\""
		tspy.ExpectLogEqual(wMsg, TestImg0.ref)
		tspy.Close()

		// --- When ---
		have := HasEnv(tspy, TestImg0.ref, xdef.EnvImgTitle, "other")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - non-existent reference", func(t *testing.T) {
		// --- Given ---
		ref := RandRef()

		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting environment variables] docker command error:\n" +
			"   cmd: docker inspect --format {{ json .Config.Env }} %s\n" +
			"   err: exit status 1\n" +
			"  eout: error: no such object: %s"
		tspy.ExpectLogEqual(wMsg, ref, ref)
		tspy.Close()

		// --- When ---
		have := HasEnv(tspy, ref, xdef.LabImgAuthors, t.Name())

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_HasNoEnv(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := HasNoEnv(tspy, TestImg0.ref, "NOT_SET")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - empty existing environment variable", func(t *testing.T) {
		// --- Given ---
		find := envTestEmpty

		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting environment variable] expected variable not to exist:\n" +
			"        ref: %s\n" +
			"       name: \"C42_TEST_EMPTY\"\n" +
			"  has value: \"\""
		tspy.ExpectLogEqual(wMsg, TestImg0.iid)
		tspy.Close()

		// --- When ---
		have := HasNoEnv(tspy, TestImg0.iid, find)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - non-existent reference", func(t *testing.T) {
		// --- Given ---
		ref := RandRef()

		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting environment variables] docker command error:\n" +
			"   cmd: docker inspect --format {{ json .Config.Env }} %s\n" +
			"   err: exit status 1\n" +
			"  eout: error: no such object: %s"
		tspy.ExpectLogEqual(wMsg, ref, ref)
		tspy.Close()

		// --- When ---
		have := HasNoEnv(tspy, ref, "C42_TEST_VALUE0")

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_HasEnvs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		want := map[string]string{
			xdef.EnvImgCreated:  "2000-01-02T03:04:05Z",
			xdef.EnvImgBaseName: TestImageBaseRef,
			xdef.EnvImgTitle:    "Image0",
			envTestEmpty:        "",
		}
		have := HasEnvs(tspy, TestImg0.iid, want)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - one missing", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting environment variables] expected the map to have keys:\n" +
			"   ref: %s\n" +
			"  keys: \"C42_TEST_MISSING0\", \"C42_TEST_VALUE\""
		tspy.ExpectLogEqual(wMsg, TestImg0.ref)
		tspy.Close()

		// --- When ---
		want := map[string]string{
			envTestEmpty:        "",
			envTestValue:        "value",
			"C42_TEST_MISSING0": "missing0",
		}
		have := HasEnvs(tspy, TestImg0.ref, want)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - missing and wrong values", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"multiple expectations violated:\n" +
			"  error: [getting environment variables] expected values to be equal\n" +
			"  trail: map[\"C42_TEST_EMPTY\"]\n" +
			"   want: \"wrong\"\n" +
			"   have: \"\"\n" +
			"      ---\n" +
			"  error: [getting environment variables] expected values to be equal\n" +
			"  trail: map[\"OCI_IMAGE_TITLE\"]\n" +
			"   want: \"wrong\"\n" +
			"   have: \"Image0\"\n" +
			"      ---\n" +
			"  error: [getting environment variables] expected the map to have keys\n" +
			"    ref: %s\n" +
			"   keys: \"C42_TEST_MISSING0\", \"C42_TEST_MISSING1\""
		tspy.ExpectLogEqual(wMsg, TestImg0.ref)
		tspy.Close()

		// --- When ---
		want := map[string]string{
			xdef.EnvImgTitle:    "wrong",
			envTestEmpty:        "wrong",
			"C42_TEST_MISSING0": "missing0",
			"C42_TEST_MISSING1": "missing1",
		}
		have := HasEnvs(tspy, TestImg0.ref, want)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - non-existent reference", func(t *testing.T) {
		// --- Given ---
		ref := RandRef()

		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[getting environment variables] docker command error:\n" +
			"   cmd: docker inspect --format {{ json .Config.Env }} %s\n" +
			"   err: exit status 1\n" +
			"  eout: error: no such object: %s"
		tspy.ExpectLogEqual(wMsg, ref, ref)
		tspy.Close()

		// --- When ---
		have := HasEnvs(tspy, ref, nil)

		// --- Then ---
		assert.False(t, have)
	})
}
