# BulkInsertReplace

Массовая вставка записей с обработкой конфликтов.

## Сигнатура

```go
func BulkInsertReplace(ctx context.Context, objs []*User, conflictIndex OnConflictIndex) error
```

## Базовое использование

```go
// Создаём объекты
users := make([]*user.User, 100)
for i := range users {
    u := user.New(ctx)
    u.SetEmail(fmt.Sprintf("user%d@example.com", i))
    u.SetName(fmt.Sprintf("User %d", i))
    users[i] = u
}

// Массовая вставка с upsert по Email
err := user.BulkInsertReplace(ctx, users, user.OnConflictEmail)
if err != nil {
    return err
}

// После вызова объекты получают ID из БД
for _, u := range users {
    fmt.Printf("User %s got ID: %d\n", u.GetEmail(), u.GetId())
}
```

## Как это работает

### 1. Выбор индекса для сопоставления

Для получения автоинкрементных значений нужен уникальный индекс **без** `init_by_db` полей:

```go
type FieldsUser struct {
    Id    int64  `ar:"primary_key;init_by_db"`  // Автоинкремент
    Email string `ar:"unique;size:256"`          // Подходит для сопоставления
    Name  string `ar:"size:256"`
}
```

Приоритет выбора:

1. `conflictIndex` (если уникальный и без init_by_db)
2. Любой другой подходящий уникальный индекс

### 2. Генерация SQL

```sql
INSERT INTO users (id, email, name, created_at)
VALUES
    (DEFAULT, 'user1@example.com', 'User 1', 1234567890),
    (DEFAULT, 'user2@example.com', 'User 2', 1234567891),
    (DEFAULT, 'user3@example.com', 'User 3', 1234567892)
ON CONFLICT (email) DO UPDATE SET
    name = EXCLUDED.name,
    created_at = EXCLUDED.created_at
RETURNING id, email;
```

### 3. Сопоставление результатов

Результаты из RETURNING сопоставляются с исходными объектами через уникальный индекс:

```go
// Внутренняя логика
key := fmt.Sprintf("%v", email)  // Ключ для сопоставления
objMap[key] = obj                 // Map для быстрого поиска
```

## OnConflict индексы

Для каждого уникального индекса генерируется константа:

```go
type FieldsUser struct {
    Id    int64  `ar:"primary_key;init_by_db"`
    Email string `ar:"unique;size:256"`
    Code  string `ar:"unique;size:64"`
}

// Генерируемые константы
const (
    OnConflictId    = ...  // Конфликт по Id
    OnConflictEmail = ...  // Конфликт по Email
    OnConflictCode  = ...  // Конфликт по Code
)
```

## Требования

### Уникальный индекс без init_by_db

Функция требует наличия уникального индекса **без** автоинкрементных полей:

```go
// ✅ Подходит — Email уникальный и не init_by_db
type FieldsUser struct {
    Id    int64  `ar:"primary_key;init_by_db"`
    Email string `ar:"unique;size:256"`
}

// ❌ Не подходит — только Id, который init_by_db
type FieldsCounter struct {
    Id    int64  `ar:"primary_key;init_by_db"`
    Value int64  `ar:""`
}
```

При отсутствии подходящего индекса:

```go
err := counter.BulkInsertReplace(ctx, counters, counter.OnConflictId)
// Вернёт ошибку:
// "BulkInsertReplace is not available for Counter: no unique index
// without InitByDB fields found"
```

## Метрики

| Метрика | Описание |
|---------|----------|
| `bulk_insertorreplace_request` | Количество вызовов |
| `bulk_insert_gen` | Ошибки генерации SQL |
| `bulk_insert_preparedb` | Ошибки подготовки соединения |
| `bulk_insert_db` | Ошибки выполнения |
| `bulk_insert_scan` | Ошибки чтения результатов |
| `bulk_insert_match` | Ошибки сопоставления |
| `bulk_insertorreplace_success` | Успешно вставленные строки |

## Альтернативы

Если уникального индекса без init_by_db нет:

### Цикл с InsertOrReplace

```go
for _, obj := range objs {
    err := obj.InsertOrReplace(ctx, OnConflictId)
    if err != nil {
        return err
    }
}
```

### Добавление подходящего индекса

```go
type FieldsCounter struct {
    Id        int64  `ar:"primary_key;init_by_db"`
    ExternalId string `ar:"unique;size:64"`  // Добавили для сопоставления
    Value     int64  `ar:""`
}
```

## Пример с составным ключом

```go
type FieldsUserRole struct {
    Id     int64 `ar:"primary_key;init_by_db"`
    UserId int64 `ar:""`
    RoleId int64 `ar:""`
    Flags  uint32 `ar:""`
}

type IndexesUserRole struct {
    UserRole bool `ar:"fields:UserId,RoleId;unique"`  // Для сопоставления
}
```

```go
roles := []*userrole.UserRole{...}
err := userrole.BulkInsertReplace(ctx, roles, userrole.OnConflictUserRole)

// Каждый объект получит свой Id
for _, r := range roles {
    fmt.Printf("Role: user=%d, role=%d, id=%d\n",
        r.GetUserId(), r.GetRoleId(), r.GetId())
}
```

## Ограничения

1. **Требуется уникальный индекс без init_by_db**
   - Иначе невозможно сопоставить результаты

2. **Конфликт только по уникальным индексам**
   - `conflictIndex` должен быть уникальным

3. **Не поддерживается DO NOTHING**
   - Только DO UPDATE (upsert)

## Сравнение с другими методами

| Метод | Запросов | Возврат ID | Upsert |
|-------|----------|------------|--------|
| Цикл Insert | N | ✅ | ❌ |
| Цикл InsertOrReplace | N | ✅ | ✅ |
| BulkInsertReplace | 1 | ✅ | ✅ |

## Следующие шаги

- [BulkUpdate](bulk-update.md) — массовое обновление
- [CRUD операции](../usage/crud.md) — обычные операции
