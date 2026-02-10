package postgres

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Educentr/go-activerecord/v3/internal/pkg/ds"
	"github.com/Educentr/go-activerecord/v3/internal/pkg/schema"
)

// PostgresSchemaGenerator генератор DDL схемы для PostgreSQL
type PostgresSchemaGenerator struct{}

// NewSchemaGenerator создаёт новый генератор схемы PostgreSQL
func NewSchemaGenerator() *PostgresSchemaGenerator {
	return &PostgresSchemaGenerator{}
}

// SupportsMigrations возвращает true - PostgreSQL поддерживает миграции
func (g *PostgresSchemaGenerator) SupportsMigrations() bool {
	return true
}

// SchemaFileExtension возвращает расширение файла схемы
func (g *PostgresSchemaGenerator) SchemaFileExtension() string {
	return ".sql"
}

// SchemaFileName возвращает имя файла схемы
func (g *PostgresSchemaGenerator) SchemaFileName() string {
	return "schema.sql"
}

// GenerateSchemaJSON генерирует JSON представление таблицы
func (g *PostgresSchemaGenerator) GenerateSchemaJSON(pkgInterface any) (*schema.Table, error) {
	pkg, ok := pkgInterface.(*ds.RecordPackage)
	if !ok {
		return nil, fmt.Errorf("expected *ds.RecordPackage, got %T", pkgInterface)
	}

	// Используем ObjectName (snake_case из //ar:namespace:) для имени таблицы
	tableName := pkg.Namespace.ObjectName
	if tableName == "" {
		// Fallback на PublicName если ObjectName не задан
		tableName = pkg.Namespace.PublicName
	}

	table := &schema.Table{
		Name:    tableName,
		Backend: "postgres",
		Columns: make([]schema.Column, 0, len(pkg.Fields)),
		Indexes: make([]schema.Index, 0, len(pkg.Indexes)),
	}

	// Конвертируем поля в колонки
	for i, field := range pkg.Fields {
		col := g.fieldToColumn(field, i)
		table.Columns = append(table.Columns, col)

		if field.PrimaryKey {
			table.PrimaryKey = append(table.PrimaryKey, strings.ToLower(field.Name))
		}
	}

	// Конвертируем индексы
	for _, idx := range pkg.Indexes {
		schemaIdx := g.indexToSchemaIndex(idx, pkg)
		table.Indexes = append(table.Indexes, schemaIdx)
	}

	// Генерируем FK из FieldsObjectMap
	for _, fo := range pkg.FieldsObjectMap {
		fk := schema.ForeignKey{
			Name:      fmt.Sprintf("%s_%s_fkey", strings.ToLower(tableName), strings.ToLower(fo.Field)),
			Column:    strings.ToLower(fo.Field),
			RefTable:  fo.ObjectName, // Уже snake_case из //ar:namespace:
			RefColumn: strings.ToLower(fo.Key),
			OnDelete:  "CASCADE", // По умолчанию CASCADE
		}
		table.ForeignKeys = append(table.ForeignKeys, fk)
	}

	return table, nil
}

// fieldToColumn конвертирует FieldDeclaration в schema.Column
func (g *PostgresSchemaGenerator) fieldToColumn(field ds.FieldDeclaration, position int) schema.Column {
	col := schema.Column{
		Name:       strings.ToLower(field.Name), // PostgreSQL case-insensitive, приводим к нижнему регистру
		GoType:     string(field.Format),
		PrimaryKey: field.PrimaryKey,
		NotNull:    true, // Все поля NOT NULL (Go не имеет null)
		Size:       field.Size,
		Position:   position,
	}

	// Определяем SQL тип (используем init_by_db как признак SERIAL)
	col.Type = g.goTypeToSQLType(field.Format, field.Size, field.InitByDB)

	// Устанавливаем DEFAULT значение (кроме SERIAL типов - их инициализирует БД)
	if !field.InitByDB {
		col.Default = g.getDefaultValue(field.Format)
	}

	return col
}

// getDefaultValue возвращает SQL DEFAULT значение для Go типа
func (g *PostgresSchemaGenerator) getDefaultValue(format ds.Format) string {
	switch format {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return "0"
	case "float32", "float64":
		return "0"
	case "bool":
		return "false"
	case "string", "[]byte":
		return "''"
	case "time.Time":
		// Значение time.Time{} это "0001-01-01 00:00:00"
		return "'0001-01-01 00:00:00'"
	default:
		return ""
	}
}

