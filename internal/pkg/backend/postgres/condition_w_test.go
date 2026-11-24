package postgres

import (
	"errors"
	"strings"
	"testing"
	"time"

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
		{name: "time", format: "time.Time", value: "1700000000", want: "time.Unix(1700000000, 0)"},
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
