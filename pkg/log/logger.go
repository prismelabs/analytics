package log

import (
	"io"
	"os"
	"time"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog"
)

var zerologToBunyanLevels = []int{
	zerolog.DebugLevel: 20,
	zerolog.InfoLevel:  30,
	zerolog.WarnLevel:  40,
	zerolog.ErrorLevel: 50,
	zerolog.FatalLevel: 60,
	zerolog.PanicLevel: 70,
	zerolog.NoLevel:    30,
	zerolog.Disabled:   0,
}

func init() {
	zerolog.MessageFieldName = "msg"
	zerolog.LevelFieldName = "" // Disable level field so bunyan hook can set it.
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }
}

// NewLogger creates a new configured zerolog logger.
func NewLogger(name string, w io.Writer, debug bool) zerolog.Logger {
	pid := os.Getpid()
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	logger := zerolog.New(w).Hook(bunyanLevelHook{}).
		With().
		Timestamp().
		Int("v", 0).
		Int("pid", pid).
		Str("hostname", hostname).
		Str("name", name).
		Logger()

	if debug {
		logger = logger.Level(zerolog.DebugLevel)
	} else {
		logger = logger.Level(zerolog.InfoLevel)
	}

	return logger
}

func TestLoggers(logger ...zerolog.Logger) {
	initialErrorHandler := zerolog.ErrorHandler

	zerolog.ErrorHandler = func(err error) {
		panic(err)
	}

	for _, l := range logger {
		l.Log().Msgf("logger ready")
	}

	zerolog.ErrorHandler = initialErrorHandler
}

type bunyanLevelHook struct{}

func (blh bunyanLevelHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Int("level", zerologToBunyanLevels[level])
}

// GoMigrateLogger wraps the given logger to implements gomigrate.Logger.
func GoMigrateLogger(logger zerolog.Logger) gomigrate.Logger {
	return &goMigrateLogger{logger}
}

type goMigrateLogger struct {
	zerolog.Logger
}

// Verbose implements migrate.Logger.
func (gml *goMigrateLogger) Verbose() bool {
	return gml.Logger.GetLevel() <= zerolog.DebugLevel
}
