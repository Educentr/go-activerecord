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

### Значения условий

Константа условия разбирается на этапе генерации и попадает в двух видах: в
сгенерированный `index.go` — литералом того типа, который объявлен у поля, и в
`WHERE` предиката индекса в `schema.sql` — литералом типа колонки. Поэтому она
должна быть записана в теге так, чтобы приводилась к этому типу:

| Тип поля | Как писать значение | Литерал в `index.go` | Литерал в DDL (`schema.sql`) |
|----------|---------------------|----------------------|------------------------------|
| `int16`, `int32`, `int64` | целое число | литерал типа поля: `int16(0)`, `int32(0)`, `int64(0)` | число как есть: `0` |
| `float32`, `float64` | число | литерал типа поля: `float32(0.5)`, `float64(0.5)` | число как есть: `0.5` |
| `bool` | `true` / `false` | `true` | `true` |
| `string` | текст, кавычки необязательны | `"active"` | `'active'` |
| `postgres.BYTEA` | текст | `postgres.BYTEA("...")` | `'...'` |
| `time.Time` | число секунд unix-времени | `time.Unix(1700000000, 0).UTC()` | `TIMESTAMP '2023-11-14 22:13:20'` |
| `uuid.UUID` | uuid в каноническом виде | `uuid.UUID{0x6b, ...}` | `'6ba7b810-9dad-11d1-80b4-00c04fd430c8'` |

Время в DDL пишется литералом без часового пояса, в UTC: предикат частичного
индекса обязан быть IMMUTABLE, а `to_timestamp(N)` возвращает `timestamptz` —
сравнение с колонкой `TIMESTAMP` зависит от `TimeZone` сессии, и PostgreSQL
такой индекс не создаст (`functions in index predicate must be marked
IMMUTABLE`).

#### Семантика времени в условии

Число секунд в условии по полю `time.Time` трактуется как **UTC wall clock
колонки `TIMESTAMP`**: `1700000000` означает `2023-11-14 22:13:20` — ровно то,
что попадёт в предикат индекса в DDL. Поэтому литерал в `index.go` приводится к
UTC (`.UTC()`): колонка `TIMESTAMP` часового пояса не хранит, и pgx при записи
`time.Time` зону отбрасывает, сохраняя wall clock как есть
(`pgtype/timestamp.go`, `discardTimeZone`). Без приведения к UTC на процессе с
`TZ != UTC` одно и то же значение из тега означало бы два разных момента: DDL —
`22:13:20`, условие запроса — `01:13:20` следующих суток для
`TZ=Europe/Moscow`. Планировщик такой запрос частичным индексом не покрывает, а
выборка отбирает не тот диапазон, что описан в декларации.

Из этого следует, что время в условии сверяется с wall clock колонки, а не с
моментом на оси времени: если приложение пишет строки на процессе с
`TZ != UTC`, pgx положит в колонку его локальный wall clock, и в предикат
попадут не те строки, которые ожидались по unix-времени. Это свойство колонки
`TIMESTAMP` (без зоны), а не условий индекса: `argen` на этапе генерации TZ
процесса не знает. Нужно сравнение по моментам — пишите в такие колонки
значения в UTC.

Если значение к типу поля не приводится, `argen` падает на проверке декларации
и указывает поле, значение и тип:

```text
ErrCheckPackageIndexDecl Pkg: `user`; Index: `Status`;
	condition on field `Status`: invalid index condition value: `abc` is not a valid integer for type `int64`
```

### Несколько значений в условии

Значения перечисляются через запятую. Такое условие попадает в DDL списком:
`=` даёт `IN (...)`, `!=` — `NOT IN (...)`:

```go
type IndexesProviderKey struct {
    Mid    bool `ar:"fields:Mid;condition:Mid[=]1,2,3"`
    Status bool `ar:"fields:Status;condition:Status[=]'active','pending'"`
    Live   bool `ar:"fields:Status;condition:Status[!=]'draft','deleted'"`
}
```

```sql
CREATE INDEX idx_provider_keys_mid ON provider_keys (mid ASC)
    WHERE mid IN (1, 2, 3);

CREATE INDEX idx_provider_keys_status ON provider_keys (status ASC)
    WHERE status IN ('active', 'pending');

CREATE INDEX idx_provider_keys_live ON provider_keys (status ASC)
    WHERE status NOT IN ('draft', 'deleted');
```

Каждое значение приводится к типу колонки по таблице выше, то есть список по
полю `time.Time` или `uuid.UUID` тоже корректен:
`WHERE createdat IN (TIMESTAMP '2023-11-14 22:13:20', TIMESTAMP '2023-11-14 23:13:20')`.

Предикат индекса и условие запроса по этому индексу строятся по одной таблице
операторов, поэтому они совпадают: запрос по индексу `Mid` получает
`WHERE mid IN ($1, $2, $3)`. Расхождение здесь означало бы, что планировщик не
применит частичный индекс к запросу, а у уникального индекса уникальность
действовала бы только для части значений.

Операторы сравнения (`>`, `<`, `>=`, `<=`) принимают только одно значение:
списка для них нет ни в предикате индекса, ни в запросе. `argen` отвергает такую
декларацию на проверке:

```text
ErrCheckPackageIndexDecl Pkg: `providerkey`; Index: `Mid`; operator value count mismatch
```

## Типы данных

| Go тип | PostgreSQL | С init_by_db |
|--------|------------|--------------|
| `int64` | BIGINT | BIGSERIAL |
| `int32` | INTEGER | SERIAL |
| `int16` | SMALLINT | SMALLSERIAL |
| `string` | VARCHAR(n) | - |
| `postgres.BYTEA` | BYTEA | - |
| `bool` | BOOLEAN | - |
| `float64` | DOUBLE PRECISION | - |
| `time.Time` | TIMESTAMP | - |
| `uuid.UUID` | UUID | - |

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
argen --path "model/repository" --schema_path "etc/database"
```

Подробнее: [Генератор схемы](schema-generator.md)

!!! tip "Совет"
    Используйте генератор схемы для создания миграций и документации БД.

## Следующие шаги

- [BulkUpdate](bulk-update.md) — массовое обновление
- [Генератор схемы](schema-generator.md) — DDL из деклараций
