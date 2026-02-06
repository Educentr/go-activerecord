# Работа с tuples

Структура данных tuple и особенности работы с Octopus.

## Что такое tuple

Tuple — это упорядоченный массив полей в Octopus/Tarantool:

```
Tuple: [field0, field1, field2, field3, ...]
        ↓       ↓       ↓       ↓
        Id      Email   Name    Status
```

## Порядок полей

!!! danger "Критично"
    Порядок полей в декларации **должен точно соответствовать** порядку в tuple!

```go
// Tuple в БД: [id, email, name, status]
type FieldsUser struct {
    Id     int64  `ar:"primary_key"`  // Поле 0
    Email  string `ar:"size:256"`      // Поле 1
    Name   string `ar:"size:256"`      // Поле 2
    Status string `ar:"size:32"`       // Поле 3
}
```

Несоответствие порядка приведёт к неправильной интерпретации данных!

## Порядок индексов

Порядок индексов также критичен:

```go
type IndexesUser struct {
    // Должен соответствовать конфигурации Octopus
    Id     bool `ar:"fields:Id;primary_key"`      // Индекс 0
    Email  bool `ar:"fields:Email;unique"`         // Индекс 1
    Status bool `ar:"fields:Status"`               // Индекс 2
}
```

## Дополнительные поля (extraFields)

### Когда появляются

Если tuple содержит больше полей, чем объявлено:

```go
// Объявлено 3 поля
type FieldsUser struct {
    Id    int64  `ar:"primary_key"`
    Email string `ar:"size:256"`
    Name  string `ar:"size:256"`
}

// Tuple в БД: [id, email, name, legacy_field, another_field]
// legacy_field и another_field попадут в extraFields
```

### Доступ к extraFields

```go
user, _ := user.SelectById(ctx, 123)

// extraFields содержит "лишние" поля
for i, field := range user.ExtraFields {
    fmt.Printf("Extra field %d: %v\n", i, field)
}
```

### Сохранение при Update

При `Update()` extraFields сохраняются:

```go
user.SetName("New Name")
user.Update(ctx)  // extraFields не затрагиваются
```

При `Replace()` extraFields теряются:

```go
user.Replace(ctx)  // Записываются только объявленные поля
```

## RepairTuple триггер

### Назначение

Вызывается при проблемах с десериализацией:

- Неверное число полей
- Неверный формат поля
- Повреждённые данные

### Объявление

```go
type TriggersUser struct {
    RepairTuple bool `ar:"pkg:github.com/myapp/repair;func:RepairUserTuple"`
}
```

### Реализация

```go
package repair

import "github.com/Educentr/go-activerecord/v3/pkg/octopus"

func RepairUserTuple(tuple *octopus.TupleData) error {
    expectedFields := 4

    // Добавляем недостающие поля
    for len(tuple.Fields) < expectedFields {
        tuple.Fields = append(tuple.Fields, []byte(""))
    }

    // Проверяем формат поля 0 (Id)
    if len(tuple.Fields[0]) != 8 {
        // Исправляем или логируем
        return fmt.Errorf("invalid Id format")
    }

    return nil
}
```

### Флаг Repaired

После восстановления:

```go
user, _ := user.SelectById(ctx, 123)

if user.Repaired {
    // Данные были восстановлены
    // Update() будет работать как Replace()
    log.Println("User data was repaired")
}
```

## Сериализация типов

### Числовые типы

| Go тип | Размер в tuple |
|--------|----------------|
| `int8`, `uint8` | 1 байт |
| `int16`, `uint16` | 2 байта |
| `int32`, `uint32` | 4 байта |
| `int64`, `uint64` | 8 байт |

### Строки

Строки хранятся как есть, размер определяется тегом `size`:

```go
Email string `ar:"size:256"`  // Максимум 256 байт
```

### Сериализаторы

Для сложных типов используйте сериализаторы:

```go
type FieldsUser struct {
    Settings string `ar:"serializer:Json;size:2048"`
}

type SerializersUser struct {
    Settings map[string]interface{} `ar:""`
}
```

## Пример миграции схемы

### Добавление поля

1. Добавьте поле в декларацию:

```go
type FieldsUser struct {
    Id     int64  `ar:"primary_key"`
    Email  string `ar:"size:256"`
    Name   string `ar:"size:256"`
    Status string `ar:"size:32"`  // Новое поле
}
```

2. Добавьте триггер для старых записей:

```go
type TriggersUser struct {
    RepairTuple bool `ar:"pkg:github.com/myapp/repair;func:RepairUserTuple"`
}
```

```go
func RepairUserTuple(tuple *octopus.TupleData) error {
    if len(tuple.Fields) < 4 {
        // Добавляем Status по умолчанию
        tuple.Fields = append(tuple.Fields, []byte("active"))
    }
    return nil
}
```

3. Перегенерируйте код:

```bash
argen --path "model/repository"
```

## Best Practices

### 1. Всегда объявляйте триггер

```go
type TriggersModel struct {
    RepairTuple bool `ar:"pkg:github.com/myapp/repair;func:Repair"`
}
```

### 2. Логируйте восстановления

```go
func RepairTuple(tuple *octopus.TupleData) error {
    if len(tuple.Fields) < expected {
        log.Printf("Repairing tuple with %d fields", len(tuple.Fields))
    }
    // ...
}
```

### 3. Не удаляйте поля

При удалении поля из декларации старые данные станут extraFields. Лучше пометить поле как deprecated.

### 4. Тестируйте миграции

```go
func TestRepairTuple(t *testing.T) {
    tuple := &octopus.TupleData{
        Fields: [][]byte{
            {0, 0, 0, 0, 0, 0, 0, 1},  // Id
            []byte("email"),           // Email
            // Name отсутствует
        },
    }

    err := RepairTuple(tuple)
    require.NoError(t, err)
    require.Len(t, tuple.Fields, 3)
}
```

## Следующие шаги

- [Хранимые процедуры](procedures.md) — вызов Lua процедур
- [Триггеры](../declaration/triggers.md) — декларация триггеров
