package schema

import (
	"strings"
	"testing"
)

func TestMigration_Format(t *testing.T) {
	m := &Migration{
		Number:      1,
		Description: "initial schema",
		Up: []string{
			"CREATE TABLE users (id BIGINT)",
		},
		Down: []string{
			"DROP TABLE users",
		},
	}

	formatted := m.Format()

	if !strings.Contains(formatted, "-- Migration 001: initial schema") {
		t.Error("expected formatted migration to contain header")
	}
	if !strings.Contains(formatted, "-- +up") {
		t.Error("expected formatted migration to contain -- +up")
	}
	if !strings.Contains(formatted, "-- +down") {
		t.Error("expected formatted migration to contain -- +down")
	}
	if !strings.Contains(formatted, "CREATE TABLE users (id BIGINT);") {
		t.Error("expected formatted migration to contain CREATE statement")
	}
	if !strings.Contains(formatted, "DROP TABLE users;") {
		t.Error("expected formatted migration to contain DROP statement")
	}
}

func TestMigration_Filename(t *testing.T) {
	tests := []struct {
		number      int
		description string
		want        string
	}{
		{
			number:      1,
			description: "initial schema",
			want:        "001_initial_schema.sql",
		},
		{
			number:      42,
			description: "Add user_email column",
			want:        "042_add_user_email_column.sql",
		},
		{
			number:      100,
			description: "Drop old tables!",
			want:        "100_drop_old_tables.sql",
		},
	}

	for _, tt := range tests {
		m := &Migration{
			Number:      tt.number,
			Description: tt.description,
		}
		got := m.Filename()
		if got != tt.want {
			t.Errorf("Filename() = %s, want %s", got, tt.want)
		}
	}
}

func TestMigrationBuilder_NextNumber(t *testing.T) {
	builder := NewMigrationBuilder(5)

	if got := builder.NextNumber(); got != 5 {
		t.Errorf("NextNumber() = %d, want 5", got)
	}
	if got := builder.NextNumber(); got != 6 {
		t.Errorf("NextNumber() = %d, want 6", got)
	}
	if got := builder.NextNumber(); got != 7 {
		t.Errorf("NextNumber() = %d, want 7", got)
	}
}
