package logger

import (
	"context"
	"log/slog"
)

type loggerKey string

const LoggerKey loggerKey = "logger"

type Logger interface {
	Debug(v ...any)
	Debugf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Warn(v ...any)
	Warnf(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
	Printf(format string, v ...any)
	Println(v ...any)
}

type nopLogger struct{}

var noplogger = &nopLogger{}

func (*nopLogger) Debug(v ...any)                 {}
func (*nopLogger) Debugf(format string, v ...any) {}
func (*nopLogger) Info(v ...any)                  {}
func (*nopLogger) Infof(format string, v ...any)  {}
func (*nopLogger) Warn(v ...any)                  {}
func (*nopLogger) Warnf(format string, v ...any)  {}
func (*nopLogger) Error(v ...any)                 {}
func (*nopLogger) Errorf(format string, v ...any) {}
func (*nopLogger) Printf(format string, v ...any) {}
func (*nopLogger) Println(v ...any)               {}

type logger struct {
	l *slog.Logger
}

var _ Logger = (*logger)(nil)

// Debug implements Logger.
func (l *logger) Debug(v ...any) {
	l.l.Debug("[DEBUG]", v...)
}

// Debugf implements Logger.
func (l *logger) Debugf(format string, v ...any) {
	l.l.Debug("[DEBUG]", v...)
}

// Error implements Logger.
func (l *logger) Error(v ...any) {
	l.l.Error("[ERROR]", v...)
}

// Errorf implements Logger.
func (l *logger) Errorf(format string, v ...any) {
	l.l.Error("[ERROR]", v...)
}

// Info implements Logger.
func (l *logger) Info(v ...any) {
	l.l.Info("[INFO]", v...)
}

// Infof implements Logger.
func (l *logger) Infof(format string, v ...any) {
	l.l.Info("[INFO]", v...)
}

// Printf implements Logger.
func (l *logger) Printf(format string, v ...any) {
	panic("unimplemented")
}

// Println implements Logger.
func (*logger) Println(v ...any) {
	panic("unimplemented")
}

// Warn implements Logger.
func (l *logger) Warn(v ...any) {
	l.l.Warn("[WARN]", v...)
}

// Warnf implements Logger.
func (l *logger) Warnf(format string, v ...any) {
	l.l.Warn("[WARN]", v...)
}

func FromContext(ctx context.Context) Logger {
	logger, ok := ctx.Value(LoggerKey).(Logger)
	if !ok {
		return noplogger
	}
	return logger
}

func NewContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}
