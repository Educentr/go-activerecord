package schema

import (
	"sort"
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

func TestComputeColumnDiff(t *testing.T) {
	base := Column{Name: "col", Type: "BIGINT", Size: 0, Default: "0", NotNull: true}

	tests := []struct {
		name           string
		old, cur       Column
		wantNil        bool
		typeChanged    bool
		sizeChanged    bool
		defaultChanged bool
		notNullChanged bool
	}{
		{
			name:    "no changes",
			old:     base,
			cur:     base,
			wantNil: true,
		},
		{
			name:        "type changed",
			old:         base,
			cur:         Column{Name: "col", Type: "INTEGER", Size: 0, Default: "0", NotNull: true},
			typeChanged: true,
		},
		{
			name:        "size changed",
			old:         Column{Name: "col", Type: "VARCHAR(100)", Size: 100, Default: "''", NotNull: true},
			cur:         Column{Name: "col", Type: "VARCHAR(100)", Size: 255, Default: "''", NotNull: true},
			sizeChanged: true,
		},
		{
			name:           "default changed",
			old:            base,
			cur:            Column{Name: "col", Type: "BIGINT", Size: 0, Default: "42", NotNull: true},
			defaultChanged: true,
		},
		{
			name:           "not null changed",
			old:            base,
			cur:            Column{Name: "col", Type: "BIGINT", Size: 0, Default: "0", NotNull: false},
			notNullChanged: true,
		},
		{
			name:           "multiple changes",
			old:            Column{Name: "col", Type: "VARCHAR(100)", Size: 100, Default: "''", NotNull: true},
			cur:            Column{Name: "col", Type: "TEXT", Size: 0, Default: "", NotNull: true},
			typeChanged:    true,
			sizeChanged:    true,
			defaultChanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := computeColumnDiff(tt.old, tt.cur)
			if tt.wantNil {
				if diff != nil {
					t.Errorf("expected nil diff, got %+v", diff)
				}

				return
			}

			if diff == nil {
				t.Fatal("expected non-nil diff")
			}

			if diff.TypeChanged != tt.typeChanged {
				t.Errorf("TypeChanged = %v, want %v", diff.TypeChanged, tt.typeChanged)
			}

			if diff.SizeChanged != tt.sizeChanged {
				t.Errorf("SizeChanged = %v, want %v", diff.SizeChanged, tt.sizeChanged)
			}

			if diff.DefaultChanged != tt.defaultChanged {
				t.Errorf("DefaultChanged = %v, want %v", diff.DefaultChanged, tt.defaultChanged)
			}

			if diff.NotNullChanged != tt.notNullChanged {
				t.Errorf("NotNullChanged = %v, want %v", diff.NotNullChanged, tt.notNullChanged)
			}
		})
	}
}

func TestComputeIndexDiff(t *testing.T) {
	base := Index{Name: "idx", Columns: []string{"a", "b"}, Order: map[string]string{"a": "ASC", "b": "ASC"}, Unique: false, Condition: ""}

	tests := []struct {
		name             string
		old, cur         Index
		wantNil          bool
		columnsChanged   bool
		uniqueChanged    bool
		conditionChanged bool
	}{
		{
			name:    "no changes",
			old:     base,
			cur:     Index{Name: "idx", Columns: []string{"a", "b"}, Order: map[string]string{"a": "ASC", "b": "ASC"}, Unique: false},
			wantNil: true,
		},
		{
			name:           "columns changed",
			old:            base,
			cur:            Index{Name: "idx", Columns: []string{"a", "c"}, Order: map[string]string{"a": "ASC", "c": "ASC"}},
			columnsChanged: true,
		},
		{
			name:           "order changed same columns different directions",
			old:            base,
			cur:            Index{Name: "idx", Columns: []string{"a", "b"}, Order: map[string]string{"a": "ASC", "b": "DESC"}},
			columnsChanged: true, // Order change is detected via ColumnsChanged
		},
		{
			name:          "unique changed",
			old:           base,
			cur:           Index{Name: "idx", Columns: []string{"a", "b"}, Order: map[string]string{"a": "ASC", "b": "ASC"}, Unique: true},
			uniqueChanged: true,
		},
		{
			name:             "condition changed",
			old:              base,
			cur:              Index{Name: "idx", Columns: []string{"a", "b"}, Order: map[string]string{"a": "ASC", "b": "ASC"}, Condition: "status = 'active'"},
			conditionChanged: true,
		},
		{
			name: "multiple changes",
			old:  base,
			cur: Index{
				Name: "idx", Columns: []string{"a"}, Order: map[string]string{"a": "DESC"},
				Unique: true, Condition: "active = true",
			},
			columnsChanged:   true,
			uniqueChanged:    true,
			conditionChanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := computeIndexDiff(tt.old, tt.cur)
			if tt.wantNil {
				if diff != nil {
					t.Errorf("expected nil diff, got %+v", diff)
				}

				return
			}

			if diff == nil {
				t.Fatal("expected non-nil diff")
			}

			if diff.ColumnsChanged != tt.columnsChanged {
				t.Errorf("ColumnsChanged = %v, want %v", diff.ColumnsChanged, tt.columnsChanged)
			}

			if diff.UniqueChanged != tt.uniqueChanged {
				t.Errorf("UniqueChanged = %v, want %v", diff.UniqueChanged, tt.uniqueChanged)
			}

			if diff.ConditionChanged != tt.conditionChanged {
				t.Errorf("ConditionChanged = %v, want %v", diff.ConditionChanged, tt.conditionChanged)
			}
		})
	}
}

