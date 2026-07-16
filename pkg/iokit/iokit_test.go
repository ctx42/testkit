// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package iokit

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/testkit/pkg/oskit"
)

func Test_ReadAll(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		rdr := strings.NewReader("content")

		// --- When ---
		have := ReadAll(tspy, rdr)

		// --- Then ---
		assert.Equal(t, "content", string(have))
	})

	t.Run("error from reader", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogEqual(ErrRead.Error())
		tspy.Close()

		rdr := ErrReader(strings.NewReader("content"), 1)

		// --- When ---
		have := ReadAll(tspy, rdr)

		// --- Then ---
		assert.Equal(t, []byte{byte('c')}, have)
	})

	t.Run("close is called", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		fil := must.Value(os.Open("testdata/file.txt"))

		// --- When ---
		have := ReadAll(tspy, fil)

		// --- Then ---
		tspy.Finish().AssertExpectations()
		assert.Equal(t, []byte("content"), have)
		wMsg := "close testdata/file.txt: file already closed"
		assert.ErrorEqual(t, wMsg, fil.Close())
	})
}

func Test_ReadAllStr(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		rdr := strings.NewReader("content")

		// --- When ---
		have := ReadAllStr(tspy, rdr)

		// --- Then ---
		assert.Equal(t, "content", have)
	})

	t.Run("error from reader", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogEqual(ErrRead.Error())
		tspy.Close()

		rdr := ErrReader(strings.NewReader("content"), 1)

		// --- When ---
		have := ReadAllStr(tspy, rdr)

		// --- Then ---
		assert.Equal(t, "c", have)
	})

	t.Run("close is called", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		fil := must.Value(os.Open("testdata/file.txt"))

		// --- When ---
		have := ReadAllStr(tspy, fil)

		// --- Then ---
		tspy.Finish().AssertExpectations()
		assert.Equal(t, "content", have)
		wMsg := "close testdata/file.txt: file already closed"
		assert.ErrorEqual(t, wMsg, fil.Close())
	})
}

func Test_ReadAllFromStart(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		fil := must.Value(os.Open("testdata/file.txt"))
		must.Value(fil.Seek(3, io.SeekStart))

		// --- When ---
		have := ReadAllFromStart(fil)

		// --- Then ---
		assert.Equal(t, []byte("content"), have)
		assert.Equal(t, int64(3), must.Value(fil.Seek(0, io.SeekCurrent)))
	})
}

func Test_Offset(t *testing.T) {
	// --- Given ---
	pth := filepath.Join(t.TempDir(), "test_file.txt")
	content := "line1\nline2\nend"
	oskit.Create(t, content, pth)

	fil := must.Value(os.Open(pth))
	must.Value(fil.Read(make([]byte, 3)))

	// --- When ---
	have := Offset(fil)

	// --- Then ---
	assert.Equal(t, int64(3), have)
}

func Test_Seek(t *testing.T) {
	// --- Given ---
	pth := filepath.Join(t.TempDir(), "test_file.txt")
	content := "line1\nline2\nend"
	oskit.Create(t, content, pth)

	fil := must.Value(os.Open(pth))

	// --- When ---
	have := Seek(fil, 4, io.SeekStart)

	// --- Then ---
	assert.Equal(t, int64(4), have)
	assert.Equal(t, "1\nline2\nend", string(must.Value(io.ReadAll(fil))))
}
