package postgres

import (
	"github.com/Educentr/go-activerecord/v3/pkg/activerecord"
)

type OnConflictAction uint8

type Operation struct {
	Field string
	Op    activerecord.OpCode
	Value any
	// SQLSerializer   string
	// SQLDeserializer string
}

type BaseField struct {
	activerecord.BaseField
	UpdateOps []Operation
}

type UpdateParams struct {
	PK  []any
	Ops []Operation
}

const MaxLimit uint16 = 10000

type BYTEA []byte
