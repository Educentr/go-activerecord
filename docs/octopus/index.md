# Octopus / Tarantool 1.5

Особенности работы с Octopus и Tarantool 1.5 бэкендом.

## Содержание раздела

- [Работа с tuples](tuples.md) — структура данных и особенности
- [Хранимые процедуры](procedures.md) — ProcFields* и Lua

## Обзор

Octopus бэкенд использует бинарный протокол iproto:

| Особенность | Описание |
|-------------|----------|
| Протокол | iproto (binary) |
| Namespace | Числовой ID спейса |
| Данные | Tuples (массивы полей) |
| Порядок полей | Критичен! |
| Порядок индексов | Критичен! |
| Хранимые процедуры | Lua |

## Декларация

```go
//ar:serverConf:octconf
//ar:namespace:5          // Числовой ID спейса
//ar:backend:octopus
type FieldsUser struct {
    Id    int64  `ar:"primary_key"`     // Поле 0 в tuple
    Email string `ar:"size:256"`         // Поле 1 в tuple
    Name  string `ar:"size:256"`         // Поле 2 в tuple
}
```

!!! danger "Порядок полей критичен"
    Порядок полей в `Fields*` должен **точно** соответствовать порядку полей в tuple базы данных!

## Порядок индексов

Порядок объявления индексов должен соответствовать конфигурации Octopus:

```go
type IndexesUser struct {
    // Индекс 0 (обычно primary)
    Id bool `ar:"fields:Id;primary_key"`
    // Индекс 1
    Email bool `ar:"fields:Email;unique"`
    // Индекс 2
    Status bool `ar:"fields:Status"`
}
```

## Дополнительные поля (extraFields)

Если tuple содержит больше полей, чем объявлено:

```go
// Объявлено 3 поля
type FieldsUser struct {
    Id    int64
    Email string
    Name  string
}

// Tuple содержит 5 полей: [id, email, name, extra1, extra2]
// extra1 и extra2 сохраняются в extraFields
```

## Триггер RepairTuple

При проблемах с десериализацией вызывается триггер:

```go
type TriggersUser struct {
    RepairTuple bool `ar:"pkg:github.com/myapp/repair;func:RepairTuple"`
}
```

```go
func RepairTuple(tuple *octopus.TupleData) error {
    // Добавляем недостающие поля
    for len(tuple.Fields) < 3 {
        tuple.Fields = append(tuple.Fields, []byte(""))
    }
    return nil
}
```

Подробнее: [Работа с tuples](tuples.md)

## Хранимые процедуры

```go
//ar:serverConf:octconf
//ar:namespace:calculate_stats  // Имя процедуры
//ar:backend:octopus
type ProcFieldsStats struct {
    UserId string `ar:"input"`
    Result int64  `ar:"output"`
}
```

Подробнее: [Хранимые процедуры](procedures.md)

## Установка

```bash
# Ubuntu/Debian
apt-get install tarantool-lts
```

Пакет: [packages.debian.org/tarantool-lts](https://packages.debian.org/ru/buster/tarantool-lts)

## Протокол iproto

Документация: [octopus/doc/silverbox-protocol.txt](https://github.com/Vespertinus/octopus/blob/master/doc/silverbox-protocol.txt)

## Следующие шаги

- [Работа с tuples](tuples.md) — детали структуры данных
- [Хранимые процедуры](procedures.md) — вызов Lua процедур
