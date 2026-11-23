package postgres

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
)

const (
	// Основание системы счисления, в которой пишутся числовые константы условий
	conditionNumBase = 10
	// Разрядность числовых типов
	conditionBits16 = 16
	conditionBits32 = 32
	conditionBits64 = 64
)

// Имена форматов, у которых константа условия требует своего разбора.
// Константы, а не литералы по месту: один и тот же формат упоминается и в
// разборе значения, и в тексте ошибки
const (
	boolFormat   ds.Format = "bool"
	stringFormat ds.Format = "string"
	byteaFormat  ds.Format = "postgres.BYTEA"
	timeFormat   ds.Format = "time.Time"
	uuidFormat   ds.Format = "uuid.UUID"
)

var (
	// ErrConditionValueInvalid возвращается, когда константу условия частичного
	// индекса нельзя привести к типу поля
	ErrConditionValueInvalid = errors.New("invalid index condition value")
	// ErrConditionTypeUnsupported возвращается для типов полей, у которых
	// условие частичного индекса не поддерживается
	ErrConditionTypeUnsupported = errors.New("index condition not supported for field type")
)

var (
	// Разрядность целочисленных типов
	conditionIntBits = map[ds.Format]int{
		"int16": conditionBits16,
		"int32": conditionBits32,
		"int64": conditionBits64,
	}
	// Разрядность типов с плавающей точкой
	conditionFloatBits = map[ds.Format]int{
		"float32": conditionBits32,
		"float64": conditionBits64,
	}
)

// ConditionLiteral преобразует значение условия частичного индекса, взятое из
// тега декларации (в теге это всегда строка), в литерал Go того типа, который
// объявлен у поля.
//
// Значение разбирается на этапе генерации, в сгенерированный пакет попадает
// готовая константа нужного типа. Разбор строки в рантайме здесь не подходит:
// значение условия — константа, а парсеры вида strconv.ParseInt возвращают два
// значения, то есть в одноместном контексте (элемент []any) не компилируются.
//
// Если значение нельзя привести к типу поля, возвращается ошибка с типом и
// значением — argen падает до того, как некорректная константа попадёт в
// сгенерированный код.
func ConditionLiteral(format ds.Format, value string) (string, error) {
	if bitSize, ex := conditionIntBits[format]; ex {
		return intConditionLiteral(format, value, bitSize)
	}

	if bitSize, ex := conditionFloatBits[format]; ex {
		return floatConditionLiteral(format, value, bitSize)
	}

	switch format {
	case boolFormat:
		return boolConditionLiteral(format, value)
	case stringFormat:
		return strconv.Quote(value), nil
	case byteaFormat:
		return conversionLiteral(format, strconv.Quote(value)), nil
	case timeFormat:
		return timeConditionLiteral(format, value)
	case uuidFormat:
		return uuidConditionLiteral(format, value)
	}

	return "", fmt.Errorf("%w: `%s`", ErrConditionTypeUnsupported, format)
}

func intConditionLiteral(format ds.Format, value string, bitSize int) (string, error) {
	parsed, err := strconv.ParseInt(value, conditionNumBase, bitSize)
	if err != nil {
		return "", conditionValueError(format, value, "integer")
	}

	return conversionLiteral(format, strconv.FormatInt(parsed, conditionNumBase)), nil
}

func floatConditionLiteral(format ds.Format, value string, bitSize int) (string, error) {
	parsed, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		return "", conditionValueError(format, value, "float")
	}

	return conversionLiteral(format, strconv.FormatFloat(parsed, 'g', -1, bitSize)), nil
}

func boolConditionLiteral(format ds.Format, value string) (string, error) {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return "", conditionValueError(format, value, string(boolFormat))
	}

	return strconv.FormatBool(parsed), nil
}

// timeConditionLiteral ожидает в теге число секунд unix-времени
func timeConditionLiteral(format ds.Format, value string) (string, error) {
	parsed, err := strconv.ParseInt(value, conditionNumBase, conditionBits64)
	if err != nil {
		return "", conditionValueError(format, value, "unix timestamp (seconds)")
	}

	return fmt.Sprintf("time.Unix(%d, 0)", parsed), nil
}

// uuidConditionLiteral собирает составной литерал, а не вызов конструктора
// конкретной реализации uuid: в декларации может быть импортирована любая
// библиотека, у которой uuid.UUID — это [16]byte
func uuidConditionLiteral(format ds.Format, value string) (string, error) {
	parsed, err := uuid.FromString(value)
	if err != nil {
		return "", conditionValueError(format, value, "uuid")
	}

	data := parsed.Bytes()
	octets := make([]string, 0, len(data))

	for _, octet := range data {
		octets = append(octets, fmt.Sprintf("%#02x", octet))
	}

	return fmt.Sprintf("%s{%s}", format, strings.Join(octets, ", ")), nil
}

// conversionLiteral оборачивает готовое значение в преобразование к типу поля
func conversionLiteral(format ds.Format, value string) string {
	return fmt.Sprintf("%s(%s)", format, value)
}

// conditionValueError описывает, какое значение и к какому типу не удалось
// привести
func conditionValueError(format ds.Format, value, expected string) error {
	return fmt.Errorf("%w: `%s` is not a valid %s for type `%s`", ErrConditionValueInvalid, value, expected, format)
}
