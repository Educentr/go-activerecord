// Package schema предоставляет структуры данных и функции для описания схемы базы данных
// и вычисления различий между версиями схемы.
package schema

import (
	"time"
)

// Generator интерфейс для генерации DDL схемы и миграций.
// Определён здесь, чтобы избежать циклических зависимостей между пакетами.
type Generator interface {
	// GenerateFullSchema генерирует полную DDL схему (SQL или конфиг)
	GenerateFullSchema(pkg interface{}) ([]byte, error)

	// GenerateSchemaJSON генерирует JSON представление схемы
	GenerateSchemaJSON(pkg interface{}) (*Table, error)

	// GenerateMigration генерирует миграцию из diff (nil если не поддерживается)
	GenerateMigration(diff *TableDiff, tableName string) (*Migration, error)

	// SupportsMigrations возвращает true если backend поддерживает миграции
	SupportsMigrations() bool

	// SchemaFileExtension возвращает расширение файла схемы (.sql, .cfg)
	SchemaFileExtension() string

	// SchemaFileName возвращает имя файла схемы (schema.sql, space.cfg)
	SchemaFileName() string
}

// Schema описывает полную схему базы данных для одного namespace
type Schema struct {
	Version   string           `json:"version"`
	Generated time.Time        `json:"generated"`
	Backend   string           `json:"backend"`
	Tables    map[string]Table `json:"tables"`
}

// NewSchema создаёт новую схему
func NewSchema(backend string) *Schema {
	return &Schema{
		Version:   "1.0",
		Generated: time.Now(),
		Backend:   backend,
		Tables:    make(map[string]Table),
	}
}

// Table описывает одну таблицу/спейс в схеме
type Table struct {
	Name        string       `json:"name"`
	Backend     string       `json:"backend"`
	SpaceID     uint         `json:"space_id,omitempty"` // Для Octopus
	Columns     []Column     `json:"columns"`
	Indexes     []Index      `json:"indexes"`
	PrimaryKey  []string     `json:"primary_key,omitempty"`
	ForeignKeys []ForeignKey `json:"foreign_keys,omitempty"`
}

// ForeignKey описывает внешний ключ таблицы
type ForeignKey struct {
	Name      string `json:"name"`                // Имя constraint (e.g., subscription_userid_fkey)
	Column    string `json:"column"`              // Локальная колонка
	RefTable  string `json:"ref_table"`           // Ссылаемая таблица
	RefColumn string `json:"ref_column"`          // Ссылаемая колонка
	OnDelete  string `json:"on_delete,omitempty"` // Действие при удалении (CASCADE, SET NULL, etc.)
}

// Column описывает одну колонку/поле в таблице
type Column struct {
	Name       string `json:"name"`
	Type       string `json:"type"`                  // SQL тип (INTEGER, VARCHAR(255), etc.)
	GoType     string `json:"go_type"`               // Оригинальный Go тип
	PrimaryKey bool   `json:"primary_key,omitempty"` // Участвует в первичном ключе
	NotNull    bool   `json:"not_null,omitempty"`    // NOT NULL constraint
	Size       int64  `json:"size,omitempty"`        // Размер для строковых типов
	Default    string `json:"default,omitempty"`     // Значение по умолчанию
	Position   int    `json:"position"`              // Позиция поля (важно для Octopus)
}

// Index описывает индекс таблицы
type Index struct {
	Name      string            `json:"name"`
	Columns   []string          `json:"columns"`
	Order     map[string]string `json:"order,omitempty"` // Направление сортировки: ASC/DESC
	Unique    bool              `json:"unique,omitempty"`
	Primary   bool              `json:"primary,omitempty"`
	Condition string            `json:"condition,omitempty"` // WHERE clause для частичных индексов
	Num       uint8             `json:"num,omitempty"`       // Номер индекса для Octopus
	Type      string            `json:"type,omitempty"`      // HASH/TREE для Octopus
}

// ColumnByName возвращает колонку по имени
func (t *Table) ColumnByName(name string) *Column {
	for i := range t.Columns {
		if t.Columns[i].Name == name {
			return &t.Columns[i]
		}
	}
	return nil
}

// IndexByName возвращает индекс по имени
func (t *Table) IndexByName(name string) *Index {
	for i := range t.Indexes {
		if t.Indexes[i].Name == name {
			return &t.Indexes[i]
		}
	}
	return nil
}

// GetPrimaryKey возвращает индекс первичного ключа
func (t *Table) GetPrimaryKey() *Index {
	for i := range t.Indexes {
		if t.Indexes[i].Primary {
			return &t.Indexes[i]
		}
	}
	return nil
}
