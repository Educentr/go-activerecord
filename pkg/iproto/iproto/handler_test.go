package iproto_test

import (
	"context"
	"testing"
	"time"

	"github.com/Educentr/go-activerecord/v3/pkg/iproto/iproto"
	"github.com/Educentr/go-activerecord/v3/pkg/iproto/iproto/internal/testutil"
)

func handler(ctx context.Context, rw iproto.Conn, pkt iproto.Packet) {
	_ = rw.Send(ctx, iproto.ResponseTo(pkt, nil))
}

func benchmarkHandler(b *testing.B, h iproto.Handler) {
	var p iproto.Packet
	rw := testutil.NewFakeResponseWriter()
	rw.DoSend = func(context.Context, iproto.Packet) error {
		// emulate some work
		time.Sleep(time.Microsecond)
		return nil
	}
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		h.ServeIProto(ctx, rw, p)
	}
}

func BenchmarkHandlerPlain(b *testing.B) {
	benchmarkHandler(b, iproto.HandlerFunc(handler))
}
