package schema

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// Детерминизм обхода: ComputeDiff строит списки таблиц обходом map, поэтому без
// сортировки порядок (а значит и порядок операторов в миграции, и её имя файла)
// меняется от прогона к прогону.
func TestComputeDiff_AddedTablesSorted(t *testing.T) {
	newSchema := &Schema{
		Version: "1.0",
		Backend: "postgres",
		Tables: map[string]Table{
			"users":    {Name: "users", Backend: "postgres"},
			"accounts": {Name: "accounts", Backend: "postgres"},
			"products": {Name: "products", Backend: "postgres"},
		},
	}

	want := []string{"accounts", "products", "users"}

	// Повторяем: порядок обхода map рандомизируется на каждом range
	for i := 0; i < 20; i++ {
		diff := ComputeDiff(nil, newSchema)
		if !reflect.DeepEqual(diff.AddedTables, want) {
			t.Fatalf("iteration %d: AddedTables = %v, want %v", i, diff.AddedTables, want)
		}
	}
}

func TestComputeDiff_AddedAndDroppedTablesSorted(t *testing.T) {
	oldSchema := &Schema{
		Version: "1.0",
		Backend: "postgres",
		Tables: map[string]Table{
			"gone_c": {Name: "gone_c"},
			"gone_a": {Name: "gone_a"},
			"gone_b": {Name: "gone_b"},
			"kept":   {Name: "kept"},
		},
	}

	newSchema := &Schema{
		Version: "1.0",
		Backend: "postgres",
		Tables: map[string]Table{
			"new_c": {Name: "new_c"},
			"new_a": {Name: "new_a"},
			"new_b": {Name: "new_b"},
			"kept":  {Name: "kept"},
		},
	}

	wantAdded := []string{"new_a", "new_b", "new_c"}
	wantDropped := []string{"gone_a", "gone_b", "gone_c"}

	for i := 0; i < 20; i++ {
		diff := ComputeDiff(oldSchema, newSchema)

		if !reflect.DeepEqual(diff.AddedTables, wantAdded) {
			t.Fatalf("iteration %d: AddedTables = %v, want %v", i, diff.AddedTables, wantAdded)
		}

		if !reflect.DeepEqual(diff.DroppedTables, wantDropped) {
			t.Fatalf("iteration %d: DroppedTables = %v, want %v", i, diff.DroppedTables, wantDropped)
		}
	}
}

// Колонки и индексы обходятся в порядке объявления, а не в порядке обхода map
func TestComputeTableDiff_ColumnsAndIndexesInDeclarationOrder(t *testing.T) {
	oldTable := &Table{
		Name: "users",
		Columns: []Column{
			{Name: "id", Type: "BIGINT", Position: 0},
			{Name: "old_z", Type: "TEXT", Position: 1},
			{Name: "old_a", Type: "TEXT", Position: 2},
		},
		Indexes: []Index{
			{Name: "users_pkey", Columns: []string{"id"}, Primary: true},
			{Name: "idx_users_old_z", Columns: []string{"old_z"}},
			{Name: "idx_users_old_a", Columns: []string{"old_a"}},
		},
	}

	newTable := &Table{
		Name: "users",
		Columns: []Column{
			{Name: "id", Type: "BIGINT", Position: 0},
			{Name: "new_z", Type: "TEXT", Position: 1},
			{Name: "new_a", Type: "TEXT", Position: 2},
		},
		Indexes: []Index{
			{Name: "users_pkey", Columns: []string{"id"}, Primary: true},
			{Name: "idx_users_new_z", Columns: []string{"new_z"}},
			{Name: "idx_users_new_a", Columns: []string{"new_a"}},
		},
	}

	wantAddedColumns := []string{"new_z", "new_a"}
	wantDroppedColumns := []string{"old_z", "old_a"}
	wantAddedIndexes := []string{"idx_users_new_z", "idx_users_new_a"}
	wantDroppedIndexes := []string{"idx_users_old_z", "idx_users_old_a"}

	for i := 0; i < 20; i++ {
		diff := computeTableDiff(oldTable, newTable)

		gotAddedColumns := make([]string, 0, len(diff.AddedColumns))
		for _, col := range diff.AddedColumns {
			gotAddedColumns = append(gotAddedColumns, col.Name)
		}

		gotAddedIndexes := make([]string, 0, len(diff.AddedIndexes))
		for _, idx := range diff.AddedIndexes {
			gotAddedIndexes = append(gotAddedIndexes, idx.Name)
		}

		if !reflect.DeepEqual(gotAddedColumns, wantAddedColumns) {
			t.Fatalf("iteration %d: AddedColumns = %v, want %v", i, gotAddedColumns, wantAddedColumns)
		}

		if !reflect.DeepEqual(diff.DroppedColumns, wantDroppedColumns) {
			t.Fatalf("iteration %d: DroppedColumns = %v, want %v", i, diff.DroppedColumns, wantDroppedColumns)
		}

		if !reflect.DeepEqual(gotAddedIndexes, wantAddedIndexes) {
			t.Fatalf("iteration %d: AddedIndexes = %v, want %v", i, gotAddedIndexes, wantAddedIndexes)
		}

		if !reflect.DeepEqual(diff.DroppedIndexes, wantDroppedIndexes) {
			t.Fatalf("iteration %d: DroppedIndexes = %v, want %v", i, diff.DroppedIndexes, wantDroppedIndexes)
		}
	}
}

