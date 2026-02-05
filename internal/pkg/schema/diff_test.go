package schema

import (
	"testing"
)

func TestComputeDiff_NilOldSchema(t *testing.T) {
	newSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name: "users",
				Columns: []Column{
					{Name: "id", Type: "BIGINT"},
					{Name: "name", Type: "VARCHAR(255)"},
				},
			},
		},
	}

	diff := ComputeDiff(nil, newSchema)

	if len(diff.AddedTables) != 1 {
		t.Errorf("expected 1 added table, got %d", len(diff.AddedTables))
	}
	if diff.AddedTables[0] != "users" {
		t.Errorf("expected added table 'users', got '%s'", diff.AddedTables[0])
	}
}

func TestComputeDiff_DroppedTable(t *testing.T) {
	oldSchema := &Schema{
		Tables: map[string]Table{
			"users":    {Name: "users"},
			"products": {Name: "products"},
		},
	}
	newSchema := &Schema{
		Tables: map[string]Table{
			"users": {Name: "users"},
		},
	}

	diff := ComputeDiff(oldSchema, newSchema)

	if len(diff.DroppedTables) != 1 {
		t.Errorf("expected 1 dropped table, got %d", len(diff.DroppedTables))
	}
	if diff.DroppedTables[0] != "products" {
		t.Errorf("expected dropped table 'products', got '%s'", diff.DroppedTables[0])
	}
}

func TestComputeDiff_AddedColumn(t *testing.T) {
	oldSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name: "users",
				Columns: []Column{
					{Name: "id", Type: "BIGINT"},
				},
			},
		},
	}
	newSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name: "users",
				Columns: []Column{
					{Name: "id", Type: "BIGINT"},
					{Name: "email", Type: "VARCHAR(255)"},
				},
			},
		},
	}

	diff := ComputeDiff(oldSchema, newSchema)

	if len(diff.ModifiedTables) != 1 {
		t.Errorf("expected 1 modified table, got %d", len(diff.ModifiedTables))
	}
	tableDiff := diff.ModifiedTables["users"]
	if len(tableDiff.AddedColumns) != 1 {
		t.Errorf("expected 1 added column, got %d", len(tableDiff.AddedColumns))
	}
	if tableDiff.AddedColumns[0].Name != "email" {
		t.Errorf("expected added column 'email', got '%s'", tableDiff.AddedColumns[0].Name)
	}
}

func TestComputeDiff_DroppedColumn(t *testing.T) {
	oldSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name: "users",
				Columns: []Column{
					{Name: "id", Type: "BIGINT"},
					{Name: "deprecated_field", Type: "VARCHAR(255)"},
				},
			},
		},
	}
	newSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name: "users",
				Columns: []Column{
					{Name: "id", Type: "BIGINT"},
				},
			},
		},
	}

	diff := ComputeDiff(oldSchema, newSchema)

	tableDiff := diff.ModifiedTables["users"]
	if len(tableDiff.DroppedColumns) != 1 {
		t.Errorf("expected 1 dropped column, got %d", len(tableDiff.DroppedColumns))
	}
	if tableDiff.DroppedColumns[0] != "deprecated_field" {
		t.Errorf("expected dropped column 'deprecated_field', got '%s'", tableDiff.DroppedColumns[0])
	}
}

func TestComputeDiff_ModifiedColumn(t *testing.T) {
	oldSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name: "users",
				Columns: []Column{
					{Name: "name", Type: "VARCHAR(100)"},
				},
			},
		},
	}
	newSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name: "users",
				Columns: []Column{
					{Name: "name", Type: "VARCHAR(255)"},
				},
			},
		},
	}

	diff := ComputeDiff(oldSchema, newSchema)

	tableDiff := diff.ModifiedTables["users"]
	if len(tableDiff.ModifiedColumns) != 1 {
		t.Errorf("expected 1 modified column, got %d", len(tableDiff.ModifiedColumns))
	}
	colDiff := tableDiff.ModifiedColumns["name"]
	if !colDiff.TypeChanged {
		t.Error("expected TypeChanged to be true")
	}
}

func TestComputeDiff_AddedIndex(t *testing.T) {
	oldSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name:    "users",
				Indexes: []Index{},
			},
		},
	}
	newSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name: "users",
				Indexes: []Index{
					{Name: "idx_users_email", Columns: []string{"email"}, Unique: true},
				},
			},
		},
	}

	diff := ComputeDiff(oldSchema, newSchema)

	tableDiff := diff.ModifiedTables["users"]
	if len(tableDiff.AddedIndexes) != 1 {
		t.Errorf("expected 1 added index, got %d", len(tableDiff.AddedIndexes))
	}
	if tableDiff.AddedIndexes[0].Name != "idx_users_email" {
		t.Errorf("expected added index 'idx_users_email', got '%s'", tableDiff.AddedIndexes[0].Name)
	}
}

func TestComputeDiff_DroppedIndex(t *testing.T) {
	oldSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name: "users",
				Indexes: []Index{
					{Name: "idx_users_old", Columns: []string{"old_field"}},
				},
			},
		},
	}
	newSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name:    "users",
				Indexes: []Index{},
			},
		},
	}

	diff := ComputeDiff(oldSchema, newSchema)

	tableDiff := diff.ModifiedTables["users"]
	if len(tableDiff.DroppedIndexes) != 1 {
		t.Errorf("expected 1 dropped index, got %d", len(tableDiff.DroppedIndexes))
	}
	if tableDiff.DroppedIndexes[0] != "idx_users_old" {
		t.Errorf("expected dropped index 'idx_users_old', got '%s'", tableDiff.DroppedIndexes[0])
	}
}

func TestSchemaDiff_IsEmpty(t *testing.T) {
	empty := &SchemaDiff{
		AddedTables:    []string{},
		DroppedTables:  []string{},
		ModifiedTables: make(map[string]TableDiff),
	}

	if !empty.IsEmpty() {
		t.Error("expected empty diff to return true for IsEmpty()")
	}

	notEmpty := &SchemaDiff{
		AddedTables:    []string{"users"},
		DroppedTables:  []string{},
		ModifiedTables: make(map[string]TableDiff),
	}

	if notEmpty.IsEmpty() {
		t.Error("expected non-empty diff to return false for IsEmpty()")
	}
}

func TestTableDiff_IsEmpty(t *testing.T) {
	empty := TableDiff{
		AddedColumns:    []Column{},
		DroppedColumns:  []string{},
		ModifiedColumns: make(map[string]ColumnDiff),
		AddedIndexes:    []Index{},
		DroppedIndexes:  []string{},
		ModifiedIndexes: make(map[string]IndexDiff),
	}

	if !empty.IsEmpty() {
		t.Error("expected empty table diff to return true for IsEmpty()")
	}

	notEmpty := TableDiff{
		AddedColumns:    []Column{{Name: "new_col"}},
		DroppedColumns:  []string{},
		ModifiedColumns: make(map[string]ColumnDiff),
		AddedIndexes:    []Index{},
		DroppedIndexes:  []string{},
		ModifiedIndexes: make(map[string]IndexDiff),
	}

	if notEmpty.IsEmpty() {
		t.Error("expected non-empty table diff to return false for IsEmpty()")
	}
}
