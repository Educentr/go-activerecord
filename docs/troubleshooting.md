# Решение проблем

Частые проблемы и их решения.

## Генерация кода

### Ошибка ".argen file not found"

**Проблема:** argen отказывается генерировать код.

**Решение:**

```bash
mkdir -p model/repository/generated
touch model/repository/generated/.argen
```

### Ошибка "package name must match regex"

**Проблема:** Имя модели не соответствует ограничениям.

**Решение:** Используйте имя из 1-20 строчных букв:

```go
// ❌ Неправильно
type FieldsUserProfile struct { ... }

// ✅ Правильно
type FieldsUser struct { ... }
```

### Файлы не генерируются

**Проблема:** После запуска argen файлы не появляются.

**Проверьте:**

1. Путь не начинается с `.`:
   ```bash
   argen --path "model"  # ✅
   argen --path "./model" # ❌
   ```

2. Декларации в правильной папке:
   ```bash
   ls model/repository/declaration/
   ```

3. Структуры имеют правильные префиксы:
   ```go
   type FieldsUser struct { ... }  // ✅
   type User struct { ... }        // ❌
   ```

## Компиляция

### Undefined: user.SelectById

**Проблема:** Сгенерированные методы не найдены.

**Решения:**

1. Перегенерируйте код:
   ```bash
   argen --path "model/repository"
   ```

2. Проверьте импорт:
   ```go
   import "yourapp/model/repository/generated/user"
   ```

3. Для тестов добавьте build tag:
   ```go
   //go:build activerecord
   ```

### Циклический импорт

**Проблема:** Циклическая зависимость между моделями.

**Решение:** Избегайте взаимных ссылок в `FieldsObject*`:

```go
// ❌ Проблема
type FieldsObjectUser struct {
    Profile int64 `ar:"object:Profile"`
}
type FieldsObjectProfile struct {
    User int64 `ar:"object:User"`
}

// ✅ Решение: убрать одну из ссылок
```

## Runtime ошибки

### ErrNotFound

**Проблема:** Запись не найдена.

**Решение:**

```go
user, err := user.SelectById(ctx, id)
if err != nil {
    if errors.Is(err, activerecord.ErrNotFound) {
        // Обработка отсутствия записи
        return nil, nil
    }
    return nil, err
}
```

### Exists must be true for Update

**Проблема:** Попытка обновить новый объект.

**Решение:**

```go
// ❌ Неправильно
u := user.New(ctx)
u.SetName("John")
u.Update(ctx)  // Ошибка: Exists = false

// ✅ Правильно — используйте Insert для нового объекта
u := user.New(ctx)
u.SetName("John")
u.Insert(ctx)
```

### Exists must be false for Insert

**Проблема:** Попытка вставить существующий объект.

**Решение:**

```go
// ❌ Неправильно
u, _ := user.SelectById(ctx, 123)
u.Insert(ctx)  // Ошибка: Exists = true

// ✅ Правильно — используйте Update или Replace
u.SetName("New Name")
u.Update(ctx)
```

### Limiter required for non-unique index

**Проблема:** Выборка по неуникальному индексу без лимита.

**Решение:**

```go
// ❌ Неправильно
users, _ := user.SelectByStatus(ctx, "active", nil)

// ✅ Правильно
limiter := activerecord.NewLimiter(100)
users, _ := user.SelectByStatus(ctx, "active", limiter)
```

## Octopus

### Неверные данные в полях

**Проблема:** Поля содержат неправильные значения.

**Причина:** Порядок полей в декларации не соответствует tuple.

**Решение:** Проверьте и исправьте порядок полей:

```go
// Порядок должен соответствовать tuple в БД
type FieldsUser struct {
    Id    int64  // Поле 0
    Email string // Поле 1
    Name  string // Поле 2
}
```

### RepairTuple вызывается постоянно

**Проблема:** Триггер вызывается при каждом SELECT.

**Причина:** Несоответствие количества полей.

**Решение:**

1. Проверьте количество полей в декларации
2. Обновите схему в Octopus
3. Или добавьте недостающие поля через RepairTuple

### Неверный порядок индексов

**Проблема:** Выборка по индексу возвращает неправильные данные.

**Решение:** Порядок индексов должен соответствовать конфигурации Octopus:

```go
type IndexesUser struct {
    Id     bool // Индекс 0 в конфиге
    Email  bool // Индекс 1 в конфиге
    Status bool // Индекс 2 в конфиге
}
```

## PostgreSQL

### BulkInsertReplace: no unique index without InitByDB fields

**Проблема:** BulkInsertReplace недоступен.

**Решение:** Добавьте уникальный индекс без автоинкрементных полей:

```go
type FieldsUser struct {
    Id    int64  `ar:"primary_key;init_by_db"`
    Email string `ar:"unique;size:256"`  // Добавьте unique
}
```

### Условный индекс не используется

**Проблема:** PostgreSQL не использует условный индекс.

**Решение:** В запросе должны быть все условия индекса:

```go
// Индекс: condition:Status[=]active&&DeletedAt[is null]
// Запрос автоматически добавит эти условия
users, _ := user.SelectByActiveIndex(ctx, key, limiter)
```

## Конфигурация

### Connection refused

**Проблема:** Нет соединения с БД.

**Проверьте:**

1. Сервер БД запущен
2. Правильный адрес в конфигурации:
   ```go
   "mydb/master": "127.0.0.1:5432"
   ```
3. Firewall разрешает соединения

### Connection timeout

**Проблема:** Соединение зависает.

**Решение:** Настройте таймауты:

```go
"mydb/Timeout": "5s"
```

И используйте контекст с таймаутом:

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
```

## Тесты

### Тесты не компилируются

**Проблема:** `undefined: user.SelectById` в тестах.

**Решение:** Добавьте build tag:

```go
//go:build activerecord

package mytest
```

Запуск:

```bash
go test -tags=activerecord ./...
```

### Фикстуры не находятся

**Проблема:** `fixture.GetUserByCode` возвращает nil.

**Проверьте:**

1. YAML файл существует: `fixture/data/user.yaml`
2. Формат YAML правильный (snake_case)
3. Code совпадает

## Логирование

Для отладки включите подробные логи:

```go
logger := NewDetailedLogger()
activerecord.RegisterLogger(logger)
```

## Получение помощи

1. Проверьте [документацию](index.md)
2. Посмотрите [примеры](https://github.com/mailru/activerecord-cookbook)
3. Создайте [issue](https://github.com/Educentr/go-activerecord/issues)
