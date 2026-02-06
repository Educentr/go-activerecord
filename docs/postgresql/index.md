# PostgreSQL

Особенности работы с PostgreSQL бэкендом.

## Содержание раздела

- [BulkUpdate](bulk-update.md) — массовое обновление записей
- [BulkInsertReplace](bulk-insert.md) — массовая вставка с upsert
- [Генератор схемы](schema-generator.md) — DDL генерация из деклараций

## Обзор

PostgreSQL бэкенд использует драйвер [pgx/v5](https://github.com/jackc/pgx) и предоставляет:

| Возможность | Описание |
|-------------|----------|
| Условные индексы | WHERE clause в индексах |
| BulkUpdate | Массовое обновление через UPDATE FROM |
| BulkInsertReplace | Массовая вставка с ON CONFLICT |
| DDL генерация | SQL схема из деклараций |

!!! note "Автокоммит"
    Все операции выполняются в режиме автокоммита. Транзакции не поддерживаются. Подробнее: [Транзакции и автокоммит](../architecture/transactions.md).

## Декларация

```go
//ar:serverConf:pgconf
//ar:namespace:users    // Имя таблицы
//ar:backend:postgres
type FieldsUser struct {
    Id        int64  `ar:"primary_key;init_by_db"`  // BIGSERIAL
    Email     string `ar:"unique;size:256"`
    Name      string `ar:"size:256"`
    CreatedAt int64  `ar:""`
}
```

## Условные индексы

Индексы с WHERE clause:

```go
type IndexesUser struct {
    // Только активные пользователи
    ActiveByDate bool `ar:"fields:CreatedAt;condition:Status[=]active&&DeletedAt[is null]"`

    // Битовая маска
    Premium bool `ar:"fields:CreatedAt;condition:Flags&4[=]4"`

    // Пустые строки
    NoEmail bool `ar:"fields:Id;condition:Email[=]''"`
}
```

Генерируемый SQL:

```sql
CREATE INDEX idx_users_activebydate ON users (created_at)
    WHERE status = 'active' AND deleted_at IS NULL;

CREATE INDEX idx_users_premium ON users (created_at)
    WHERE flags & 4 = 4;

CREATE INDEX idx_users_noemail ON users (id)
    WHERE email = '';
```

## Типы данных

| Go тип | PostgreSQL | С init_by_db |
|--------|------------|--------------|
| `int64` | BIGINT | BIGSERIAL |
| `int32` | INTEGER | SERIAL |
| `int16` | SMALLINT | SMALLSERIAL |
| `string` | VARCHAR(n) | - |
| `[]byte` | BYTEA | - |
| `bool` | BOOLEAN | - |
| `float64` | DOUBLE PRECISION | - |
| `time.Time` | TIMESTAMP | - |

## Bulk операции

### BulkUpdate

```go
users, _ := user.SelectByIds(ctx, ids)

for _, u := range users {
    u.SetStatus("active")
}

updatedCount, err := user.BulkUpdate(ctx, users)
```

Подробнее: [BulkUpdate](bulk-update.md)

### BulkInsertReplace

```go
users := make([]*user.User, 100)
for i := range users {
    u := user.New(ctx)
    u.SetEmail(fmt.Sprintf("user%d@example.com", i))
    users[i] = u
}

err := user.BulkInsertReplace(ctx, users, user.OnConflictEmail)
```

Подробнее: [BulkInsertReplace](bulk-insert.md)

## Генерация DDL

```bash
argen --path "model/repository" --schema-output schema.sql
```

Подробнее: [Генератор схемы](schema-generator.md)

!!! tip "Совет"
    Используйте генератор схемы для создания миграций и документации БД.

## Следующие шаги

- [BulkUpdate](bulk-update.md) — массовое обновление
- [Генератор схемы](schema-generator.md) — DDL из деклараций
