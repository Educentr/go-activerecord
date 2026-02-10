package schema

import (
	"testing"
	"time"
)

func TestNewSchema(t *testing.T) {
	before := time.Now()
	s := NewSchema("postgres")
	after := time.Now()

	if s.Version != "1.0" {
		t.Errorf("expected version '1.0', got '%s'", s.Version)
	}

	if s.Backend != "postgres" {
		t.Errorf("expected backend 'postgres', got '%s'", s.Backend)
	}

	if s.Tables == nil {
		t.Error("expected Tables to be initialized, got nil")
	}

	if len(s.Tables) != 0 {
		t.Errorf("expected empty Tables, got %d entries", len(s.Tables))
	}

	if s.Generated.Before(before) || s.Generated.After(after) {
		t.Errorf("Generated time %v should be between %v and %v", s.Generated, before, after)
	}
}

func TestTable_ColumnByName(t *testing.T) {
	table := &Table{
		Columns: []Column{
			{Name: "id", Type: "BIGINT"},
			{Name: "email", Type: "VARCHAR(255)"},
			{Name: "status", Type: "INTEGER"},
		},
	}

	tests := []struct {
		name     string
		colName  string
		wantNil  bool
		wantType string
	}{
		{"found", "email", false, "VARCHAR(255)"},
		{"not found", "nonexistent", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := table.ColumnByName(tt.colName)
			if tt.wantNil {
				if col != nil {
					t.Errorf("expected nil, got %+v", col)
				}

				return
			}

			if col == nil {
				t.Fatal("expected non-nil column")
			}

			if col.Type != tt.wantType {
				t.Errorf("expected type %q, got %q", tt.wantType, col.Type)
			}
		})
	}

	t.Run("empty columns", func(t *testing.T) {
		emptyTable := &Table{Columns: []Column{}}
		if col := emptyTable.ColumnByName("anything"); col != nil {
			t.Errorf("expected nil for empty columns, got %+v", col)
		}
	})
}

func TestTable_IndexByName(t *testing.T) {
	table := &Table{
		Indexes: []Index{
			{Name: "pk_users", Primary: true},
			{Name: "idx_email", Unique: true},
		},
	}

	tests := []struct {
		name     string
		idxName  string
		wantNil  bool
		wantPrim bool
	}{
		{"found", "idx_email", false, false},
		{"not found", "idx_nonexistent", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := table.IndexByName(tt.idxName)
			if tt.wantNil {
				if idx != nil {
					t.Errorf("expected nil, got %+v", idx)
				}

				return
			}

			if idx == nil {
				t.Fatal("expected non-nil index")
			}
		})
	}
}

func TestTable_GetPrimaryKey(t *testing.T) {
	tests := []struct {
		name     string
		table    *Table
		wantNil  bool
		wantName string
	}{
		{
			name: "has primary",
			table: &Table{
				Indexes: []Index{
					{Name: "idx_email", Unique: true},
					{Name: "pk_users", Primary: true},
				},
			},
			wantNil:  false,
			wantName: "pk_users",
		},
		{
			name: "no primary",
			table: &Table{
				Indexes: []Index{
					{Name: "idx_email", Unique: true},
					{Name: "idx_status"},
				},
			},
			wantNil: true,
		},
		{
			name: "empty indexes",
			table: &Table{
				Indexes: []Index{},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pk := tt.table.GetPrimaryKey()
			if tt.wantNil {
				if pk != nil {
					t.Errorf("expected nil, got %+v", pk)
				}

				return
			}

			if pk == nil {
				t.Fatal("expected non-nil primary key")
			}

			if pk.Name != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, pk.Name)
			}
		})
	}
}
