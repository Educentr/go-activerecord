package logger

import "context"

type ValueLogPrefix map[string]interface{}

type LogLevel uint32

const (
	PanicLoggerLevel LogLevel = iota
	FatalLoggerLevel
	ErrorLoggerLevel
	WarnLoggerLevel
	InfoLoggerLevel
	DebugLoggerLevel
	TraceLoggerLevel
)

type LoggerInterface interface {
	SetLoggerValueToContext(ctx context.Context, addVal ValueLogPrefix) context.Context

	SetLogLevel(level LogLevel)
	Fatal(ctx context.Context, args ...interface{})
	Error(ctx context.Context, args ...interface{})
	Warn(ctx context.Context, args ...interface{})
	Info(ctx context.Context, args ...interface{})
	Debug(ctx context.Context, args ...interface{})
	Trace(ctx context.Context, args ...interface{})

	CollectQueries(ctx context.Context, f func() (MockerLogger, error))
}
