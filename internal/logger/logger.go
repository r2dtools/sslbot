package logger

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/r2dtools/sslbot/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Error(message string, args ...any)
	Warning(message string, args ...any)
	Info(message string, args ...any)
	Debug(message string, args ...any)
}

type logger struct {
	zapLogger *zap.SugaredLogger
}

func (l *logger) Error(message string, args ...interface{}) {
	l.zapLogger.Errorf(message, args...)
}

func (l *logger) Warning(message string, args ...interface{}) {
	l.zapLogger.Warnf(message, args...)
}

func (l *logger) Info(message string, args ...interface{}) {
	l.zapLogger.Infof(message, args...)
}

func (l *logger) Debug(message string, args ...interface{}) {
	l.zapLogger.Debugf(message, args...)
}

func NewLogger(config *config.Config) (Logger, error) {
	logDir := path.Dir(config.LogFile)

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, 0755)

		if err != nil {
			return nil, err
		}
	}

	var loggerConfig zap.Config
	outputPaths := []string{}

	if config.IsDevMode {
		loggerConfig = zap.NewDevelopmentConfig()
		outputPaths = append(outputPaths, "stderr")
	} else {
		outputPaths = append(outputPaths, config.LogFile)
		loggerConfig = zap.NewProductionConfig()
	}

	loggerConfig.OutputPaths = outputPaths
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	zLogger, err := loggerConfig.Build()

	if err != nil {
		return nil, err
	}

	return &logger{zapLogger: zLogger.Sugar()}, nil
}

type NilLogger struct{}

func (l *NilLogger) Error(message string, args ...any) {
}

func (l *NilLogger) Warning(message string, args ...any) {
}

func (l *NilLogger) Info(message string, args ...any) {
}

func (l *NilLogger) Debug(message string, args ...any) {
}

type TestLogger struct {
	T *testing.T
}

func (l *TestLogger) Error(message string, args ...any) {
	l.T.Logf(message, args...)
}

func (l *TestLogger) Warning(message string, args ...any) {
	l.T.Logf(message, args...)
}

func (l *TestLogger) Info(message string, args ...any) {
	l.T.Logf(message, args...)
}

func (l *TestLogger) Debug(message string, args ...any) {
	l.T.Logf(message, args...)
}
