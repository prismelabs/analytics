package log

import (
	"io"
	"os"
	"time"

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

type Logger struct {
	zerolog.Logger

	name string
}

func NewLogger(name string, w io.Writer, debug bool) Logger {
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

	return Logger{
		Logger: logger,
		name:   name,
	}
}

// Verbose implements migrate.Logger.
func (l *Logger) Verbose() bool {
	return l.Logger.GetLevel() <= zerolog.DebugLevel
}

func TestLoggers(logger ...Logger) {
	zerolog.ErrorHandler = func(err error) {
		panic(err)
	}

	for _, l := range logger {
		l.Log().Msgf("logger %q ready", l.name)
	}

	zerolog.ErrorHandler = nil
}

type bunyanLevelHook struct{}

func (blh bunyanLevelHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Int("level", zerologToBunyanLevels[level])
}
