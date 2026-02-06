# Индексы

Структуры `Indexes*` и `IndexParts*` описывают составные индексы и частичные селекторы.

## Indexes*

### Базовый синтаксис

```go
type IndexesUser struct {
    EmailStatus   bool `ar:"fields:Email,Status;unique"`
    StatusCreated bool `ar:"fields:Status,CreatedAt;order:Status asc,CreatedAt desc"`
}
```

### Теги индексов

| Тег | Описание | Пример |
|-----|----------|--------|
| `fields` | Поля индекса через запятую | `fields:Email,Status` |
| `unique` | Уникальный индекс | `unique` |
| `primary_key` | Составной первичный ключ | `primary_key` |
| `selector` | Имя селектора | `selector:SelectByEmailStatus` |
| `order` | Направление сортировки | `order:Status asc,CreatedAt desc` |
| `condition` | Условие индекса (WHERE) | `condition:Status[=]1` |

### Уникальные индексы

```go
type IndexesUser struct {
    // Уникальная комбинация email + type
    EmailType bool `ar:"fields:Email,Type;unique"`
}
```

Генерируемый селектор возвращает один объект:

```go
user, err := user.SelectByEmailType(ctx, user.EmailTypeIndexType{
    Email: "john@example.com",
    Type:  "admin",
})
```

### Неуникальные индексы

```go
type IndexesUser struct {
    StatusCreated bool `ar:"fields:Status,CreatedAt"`
}
```

Селектор требует лимитер:

```go
limiter := activerecord.NewLimiter(100)
users, err := user.SelectByStatusCreated(ctx, user.StatusCreatedIndexType{
    Status:    "active",
    CreatedAt: 1234567890,
}, limiter)
```

### Направление сортировки

```go
type IndexesUser struct {
    // Status по возрастанию, CreatedAt по убыванию
    RecentActive bool `ar:"fields:Status,CreatedAt;order:Status asc,CreatedAt desc"`
}
```

!!! note "Для Octopus"
    Устаревший тег `orderdesc` заменён на `order`. Используйте `order` для совместимости.

### Условные индексы (PostgreSQL)

```go
type IndexesUser struct {
    // Индекс только для активных пользователей
    ActiveUsers bool `ar:"fields:Status,CreatedAt;condition:Status[=]1&&DeletedAt[is null]"`

    // Индекс для пустых строк
    EmptyEmail bool `ar:"fields:Email;condition:Email[=]''"`

    // Битовая маска
    PremiumUsers bool `ar:"fields:CreatedAt;condition:Flags&4[=]4"`
}
```

**Синтаксис условий:**

| Условие | Пример | SQL |
|---------|--------|-----|
| Равенство | `Status[=]1` | `status = 1` |
| Не равно | `Status[<>]0` | `status <> 0` |
| IS NULL | `DeletedAt[is null]` | `deleted_at IS NULL` |
| IS NOT NULL | `DeletedAt[is not null]` | `deleted_at IS NOT NULL` |
| Пустая строка | `Email[=]''` | `email = ''` |
| Битовая маска | `Flags&4[=]4` | `flags & 4 = 4` |

Условия объединяются через `&&`:

```go
condition:Status[=]1&&Flags&1[=]1&&DeletedAt[is null]
```

!!! note "Поддержка в бэкендах"
    Условные индексы поддерживаются в PostgreSQL и Octopus. Условия автоматически добавляются во все запросы по этому индексу.

## IndexParts*

Частичные индексы позволяют делать выборки по части составного индекса.

### Синтаксис

```go
type IndexesUser struct {
    StatusCreatedId bool `ar:"fields:Status,CreatedAt,Id"`
}

type IndexPartsUser struct {
    // Выборка только по Status (первое поле)
    StatusPart bool `ar:"index:StatusCreatedId;fieldnum:1;selector:SelectByStatus"`

    // Выборка по Status и CreatedAt (первые 2 поля)
    StatusCreatedPart bool `ar:"index:StatusCreatedId;fieldnum:2;selector:SelectByStatusCreated"`
}
```

### Теги IndexParts

| Тег | Описание | Пример |
|-----|----------|--------|
| `index` | Имя родительского индекса | `index:StatusCreatedId` |
| `fieldnum` | Количество используемых полей | `fieldnum:2` |
| `selector` | Имя селектора | `selector:SelectByStatus` |

### Использование

```go
// Полный индекс (все 3 поля)
users, err := user.SelectByStatusCreatedId(ctx, user.StatusCreatedIdIndexType{
    Status:    "active",
    CreatedAt: 1234567890,
    Id:        123,
}, limiter)

// Частичный индекс (1 поле)
users, err := user.SelectByStatus(ctx, "active", limiter)

// Частичный индекс (2 поля)
users, err := user.SelectByStatusCreated(ctx, user.StatusCreatedPartIndexType{
    Status:    "active",
    CreatedAt: 1234567890,
}, limiter)
```

!!! tip "Наследование условий"
    Частичные индексы наследуют условия (`condition`) от родительского индекса.

## Структуры типов индексов

Для составных индексов генерируются структуры:

```go
// Для индекса EmailType
type EmailTypeIndexType struct {
    Email string
    Type  string
}

// Для индекса StatusCreatedId
type StatusCreatedIdIndexType struct {
    Status    string
    CreatedAt uint32
    Id        int64
}
```

Использование:

```go
key := user.EmailTypeIndexType{
    Email: "john@example.com",
    Type:  "admin",
}
user, err := user.SelectByEmailType(ctx, key)
```

## Составной первичный ключ

```go
type FieldsUserRole struct {
    UserId int64  `ar:""`
    RoleId int64  `ar:""`
    Flags  uint32 `ar:""`
}

type IndexesUserRole struct {
    UserRole bool `ar:"fields:UserId,RoleId;primary_key"`
}
```

## Порядок индексов (Octopus)

!!! danger "Критично для Octopus"
    Порядок объявления индексов должен соответствовать конфигурации Octopus!

```go
type IndexesUser struct {
    // Индекс 0 (primary)
    Id bool `ar:"fields:Id;primary_key"`
    // Индекс 1
    Email bool `ar:"fields:Email;unique"`
    // Индекс 2
    Status bool `ar:"fields:Status"`
}
```

## Полный пример

```go
package repository

//ar:serverConf:mydb
//ar:namespace:users
//ar:backend:postgres
type FieldsUser struct {
    Id        int64  `ar:"primary_key;init_by_db"`
    Email     string `ar:"size:256"`
    Status    string `ar:"size:32"`
    Type      string `ar:"size:32"`
    Flags     uint32 `ar:""`
    CreatedAt uint32 `ar:""`
    DeletedAt uint32 `ar:""`
}

type IndexesUser struct {
    // Уникальный составной индекс
    EmailType bool `ar:"fields:Email,Type;unique"`

    // Условный индекс для активных
    ActiveByDate bool `ar:"fields:Status,CreatedAt;condition:Status[=]active&&DeletedAt[is null];order:CreatedAt desc"`

    // Индекс с битовой маской
    PremiumUsers bool `ar:"fields:CreatedAt;condition:Flags&4[=]4"`
}

type IndexPartsUser struct {
    // Частичный селектор по Status
    StatusPart bool `ar:"index:ActiveByDate;fieldnum:1;selector:SelectByActiveStatus"`
}
```

## Следующие шаги

- [Сериализаторы](serializers.md) — преобразование типов
- [Селекторы](../usage/selectors.md) — использование селекторов
