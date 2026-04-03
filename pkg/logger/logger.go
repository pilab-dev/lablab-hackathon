package logger

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	consoleLevel zerolog.Level = zerolog.InfoLevel
	consoleOnce  sync.Once
	consoleW     zerolog.ConsoleWriter
	fileW        *os.File
	mu           sync.RWMutex
)

func init() {
	consoleW = zerolog.ConsoleWriter{Out: os.Stdout}

	logFile, err := os.OpenFile("trader.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Logger = zerolog.New(consoleW).With().Caller().Timestamp().Logger()
	} else {
		fileW = logFile
		log.Logger = zerolog.New(zerolog.MultiLevelWriter(consoleW, logFile)).With().Caller().Timestamp().Logger()
	}

	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Logger.Level(zerolog.InfoLevel)
}

func SetConsoleLevel(level string) error {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return err
	}

	mu.Lock()
	consoleLevel = lvl
	mu.Unlock()

	consoleOnce.Do(func() {
		consoleW = zerolog.ConsoleWriter{Out: os.Stdout}
	})

	log.Logger = zerolog.New(zerolog.MultiLevelWriter(consoleW, fileW)).With().Caller().Timestamp().Logger().Level(consoleLevel)

	log.Info().Str("level", level).Msg("Console log level updated")
	return nil
}

func GetConsoleLevel() string {
	mu.RLock()
	defer mu.RUnlock()
	return consoleLevel.String()
}
