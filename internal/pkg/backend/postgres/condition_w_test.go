package postgres

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
)

func TestConditionLiteral(t *testing.T) {
	tests := []struct {
		name   string
		format ds.Format
		value  string
		want   string
	}{
		{name: "int16", format: "int16", value: "0", want: "int16(0)"},
		{name: "int32", format: "int32", value: "-42", want: "int32(-42)"},
		{name: "int64", format: "int64", value: "0", want: "int64(0)"},
		{name: "int64 max", format: "int64", value: "9223372036854775807", want: "int64(9223372036854775807)"},
		{name: "float32", format: "float32", value: "0.5", want: "float32(0.5)"},
		{name: "float64", format: "float64", value: "1e-3", want: "float64(0.001)"},
		{name: "float64 integral", format: "float64", value: "5", want: "float64(5)"},
		{name: "bool true", format: "bool", value: "true", want: "true"},
		{name: "bool false", format: "bool", value: "false", want: "false"},
		{name: "bool numeric", format: "bool", value: "1", want: "true"},
		{name: "string", format: "string", value: "active", want: `"active"`},
		{name: "string empty", format: "string", value: "", want: `""`},
		{name: "string with quote", format: "string", value: `a"b`, want: `"a\"b"`},
		{name: "bytea", format: "postgres.BYTEA", value: "raw", want: `postgres.BYTEA("raw")`},
		{name: "time", format: "time.Time", value: "1700000000", want: "time.Unix(1700000000, 0).UTC()"},
		{
			name:   "uuid",
			format: "uuid.UUID",
			value:  "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			want:   "uuid.UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConditionLiteral(tt.format, tt.value)
			if err != nil {
				t.Fatalf("ConditionLiteral(%q, %q) unexpected error: %s", tt.format, tt.value, err)
			}

			if got != tt.want {
				t.Errorf("ConditionLiteral(%q, %q) = %q, want %q", tt.format, tt.value, got, tt.want)
			}
		})
	}
}

func TestConditionLiteralError(t *testing.T) {
	tests := []struct {
		name    string
		format  ds.Format
		value   string
		wantErr error
	}{
		{name: "not a number", format: "int64", value: "abc", wantErr: ErrConditionValueInvalid},
		{name: "float in int field", format: "int64", value: "0.5", wantErr: ErrConditionValueInvalid},
		{name: "int16 overflow", format: "int16", value: "40000", wantErr: ErrConditionValueInvalid},
		{name: "empty int", format: "int64", value: "", wantErr: ErrConditionValueInvalid},
		{name: "not a float", format: "float64", value: "abc", wantErr: ErrConditionValueInvalid},
		{name: "not a bool", format: "bool", value: "yes", wantErr: ErrConditionValueInvalid},
		{name: "not a timestamp", format: "time.Time", value: "2024-01-01", wantErr: ErrConditionValueInvalid},
		{name: "not a uuid", format: "uuid.UUID", value: "zzz", wantErr: ErrConditionValueInvalid},
		{name: "unknown type", format: "complex128", value: "0", wantErr: ErrConditionTypeUnsupported},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConditionLiteral(tt.format, tt.value)
			if err == nil {
				t.Fatalf("ConditionLiteral(%q, %q) = %q, want error", tt.format, tt.value, got)
			}

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ConditionLiteral(%q, %q) error = %v, want %v", tt.format, tt.value, err, tt.wantErr)
			}

			// В сообщении должен быть тип поля, а для неподходящей константы —
			// и она сама: по сообщению автор декларации должен понять, что
			// именно поправить в теге
			if !strings.Contains(err.Error(), string(tt.format)) {
				t.Errorf("ConditionLiteral(%q, %q) error = %q, want it to mention the field type", tt.format, tt.value, err)
			}

			if errors.Is(err, ErrConditionValueInvalid) && tt.value != "" && !strings.Contains(err.Error(), tt.value) {
				t.Errorf("ConditionLiteral(%q, %q) error = %q, want it to mention the value", tt.format, tt.value, err)
			}
		})
	}
}

