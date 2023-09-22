package logger

import (
	"context"
	"net/http"

	"github.com/qbox/pili/base/qiniu/http/rpcutil.v1"
	"github.com/qbox/pili/base/qiniu/xlog.v1"
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

func New() Logger {
	return xlog.NewDummy()
}

func NewWith(reqId string) Logger {
	return xlog.NewWith(reqId)
}

func SetOutputLevel(level int) {
	xlog.SetOutputLevel(level)
}

func NewWithEnv(env *rpcutil.Env) Logger {
	return xlog.New(env.W, env.Req)
}

func NewWithRequest(req *http.Request) Logger {
	return xlog.NewWithReq(req)
}

func FromContext(ctx context.Context) Logger {
	logger, ok := ctx.Value(LoggerKey).(Logger)
	if !ok {
		if logger, ok := xlog.FromContext(ctx); ok {
			return logger
		}
		return noplogger
	}
	return logger
}

func NewContext(parent context.Context, logger Logger) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, LoggerKey, logger)
}

func IdFromContext(ctx context.Context) string {
	if v, ok := FromContext(ctx).(interface{ ReqId() string }); ok {
		return v.ReqId()
	}
	return ""
}