func TestSchemaDiff_ModifiedTableNames(t *testing.T) {
	diff := &SchemaDiff{
		ModifiedTables: map[string]TableDiff{
			"users":    {},
			"accounts": {},
			"products": {},
		},
	}

	want := []string{"accounts", "products", "users"}

	for i := 0; i < 20; i++ {
		if got := diff.ModifiedTableNames(); !reflect.DeepEqual(got, want) {
			t.Fatalf("iteration %d: ModifiedTableNames() = %v, want %v", i, got, want)
		}
	}
}

func TestTableDiff_ModifiedColumnAndIndexNames(t *testing.T) {
	diff := &TableDiff{
		ModifiedColumns: map[string]ColumnDiff{
			"status":  {},
			"email":   {},
			"zipcode": {},
		},
		ModifiedIndexes: map[string]IndexDiff{
			"idx_c": {},
			"idx_a": {},
			"idx_b": {},
		},
	}

	wantColumns := []string{"email", "status", "zipcode"}
	wantIndexes := []string{"idx_a", "idx_b", "idx_c"}

	for i := 0; i < 20; i++ {
		if got := diff.ModifiedColumnNames(); !reflect.DeepEqual(got, wantColumns) {
			t.Fatalf("iteration %d: ModifiedColumnNames() = %v, want %v", i, got, wantColumns)
		}

		if got := diff.ModifiedIndexNames(); !reflect.DeepEqual(got, wantIndexes) {
			t.Fatalf("iteration %d: ModifiedIndexNames() = %v, want %v", i, got, wantIndexes)
		}
	}
}

func TestTableDiff_ModifiedNamesEmpty(t *testing.T) {
	diff := &TableDiff{}

	if got := diff.ModifiedColumnNames(); len(got) != 0 {
		t.Errorf("ModifiedColumnNames() = %v, want empty", got)
	}

	if got := diff.ModifiedIndexNames(); len(got) != 0 {
		t.Errorf("ModifiedIndexNames() = %v, want empty", got)
	}
}