// goTypeToSQLType конвертирует Go тип в SQL тип
// initByDB используется для определения SERIAL типов (автоинкремент)
func (g *PostgresSchemaGenerator) goTypeToSQLType(format ds.Format, size int64, initByDB bool) string {
	// Для SERIAL типов (когда поле инициализируется БД)
	if initByDB {
		switch format {
		case "int64":
			return "BIGSERIAL"
		case "int32", "int":
			return "SERIAL"
		case "int16":
			return "SMALLSERIAL"
		}
	}

	formatType, err := GetFormat(format)
	if err != nil {
		// Fallback для неизвестных типов
		return "TEXT"
	}

	dbType := formatType.DBType()

	// Для VARCHAR с указанным размером
	if strings.Contains(dbType, "(n)") && size > 0 {
		return strings.Replace(dbType, "(n)", fmt.Sprintf("(%d)", size), 1)
	}

	// Для VARCHAR без размера используем TEXT
	if strings.Contains(dbType, "(n)") {
		return "TEXT"
	}

	return dbType
}

// indexToSchemaIndex конвертирует IndexDeclaration в schema.Index
func (g *PostgresSchemaGenerator) indexToSchemaIndex(idx ds.IndexDeclaration, pkg *ds.RecordPackage) schema.Index {
	// Используем ObjectName для имени индекса
	tableName := pkg.Namespace.ObjectName
	if tableName == "" {
		tableName = pkg.Namespace.PublicName
	}

	schemaIdx := schema.Index{
		Name:    g.generateIndexName(tableName, idx),
		Columns: make([]string, 0, len(idx.Fields)),
		Order:   make(map[string]string),
		Unique:  idx.Unique,
		Primary: idx.Primary,
	}

	// Добавляем колонки индекса
	for _, fieldNum := range idx.Fields {
		if fieldNum < len(pkg.Fields) {
			fieldName := pkg.Fields[fieldNum].Name
			colName := strings.ToLower(fieldName)
			schemaIdx.Columns = append(schemaIdx.Columns, colName)

			// Добавляем направление сортировки
			if idxField, ok := idx.FieldsMap[fieldName]; ok {
				if idxField.Order == ds.IndexOrderDesc {
					schemaIdx.Order[colName] = "DESC"
				} else {
					schemaIdx.Order[colName] = "ASC"
				}
			}
		}
	}

	// Генерируем WHERE clause для условного индекса
	if len(idx.Conditions) > 0 {
		schemaIdx.Condition = g.generateConditionClause(idx.Conditions, pkg)
	}

	return schemaIdx
}

// generateIndexName генерирует имя индекса
func (g *PostgresSchemaGenerator) generateIndexName(tableName string, idx ds.IndexDeclaration) string {
	if idx.Primary {
		return fmt.Sprintf("%s_pkey", strings.ToLower(tableName))
	}
	return fmt.Sprintf("idx_%s_%s", strings.ToLower(tableName), strings.ToLower(idx.Name))
}

// generateConditionClause генерирует WHERE clause для условного индекса
func (g *PostgresSchemaGenerator) generateConditionClause(conditions map[int]ds.IndexCondition, pkg *ds.RecordPackage) string {
	var parts []string

	for fieldNum, cond := range conditions {
		var fieldName string
		var fieldFormat ds.Format
		if fieldNum < len(pkg.Fields) {
			fieldName = pkg.Fields[fieldNum].Name
			fieldFormat = pkg.Fields[fieldNum].Format
		} else {
			continue
		}

		// Приводим имя поля к нижнему регистру
		colName := strings.ToLower(fieldName)

		// Используем выражение поля если указано (например, "Flags&1")
		// и заменяем имя поля на lowercase версию
		fieldExpr := colName
		if cond.FieldExpression != "" {
			// Заменяем имя поля в выражении на lowercase
			fieldExpr = strings.Replace(cond.FieldExpression, fieldName, colName, 1)
		}

		var part string
		if cond.IsNullCheck {
			part = fmt.Sprintf("%s %s", fieldExpr, cond.ConditionType)
		} else if len(cond.Value) > 0 {
			value := cond.Value[0]
			// Определяем нужно ли экранировать значение как строку
			needQuote := g.isStringType(fieldFormat)

			// Если значение пустое, используем пустую строку в SQL
			if value == "" || value == "," {
				value = "''"
			} else if needQuote && !strings.HasPrefix(value, "'") {
				// Экранируем строковое значение если оно ещё не экранировано
				value = fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "''"))
			}
			part = fmt.Sprintf("%s %s %s", fieldExpr, cond.ConditionType, value)
		}

		if part != "" {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, " AND ")
}