func TestConditionSQLLiteral(t *testing.T) {
	tests := []struct {
		name   string
		format ds.Format
		value  string
		want   string
	}{
		{name: "int64 as is", format: "int64", value: "42", want: "42"},
		{name: "int64 negative", format: "int64", value: "-42", want: "-42"},
		{name: "float64 as is", format: "float64", value: "0.5", want: "0.5"},
		{name: "bool as is", format: "bool", value: "true", want: "true"},
		{name: "string quoted", format: "string", value: "active", want: "'active'"},
		{name: "string with quote escaped", format: "string", value: "o'hara", want: "'o''hara'"},
		{name: "bytea quoted", format: "postgres.BYTEA", value: "abc", want: "'abc'"},
		{name: "uuid quoted", format: "uuid.UUID", value: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", want: "'6ba7b810-9dad-11d1-80b4-00c04fd430c8'"},
		{name: "time as timestamp literal in utc", format: "time.Time", value: "1700000000", want: "TIMESTAMP '2023-11-14 22:13:20'"},
		{name: "time epoch", format: "time.Time", value: "0", want: "TIMESTAMP '1970-01-01 00:00:00'"},
		// Формат, которого нет в type.go: значение отдаётся как есть, как и до
		// появления ConditionSQLLiteral
		{name: "unknown format as is", format: "complex128", value: "0", want: "0"},
		// Значение проверяет checker до генерации схемы, но на мусоре функция
		// не должна собирать битый TIMESTAMP-литерал
		{name: "time with unparsable value as is", format: "time.Time", value: "2023-11-14", want: "2023-11-14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConditionSQLLiteral(tt.format, tt.value); got != tt.want {
				t.Errorf("ConditionSQLLiteral(%q, %q) = %q, want %q", tt.format, tt.value, got, tt.want)
			}
		})
	}
}

// Литерал времени в DDL не зависит от TZ процесса: предикат частичного индекса
// обязан быть IMMUTABLE, значение в теге — unix-секунды в UTC
func TestConditionSQLLiteralTimeIndependentOfLocalZone(t *testing.T) {
	saved := time.Local

	defer func() { time.Local = saved }()

	const want = "TIMESTAMP '2023-11-14 22:13:20'"

	for _, name := range []string{"UTC", "Europe/Moscow", "America/Los_Angeles"} {
		loc, err := time.LoadLocation(name)
		if err != nil {
			t.Skipf("timezone database is not available: %s", err)
		}

		time.Local = loc

		if got := ConditionSQLLiteral(timeFormat, "1700000000"); got != want {
			t.Errorf("ConditionSQLLiteral with TZ %s = %q, want %q", name, got, want)
		}
	}
}

// Литерал условия для time.Time в index.go и TIMESTAMP-литерал предиката в DDL
// обязаны означать один и тот же wall clock при любом TZ процесса: pgx при
// записи time.Time в колонку TIMESTAMP отбрасывает зону, сохраняя wall clock
// как есть, а предикат частичного индекса в DDL всегда в UTC. Расходятся —
// планировщик не может применить частичный индекс к запросу, а выборка
// отбирает не тот диапазон, что описан в декларации
func TestConditionLiteralTimeAgreesWithSQLLiteral(t *testing.T) {
	saved := time.Local

	defer func() { time.Local = saved }()

	for _, zone := range []string{"UTC", "Europe/Moscow", "America/Los_Angeles"} {
		loc, err := time.LoadLocation(zone)
		if err != nil {
			t.Skipf("timezone database is not available: %s", err)
		}

		time.Local = loc

		for _, value := range []string{"1700000000", "0", "-1"} {
			literal, err := ConditionLiteral(timeFormat, value)
			if err != nil {
				t.Fatalf("ConditionLiteral(%q, %q) unexpected error: %s", timeFormat, value, err)
			}

			// то же значение time.Time, которое получит сгенерированный index.go
			runtimeValue := evalTimeConditionLiteral(t, literal)

			stored := "TIMESTAMP '" + encodeTimestampAsPgx(t, runtimeValue) + "'"
			if want := ConditionSQLLiteral(timeFormat, value); stored != want {
				t.Errorf("TZ %s, condition `%s`: index.go literal %s пишется в TIMESTAMP как %s, в DDL предикате %s",
					zone, value, literal, stored, want)
			}
		}
	}
}

