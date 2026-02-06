# CLI аргументы argen

Справочник аргументов командной строки argen.

## Синтаксис

```bash
argen [OPTIONS]
```

## Обязательные аргументы

| Аргумент | Описание |
|----------|----------|
| `--path` | Путь к модели (не может начинаться с `.`) |

## Опциональные аргументы

| Аргумент | По умолчанию | Описание |
|----------|--------------|----------|
| `--declaration` | `declaration` | Папка с декларациями |
| `--destination` | `generated` | Папка для сгенерированного кода |
| `--module` | из `go.mod` | Имя модуля |
| `--fixture_path` | - | Путь для тестовых фикстур |
| `--schema-output` | - | Файл для DDL схемы |

## Примеры

### Базовое использование

```bash
argen --path "model/repository" \
      --declaration "declaration" \
      --destination "generated"
```

### С фикстурами

```bash
argen --path "model/repository" \
      --declaration "decl" \
      --destination "cmpl" \
      --fixture_path "testutil/fixture"
```

### С явным модулем

```bash
argen --path "model/repository" \
      --module "github.com/myorg/myapp"
```

### Генерация DDL

```bash
argen --path "model/repository" \
      --schema-output database/schema.sql
```

## Структура путей

```
project/
├── go.mod                    # Модуль определяется отсюда
├── model/
│   └── repository/           # --path
│       ├── declaration/      # --declaration (относительно path)
│       │   ├── user.go
│       │   └── product.go
│       └── generated/        # --destination (относительно path)
│           ├── .argen        # Маркер-файл
│           ├── user/
│           └── product/
└── testutil/
    └── fixture/              # --fixture_path
        └── fixture/
            └── data/
```

## Маркер-файл .argen

При первой генерации создаётся файл `.argen` в папке назначения:

```
model/repository/generated/.argen
```

**Назначение:**

- Защита от случайной генерации в неправильную директорию
- При отсутствии argen откажется генерировать код

!!! warning "Не удаляйте .argen"
    Удаление маркера заблокирует генерацию.

## Поведение

### Удаление файлов

При каждом запуске **все `.go` файлы** в папке назначения удаляются и создаются заново.

!!! danger "Не редактируйте сгенерированный код"
    Изменения будут потеряны при следующей генерации.

### Именование пакетов

Имя сгенерированного пакета берётся из имени структуры `Fields*`:

```go
type FieldsUser struct { ... }   // → пакет "user"
type FieldsProduct struct { ... } // → пакет "product"
```

**Ограничения:**

- Regex: `^[a-z]{1,20}$`
- Только строчные буквы
- Максимум 20 символов

## Проверка версии

```bash
argen --version
```

Вывод:

```
argen version v3.1.21 (abc12345) built 20240115-1200 on linux
```

## Справка

```bash
argen --help
```

## Интеграция с Make

```makefile
.PHONY: generate-models
generate-models: ## Generate ActiveRecord models
	argen --path "model/repository" \
	      --declaration "declaration" \
	      --destination "generated"

.PHONY: generate-fixtures
generate-fixtures: ## Generate test fixtures
	argen --path "model/repository" \
	      --declaration "declaration" \
	      --destination "generated" \
	      --fixture_path "testutil/fixture"

.PHONY: generate-schema
generate-schema: ## Generate DDL schema
	argen --path "model/repository" \
	      --schema-output database/schema.sql
```

## Интеграция с go:generate

```go
//go:generate argen --path "../../model/repository" --declaration "declaration" --destination "generated"

package main
```

Запуск:

```bash
go generate ./...
```

## Ошибки

### "path cannot start with `.`"

```bash
# ❌ Неправильно
argen --path "./model"

# ✅ Правильно
argen --path "model"
```

### ".argen file not found"

Создайте маркер вручную или убедитесь, что папка назначения существует:

```bash
mkdir -p model/repository/generated
touch model/repository/generated/.argen
```

### "package name must match regex"

Имя модели должно соответствовать `^[a-z]{1,20}$`:

```go
// ❌ Неправильно
type FieldsUserProfile struct { ... }  // "userprofile" > 20 символов

// ✅ Правильно
type FieldsUser struct { ... }
```

## Следующие шаги

- [Установка](../getting-started/installation.md) — установка argen
- [Быстрый старт](../getting-started/quickstart.md) — первая модель
