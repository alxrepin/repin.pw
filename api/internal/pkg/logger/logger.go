package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Debug bool `env:"LOGGER_DEBUG" envDefault:"true"`
}

func MustLoad(cfg Config) *zerolog.Logger {
	level := zerolog.InfoLevel
	if cfg.Debug {
		level = zerolog.DebugLevel
	}

	writer := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	l := zerolog.New(writer).With().Timestamp().Logger().Level(level)

	return &l
}
