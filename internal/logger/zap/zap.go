package zap

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger *zap.Logger
}

func New() *ZapLogger {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true

	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
    	enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}

	logger, _ := cfg.Build()

	return &ZapLogger{
		logger: logger,
	}
}

func (z ZapLogger) Debug(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	z.logger.Debug(msg)
}

func (z ZapLogger) Info(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	z.logger.Info(msg)
}

func (z ZapLogger) Warn(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	z.logger.Warn(msg)
}

func (z ZapLogger) Error(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	z.logger.Error(msg)
}

func (z ZapLogger) Fatal(v ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(v...), "\n")
	z.logger.Fatal(msg)
}

func (l ZapLogger) Printf(level string, format string, v ...any) {
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
func (z ZapLogger) Close() error {
	err := z.logger.Sync()
	return err
}