// isStringType проверяет является ли тип строковым
func (g *PostgresSchemaGenerator) isStringType(format ds.Format) bool {
	switch format {
	case "string", "[]byte":
		return true
	default:
		return false
	}
}

// GenerateFullSchema генерирует полную DDL схему
func (g *PostgresSchemaGenerator) GenerateFullSchema(pkgInterface any) ([]byte, error) {
	table, err := g.GenerateSchemaJSON(pkgInterface)
	if err != nil {
		return nil, err
	}

	return g.generateDDL(table)
}

// generateDDL генерирует SQL DDL из schema.Table
func (g *PostgresSchemaGenerator) generateDDL(table *schema.Table) ([]byte, error) {
	tmpl := `-- DDL Schema for table {{ .Name }}
-- Generated by argen

CREATE TABLE IF NOT EXISTS {{ .Name }} (
{{- $lastColIdx := sub (len .Columns) 1 }}
{{- $hasPK := gt (len .PrimaryKey) 0 }}
{{- $hasFK := gt (len .ForeignKeys) 0 }}
{{- $lastFKIdx := sub (len .ForeignKeys) 1 }}
{{- range $i, $col := .Columns }}
    {{ $col.Name }} {{ $col.Type }}{{ if $col.NotNull }} NOT NULL{{ end }}{{ if $col.Default }} DEFAULT {{ $col.Default }}{{ end }}{{ if or (ne $i $lastColIdx) $hasPK $hasFK }},{{ end }}
{{- end }}
{{- if .PrimaryKey }}
    PRIMARY KEY ({{ join .PrimaryKey ", " }}){{ if $hasFK }},{{ end }}
{{- end }}
{{- range $i, $fk := .ForeignKeys }}
    CONSTRAINT {{ $fk.Name }} FOREIGN KEY ({{ $fk.Column }})
        REFERENCES {{ $fk.RefTable }} ({{ $fk.RefColumn }}){{ if $fk.OnDelete }} ON DELETE {{ $fk.OnDelete }}{{ end }}{{ if ne $i $lastFKIdx }},{{ end }}
{{- end }}
);
{{ range .Indexes }}
{{- if not .Primary }}

CREATE {{ if .Unique }}UNIQUE {{ end }}INDEX IF NOT EXISTS {{ .Name }} ON {{ $.Name }} ({{ indexColumns . }}){{ if .Condition }}
    WHERE {{ .Condition }}{{ end }};
{{- end }}
{{- end }}
`

	funcMap := template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"join": strings.Join,
		"indexColumns": func(idx schema.Index) string {
			var parts []string
			for _, col := range idx.Columns {
				part := col
				if o, ok := idx.Order[col]; ok && o != "" {
					part += " " + o
				}
				parts = append(parts, part)
			}
			return strings.Join(parts, ", ")
		},
	}

	t, err := template.New("ddl").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse DDL template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, table); err != nil {
		return nil, fmt.Errorf("execute DDL template: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateMigration генерирует миграцию из diff
func (g *PostgresSchemaGenerator) GenerateMigration(diff *schema.TableDiff, tableName string) (*schema.Migration, error) {
	if diff == nil || diff.IsEmpty() {
		return nil, nil
	}

	migration := &schema.Migration{
		Up:   []string{},
		Down: []string{},
	}

	var descriptions []string

	// Добавленные колонки
	for _, col := range diff.AddedColumns {
		stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, col.Name, col.Type)
		if col.NotNull {
			stmt += " NOT NULL"
		}
		if col.Default != "" {
			stmt += " DEFAULT " + col.Default
		}
		migration.Up = append(migration.Up, stmt)
		migration.Down = append(migration.Down, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, col.Name))
		descriptions = append(descriptions, fmt.Sprintf("add column %s", col.Name))
	}

	// Удалённые колонки
	for _, colName := range diff.DroppedColumns {
		migration.Up = append(migration.Up, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, colName))
		// Для down нужны данные о типе - оставляем комментарий
		migration.Down = append(migration.Down, fmt.Sprintf("-- FIXME: Restore column %s (see docs/roadmap.md)", colName))
		descriptions = append(descriptions, fmt.Sprintf("drop column %s", colName))
	}

	// Изменённые колонки
	for colName, colDiff := range diff.ModifiedColumns {
		if colDiff.TypeChanged {
			migration.Up = append(migration.Up,
				fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", tableName, colName, colDiff.NewColumn.Type))
			migration.Down = append(migration.Down,
				fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", tableName, colName, colDiff.OldColumn.Type))
			descriptions = append(descriptions, fmt.Sprintf("change type of %s", colName))
		}

		if colDiff.NotNullChanged {
			if colDiff.NewColumn.NotNull {
				migration.Up = append(migration.Up,
					fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", tableName, colName))
				migration.Down = append(migration.Down,
					fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", tableName, colName))
			} else {
				migration.Up = append(migration.Up,
					fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", tableName, colName))
				migration.Down = append(migration.Down,
					fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", tableName, colName))
			}
		}

		if colDiff.DefaultChanged {
			if colDiff.NewColumn.Default != "" {
				migration.Up = append(migration.Up,
					fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s", tableName, colName, colDiff.NewColumn.Default))
			} else {
				migration.Up = append(migration.Up,
					fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT", tableName, colName))
			}
			if colDiff.OldColumn.Default != "" {
				migration.Down = append(migration.Down,
					fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s", tableName, colName, colDiff.OldColumn.Default))
			} else {
				migration.Down = append(migration.Down,
					fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT", tableName, colName))
			}
		}
	}

	// Добавленные индексы
	for _, idx := range diff.AddedIndexes {
		stmt := g.generateCreateIndexStatement(tableName, idx)
		migration.Up = append(migration.Up, stmt)
		migration.Down = append(migration.Down, fmt.Sprintf("DROP INDEX IF EXISTS %s", idx.Name))
		descriptions = append(descriptions, fmt.Sprintf("add index %s", idx.Name))
	}

	// Удалённые индексы
	for _, idxName := range diff.DroppedIndexes {
		migration.Up = append(migration.Up, fmt.Sprintf("DROP INDEX IF EXISTS %s", idxName))
		migration.Down = append(migration.Down, fmt.Sprintf("-- FIXME: Restore index %s (see docs/roadmap.md)", idxName))
		descriptions = append(descriptions, fmt.Sprintf("drop index %s", idxName))
	}

	// Изменённые индексы (drop + create)
	for _, idxDiff := range diff.ModifiedIndexes {
		migration.Up = append(migration.Up, fmt.Sprintf("DROP INDEX IF EXISTS %s", idxDiff.OldIndex.Name))
		migration.Up = append(migration.Up, g.generateCreateIndexStatement(tableName, idxDiff.NewIndex))

		migration.Down = append(migration.Down, fmt.Sprintf("DROP INDEX IF EXISTS %s", idxDiff.NewIndex.Name))
		migration.Down = append(migration.Down, g.generateCreateIndexStatement(tableName, idxDiff.OldIndex))

		descriptions = append(descriptions, fmt.Sprintf("modify index %s", idxDiff.NewIndex.Name))
	}

	if len(descriptions) > 0 {
		migration.Description = strings.Join(descriptions, ", ")
	} else {
		migration.Description = "schema changes"
	}

	return migration, nil
}

// generateCreateIndexStatement генерирует CREATE INDEX statement
func (g *PostgresSchemaGenerator) generateCreateIndexStatement(tableName string, idx schema.Index) string {
	var sb strings.Builder

	sb.WriteString("CREATE ")
	if idx.Unique {
		sb.WriteString("UNIQUE ")
	}
	sb.WriteString("INDEX IF NOT EXISTS ")
	sb.WriteString(idx.Name)
	sb.WriteString(" ON ")
	sb.WriteString(tableName)
	sb.WriteString(" (")

	for i, col := range idx.Columns {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(col)
		if order, ok := idx.Order[col]; ok && order != "" {
			sb.WriteString(" ")
			sb.WriteString(order)
		}
	}

	sb.WriteString(")")

	if idx.Condition != "" {
		sb.WriteString(" WHERE ")
		sb.WriteString(idx.Condition)
	}

	return sb.String()
}
