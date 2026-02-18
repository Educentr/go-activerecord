package octopus

import (
	"embed"
	"log"
	"text/template"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/schema"
	octopusPkg "github.com/Educentr/go-activerecord/v3/pkg/octopus"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type BackendGenerator struct {
	availFormat     map[ds.Format]struct{}
	procAvailFormat map[ds.Format]struct{}
}

func (b *BackendGenerator) Init() {
	b.availFormat = make(map[ds.Format]struct{}, len(AllFormatT))
	b.procAvailFormat = make(map[ds.Format]struct{}, len(AllProcFormatT))

	for _, form := range AllFormatT {
		b.availFormat[form.TypeName] = struct{}{}
	}

	for _, form := range AllProcFormatT {
		b.procAvailFormat[form.TypeName] = struct{}{}
	}
}

func (b BackendGenerator) Name() ds.Backend {
	return "octopus"
}

func (b BackendGenerator) Aliases() []ds.Backend {
	return []ds.Backend{"tarantool15"}
}

func (b BackendGenerator) Templates() embed.FS {
	return Templates
}

func (b BackendGenerator) TemplatePath() string {
	return "tmpl/pkg"
}

func (b BackendGenerator) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"packerParam": func(format ds.Format) ds.FormatParam {
			for _, f := range AllFormatT {
				if f.TypeName == format {
					return f
				}
			}

			log.Fatalf("packer for type `%s` not found", format)

			return nil
		},
		"mutatorParam": func(mut string, format ds.Format) MutatorParam {
			ret, ex := MutatorMapper[mut]
			if !ex {
				log.Fatalf("mutator packer for type `%s` not found", format)
			}

			for _, availFormat := range ret.AvailableType {
				if availFormat.TypeName == format {
					return ret
				}
			}

			log.Fatalf("Mutator `%s` not available for type `%s`", mut, format)

			return MutatorParam{}
		},
	}
}

// SchemaGenerator возвращает генератор конфигурации для Octopus
func (b BackendGenerator) SchemaGenerator() schema.Generator {
	return NewConfigGenerator()
}

//go:embed tmpl/pkg/*
var Templates embed.FS

var ToLower = cases.Title(language.English)

type MutatorParam struct {
	Name          string
	AvailableType []FormatType
	ArgType       string
}

// ToDo перенести в описание форматов (type.go)
var MutatorMapper = map[string]MutatorParam{
	ds.IncMutator:      {Name: "Inc", AvailableType: NumericFormatT},
	ds.DecMutator:      {Name: "Dec", AvailableType: NumericFormatT},
	ds.AndMutator:      {Name: "And", AvailableType: UnsignedFormatT},
	ds.OrMutator:       {Name: "Or", AvailableType: UnsignedFormatT},
	ds.XorMutator:      {Name: "Xor", AvailableType: UnsignedFormatT},
	ds.ClearBitMutator: {Name: "ClearBit", AvailableType: NumericFormatT},
	ds.SetBitMutator:   {Name: "SetBit", AvailableType: NumericFormatT},
}

var UnsignedFormatT = []FormatType{
	{TypeName: "uint8", Name: "Uint8", len: 2, tostr: "strconv.FormatUint(uint64(%%), 10)"},
	{TypeName: "uint16", Name: "Uint16", len: 3, tostr: "strconv.FormatUint(uint64(%%), 10)"},
	{TypeName: "uint32", Name: "Uint32", len: 5, tostr: "strconv.FormatUint(uint64(%%), 10)"},
	{TypeName: "uint64", Name: "Uint64", len: 9, tostr: "strconv.FormatUint(%%, 10)"},
	{TypeName: "uint", Name: "Uint32", len: 5, tostr: "strconv.FormatUint(uint64(%%), 10)", packConvFunc: "uint32", UnpackConvFunc: "uint"},
}

var NumericFormatT = append(UnsignedFormatT,
	FormatType{TypeName: "int8", Name: "Uint8", len: 2, tostr: "strconv.FormatInt(int64(%%), 10)", packConvFunc: "uint8", UnpackConvFunc: "int8", minValue: "math.MinInt8", maxValue: "math.MaxInt8"},
	FormatType{TypeName: "int16", Name: "Uint16", len: 3, tostr: "strconv.FormatInt(int64(%%), 10)", packConvFunc: "uint16", UnpackConvFunc: "int16", minValue: "math.MinInt16", maxValue: "math.MaxInt16"},
	FormatType{TypeName: "int32", Name: "Uint32", len: 5, tostr: "strconv.FormatInt(int64(%%), 10)", packConvFunc: "uint32", UnpackConvFunc: "int32", minValue: "math.MinInt32", maxValue: "math.MaxInt32"},
	FormatType{TypeName: "int64", Name: "Uint64", len: 9, tostr: "strconv.FormatInt(%%, 10)", packConvFunc: "uint64", UnpackConvFunc: "int64", minValue: "math.MinInt64", maxValue: "math.MaxInt64"},
	FormatType{TypeName: "int", Name: "Uint32", len: 5, tostr: "strconv.FormatInt(int64(%%), 10)", packConvFunc: "uint32", UnpackConvFunc: "int", minValue: "math.MinInt32", maxValue: "math.MaxInt32"},
)

var FloatFormatT = []FormatType{
	{TypeName: "float32", Name: "Uint32", len: 5, tostr: "strconv.FormatFloat(%%, 32)", packConvFunc: "math.Float32bits", UnpackConvFunc: "math.Float32frombits", unpackType: "uint32", minValue: "math.MinFloat32", maxValue: "math.MaxFloat32"},
	{TypeName: "float64", Name: "Uint64", len: 9, tostr: "strconv.FormatFloat(%%, 64)", packConvFunc: "math.Float64bits", UnpackConvFunc: "math.Float64frombits", unpackType: "uint64", minValue: "math.MinFloat64", maxValue: "math.MaxFloat64"},
}

var DataFormatT = []FormatType{
	{
		TypeName: "string", Name: "String", tostr: " %% ", lenFunc: octopusPkg.ByteLen,
		packFunc: "octopus.PackString", unpackFunc: "octopus.UnpackString", omitModeParam: true,
		minValue: "0", maxValue: "4096", unpackType: "string",
	},
}

var AllFormatT = append(append(append(
	NumericFormatT,
	FloatFormatT...),
	DataFormatT...),
	FormatType{TypeName: "bool", Name: "Uint8", len: 2, tostr: "strconv.FormatBool(%%)", packConvFunc: "octopus.BoolToUint", UnpackConvFunc: "octopus.UintToBool", unpackType: "uint8"},
)

var AllProcFormatT = append(AllFormatT,
	// ToDo Очень странные типы, кажется, что не будут работать пакеры и что то еще. Надо проверить и либо доделать либо переосмыслить
	FormatType{TypeName: "[]string", Name: "StringArray"},
	FormatType{TypeName: "[]byte", Name: "ByteArray"},
)
