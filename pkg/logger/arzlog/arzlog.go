package arzlog

import (
	"context"
	"fmt"
	"sync"

	"github.com/Educentr/go-activerecord/pkg/logger"
	zlog "github.com/rs/zerolog"
)

type cxtField int8

const (
	ctxEventName cxtField = iota
	ctxQueries
)

type AggrCollectedMocker map[string][]string
type AggrCollectedFixtures map[string][]any

type AggrCollectedLogger struct {
	Mocked   AggrCollectedMocker
	Fixtures AggrCollectedFixtures
	sync.Mutex
}

var Collected *AggrCollectedLogger

type ARLogger struct {
	level          zlog.Level
	collectFixture bool
}

// Todo onlineconf
func NewARLogger(level ...logger.LogLevel) *ARLogger {
	nLog := &ARLogger{
		level: zlog.InfoLevel,
	}

	if len(level) > 0 {
		nLog.SetLogLevel(level[0])
	}

	return nLog
}

func NewCollectARLogger(level ...logger.LogLevel) *ARLogger {
	nLog := &ARLogger{
		level:          zlog.InfoLevel,
		collectFixture: true,
	}

	if len(level) > 0 {
		nLog.SetLogLevel(level[0])
	}

	Collected = &AggrCollectedLogger{
		Mocked:   make(AggrCollectedMocker, 0),
		Fixtures: make(AggrCollectedFixtures, 0),
	}

	return nLog
}

func (l *ARLogger) getLoggerContext(ctx context.Context) *zlog.Context {
	ctxEv := ctx.Value(ctxEventName)

	var logCtx zlog.Context

	var existsLogCtx bool

	if ctxEv != nil {
		var ok bool
		if logCtx, ok = ctxEv.(zlog.Context); ok {
			existsLogCtx = true
		}
	}

	if !existsLogCtx {
		logCtx = zlog.Ctx(ctx).Level(l.level).With()
	}

	return &logCtx
}

func (l *ARLogger) SetLoggerValueToContext(ctx context.Context, vals logger.ValueLogPrefix) context.Context {
	logEv := *(l.getLoggerContext(ctx))

	for k, v := range vals {
		logEv = logEv.Interface(k, v)
	}

	return context.WithValue(ctx, ctxEventName, logEv)
}

func (l *ARLogger) SetLogLevel(level logger.LogLevel) {
	switch level {
	case logger.TraceLoggerLevel:
		l.level = zlog.TraceLevel
	case logger.DebugLoggerLevel:
		l.level = zlog.DebugLevel
	case logger.InfoLoggerLevel:
		l.level = zlog.InfoLevel
	case logger.WarnLoggerLevel:
		l.level = zlog.WarnLevel
	case logger.ErrorLoggerLevel:
		l.level = zlog.ErrorLevel
	case logger.FatalLoggerLevel:
		l.level = zlog.FatalLevel
	default:
		l.level = zlog.InfoLevel
	}
}

func (l *ARLogger) do(event *zlog.Event, args ...interface{}) {
	msg := ""
	first := 0

	last := len(args)
	if last > 0 {
		ok := false
		if msg, ok = args[0].(string); ok {
			first = 1
		}
	}

	for i := first; i < last; i++ {
		switch a := args[i].(type) {
		case error:
			event.Err(a)
		default:
			msg = fmt.Sprintf("%s %v", msg, a)
		}
	}

	if msg == "" {
		msg = "empty msg"
	}

	// Вот тут прям очень плохо выглядит migic num и не всегда корректно работает
	// выставить какой то конкретный skip не получится там есть несколько уровней вложенности
	// как выбирать необходимый пока не понятно...
	event.Stack().CallerSkipFrame(2).Msg(msg)
}

func (l *ARLogger) Fatal(ctx context.Context, args ...interface{}) {
	logger := l.getLoggerContext(ctx).Logger()
	l.do((&logger).Fatal(), args)
}
func (l *ARLogger) Error(ctx context.Context, args ...interface{}) {
	logger := l.getLoggerContext(ctx).Logger()
	l.do((&logger).Error(), args)
}
func (l *ARLogger) Warn(ctx context.Context, args ...interface{}) {
	logger := l.getLoggerContext(ctx).Logger()
	l.do((&logger).Warn(), args)
}
func (l *ARLogger) Info(ctx context.Context, args ...interface{}) {
	logger := l.getLoggerContext(ctx).Logger()
	l.do((&logger).Info(), args)
}
func (l *ARLogger) Debug(ctx context.Context, args ...interface{}) {
	logger := l.getLoggerContext(ctx).Logger()
	l.do((&logger).Debug(), args)
}
func (l *ARLogger) Trace(ctx context.Context, args ...interface{}) {
	logger := l.getLoggerContext(ctx).Logger()
	l.do((&logger).Trace(), args)
}

func GenerateMocker(m logger.MockerLogger) string {
	return m.MockerName + " := " + m.Mockers
}

func Generate(mockers []logger.MockerLogger) string {
	code := ""

	aggr := map[string][]string{}
	for _, mock := range mockers {
		gMock := GenerateMocker(mock)
		aggr[gMock] = aggr[mock.MockerName]
		aggr[gMock] = append(aggr[gMock], mock.FixturesSelector)
	}

	for name, fixtures := range aggr {
		code += name + "\n"
		for _, fix := range fixtures {
			code += fix + "\n"
		}
	}

	return code
}

func (l *ARLogger) CollectQueries(ctx context.Context, f func() (logger.MockerLogger, error)) {
	// ToDo Сбор данных по запросам в БД для упрощения создания списка фикстур
	if l.collectFixture {
		col, err := f()
		if err != nil {
			logger := l.getLoggerContext(ctx).Logger()
			l.do((&logger).Warn(), "Error collect mock")

			return
		}

		Collected.Lock()
		if Collected.Mocked[GenerateMocker(col)] == nil {
			Collected.Mocked[GenerateMocker(col)] = []string{col.FixturesSelector}
			Collected.Fixtures[col.ResultName] = []any{col.Results}
		} else {
			Collected.Mocked[GenerateMocker(col)] = append(Collected.Mocked[GenerateMocker(col)], col.FixturesSelector)
			Collected.Fixtures[col.ResultName] = append(Collected.Fixtures[col.ResultName], col.Results)
		}

		Collected.Unlock()
	}
}
