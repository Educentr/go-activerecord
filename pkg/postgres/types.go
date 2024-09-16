package postgres

import "github.com/mailru/activerecord/pkg/activerecord"

// ToDo numeric or numeric(p,s)
// ToDo DATE TIME TIMESTAMP TIMESTAMPTZ INTERVAL
const (
	Bool    activerecord.Format = "bool"    // BOOLEAN
	String  activerecord.Format = "string"  // CHARACTER(n) or CHAR(n) VARYING(n) VARCHAR(n) TEXT
	Int16   activerecord.Format = "int16"   // SMALLINT
	Int32   activerecord.Format = "int32"   // INTEGER
	Int64   activerecord.Format = "int64"   // BIGINT
	Float32 activerecord.Format = "float32" // real float8
	Float64 activerecord.Format = "float64" // float(n)
// ByteArray   Format = "[]byte"
)

var NumericFormat = []activerecord.Format{Int16, Int32, Int64}
var FloatFormat = []activerecord.Format{Float32, Float64}
var DataFormat = []activerecord.Format{String}
var AllFormat = append(append(append(
	NumericFormat,
	FloatFormat...),
	DataFormat...),
	Bool,
)

const MaxLimit uint16 = 10000
