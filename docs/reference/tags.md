# Справочник тегов

Полный справочник всех доступных тегов и комментариев.

## AR комментарии

Комментарии к структуре `Fields*`:

| Комментарий | Описание | Пример |
|-------------|----------|--------|
| `serverConf` | Ключ конфигурации | `//ar:serverConf:mydb` |
| `namespace` | Таблица (PG) или спейс (Octopus) | `//ar:namespace:users` |
| `backend` | Тип БД | `//ar:backend:postgres` |
| `enableSelectAll` | Генерация SelectAll | `//ar:enableSelectAll:10000` |

### Синтаксис

```go
// Отдельные строки
//ar:serverConf:mydb
//ar:namespace:users
//ar:backend:postgres

// Или через точку с запятой
//ar:serverConf:mydb;namespace:users;backend:postgres
```

## Теги Fields*

| Тег | Описание | Пример |
|-----|----------|--------|
| `primary_key` | Первичный ключ | `ar:"primary_key"` |
| `unique` | Уникальный индекс | `ar:"unique"` |
| `selector` | Имя селектора | `ar:"selector:SelectByEmail"` |
| `size` | Размер поля (байты) | `ar:"size:256"` |
| `serializer` | Сериализатор | `ar:"serializer:Json"` |
| `mutators` | Мутаторы | `ar:"mutators:inc,dec"` |
| `init_by_db` | Инициализация БД | `ar:"init_by_db"` |

### Примеры

```go
type FieldsUser struct {
    Id        int64  `ar:"primary_key;init_by_db"`
    Email     string `ar:"unique;size:256;selector:SelectByEmail"`
    Counter   uint32 `ar:"mutators:inc,dec"`
    Settings  string `ar:"serializer:Json;size:4096"`
}
```

## Теги Indexes*

| Тег | Описание | Пример |
|-----|----------|--------|
| `fields` | Поля индекса | `ar:"fields:Email,Status"` |
| `unique` | Уникальный | `ar:"unique"` |
| `primary_key` | Составной PK | `ar:"primary_key"` |
| `selector` | Имя селектора | `ar:"selector:SelectByEmailStatus"` |
| `order` | Сортировка | `ar:"order:Email asc,Status desc"` |
| `condition` | Условие (WHERE) | `ar:"condition:Status[=]1"` |

### Синтаксис условий

| Условие | Синтаксис | SQL |
|---------|-----------|-----|
| Равенство | `Field[=]value` | `field = value` |
| Не равно | `Field[<>]value` | `field <> value` |
| Больше | `Field[>]value` | `field > value` |
| Меньше | `Field[<]value` | `field < value` |
| IS NULL | `Field[is null]` | `field IS NULL` |
| IS NOT NULL | `Field[is not null]` | `field IS NOT NULL` |
| Пустая строка | `Field[=]''` | `field = ''` |
| Битовая маска | `Flags&4[=]4` | `flags & 4 = 4` |

Объединение через `&&`:

```go
ar:"condition:Status[=]1&&DeletedAt[is null]&&Flags&4[=]4"
```

## Теги IndexParts*

| Тег | Описание | Пример |
|-----|----------|--------|
| `index` | Родительский индекс | `ar:"index:StatusCreated"` |
| `fieldnum` | Количество полей | `ar:"fieldnum:1"` |
| `selector` | Имя селектора | `ar:"selector:SelectByStatus"` |

```go
type IndexPartsUser struct {
    StatusOnly bool `ar:"index:StatusCreated;fieldnum:1;selector:SelectByStatusOnly"`
}
```

## Теги Serializers*

| Тег | Описание | Пример |
|-----|----------|--------|
| `pkg` | Пакет сериализатора | `ar:"pkg:github.com/app/ser"` |
| `object` | Имя объекта | `ar:"object:MySerializer"` |
| `marshaler` | Функция сериализации | `ar:"marshaler:CustomMarshal"` |
| `unmarshaler` | Функция десериализации | `ar:"unmarshaler:CustomUnmarshal"` |

```go
type SerializersUser struct {
    Settings map[string]interface{} `ar:""`  // Встроенный
    Profile  *types.Profile         `ar:"pkg:github.com/app/types;object:ProfileS"`
}
```

## Теги Mutators*

| Тег | Описание | Пример |
|-----|----------|--------|
| `update` | Функция/процедура UPDATE | `ar:"update:updateProfile"` |
| `replace` | Функция/процедура REPLACE | `ar:"replace:replaceProfile"` |
| `pkg` | Пакет с функцией | `ar:"pkg:github.com/app/conv"` |

```go
type MutatorsUser struct {
    ProfilePart bool `ar:"update:updateProfile;pkg:github.com/app/conv"`
}
```

## Теги Flags*

| Тег | Описание | Пример |
|-----|----------|--------|
| `flags` | Список флагов | `ar:"flags:Active,Verified,_,Admin"` |

- `_` пропускает позицию бита
- Генерирует константы `{Field}{Flag}Flag`

```go
type FlagsUser struct {
    Status string `ar:"flags:Active,Verified,Premium,_,Admin"`
}

// Генерирует:
// StatusActiveFlag   = 1 << 0
// StatusVerifiedFlag = 1 << 1
// StatusPremiumFlag  = 1 << 2
// StatusAdminFlag    = 1 << 4
```

## Теги Triggers*

| Тег | Описание | Пример |
|-----|----------|--------|
| `pkg` | Пакет с функцией | `ar:"pkg:github.com/app/repair"` |
| `func` | Имя функции | `ar:"func:RepairTuple"` |

```go
type TriggersUser struct {
    RepairTuple bool `ar:"pkg:github.com/app/repair;func:RepairUserTuple"`
}
```

## Теги FieldsObject*

| Тег | Описание | Пример |
|-----|----------|--------|
| `key` | Поле в связанном объекте | `ar:"key:Id"` |
| `object` | Имя связанной модели | `ar:"object:User"` |
| `field` | Поле в текущем объекте | `ar:"field:UserId"` |
| `shard_by` | Функция шардирования | `ar:"shard_by:*"` |

```go
type FieldsObjectOrder struct {
    User int64 `ar:"key:Id;object:User;field:UserId"`
}
```

## Теги ProcFields*

| Тег | Описание | Пример |
|-----|----------|--------|
| `input` | Входной параметр | `ar:"input"` |
| `output` | Выходной параметр | `ar:"output"` |
| `output:N` | Выходной с номером | `ar:"output:2"` |
| `serializer` | Сериализатор | `ar:"serializer:Json"` |
| `size` | Размер параметра | `ar:"size:256"` |

```go
type ProcFieldsStats struct {
    UserId string `ar:"input;size:32"`
    Result int64  `ar:"output"`
}
```

## Встроенные мутаторы

| Мутатор | Метод | Операция |
|---------|-------|----------|
| `inc` | `Inc{Field}(delta)` | `field += delta` |
| `dec` | `Dec{Field}(delta)` | `field -= delta` |
| `and` | `And{Field}(mask)` | `field &= mask` |
| `or` | `Or{Field}(mask)` | `field \|= mask` |
| `xor` | `Xor{Field}(mask)` | `field ^= mask` |
| `set_bit` | `SetBit{Field}(mask)` | `field \|= mask` |
| `clear_bit` | `ClearBit{Field}(mask)` | `field &= ~mask` |

## Встроенные сериализаторы

| Сериализатор | Описание | Пример |
|--------------|----------|--------|
| `Json` | encoding/json | `serializer:Json` |
| `Printf` | Форматирование | `serializer:Printf,%d.%d.%d` |
| `Mapstructure` | mapstructure | `serializer:Mapstructure` |