func TestSchema_SameContent(t *testing.T) {
	base := func() *Schema {
		return &Schema{
			Version:   "1.0",
			Generated: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Backend:   "postgres",
			Tables: map[string]Table{
				"users": {
					Name:    "users",
					Backend: "postgres",
					Columns: []Column{{Name: "id", Type: "BIGINT"}},
					Indexes: []Index{{Name: "users_pkey", Columns: []string{"id"}, Primary: true}},
				},
			},
		}
	}

	t.Run("другая метка времени — то же содержимое", func(t *testing.T) {
		left, right := base(), base()
		right.Generated = time.Date(2026, 6, 15, 12, 30, 0, 0, time.UTC)

		if !left.SameContent(right) {
			t.Error("схемы отличаются только меткой времени, ожидалось SameContent() == true")
		}
	})

	// Загруженная из JSON схема несёт nil вместо пустых коллекций: сравнение
	// структур через reflect.DeepEqual считало бы такие схемы разными
	t.Run("nil и пустая коллекция эквивалентны", func(t *testing.T) {
		left, right := base(), base()

		table := right.Tables["users"]
		table.Indexes = []Index{{Name: "users_pkey", Columns: []string{"id"}, Primary: true, Order: map[string]string{}}}
		table.ForeignKeys = []ForeignKey{}
		right.Tables["users"] = table

		if !left.SameContent(right) {
			t.Error("nil и пустая коллекция должны считаться одинаковыми")
		}
	})

	t.Run("изменение содержимого", func(t *testing.T) {
		left, right := base(), base()

		table := right.Tables["users"]
		table.Columns = append(table.Columns, Column{Name: "email", Type: "TEXT"})
		right.Tables["users"] = table

		if left.SameContent(right) {
			t.Error("схемы отличаются набором колонок, ожидалось SameContent() == false")
		}
	})

	t.Run("nil вместо схемы", func(t *testing.T) {
		if base().SameContent(nil) {
			t.Error("сравнение с nil должно давать false")
		}
	})
}

func TestSchema_PreserveGenerated(t *testing.T) {
	oldStamp := time.Date(2026, 1, 1, 10, 20, 30, 0, time.UTC)
	newStamp := time.Date(2026, 11, 22, 19, 0, 0, 0, time.UTC)

	table := Table{
		Name:    "users",
		Backend: "postgres",
		Columns: []Column{{Name: "id", Type: "BIGINT"}},
	}

	t.Run("схема не изменилась — метка переносится", func(t *testing.T) {
		oldSchema := &Schema{Version: "1.0", Generated: oldStamp, Backend: "postgres", Tables: map[string]Table{"users": table}}
		newSchema := &Schema{Version: "1.0", Generated: newStamp, Backend: "postgres", Tables: map[string]Table{"users": table}}

		newSchema.PreserveGenerated(oldSchema)

		if !newSchema.Generated.Equal(oldStamp) {
			t.Errorf("Generated = %s, want %s", newSchema.Generated, oldStamp)
		}
	})

	t.Run("схема изменилась — метка обновляется", func(t *testing.T) {
		changed := table
		changed.Columns = append([]Column{}, table.Columns...)
		changed.Columns = append(changed.Columns, Column{Name: "email", Type: "TEXT"})

		oldSchema := &Schema{Version: "1.0", Generated: oldStamp, Backend: "postgres", Tables: map[string]Table{"users": table}}
		newSchema := &Schema{Version: "1.0", Generated: newStamp, Backend: "postgres", Tables: map[string]Table{"users": changed}}

		newSchema.PreserveGenerated(oldSchema)

		if !newSchema.Generated.Equal(newStamp) {
			t.Errorf("Generated = %s, want %s", newSchema.Generated, newStamp)
		}
	})

	t.Run("первая генерация — предыдущей схемы нет", func(t *testing.T) {
		newSchema := &Schema{Version: "1.0", Generated: newStamp, Backend: "postgres", Tables: map[string]Table{"users": table}}

		newSchema.PreserveGenerated(nil)

		if !newSchema.Generated.Equal(newStamp) {
			t.Errorf("Generated = %s, want %s", newSchema.Generated, newStamp)
		}
	})

	// Метка времени проходит через JSON без искажений: иначе schema.json
	// менялся бы побайтно даже с перенесённой меткой
	t.Run("метка времени стабильна при round-trip через JSON", func(t *testing.T) {
		original := &Schema{Version: "1.0", Generated: time.Now(), Backend: "postgres", Tables: map[string]Table{"users": table}}

		data, err := json.MarshalIndent(original, "", "  ")
		if err != nil {
			t.Fatalf("MarshalIndent error: %v", err)
		}

		var loaded Schema
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		reMarshaled, err := json.MarshalIndent(&loaded, "", "  ")
		if err != nil {
			t.Fatalf("MarshalIndent error: %v", err)
		}

		if string(data) != string(reMarshaled) {
			t.Errorf("JSON изменился после round-trip:\n%s\n---\n%s", data, reMarshaled)
		}
	})
}
