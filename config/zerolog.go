package config

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewLogger() {
	// pretty console logging
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	log.Logger = zerolog.New(output).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()

	// global config
	zerolog.TimeFieldFormat = time.RFC3339
}
