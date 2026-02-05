package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// LoadSchema загружает схему из JSON файла
func LoadSchema(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Файл не существует - это нормально для первой генерации
		}
		return nil, fmt.Errorf("read schema file: %w", err)
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parse schema JSON: %w", err)
	}

	return &schema, nil
}

// SaveSchema сохраняет схему в JSON файл
func SaveSchema(schema *Schema, path string) error {
	// Создаём директорию если не существует
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write schema file: %w", err)
	}

	return nil
}

// SaveDDL сохраняет DDL в файл
func SaveDDL(ddl []byte, path string) error {
	// Создаём директорию если не существует
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, ddl, 0644); err != nil {
		return fmt.Errorf("write DDL file: %w", err)
	}

	return nil
}

// SaveMigration сохраняет миграцию в файл
func SaveMigration(migration *Migration, dir string) error {
	// Создаём директорию migrations если не существует
	migrationsDir := filepath.Join(dir, "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("create migrations directory: %w", err)
	}

	path := filepath.Join(migrationsDir, migration.Filename())
	content := migration.Format()

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write migration file: %w", err)
	}

	return nil
}

// GetNextMigrationNumber определяет следующий номер миграции в директории
func GetNextMigrationNumber(dir string) (int, error) {
	migrationsDir := filepath.Join(dir, "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil // Директория не существует - начинаем с 1
		}
		return 0, fmt.Errorf("read migrations directory: %w", err)
	}

	// Ищем максимальный номер миграции
	maxNum := 0
	re := regexp.MustCompile(`^(\d+)_.*\.sql$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(entry.Name())
		if len(matches) < 2 {
			continue
		}

		num, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		if num > maxNum {
			maxNum = num
		}
	}

	return maxNum + 1, nil
}
