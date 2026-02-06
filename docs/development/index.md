# Разработка

Руководство для разработчиков go-activerecord.

## Содержание раздела

- [Шаблоны](templates.md) — модификация шаблонов генерации

## Структура проекта

```
go-activerecord/
├── cmd/
│   └── argen/              # CLI точка входа
├── internal/
│   ├── app/                # Логика приложения argen
│   └── pkg/
│       ├── parser/         # Парсинг деклараций
│       ├── checker/        # Валидация
│       ├── generator/      # Генерация кода
│       ├── ds/             # Структуры данных
│       ├── schema/         # Схема БД
│       └── backend/        # Бэкенды
│           ├── octopus/    # Octopus бэкенд
│           └── postgres/   # PostgreSQL бэкенд
├── pkg/
│   ├── activerecord/       # Core runtime
│   ├── octopus/            # Octopus драйвер
│   ├── postgres/           # PostgreSQL интеграция
│   ├── iproto/             # Протокол iproto
│   ├── serializer/         # Встроенные сериализаторы
│   └── logger/             # Логирование
├── docs/                   # Документация
└── scripts/                # Скрипты сборки
```

## Команды Make

```bash
make build          # Сборка argen
make install        # Установка в $GOPATH/bin
make test           # Запуск тестов
make cover          # Покрытие тестами
make lint           # Линтер (изменения от main)
make full-lint      # Линтер (весь код)
make generate       # Генерация моков
```

## Добавление нового бэкенда

### 1. Создание директории

```
internal/pkg/backend/mybackend/
├── backend.go          # Реализация Backend interface
├── checker.go          # Валидация
├── generator.go        # Генератор кода
├── schema_generator.go # Генератор схемы (опционально)
└── tmpl/
    └── pkg/
        ├── main.tmpl   # Основной шаблон
        ├── select.tmpl # Селекторы
        └── ...
```

### 2. Интерфейс Backend

```go
type Backend interface {
    Name() string
    Generate(pkg *ds.RecordPackage) error
    Validate(pkg *ds.RecordPackage) error
}
```

### 3. Регистрация

В `internal/pkg/backend/backend.go`:

```go
func GetBackend(name string) Backend {
    switch name {
    case "postgres":
        return postgres.New()
    case "octopus":
        return octopus.New()
    case "mybackend":
        return mybackend.New()
    default:
        return nil
    }
}
```

## Добавление нового тега

### 1. Обновление парсера

`internal/pkg/parser/field.go`:

```go
func parseFieldTag(tag string) *ds.FieldDeclaration {
    // Добавить парсинг нового тега
    if v, ok := parseTag(tag, "mytag"); ok {
        field.MyTag = v
    }
}
```

### 2. Обновление структуры данных

`internal/pkg/ds/field.go`:

```go
type FieldDeclaration struct {
    // ...
    MyTag string
}
```

### 3. Обновление валидатора

`internal/pkg/checker/field.go`:

```go
func validateField(field *ds.FieldDeclaration) error {
    if field.MyTag != "" {
        // Валидация
    }
}
```

### 4. Обновление шаблонов

```
{{ if .MyTag }}
// Использование MyTag
{{ end }}
```

## Тестирование

### Запуск тестов

```bash
# Все тесты
make test

# С покрытием
make cover

# Конкретный пакет
go test -v ./internal/pkg/parser/...
```

### Build tag

```go
//go:build activerecord

package mytest
```

## Git workflow

### Pre-commit hook

```bash
make pre-commit-hook
```

Устанавливает хук:
1. `make generate`
2. `make lint`

### Pre-push hook

```bash
make pre-push-hook
```

Устанавливает хук:
1. `make cover`

## Следующие шаги

- [Шаблоны](templates.md) — работа с шаблонами генерации
