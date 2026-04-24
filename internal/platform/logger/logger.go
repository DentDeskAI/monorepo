package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(level string) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil || lvl == zerolog.NoLevel {
		lvl = zerolog.InfoLevel
	}
	zerolog.TimeFieldFormat = time.RFC3339
	return zerolog.New(os.Stdout).
		Level(lvl).
		With().
		Timestamp().
		Str("service", "dentdesk").
		Logger()
}
