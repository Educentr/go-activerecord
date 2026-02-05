package schema

import (
	"reflect"
)

// SchemaDiff описывает различия между двумя версиями схемы
type SchemaDiff struct {
	AddedTables    []string
	DroppedTables  []string
	ModifiedTables map[string]TableDiff
}

// IsEmpty возвращает true, если нет никаких изменений
func (d *SchemaDiff) IsEmpty() bool {
	return len(d.AddedTables) == 0 &&
		len(d.DroppedTables) == 0 &&
		len(d.ModifiedTables) == 0
}

// TableDiff описывает изменения в одной таблице
type TableDiff struct {
	AddedColumns    []Column
	DroppedColumns  []string
	ModifiedColumns map[string]ColumnDiff
	AddedIndexes    []Index
	DroppedIndexes  []string
	ModifiedIndexes map[string]IndexDiff
}

// IsEmpty возвращает true, если нет изменений в таблице
func (d *TableDiff) IsEmpty() bool {
	return len(d.AddedColumns) == 0 &&
		len(d.DroppedColumns) == 0 &&
		len(d.ModifiedColumns) == 0 &&
		len(d.AddedIndexes) == 0 &&
		len(d.DroppedIndexes) == 0 &&
		len(d.ModifiedIndexes) == 0
}

// ColumnDiff описывает изменения в одной колонке
type ColumnDiff struct {
	OldColumn Column
	NewColumn Column
	TypeChanged    bool
	SizeChanged    bool
	DefaultChanged bool
	NotNullChanged bool
}

// IndexDiff описывает изменения в одном индексе
type IndexDiff struct {
	OldIndex Index
	NewIndex Index
	ColumnsChanged   bool
	UniqueChanged    bool
	ConditionChanged bool
}

// ComputeDiff вычисляет различия между двумя схемами
func ComputeDiff(old, new *Schema) *SchemaDiff {
	diff := &SchemaDiff{
		AddedTables:    []string{},
		DroppedTables:  []string{},
		ModifiedTables: make(map[string]TableDiff),
	}

	if old == nil {
		// Все таблицы новые
		for name := range new.Tables {
			diff.AddedTables = append(diff.AddedTables, name)
		}
		return diff
	}

	// Найти удалённые таблицы
	for name := range old.Tables {
		if _, exists := new.Tables[name]; !exists {
			diff.DroppedTables = append(diff.DroppedTables, name)
		}
	}

	// Найти добавленные и изменённые таблицы
	for name, newTable := range new.Tables {
		oldTable, exists := old.Tables[name]
		if !exists {
			diff.AddedTables = append(diff.AddedTables, name)
			continue
		}

		tableDiff := computeTableDiff(&oldTable, &newTable)
		if !tableDiff.IsEmpty() {
			diff.ModifiedTables[name] = tableDiff
		}
	}

	return diff
}

// computeTableDiff вычисляет различия между двумя таблицами
func computeTableDiff(old, new *Table) TableDiff {
	diff := TableDiff{
		AddedColumns:    []Column{},
		DroppedColumns:  []string{},
		ModifiedColumns: make(map[string]ColumnDiff),
		AddedIndexes:    []Index{},
		DroppedIndexes:  []string{},
		ModifiedIndexes: make(map[string]IndexDiff),
	}

	// Создаём map для быстрого поиска колонок
	oldColumns := make(map[string]Column)
	for _, col := range old.Columns {
		oldColumns[col.Name] = col
	}

	newColumns := make(map[string]Column)
	for _, col := range new.Columns {
		newColumns[col.Name] = col
	}

	// Найти удалённые колонки
	for name := range oldColumns {
		if _, exists := newColumns[name]; !exists {
			diff.DroppedColumns = append(diff.DroppedColumns, name)
		}
	}

	// Найти добавленные и изменённые колонки
	for name, newCol := range newColumns {
		oldCol, exists := oldColumns[name]
		if !exists {
			diff.AddedColumns = append(diff.AddedColumns, newCol)
			continue
		}

		colDiff := computeColumnDiff(oldCol, newCol)
		if colDiff != nil {
			diff.ModifiedColumns[name] = *colDiff
		}
	}

	// Создаём map для быстрого поиска индексов
	oldIndexes := make(map[string]Index)
	for _, idx := range old.Indexes {
		oldIndexes[idx.Name] = idx
	}

	newIndexes := make(map[string]Index)
	for _, idx := range new.Indexes {
		newIndexes[idx.Name] = idx
	}

	// Найти удалённые индексы
	for name := range oldIndexes {
		if _, exists := newIndexes[name]; !exists {
			diff.DroppedIndexes = append(diff.DroppedIndexes, name)
		}
	}

	// Найти добавленные и изменённые индексы
	for name, newIdx := range newIndexes {
		oldIdx, exists := oldIndexes[name]
		if !exists {
			diff.AddedIndexes = append(diff.AddedIndexes, newIdx)
			continue
		}

		idxDiff := computeIndexDiff(oldIdx, newIdx)
		if idxDiff != nil {
			diff.ModifiedIndexes[name] = *idxDiff
		}
	}

	return diff
}

// computeColumnDiff вычисляет различия между двумя колонками
func computeColumnDiff(old, new Column) *ColumnDiff {
	diff := &ColumnDiff{
		OldColumn: old,
		NewColumn: new,
	}

	changed := false

	if old.Type != new.Type {
		diff.TypeChanged = true
		changed = true
	}

	if old.Size != new.Size {
		diff.SizeChanged = true
		changed = true
	}

	if old.Default != new.Default {
		diff.DefaultChanged = true
		changed = true
	}

	if old.NotNull != new.NotNull {
		diff.NotNullChanged = true
		changed = true
	}

	if !changed {
		return nil
	}

	return diff
}

// computeIndexDiff вычисляет различия между двумя индексами
func computeIndexDiff(old, new Index) *IndexDiff {
	diff := &IndexDiff{
		OldIndex: old,
		NewIndex: new,
	}

	changed := false

	if !reflect.DeepEqual(old.Columns, new.Columns) || !reflect.DeepEqual(old.Order, new.Order) {
		diff.ColumnsChanged = true
		changed = true
	}

	if old.Unique != new.Unique {
		diff.UniqueChanged = true
		changed = true
	}

	if old.Condition != new.Condition {
		diff.ConditionChanged = true
		changed = true
	}

	if !changed {
		return nil
	}

	return diff
}
