package arsimplelog

import (
	"context"
	"fmt"
	"log"

	"github.com/Educentr/go-activerecord/pkg/logger"
	// ToDo не очень правильная зависимость, с такими успехами придётся тащить все логгеры для всех бекендов, непонятно зачем
)

type ctxKey uint8
type DefaultLogger struct {
	level  logger.LogLevel
	Fields logger.ValueLogPrefix
}

const (
	ContextLogPrefix ctxKey = iota
)

const (
	ValueContextErrorField = "context"
)

func NewLogger() *DefaultLogger {
	return &DefaultLogger{
		level:  logger.InfoLoggerLevel,
		Fields: logger.ValueLogPrefix{"orm": "activerecord"},
	}
}

func (l *DefaultLogger) getLoggerFromContext(ctx context.Context) logger.LoggerInterface {
	return l.getLoggerFromContextAndValue(ctx, logger.ValueLogPrefix{})
}

func (l *DefaultLogger) SetLoggerValueToContext(ctx context.Context, val logger.ValueLogPrefix) context.Context {
	ctxVal := ctx.Value(ContextLogPrefix)
	if ctxVal != nil {
		lprefix, ok := ctxVal.(logger.ValueLogPrefix)
		if !ok {
			val["logger.context.error"] = ValueContextErrorField
			val["logger.context.valueType"] = fmt.Sprintf("%T", ctxVal)
		} else {
			for k, v := range lprefix {
				if _, ok := val[k]; !ok {
					val[k] = v
				}
			}
		}
	}

	return context.WithValue(ctx, ContextLogPrefix, val)
}

func (l *DefaultLogger) getLoggerFromContextAndValue(ctx context.Context, addVal logger.ValueLogPrefix) logger.LoggerInterface {
	// Думаю что надо закешировать один раз инстанс логгера для контекста
	// Но надо учитывать, что мог измениться уровень логирования хотим ли мы в рамках одного запроса
	// менять уровни логирования?
	// Еще надо добавить в логгер конфигурацию, что бы уровни логирования можно было
	// настраивать на уровне моделей
	nl := NewLogger()
	nl.level = l.level

	for k, v := range l.Fields {
		nl.Fields[k] = v
	}

	for k, v := range addVal {
		nl.Fields[k] = v
	}

	ctxVal := ctx.Value(ContextLogPrefix)
	if ctxVal == nil {
		nl.Fields["logger.context"] = "empty"
	} else {
		lprefix, ok := ctxVal.(logger.ValueLogPrefix)
		if !ok {
			nl.Fields["logger.context.error"] = ValueContextErrorField
			nl.Fields["logger.context.valueType"] = fmt.Sprintf("%T", ctxVal)
		} else {
			for k, v := range lprefix {
				nl.Fields[k] = v
			}
		}
	}

	return nl
}

func (l *DefaultLogger) SetLogLevel(level logger.LogLevel) {
	l.level = level
}

func (l *DefaultLogger) loggerPrint(level logger.LogLevel, lprefix string, args ...interface{}) {
	if l.level < level {
		return
	}

	log.Print(lprefix, l.Fields, args)
}

func (l *DefaultLogger) Debug(ctx context.Context, args ...interface{}) {
	l.getLoggerFromContext(ctx).(*DefaultLogger).loggerPrint(logger.DebugLoggerLevel, "DEBUG: ", args)
}

func (l *DefaultLogger) Trace(ctx context.Context, args ...interface{}) {
	l.getLoggerFromContext(ctx).(*DefaultLogger).loggerPrint(logger.TraceLoggerLevel, "TRACE: ", args)
}

func (l *DefaultLogger) Info(ctx context.Context, args ...interface{}) {
	l.getLoggerFromContext(ctx).(*DefaultLogger).loggerPrint(logger.InfoLoggerLevel, "INFO: ", args)
}

func (l *DefaultLogger) Error(ctx context.Context, args ...interface{}) {
	l.getLoggerFromContext(ctx).(*DefaultLogger).loggerPrint(logger.ErrorLoggerLevel, "ERROR: ", args)
}

func (l *DefaultLogger) Warn(ctx context.Context, args ...interface{}) {
	l.getLoggerFromContext(ctx).(*DefaultLogger).loggerPrint(logger.WarnLoggerLevel, "WARN: ", args)
}

func (l *DefaultLogger) Fatal(ctx context.Context, args ...interface{}) {
	log.Fatal("FATAL: ", l.Fields, args)
}

func (l *DefaultLogger) Panic(ctx context.Context, args ...interface{}) {
	log.Panic("PANIC; ", l.Fields, args)
}

func (l *DefaultLogger) CollectQueries(ctx context.Context, f func() (logger.MockerLogger, error)) {
}
