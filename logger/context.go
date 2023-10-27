package logger

import "context"

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