// evalTimeConditionLiteral вычисляет литерал, который ConditionLiteral отдаёт
// для time.Time: разбирает его как выражение Go и применяет те же вызовы, что
// попали в сгенерированный код. Тест сверяет не текст литерала, а момент,
// который этот литерал означает в рантайме, — пропавший `.UTC()` изменит
// результат, а не только текст
func evalTimeConditionLiteral(t *testing.T, literal string) time.Time {
	t.Helper()

	expr, err := parser.ParseExpr(literal)
	if err != nil {
		t.Fatalf("condition literal `%s` is not a valid go expression: %s", literal, err)
	}

	toUTC := false

	if call, ok := expr.(*ast.CallExpr); ok {
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "UTC" {
			toUTC = true
			expr = sel.X
		}
	}

	sec, nsec := timeUnixArgs(t, literal, expr)

	value := time.Unix(sec, nsec)
	if toUTC {
		value = value.UTC()
	}

	return value
}

// timeUnixArgs проверяет, что выражение — вызов time.Unix, и отдаёт его аргументы
func timeUnixArgs(t *testing.T, literal string, expr ast.Expr) (int64, int64) {
	t.Helper()

	call, ok := expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("condition literal `%s` is not a time.Unix call", literal)
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		t.Fatalf("condition literal `%s` is not a time.Unix call", literal)
	}

	if pkg, ok := sel.X.(*ast.Ident); !ok || pkg.Name != "time" || sel.Sel.Name != "Unix" {
		t.Fatalf("condition literal `%s` is not a time.Unix call", literal)
	}

	if len(call.Args) != 2 {
		t.Fatalf("condition literal `%s`: time.Unix takes 2 arguments, got %d", literal, len(call.Args))
	}

	return intLiteralArg(t, literal, call.Args[0]), intLiteralArg(t, literal, call.Args[1])
}

// intLiteralArg разбирает целочисленный аргумент литерала, в том числе
// отрицательный (в AST это унарный минус, а не часть константы)
func intLiteralArg(t *testing.T, literal string, arg ast.Expr) int64 {
	t.Helper()

	sign := int64(1)

	if unary, ok := arg.(*ast.UnaryExpr); ok && unary.Op == token.SUB {
		sign = -1
		arg = unary.X
	}

	lit, ok := arg.(*ast.BasicLit)
	if !ok || lit.Kind != token.INT {
		t.Fatalf("condition literal `%s`: time.Unix argument is not an integer literal", literal)
	}

	parsed, err := strconv.ParseInt(lit.Value, conditionNumBase, conditionBits64)
	if err != nil {
		t.Fatalf("condition literal `%s`: cannot parse time.Unix argument `%s`: %s", literal, lit.Value, err)
	}

	return sign * parsed
}

// encodeTimestampAsPgx отдаёт wall clock, который pgx отправит в колонку
// TIMESTAMP для этого значения. Нормализацию (отбрасывание зоны,
// pgtype/timestamp.go, discardTimeZone) делает сам драйвер: повторять её в
// тесте — значит проверять копию, а не поведение pgx
func encodeTimestampAsPgx(t *testing.T, value time.Time) string {
	t.Helper()

	encoded, err := pgtype.NewMap().Encode(pgtype.TimestampOID, pgtype.TextFormatCode, value, nil)
	if err != nil {
		t.Fatalf("pgx cannot encode %s as TIMESTAMP: %s", value, err)
	}

	return string(encoded)
}
