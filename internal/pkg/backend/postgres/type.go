package postgres

import (
	"fmt"
	"log"
	"strings"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
)

type Mutator struct {
	Name          string
	sqlSerializer string
}

func (m Mutator) DBSerializer() string {
	if m.sqlSerializer != "" {
		return m.sqlSerializer
	}

	return ""
}

type MutatorParam struct {
	Name          string
	AvailableType []FormatType
	ArgType       string
}

// ToDo перенести в описание форматов (type.go)
var MutatorMapper = map[string]MutatorParam{
	ds.IncMutator: {Name: "Inc", AvailableType: NumericFormatT},
	ds.DecMutator: {Name: "Dec", AvailableType: NumericFormatT},
}

type FormatType struct {
	TypeName   ds.Format
	Name       string
	DBTypeName string
	origType   string
	minValue   string
	maxValue   string

	len               uint
	lenFunc           func(uint32) uint32
	sqlSerializer     string
	sqlDeserializer   string
	availableMutators []Mutator
	fromstr           string
}

func (p FormatType) OrigType() string {
	if p.origType != "" {
		return p.origType
	}

	return ""
}

// ToDo Перенести метод в интерфейс вместе с переделкой мутаторов в octopus-е
func (p FormatType) GetMutatorByName(name string) Mutator {
	for _, m := range p.availableMutators {
		if m.Name == name {
			return m
		}
	}

	log.Fatalf("Mutator %s not available for type %s", name, p.Name)

	return Mutator{}
}

func (p FormatType) PackConvFunc(fieldname string) string {
	log.Fatal("PackConvFunc not implemented")

	return ""
}

func (p FormatType) StringDeserializer() []string {
	if p.fromstr != "" {
		return strings.SplitN(p.fromstr, `%%`, 2)
	}

	return []string{}
}

func (p FormatType) UnpackFunc() string {
	if p.sqlDeserializer != "" {
		return p.sqlDeserializer
	}

	return ""
}

func (p FormatType) PackFunc() string {
	if p.sqlSerializer != "" {
		return p.sqlSerializer
	}

	return ""
}

func (p FormatType) DefaultValue() string {
	log.Fatal("please use DEFAULT placeholder into postgres")

	return ""
}

func (p FormatType) UnpackType() string {
	log.Fatal("UnpackType not implemented")

	return ""
}

func (p FormatType) Len(l uint32) uint {
	if p.len != 0 {
		return p.len
	}

	return uint(p.lenFunc(l))
}

func (p FormatType) MinValue() string {
	if p.minValue != "" {
		return p.minValue
	}

	// ToDo унести из рантайма в генератор
	return "math.Min" + p.Name
}

func (p FormatType) MaxValue() string {
	if p.maxValue != "" {
		return p.maxValue
	}

	// ToDo унести из рантайма в генератор
	return "math.Max" + p.Name
}

func (p FormatType) ToString() []string {
	log.Fatal("ToString not implemented")

	return nil
}

func (p FormatType) MutatorTypeConv() string {
	log.Fatal("MutatorTypeConv not implemented")

	return ""
}

// ToDo numeric or numeric(p,s)
// ByteArray   Format = "[]byte"
var NumericFormatT = []FormatType{
	{TypeName: "int16", Name: "Int16", len: 3, minValue: "math.MinInt16", maxValue: "math.MaxInt16", fromstr: "int16(strconv.ParseInt(%%, 10, 16))", DBTypeName: "SMALLINT"},
	{TypeName: "int32", Name: "Int32", len: 5, minValue: "math.MinInt32", maxValue: "math.MaxInt32", fromstr: "int32(strconv.ParseInt(%%, 10, 32))", DBTypeName: "INTEGER"},
	{TypeName: "int64", Name: "Int64", len: 9, minValue: "math.MinInt64", maxValue: "math.MaxInt64", fromstr: "int64(strconv.ParseInt(%%, 10, 64))", DBTypeName: "BIGINT"},
}
var FloatFormatT = []FormatType{
	{TypeName: "float32", Name: "Float32", len: 5, minValue: "math.MinFloat32", maxValue: "math.MaxFloat32", fromstr: "float32(strconv.ParseFloat(%%, 32))", DBTypeName: "REAL"},
	{TypeName: "float64", Name: "Float64", len: 9, minValue: "math.MinFloat64", maxValue: "math.MaxFloat64", fromstr: "float64(strconv.ParseFloat(%%, 64))", DBTypeName: "DOUBLE PRECISION"},
}
var DataFormatT = []FormatType{
	// ToDo тип в бд может быть разный CHARACTER(n) or CHAR(n) VARYING(n) VARCHAR(n) TEXT
	{TypeName: "string", Name: "String", minValue: "0", maxValue: "4096", DBTypeName: "VARCHAR(n)"},
	{TypeName: "postgres.BYTEA", Name: "Bytea", minValue: "0", maxValue: "65545", DBTypeName: "BYTEA"},
}

var DateFormatT = []FormatType{
	{TypeName: "time.Time", Name: "DateTime", len: 9, minValue: "math.MinInt64", maxValue: "math.MaxInt64", fromstr: "time.Unix(strconv.ParseInt(%%, 10, 64), 0)", DBTypeName: "TIMESTAMP", availableMutators: []Mutator{{Name: ds.IncMutator, sqlSerializer: "INTERVAL '$%d sec'"}}},
}

var AllFormatT = append(append(append(append(
	NumericFormatT,
	FloatFormatT...),
	DataFormatT...),
	DateFormatT...),
	FormatType{TypeName: "bool", Name: "Uint8", len: 2, fromstr: "strconv.ParseBool(%%)", DBTypeName: "BOOLEAN"},
	FormatType{TypeName: "uuid.UUID", Name: "UUID", len: 16, fromstr: "uuid.Parse(%%)", DBTypeName: "UUID"},
)

func GetFormat(format ds.Format) (FormatType, error) {
	for _, f := range AllFormatT {
		if f.TypeName == format {
			return f, nil
		}
	}

	return FormatType{}, fmt.Errorf("packer for type `%s` not found", format)
}
