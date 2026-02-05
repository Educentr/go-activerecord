package schema

import (
	"fmt"
	"strings"
)

// Migration описывает одну миграцию базы данных
type Migration struct {
	Number      int      // Порядковый номер миграции
	Description string   // Описание изменений
	Up          []string // SQL statements для применения миграции
	Down        []string // SQL statements для отката миграции
}

// Format форматирует миграцию в SQL файл совместимый с golang-migrate/goose
func (m *Migration) Format() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("-- Migration %03d: %s\n\n", m.Number, m.Description))

	sb.WriteString("-- +up\n")
	for _, stmt := range m.Up {
		sb.WriteString(stmt)
		if !strings.HasSuffix(stmt, ";") {
			sb.WriteString(";")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n-- +down\n")
	for _, stmt := range m.Down {
		sb.WriteString(stmt)
		if !strings.HasSuffix(stmt, ";") {
			sb.WriteString(";")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Filename возвращает имя файла миграции
func (m *Migration) Filename() string {
	// Нормализуем описание для имени файла
	desc := strings.ToLower(m.Description)
	desc = strings.ReplaceAll(desc, " ", "_")
	// Удаляем специальные символы
	var cleaned strings.Builder
	for _, r := range desc {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			cleaned.WriteRune(r)
		}
	}
	return fmt.Sprintf("%03d_%s.sql", m.Number, cleaned.String())
}

// MigrationBuilder помогает строить миграции из diff
type MigrationBuilder struct {
	number int
}

// NewMigrationBuilder создаёт новый builder с указанным номером миграции
func NewMigrationBuilder(startNumber int) *MigrationBuilder {
	return &MigrationBuilder{number: startNumber}
}

// NextNumber возвращает следующий номер миграции
func (b *MigrationBuilder) NextNumber() int {
	n := b.number
	b.number++
	return n
}
