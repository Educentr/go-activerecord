package activerecord

import (
	"context"
	"hash"
)

type (
	Format  string
	Backend string
	OpCode  uint8
)

type ModelStruct interface {
	Insert(ctx context.Context) error
	Replace(ctx context.Context) error
	InsertOrReplace(ctx context.Context) error
	Update(ctx context.Context) error
	Delete(ctx context.Context) error
}

type BaseField struct {
	Collection []ModelStruct
	Objects    map[string][]ModelStruct
	Exists     bool
	ShardNum   uint32
	IsReplica  bool
	Readonly   bool
}

// BaseConnectionOptions - опции используемые для подключения
type BaseConnectionOptions struct {
	Mode           ServerModeType
	ConnectionHash hash.Hash32
	Calculated     bool
}

const (
	OpSet    OpCode = iota // Set field value
	OpAdd                  // Atomic increment
	OpAnd                  // Atomic logic values
	OpXor                  // Atomic logic xor
	OpOr                   // Atomic logic or
	OpSplice               // Atomic splice array
	OpDelete               // Atomic delete array value
	OpInsert
	OpUpdate
)
