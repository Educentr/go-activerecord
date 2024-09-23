package octopus

import (
	"github.com/Educentr/go-activerecord/pkg/activerecord"
)

type (
	CountFlags uint32
	RetCode    uint32
)

type TupleData struct {
	Cnt  uint32
	Data [][]byte
}

type Ops struct {
	Field uint32
	Op    activerecord.OpCode
	Value []byte
}

type BaseField struct {
	activerecord.BaseField
	UpdateOps       []Ops
	ExtraFields     [][]byte
	FieldsetAltered bool
	Repaired        bool
}

type MutatorField struct {
	OpFunc        map[activerecord.OpCode]string
	PartialFields map[string]any
	UpdateOps     []Ops
}

type RequetsTypeType uint8

const (
	Backend          activerecord.Backend = "octopus"
	BackendTarantool activerecord.Backend = "tarantool15"
)

const (
	RequestTypeInsert RequetsTypeType = 13
	RequestTypeSelect RequetsTypeType = 17
	RequestTypeUpdate RequetsTypeType = 19
	RequestTypeDelete RequetsTypeType = 21
	RequestTypeCall   RequetsTypeType = 22
)

func (r RequetsTypeType) String() string {
	switch r {
	case RequestTypeInsert:
		return "Insert"
	case RequestTypeSelect:
		return "Select"
	case RequestTypeUpdate:
		return "Update"
	case RequestTypeDelete:
		return "Delete"
	case RequestTypeCall:
		return "Call"
	default:
		return "(unknown)"
	}
}

type InsertMode uint8

const (
	InsertModeInserOrReplace InsertMode = iota
	InsertModeInsert
	InsertModeReplace
)

const (
	SpaceLen uint32 = 4
	IndexLen
	LimitLen
	OffsetLen
	FlagsLen
	FieldNumLen
	OpsLen
	OpFieldNumLen
	OpOpLen = 1
)

type BoxMode uint8

const (
	ReplicaMaster BoxMode = iota
	MasterReplica
	ReplicaOnly
	MasterOnly
	SelectModeDefault = ReplicaMaster
)

const (
	UniqRespFlag CountFlags = 1 << iota
	NeedRespFlag
)

const (
	RcOK                   = RetCode(0x0)
	RcReadOnly             = RetCode(0x0401)
	RcLocked               = RetCode(0x0601)
	RcMemoryIssue          = RetCode(0x0701)
	RcNonMaster            = RetCode(0x0102)
	RcIllegalParams        = RetCode(0x0202)
	RcSecondaryPort        = RetCode(0x0301)
	RcBadIntegrity         = RetCode(0x0801)
	RcUnsupportedCommand   = RetCode(0x0a02)
	RcDuplicate            = RetCode(0x2002)
	RcWrongField           = RetCode(0x1e02)
	RcWrongNumber          = RetCode(0x1f02)
	RcWrongVersion         = RetCode(0x2602)
	RcWalIO                = RetCode(0x2702)
	RcDoesntExists         = RetCode(0x3102)
	RcStoredProcNotDefined = RetCode(0x3202)
	RcLuaError             = RetCode(0x3302)
	RcTupleExists          = RetCode(0x3702)
	RcDuplicateKey         = RetCode(0x3802)
)

const (
	Uint8       activerecord.Format = "uint8"
	Uint16      activerecord.Format = "uint16"
	Uint32      activerecord.Format = "uint32"
	Uint64      activerecord.Format = "uint64"
	Uint        activerecord.Format = "uint"
	Int8        activerecord.Format = "int8"
	Int16       activerecord.Format = "int16"
	Int32       activerecord.Format = "int32"
	Int64       activerecord.Format = "int64"
	Int         activerecord.Format = "int"
	String      activerecord.Format = "string"
	Bool        activerecord.Format = "bool"
	Float32     activerecord.Format = "float32"
	Float64     activerecord.Format = "float64"
	StringArray activerecord.Format = "[]string"
	ByteArray   activerecord.Format = "[]byte"
)

var UnsignedFormat = []activerecord.Format{Uint8, Uint16, Uint32, Uint64, Uint}
var NumericFormat = append(UnsignedFormat, Int8, Int16, Int32, Int64, Int)
var FloatFormat = []activerecord.Format{Float32, Float64}
var DataFormat = []activerecord.Format{String}
var AllFormat = append(append(append(
	NumericFormat,
	FloatFormat...),
	DataFormat...),
	Bool,
)
var AllProcFormat = append(append(append(
	NumericFormat,
	FloatFormat...),
	DataFormat...),
	Bool, StringArray, ByteArray,
)

func GetOpCodeName(op activerecord.OpCode) string {
	switch op {
	case activerecord.OpSet:
		return "Set"
	case activerecord.OpAdd:
		return "Add"
	case activerecord.OpAnd:
		return "And"
	case activerecord.OpXor:
		return "Xor"
	case activerecord.OpOr:
		return "Or"
	case activerecord.OpSplice:
		return "Splice"
	case activerecord.OpDelete:
		return "Delete"
	case activerecord.OpInsert:
		return "Insert"
	default:
		return "invalid opcode"
	}
}

func GetInsertModeName(mode InsertMode) string {
	switch mode {
	case InsertMode(0):
		return "InsertOrReplaceMode"
	case InsertModeInsert:
		return "InsertMode"
	case InsertModeReplace:
		return "ReplaceMode"
	default:
		return "Invalid mode"
	}
}