func TestComputeDiff_IdenticalSchemas(t *testing.T) {
	s := &Schema{
		Tables: map[string]Table{
			"users": {
				Name:    "users",
				Columns: []Column{{Name: "id", Type: "BIGINT"}},
				Indexes: []Index{{Name: "pk", Columns: []string{"id"}, Order: map[string]string{"id": "ASC"}, Primary: true}},
			},
		},
	}

	diff := ComputeDiff(s, s)
	if !diff.IsEmpty() {
		t.Errorf("expected empty diff for identical schemas, got added=%v dropped=%v modified=%v",
			diff.AddedTables, diff.DroppedTables, diff.ModifiedTables)
	}
}

func TestComputeDiff_MixedChanges(t *testing.T) {
	oldSchema := &Schema{
		Tables: map[string]Table{
			"users":    {Name: "users", Columns: []Column{{Name: "id", Type: "BIGINT"}}},
			"products": {Name: "products", Columns: []Column{{Name: "id", Type: "BIGINT"}}},
			"sessions": {Name: "sessions", Columns: []Column{{Name: "id", Type: "BIGINT"}}},
		},
	}

	curSchema := &Schema{
		Tables: map[string]Table{
			"users":  {Name: "users", Columns: []Column{{Name: "id", Type: "BIGINT"}, {Name: "email", Type: "VARCHAR(255)"}}},
			"orders": {Name: "orders", Columns: []Column{{Name: "id", Type: "BIGINT"}}},
		},
	}

	diff := ComputeDiff(oldSchema, curSchema)

	if len(diff.AddedTables) != 1 || diff.AddedTables[0] != "orders" {
		t.Errorf("AddedTables = %v, want [orders]", diff.AddedTables)
	}

	sort.Strings(diff.DroppedTables)

	if len(diff.DroppedTables) != 2 {
		t.Fatalf("DroppedTables len = %d, want 2", len(diff.DroppedTables))
	}

	if diff.DroppedTables[0] != "products" || diff.DroppedTables[1] != "sessions" {
		t.Errorf("DroppedTables = %v, want [products, sessions]", diff.DroppedTables)
	}

	if _, ok := diff.ModifiedTables["users"]; !ok {
		t.Error("expected 'users' in ModifiedTables")
	}
}

func TestComputeDiff_EmptyNewSchema(t *testing.T) {
	oldSchema := &Schema{
		Tables: map[string]Table{
			"users":    {Name: "users"},
			"products": {Name: "products"},
		},
	}

	curSchema := &Schema{
		Tables: map[string]Table{},
	}

	diff := ComputeDiff(oldSchema, curSchema)

	sort.Strings(diff.DroppedTables)

	if len(diff.DroppedTables) != 2 {
		t.Fatalf("DroppedTables len = %d, want 2", len(diff.DroppedTables))
	}

	if diff.DroppedTables[0] != "products" || diff.DroppedTables[1] != "users" {
		t.Errorf("DroppedTables = %v, want [products, users]", diff.DroppedTables)
	}

	if len(diff.AddedTables) != 0 {
		t.Errorf("AddedTables = %v, want empty", diff.AddedTables)
	}
}

func TestComputeDiff_ModifiedIndex(t *testing.T) {
	oldSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name:    "users",
				Indexes: []Index{{Name: "idx_email", Columns: []string{"email"}, Order: map[string]string{"email": "ASC"}, Unique: false}},
			},
		},
	}

	curSchema := &Schema{
		Tables: map[string]Table{
			"users": {
				Name:    "users",
				Indexes: []Index{{Name: "idx_email", Columns: []string{"email"}, Order: map[string]string{"email": "ASC"}, Unique: true}},
			},
		},
	}

	diff := ComputeDiff(oldSchema, curSchema)

	tableDiff, ok := diff.ModifiedTables["users"]
	if !ok {
		t.Fatal("expected 'users' in ModifiedTables")
	}

	idxDiff, ok := tableDiff.ModifiedIndexes["idx_email"]
	if !ok {
		t.Fatal("expected 'idx_email' in ModifiedIndexes")
	}

	if !idxDiff.UniqueChanged {
		t.Error("expected UniqueChanged to be true")
	}
}
