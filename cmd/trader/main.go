package main

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var defaultLogger = zerolog.New(zerolog.ConsoleWriter{
	Out: os.Stdout,
}).Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Caller().Timestamp().Logger()

func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = defaultLogger
}

func main() {
	if err := Execute(); err != nil {
		log.Fatal().Err(err).Msg("trader exited with error")
	}
}
