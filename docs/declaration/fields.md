# Поля (Fields)

Структура `Fields*` описывает все поля таблицы или спейса.

## Базовый синтаксис

```go
type FieldsUser struct {
    Id        int64  `ar:"primary_key"`
    Email     string `ar:"unique;size:256;selector:SelectByEmail"`
    Name      string `ar:"size:256"`
    Counter   uint32 `ar:"mutators:inc,dec"`
    Flags     uint32 `ar:""`
    CreatedAt uint32 `ar:""`
}
```

## Теги полей

| Тег | Описание | Пример |
|-----|----------|--------|
| `primary_key` | Первичный ключ (уникальный, один на модель) | `ar:"primary_key"` |
| `unique` | Уникальный индекс | `ar:"unique"` |
| `selector` | Имя генерируемого селектора | `ar:"selector:SelectByEmail"` |
| `size` | Размер поля в байтах | `ar:"size:256"` |
| `serializer` | Сериализатор для поля | `ar:"serializer:Json"` |
| `mutators` | Список мутаторов | `ar:"mutators:inc,dec"` |
| `init_by_db` | Поле инициализируется БД (autoincrement) | `ar:"init_by_db"` |

## Поддерживаемые типы

| Go тип | PostgreSQL | Octopus | Размер |
|--------|------------|---------|--------|
| `int8` | SMALLINT | int8 | 1 |
| `int16` | SMALLINT | int16 | 2 |
| `int32`, `int` | INTEGER | int32 | 4 |
| `int64` | BIGINT | int64 | 8 |
| `uint8` | SMALLINT | uint8 | 1 |
| `uint16` | INTEGER | uint16 | 2 |
| `uint32` | BIGINT | uint32 | 4 |
| `uint64` | BIGINT | uint64 | 8 |
| `float32` | REAL | float32 | 4 |
| `float64` | DOUBLE PRECISION | float64 | 8 |
| `bool` | BOOLEAN | bool | 1 |
| `string` | VARCHAR(n) / TEXT | string | size |
| `[]byte` | BYTEA | bytes | size |
| `time.Time` | TIMESTAMP | - | 8 |

## Первичный ключ

Каждая модель должна иметь **ровно один** первичный ключ:

```go
type FieldsUser struct {
    Id int64 `ar:"primary_key"`  // Обязательно
    // ...
}
```

!!! warning "Ограничения PK"
    - Только одно поле может быть primary_key
    - Сеттер для PK защищён от изменения существующих записей
    - Для составного PK используйте `Indexes*` с `primary_key`

## Автоинкремент (init_by_db)

Для полей, значение которых генерируется базой данных:

```go
type FieldsUser struct {
    Id        int64  `ar:"primary_key;init_by_db"`  // BIGSERIAL в PostgreSQL
    CreatedAt int64  `ar:"init_by_db"`               // DEFAULT NOW()
}
```

В PostgreSQL генерируется:
- `int64` + `init_by_db` → `BIGSERIAL`
- `int32` + `init_by_db` → `SERIAL`
- `int16` + `init_by_db` → `SMALLSERIAL`

## Селекторы

Тег `selector` создаёт метод для поиска по полю:

```go
type FieldsUser struct {
    Id    int64  `ar:"primary_key"`
    Email string `ar:"unique;selector:SelectByEmail"`
    Code  string `ar:"selector:SelectByCode"`
}
```

Генерируемые методы:

```go
// По уникальному полю — возвращает один объект
user, err := user.SelectByEmail(ctx, "john@example.com")

// По неуникальному полю — требуется лимитер
users, err := user.SelectByCode(ctx, "ABC123", activerecord.NewLimiter(100))

// Множественная выборка (суффикс 's')
users, err := user.SelectByEmails(ctx, []string{"a@b.com", "c@d.com"})
```

## Размер поля (size)

Для строковых полей указывает максимальную длину:

```go
type FieldsUser struct {
    Email    string `ar:"size:256"`   // VARCHAR(256)
    Bio      string `ar:"size:4096"`  // VARCHAR(4096)
    Content  string `ar:""`           // TEXT (без ограничения)
}
```

!!! tip "Совет"
    Всегда указывайте `size` для строковых полей — это влияет на валидацию и прогнозирование объёма хранилища.

## Мутаторы

Тег `mutators` добавляет методы для атомарных операций:

```go
type FieldsUser struct {
    Counter uint32 `ar:"mutators:inc,dec"`
    Flags   uint32 `ar:"mutators:and,or,xor,set_bit,clear_bit"`
}
```

| Мутатор | Метод | Операция |
|---------|-------|----------|
| `inc` | `Inc*` | `+= arg` |
| `dec` | `Dec*` | `-= arg` |
| `and` | `And*` | `&= arg` |
| `or` | `Or*` | `\|= arg` |
| `xor` | `Xor*` | `^= arg` |
| `set_bit` | `SetBit*` | `\|= arg` |
| `clear_bit` | `ClearBit*` | `&= ^arg` |

Подробнее в разделе [Мутаторы и флаги](mutators.md).

## Сериализаторы

Тег `serializer` применяет преобразование к полю:

```go
type FieldsConfig struct {
    Settings string `ar:"serializer:Json;size:2048"`
    Version  string `ar:"serializer:Printf,%d.%d.%d;size:16"`
}

type SerializersConfig struct {
    Settings map[string]interface{} `ar:""`
}
```

Подробнее в разделе [Сериализаторы](serializers.md).

## Порядок полей (Octopus)

!!! danger "Критично для Octopus"
    Порядок полей в `Fields*` должен **точно** соответствовать порядку полей в tuple базы данных!

```go
// Порядок ВАЖЕН для Octopus!
type FieldsUser struct {
    Id    int64  `ar:"primary_key"`  // Поле 0 в tuple
    Email string `ar:"size:256"`      // Поле 1 в tuple
    Name  string `ar:"size:256"`      // Поле 2 в tuple
}
```

Для PostgreSQL порядок полей не критичен.

## Полный пример

```go
package repository

//ar:serverConf:mydb
//ar:namespace:users
//ar:backend:postgres
type FieldsUser struct {
    // Первичный ключ с автоинкрементом
    Id int64 `ar:"primary_key;init_by_db"`

    // Уникальные поля с селекторами
    Email    string `ar:"unique;size:256;selector:SelectByEmail"`
    Username string `ar:"unique;size:64;selector:SelectByUsername"`

    // Обычные поля
    Name      string `ar:"size:256"`
    Bio       string `ar:"size:4096"`

    // Поле с сериализацией
    Settings string `ar:"serializer:Json;size:2048"`

    // Счётчики с мутаторами
    LoginCount uint32 `ar:"mutators:inc,dec"`
    Points     int64  `ar:"mutators:inc,dec"`

    // Битовые флаги
    Flags uint32 `ar:""`

    // Временные метки
    CreatedAt uint32 `ar:""`
    UpdatedAt uint32 `ar:""`
}
```

## Следующие шаги

- [Индексы](indexes.md) — составные индексы
- [Сериализаторы](serializers.md) — преобразование типов
- [Мутаторы и флаги](mutators.md) — атомарные операции
