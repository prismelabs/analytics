package log

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("WithoutDebug", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := New("test_logger_1", buf, false)

		logger.Debug("debug log")
		logger.Trace("trace log")
		require.Len(t, buf.String(), 0)
	})

	t.Run("WithDebug", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := New("test_logger_1", buf, true)

		logger.Debug("debug log")
		logger.Trace("trace log")
		require.Len(t, strings.Split(buf.String(), "\n"), 2)
	})

	t.Run("Format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := New("test_logger_1", buf, true)

		logger.Info("info log", "foo", "bar")

		require.Regexp(t,
			regexp.MustCompile(
				`{"time":"((?:(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:\d{2})?)",`+
					`"level":30,`+
					`"msg":"info log",`+
					`"v":0,`+
					`"pid":\d+,`+
					`"hostname":".+",`+
					`"name":"test_logger_1",`+
					`"foo":"bar"`+
					`}\n$`),
			buf.String(),
		)
	})
}

type errWriter struct {
	err error
}

func (ew errWriter) Write(data []byte) (int, error) {
	return 0, ew.err
}

func TestTestLoggers(t *testing.T) {
	t.Run("SingleLogger/WithError/Panics", func(t *testing.T) {
		logger := New("test_logger_1", errWriter{errors.New("unexpected error")}, false)
		err := logger.TestOutput()
		require.Error(t, err)
	})
	t.Run("SingleLogger/NoError/Panics", func(t *testing.T) {
		logger := New("test_logger_1", io.Discard, false)

		err := logger.TestOutput()
		require.NoError(t, err)
	})
}
