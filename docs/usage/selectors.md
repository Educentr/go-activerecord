# Селекторы

Селекторы — методы для поиска записей по индексам.

## Типы селекторов

### По уникальному индексу

Возвращает один объект:

```go
// Единственная запись
user, err := user.SelectById(ctx, 123)
if err != nil {
    return err
}
```

### По неуникальному индексу

Требует лимитер, возвращает слайс:

```go
limiter := activerecord.NewLimiter(100)
users, err := user.SelectByStatus(ctx, "active", limiter)
if err != nil {
    return err
}
```

### Множественная выборка

Суффикс `s` — выборка по нескольким ключам:

```go
// Несколько ID
ids := []int64{1, 2, 3, 4, 5}
users, err := user.SelectByIds(ctx, ids)

// Несколько email
emails := []string{"a@b.com", "c@d.com"}
users, err := user.SelectByEmails(ctx, emails)
```

## Лимитеры

### NewLimiter

Ограничение количества записей:

```go
limiter := activerecord.NewLimiter(100)
users, err := user.SelectByStatus(ctx, "active", limiter)
```

### NewLimitOffset

Пагинация с offset:

```go
// Страница 2 (записи 100-199)
limiter := activerecord.NewLimitOffset(100, 100)
users, err := user.SelectByStatus(ctx, "active", limiter)
```

### NewThreshold

С предупреждением при достижении лимита:

```go
limiter := activerecord.NewThreshold(1000)
users, err := user.SelectByStatus(ctx, "active", limiter)

// Если найдено 1000 записей — будет warning в логах
```

### Интерфейс SelectorLimiter

```go
type SelectorLimiter interface {
    Limit() uint32       // Максимум записей
    Offset() uint32      // Смещение
    FulfillWarn() bool   // Предупреждать при достижении
    String() string
}
```

## Составные индексы

Для составных индексов генерируются структуры типов:

```go
type IndexesUser struct {
    StatusCreated bool `ar:"fields:Status,CreatedAt"`
}
```

Использование:

```go
key := user.StatusCreatedIndexType{
    Status:    "active",
    CreatedAt: 1234567890,
}

limiter := activerecord.NewLimiter(100)
users, err := user.SelectByStatusCreated(ctx, key, limiter)
```

### Множественная выборка по составному индексу

```go
keys := []user.StatusCreatedIndexType{
    {Status: "active", CreatedAt: 1234567890},
    {Status: "pending", CreatedAt: 1234567891},
}

users, err := user.SelectByStatusCreateds(ctx, keys, limiter)
```

## SelectAll

Выборка всех записей (если включено):

```go
//ar:enableSelectAll:10000
type FieldsDict struct {
    // ...
}
```

```go
// Выборка всех с лимитом
limiter := activerecord.NewLimiter(10000)
items, err := dict.SelectAll(ctx, limiter)
```

!!! warning "Осторожно"
    `SelectAll` может вернуть много данных. Максимальный лимит — 20000.

## Обработка результатов

### Уникальный индекс

```go
user, err := user.SelectByEmail(ctx, "john@example.com")
if err != nil {
    if errors.Is(err, activerecord.ErrNotFound) {
        return nil, fmt.Errorf("user not found")
    }
    return nil, err
}

// user гарантированно не nil
fmt.Println(user.GetName())
```

### Неуникальный индекс

```go
limiter := activerecord.NewLimiter(100)
users, err := user.SelectByStatus(ctx, "active", limiter)
if err != nil {
    return err
}

// users может быть пустым слайсом
if len(users) == 0 {
    fmt.Println("No active users")
    return nil
}

for _, u := range users {
    fmt.Println(u.GetName())
}
```

## Частичные индексы (IndexParts)

Выборка по части составного индекса:

```go
type IndexesUser struct {
    StatusCreatedId bool `ar:"fields:Status,CreatedAt,Id"`
}

type IndexPartsUser struct {
    StatusPart bool `ar:"index:StatusCreatedId;fieldnum:1;selector:SelectByStatusOnly"`
}
```

```go
// Выборка только по Status
users, err := user.SelectByStatusOnly(ctx, "active", limiter)
```

## Условные индексы

Условия автоматически добавляются в запрос:

```go
type IndexesUser struct {
    ActiveUsers bool `ar:"fields:CreatedAt;condition:Status[=]active&&DeletedAt[is null]"`
}
```

```go
// Условие Status='active' AND DeletedAt IS NULL добавится автоматически
users, err := user.SelectByActiveUsers(ctx,
    user.ActiveUsersIndexType{CreatedAt: 1234567890},
    limiter,
)
```

## Производительность

### Количество ключей

```go
// ✅ Хорошо — разумное количество
ids := []int64{1, 2, 3, 4, 5}
users, _ := user.SelectByIds(ctx, ids)

// ⚠️ Плохо — слишком много ключей
ids := make([]int64, 10000)
users, _ := user.SelectByIds(ctx, ids)  // IN с 10000 значений
```

### Пагинация

```go
// Offset-based пагинация
func GetPage(ctx context.Context, page, pageSize int) ([]*user.User, error) {
    offset := (page - 1) * pageSize
    limiter := activerecord.NewLimitOffset(uint32(pageSize), uint32(offset))
    return user.SelectByStatus(ctx, "active", limiter)
}
```

## Best Practices

### 1. Всегда используйте лимитеры

```go
// ❌ Плохо — неуникальный индекс без лимита
users, _ := user.SelectByStatus(ctx, "active", nil)

// ✅ Хорошо
users, _ := user.SelectByStatus(ctx, "active", activerecord.NewLimiter(100))
```

### 2. Проверяйте ошибки

```go
user, err := user.SelectById(ctx, 123)
if err != nil {
    // Обработка ошибки
    return err
}
```

### 3. Используйте правильный селектор

```go
// ❌ Выборка всех + фильтрация в коде
users, _ := user.SelectAll(ctx, limiter)
active := filterActive(users)

// ✅ Выборка по индексу
users, _ := user.SelectByStatus(ctx, "active", limiter)
```

### 4. Контекст с таймаутом

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

users, err := user.SelectByStatus(ctx, "active", limiter)
```

## Следующие шаги

- [Атомарные операции](atomic-ops.md) — мутаторы
- [Индексы](../declaration/indexes.md) — декларация индексов
