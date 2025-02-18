package octopus

import (
	"context"
	"fmt"

	"github.com/Educentr/go-activerecord/pkg/activerecord"
	"github.com/Educentr/go-activerecord/pkg/iproto/iproto"
)

// ToDo подумать как унести в пакет octopus но так что бы октопус на начал зависеть от activerecord
var _ iproto.Logger = IprotoLogger{}

type IprotoLogger struct{}

func (il IprotoLogger) Printf(ctx context.Context, fmtStr string, v ...interface{}) {
	ctx = activerecord.Logger().SetLoggerValueToContext(ctx, map[string]interface{}{"iproto": "client"})
	activerecord.Logger().Info(ctx, fmt.Sprintf(fmtStr, v...))
}

func (il IprotoLogger) Debugf(ctx context.Context, fmtStr string, v ...interface{}) {
	ctx = activerecord.Logger().SetLoggerValueToContext(ctx, map[string]interface{}{"iproto": "client"})
	activerecord.Logger().Debug(ctx, fmt.Sprintf(fmtStr, v...))
}
