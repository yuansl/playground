package logger

import (
	"fmt"
	"time"
)

type Level int

const (
	LFatal Level = iota
	LError Level = iota
	LWarn
	LInfo
	LDebug
)

type Logger interface {
	Debugf(format string, v ...any)
	Infof(format string, v ...any)
	Warnf(format string, v ...any)
	Errorf(format string, v ...any)
	Fatalf(format string, v ...any)
}

type logger struct {
	level Level
}

func (l logger) Debugf(format string, v ...any) {
	if l.level <= LDebug {
		prefix := fmt.Sprintf("[%s][DEBUG] ", time.Now().Format(time.RFC3339Nano))
		fmt.Printf(prefix+format, v...)
	}
}

func (l logger) Infof(format string, v ...any) {
	if l.level <= LInfo {
		prefix := fmt.Sprintf("[%s][INFO] ", time.Now().Format(time.RFC3339Nano))
		fmt.Printf(prefix+format, v...)
	}
}

func (l logger) Warnf(format string, v ...any) {
	if l.level <= LWarn {
		prefix := fmt.Sprintf("[%s][WARN] ", time.Now().Format(time.RFC3339Nano))
		fmt.Printf(prefix+format, v...)
	}
}

func (l logger) Errorf(format string, v ...any) {
	if l.level <= LError {
		prefix := fmt.Sprintf("[%s][ERROR] ", time.Now().Format(time.RFC3339Nano))
		fmt.Printf(prefix+format, v...)
	}
}

func (l *logger) Fatalf(format string, v ...any) {
	if l.level <= LFatal {
		prefix := fmt.Sprintf("[%s][FATAL] ", time.Now().Format(time.RFC3339Nano))
		fmt.Printf(prefix+format, v...)
	}
}

func (l *logger) Panicf(format string, v ...any) {
	if l.level <= LFatal {
		prefix := fmt.Sprintf("[%s][FATAL] ", time.Now().Format(time.RFC3339Nano))
		fmt.Printf(prefix+format, v...)
	}
}

func New(level Level) Logger {
	return &logger{level: level}
}
