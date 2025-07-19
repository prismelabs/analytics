// Package log provides structured logging on top of slog package from standard
// library.
package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"time"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Level = slog.Level

const (
	LevelTrace Level = LevelDebug - 4
	LevelDebug       = slog.LevelDebug
	LevelInfo        = slog.LevelInfo
	LevelWarn        = slog.LevelWarn
	LevelError       = slog.LevelError
	LevelFatal Level = LevelError + 4
)

var levelMap = map[Level]int{
	LevelTrace:      10,
	slog.LevelDebug: 20,
	slog.LevelInfo:  30,
	slog.LevelWarn:  40,
	slog.LevelError: 50,
	LevelFatal:      60,
}

// Private alias.
type slogger = slog.Logger

// Logger define a structured logger built on top of `log/slog` package.
type Logger struct {
	*slogger
}

// New creates a new configured Logger.
func New(name string, w io.Writer, debug bool) Logger {
	pid := os.Getpid()
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.LevelKey:
				return slog.Int(slog.LevelKey, levelMap[a.Value.Any().(Level)])
			default:
				return a
			}
		},
	})).With(
		"v", 0,
		"pid", pid,
		"hostname", hostname,
		"name", name,
	)

	return Logger{logger}
}

func (l Logger) TestOutput() error {
	return l.Handler().Handle(
		context.Background(),
		slog.NewRecord(time.Now(), slog.LevelInfo, "logger ready", 0),
	)
}

// With returns a Logger that includes the given attributes in each output
// operation. Arguments are converted to attributes as if by Logger.Log.
func (l Logger) With(args ...any) Logger {
	return Logger{l.slogger.With(args...)}
}

// With returns a Logger that includes the given attributes in each output
// operation. Arguments are converted to attributes as if by Logger.Log.
func (l Logger) WithGroup(name string) Logger {
	return Logger{l.slogger.WithGroup(name)}
}

// Trace logs at LevelTrace.
func (l Logger) Trace(msg string, args ...any) {
	l.Log(context.Background(), LevelTrace, msg, args...)
}

// Err logs at LevelError if error is not nil.
func (l Logger) Err(msg string, err error, args ...any) {
	if err == nil {
		return
	}

	args = slices.Insert[[]any, any](args, 0, "error", err)
	l.Log(context.Background(), LevelError, msg, args...)
}

// Fatal logs at LevelFatal then panic if error is not nil.
func (l Logger) Fatal(msg string, err error, args ...any) {
	if err == nil {
		return
	}

	args = slices.Insert[[]any, any](args, 0, "error", err)
	l.Log(context.Background(), LevelFatal, msg, args...)
	panic(err)
}

// GoMigrateLogger wraps the given logger to implements gomigrate.Logger.
func GoMigrateLogger(logger Logger) gomigrate.Logger {
	return &goMigrateLogger{logger}
}

type goMigrateLogger struct {
	Logger
}

// Printf implements migrate.Logger.
func (gml *goMigrateLogger) Printf(format string, v ...any) {
	gml.Log(context.Background(), LevelDebug, fmt.Sprintf(format, v...))
}

// Verbose implements migrate.Logger.
func (gml *goMigrateLogger) Verbose() bool {
	return gml.Enabled(context.Background(), LevelDebug)
}

type promLogger struct {
	Logger
}

// PrometheusLogger returns a wrapped around given logger that implements
// promhttp.Logger.
func PrometheusLogger(logger Logger) promhttp.Logger {
	return promLogger{logger}
}

// Println implements promhttp.Logger.
func (pl promLogger) Println(v ...interface{}) {
	pl.Log(context.Background(), LevelDebug, fmt.Sprint(v...))
}
