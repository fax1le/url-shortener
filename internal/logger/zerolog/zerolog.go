package zerolog

import (
	"os"
	"time"
	"strings"
	"fmt"

	"github.com/rs/zerolog"
)

type ZerologLogger struct {
	logger *zerolog.Logger
}

func New() *ZerologLogger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	logger := zerolog.New(output).With().Timestamp().Logger()

	return &ZerologLogger{
		logger: &logger,
	}
}

func (z ZerologLogger) Debug(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	z.logger.Debug().Msg(msg)
}

func (z ZerologLogger) Info(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	z.logger.Info().Msg(msg)
}

func (z ZerologLogger) Warn(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	z.logger.Warn().Msg(msg)
}

func (z ZerologLogger) Error(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	z.logger.Error().Msg(msg)
}

func (z ZerologLogger) Fatal(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	z.logger.Fatal().Msg(msg)
}

func (z ZerologLogger) Close() error {
	return nil
}
