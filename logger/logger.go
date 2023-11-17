package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Debug(v ...any)
	Debugf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Warn(v ...any)
	Warnf(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
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

func New() Logger             { return &logger{log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)} }
func NewWith(v ...any) Logger { return New() }

type logger struct{ *log.Logger }

// Debug implements Logger.
func (*logger) Debug(v ...any) {
	panic("unimplemented")
}

// Debugf implements Logger.
func (*logger) Debugf(format string, v ...any) {
	panic("unimplemented")
}

// Error implements Logger.
func (*logger) Error(v ...any) {
	panic("unimplemented")
}

// Errorf implements Logger.
func (*logger) Errorf(format string, v ...any) {
	panic("unimplemented")
}

// Info implements Logger.
func (*logger) Info(v ...any) {
	panic("unimplemented")
}

// Infof implements Logger.
func (l *logger) Infof(format string, v ...any) {
	l.Output(2, fmt.Sprintf(format, v...))
}

// Warn implements Logger.
func (*logger) Warn(v ...any) {
	panic("unimplemented")
}

// Warnf implements Logger.
func (l *logger) Warnf(format string, v ...any) {
	l.Output(2, fmt.Sprintf(format, v...))
}
