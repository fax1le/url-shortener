package logrus

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Change this 

type LogrusLogger struct {
	logger *logrus.Logger
}

func New() *LogrusLogger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true, 
		FullTimestamp: true, 
	})

	logger.SetOutput(os.Stdout)

	return &LogrusLogger{
		logger: logger,
	}
}

func (l LogrusLogger) Debug(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	l.logger.Debug(msg)
}

func (l LogrusLogger) Info(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	l.logger.Info(msg)
}

func (l LogrusLogger) Warn(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	l.logger.Warn(msg)
}

func (l LogrusLogger) Error(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	l.logger.Error(msg)
}

func (l LogrusLogger) Fatal(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	l.logger.Fatal(msg)
}

func (l LogrusLogger) Printf(level string, format string, v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintf(format, v ...), "\n")

	switch level {
	case "DEBUG":
		l.logger.Debug(msg)
	case "INFO":
		l.logger.Info(msg)
	case "WARN":
		l.logger.Warn(msg)
	case "ERROR":
		l.logger.Error(msg)
	default:
		l.logger.Info(msg)
	}
}

func (l LogrusLogger) Close() error {
	return nil
}



