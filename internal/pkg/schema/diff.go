package schema

import (
	"reflect"
	"sort"
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

// ModifiedTableNames возвращает имена изменённых таблиц в алфавитном порядке.
// Прямой обход ModifiedTables даёт случайный порядок, из-за которого порядок
// операторов в миграции меняется от прогона к прогону.
func (d *SchemaDiff) ModifiedTableNames() []string {
	return sortedNames(d.ModifiedTables)
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

// ModifiedColumnNames возвращает имена изменённых колонок в алфавитном порядке.
// Нужен для детерминированного обхода ModifiedColumns.
func (d *TableDiff) ModifiedColumnNames() []string {
	return sortedNames(d.ModifiedColumns)
}

// ModifiedIndexNames возвращает имена изменённых индексов в алфавитном порядке.
// Нужен для детерминированного обхода ModifiedIndexes.
func (d *TableDiff) ModifiedIndexNames() []string {
	return sortedNames(d.ModifiedIndexes)
}

// ColumnDiff описывает изменения в одной колонке
type ColumnDiff struct {
	OldColumn      Column
	NewColumn      Column
	TypeChanged    bool
	SizeChanged    bool
	DefaultChanged bool
	NotNullChanged bool
}

// IndexDiff описывает изменения в одном индексе
type IndexDiff struct {
	OldIndex         Index
	NewIndex         Index
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

		sort.Strings(diff.AddedTables)

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

	// Обход map выше даёт случайный порядок, поэтому сортируем результат:
	// от него зависит порядок операторов в миграции и её имя файла
	sort.Strings(diff.AddedTables)
	sort.Strings(diff.DroppedTables)

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

	// Найти удалённые колонки (в порядке объявления в старой схеме)
	for _, oldCol := range old.Columns {
		if _, exists := newColumns[oldCol.Name]; !exists {
			diff.DroppedColumns = append(diff.DroppedColumns, oldCol.Name)
		}
	}

	// Найти добавленные и изменённые колонки (в порядке объявления в новой схеме)
	for _, newCol := range new.Columns {
		oldCol, exists := oldColumns[newCol.Name]
		if !exists {
			diff.AddedColumns = append(diff.AddedColumns, newCol)

			continue
		}

		colDiff := computeColumnDiff(oldCol, newCol)
		if colDiff != nil {
			diff.ModifiedColumns[newCol.Name] = *colDiff
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

	// Найти удалённые индексы (в порядке объявления в старой схеме)
	for _, oldIdx := range old.Indexes {
		if _, exists := newIndexes[oldIdx.Name]; !exists {
			diff.DroppedIndexes = append(diff.DroppedIndexes, oldIdx.Name)
		}
	}

	// Найти добавленные и изменённые индексы (в порядке объявления в новой схеме)
	for _, newIdx := range new.Indexes {
		oldIdx, exists := oldIndexes[newIdx.Name]
		if !exists {
			diff.AddedIndexes = append(diff.AddedIndexes, newIdx)

			continue
		}

		idxDiff := computeIndexDiff(oldIdx, newIdx)
		if idxDiff != nil {
			diff.ModifiedIndexes[newIdx.Name] = *idxDiff
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

// sortedNames возвращает ключи map в алфавитном порядке
func sortedNames[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))

	for name := range m {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}
