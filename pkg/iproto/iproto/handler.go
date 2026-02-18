package iproto

import (
	"context"
	"fmt"
	"net"
	"sync"
)

// Handler represents IProto packets handler.
type Handler interface {
	// ServeIProto called on each incoming non-technical Packet.
	// It called with underlying channel context. It is handler responsibility to make sub context if needed.
	ServeIProto(ctx context.Context, c Conn, p Packet)
}

type HandlerFunc func(context.Context, Conn, Packet)

func (f HandlerFunc) ServeIProto(ctx context.Context, c Conn, p Packet) { f(ctx, c, p) }

// Sender represetns iproto packets sender in different forms.
type Sender interface {
	Call(ctx context.Context, message uint32, data []byte) ([]byte, error)
	Notify(ctx context.Context, message uint32, data []byte) error
	Send(ctx context.Context, packet Packet) error
}

// Closer represents channel that could be closed.
type Closer interface {
	Close()
	Shutdown()
	Done() <-chan struct{}
	OnClose(func())
}

// Conn represents channel that has ability to reply to received packets.
type Conn interface {
	Sender
	Closer

	// GetBytes obtains bytes from the Channel's byte pool.
	GetBytes(n int) []byte
	// PutBytes reclaims bytes to the Channel's byte pool.
	PutBytes(p []byte)

	RemoteAddr() net.Addr
	LocalAddr() net.Addr
}

var emptyHandler = HandlerFunc(func(context.Context, Conn, Packet) {})

type ServeMux struct {
	mu       sync.RWMutex
	handlers map[uint32]Handler
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		handlers: make(map[uint32]Handler),
	}
}

func (s *ServeMux) Handle(message uint32, handler Handler) {
	s.mu.Lock()
	if _, ok := s.handlers[message]; ok {
		panic(fmt.Sprintf("iproto: multiple handlers for %x", message))
	}

	s.handlers[message] = handler

	s.mu.Unlock()
}

func (s *ServeMux) Handler(message uint32) Handler {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if h, ok := s.handlers[message]; ok {
		return h
	}

	return emptyHandler
}

func (s *ServeMux) ServeIProto(ctx context.Context, c Conn, p Packet) {
	s.Handler(p.Header.Msg).ServeIProto(ctx, c, p)
}
