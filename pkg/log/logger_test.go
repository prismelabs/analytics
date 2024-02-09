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
		logger := NewLogger("test_logger_1", buf, false)

		logger.Debug().Msg("debug log")
		logger.Trace().Msg("trace log")
		require.Len(t, buf.String(), 0)
	})

	t.Run("WithDebug", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := NewLogger("test_logger_1", buf, true)

		logger.Debug().Msg("debug log")
		logger.Trace().Msg("trace log")
		require.Len(t, strings.Split(buf.String(), "\n"), 2)
	})

	t.Run("Format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := NewLogger("test_logger_1", buf, true)

		logger.Info().Str("foo", "bar").Msg("info log")

		require.Regexp(t,
			regexp.MustCompile(`^{"v":0,"pid":\d+,"hostname":".+","name":"test_logger_1","foo":"bar","level":30,"time":"20\d{2}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z","msg":"info log"}\n$`),
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
		logger := NewLogger("test_logger_1", errWriter{errors.New("unexpected error")}, false)

		require.Panics(t, func() {
			TestLoggers(logger)
		})
	})
	t.Run("SingleLogger/NoError/Panics", func(t *testing.T) {
		logger := NewLogger("test_logger_1", io.Discard, false)

		TestLoggers(logger)
	})

	t.Run("MultipleLogger/NoError/Panics", func(t *testing.T) {
		logger1 := NewLogger("test_logger_1", io.Discard, false)
		logger2 := NewLogger("test_logger_2", io.Discard, false)

		TestLoggers(logger1, logger2)
	})
	t.Run("MultipleLogger/SingleError/Panics", func(t *testing.T) {
		logger1 := NewLogger("test_logger_1", io.Discard, false)
		logger2 := NewLogger("test_logger_2", errWriter{errors.New("unexpected error 2")}, false)

		require.Panics(t, func() {
			TestLoggers(logger1, logger2)
		})
	})
	t.Run("MultipleLogger/WithErrors/Panics", func(t *testing.T) {
		logger1 := NewLogger("test_logger_1", errWriter{errors.New("unexpected error 1")}, false)
		logger2 := NewLogger("test_logger_2", errWriter{errors.New("unexpected error 2")}, false)

		require.Panics(t, func() {
			TestLoggers(logger1, logger2)
		})
	})

}
