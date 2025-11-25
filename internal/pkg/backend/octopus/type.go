package octopus

import (
	"log"
	"strings"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
)

type FormatType struct {
	TypeName ds.Format
	Name     string
	minValue string
	maxValue string

	Default        []byte
	packFunc       string
	unpackFunc     string
	packConvFunc   string
	UnpackConvFunc string
	unpackType     string
	len            uint
	lenFunc        func(uint32) uint32
	tostr          string
}

func (p FormatType) PackConvFunc(fieldname string) string {
	if p.packConvFunc != "" {
		return p.packConvFunc + "(" + fieldname + ")"
	}

	return fieldname
}

func (p FormatType) UnpackFunc() string {
	if p.unpackFunc != "" {
		return p.unpackFunc
	}

	return "iproto.Unpack" + p.Name
}

func (p FormatType) PackFunc() string {
	if p.packFunc != "" {
		return p.packFunc
	}

	return "iproto.Pack" + p.Name
}

func (p FormatType) StringDeserializer() []string {
	log.Fatal("StringDeserializer not implemented")

	return nil
}

func (p FormatType) DefaultValue() string {
	fname := "iproto.Pack" + p.Name
	if p.packFunc != "" {
		fname = p.packFunc
	}

	switch {
	case strings.HasPrefix(p.Name, "Uint"):
		return fname + "([]byte{}, 0, iproto.ModeDefault)"
	case strings.HasPrefix(p.Name, "String"):
		return fname + `([]byte{}, "", iproto.ModeDefault)`
	default:
		return "can't detect type"
	}
}

func (p FormatType) UnpackType() string {
	if p.unpackType != "" {
		return p.unpackType
	} else if p.packConvFunc != "" {
		return p.packConvFunc
	}

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

	return "math.Min" + p.Name
}

func (p FormatType) MaxValue() string {
	if p.maxValue != "" {
		return p.maxValue
	}

	return "math.Max" + p.Name
}

func (p FormatType) ToString() []string {
	return strings.SplitN(p.tostr, `%%`, 2)
}

func (p FormatType) DBType() string {
	// octopus не использует SQL типы, возвращаем пустую строку
	return ""
}

func (p FormatType) MutatorTypeConv() string {
	if p.UnpackConvFunc != "" {
		return p.UnpackConvFunc
	}

	return ToLower.String(p.Name)
}
