# Сериализаторы

Структура `Serializers*` описывает преобразование полей в сложные типы.

## Обзор

Сериализаторы позволяют хранить в БД строку, а работать в коде с типизированными структурами:

```mermaid
graph LR
    A[map[string]any<br/>в коде] -->|Marshal| B[JSON строка<br/>в БД]
    B -->|Unmarshal| A
```

## Базовый синтаксис

```go
type FieldsConfig struct {
    Settings string `ar:"serializer:Json;size:2048"`
}

type SerializersConfig struct {
    Settings map[string]interface{} `ar:""`
}
```

В коде работаем с `map[string]interface{}`:

```go
cfg.SetSettings(map[string]interface{}{
    "theme": "dark",
    "lang":  "ru",
})

settings := cfg.GetSettings()
theme := settings["theme"].(string)
```

## Встроенные сериализаторы

### Json

Использует стандартный `encoding/json`:

```go
type FieldsUser struct {
    Preferences string `ar:"serializer:Json;size:4096"`
}

type SerializersUser struct {
    Preferences map[string]interface{} `ar:""`
}
```

### Printf

Форматирование по шаблону:

```go
type FieldsApp struct {
    Version string `ar:"serializer:Printf,%d.%d.%d;size:16"`
}

type SerializersApp struct {
    Version []int `ar:""`  // [1, 2, 3] → "1.2.3"
}
```

### Mapstructure

Использует [mapstructure](https://pkg.go.dev/github.com/mitchellh/mapstructure) для гибкой десериализации:

```go
type FieldsConfig struct {
    Data string `ar:"serializer:Mapstructure;size:8192"`
}

type SerializersConfig struct {
    Data *MyComplexStruct `ar:""`
}
```

## Пользовательские сериализаторы

### Объявление

```go
type FieldsProduct struct {
    Tags string `ar:"serializer:Tags;size:512"`
}

type SerializersProduct struct {
    Tags []string `ar:"pkg:github.com/myapp/serializers;object:TagsSerializer"`
}
```

### Реализация

```go
package serializers

import "strings"

// TagsSerializerMarshal конвертирует []string в строку для БД
func TagsSerializerMarshal(tags []string) (string, error) {
    return strings.Join(tags, ","), nil
}

// TagsSerializerUnmarshal конвертирует строку из БД в []string
func TagsSerializerUnmarshal(data string) ([]string, error) {
    if data == "" {
        return []string{}, nil
    }
    return strings.Split(data, ","), nil
}
```

### Теги Serializers*

| Тег | Описание | Пример |
|-----|----------|--------|
| `pkg` | Пакет с сериализатором | `pkg:github.com/myapp/serializers` |
| `object` | Имя объекта сериализатора | `object:TagsSerializer` |
| `marshaler` | Имя функции сериализации | `marshaler:CustomMarshal` |
| `unmarshaler` | Имя функции десериализации | `unmarshaler:CustomUnmarshal` |

По умолчанию:
- `marshaler` = `{Name}Marshal`
- `unmarshaler` = `{Name}Unmarshal`

## Параметры сериализатора

Параметры передаются через запятую после имени:

```go
type FieldsLog struct {
    // Параметр "%s: %d" передаётся в Marshal/Unmarshal
    Entry string `ar:"serializer:Printf,%s: %d;size:256"`
}
```

Сигнатура функций с параметрами:

```go
func PrintfMarshal(format string, value []interface{}) (string, error)
func PrintfUnmarshal(format string, data string) ([]interface{}, error)
```

## Сериализаторы со структурами

### Объявление

```go
// Тип данных
package types

type Product struct {
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

// Декларация
type FieldsOrder struct {
    Product string `ar:"serializer:Product;size:1024"`
}

type SerializersOrder struct {
    Product *types.Product `ar:"pkg:github.com/myapp/types;object:ProductS"`
}
```

### Реализация

```go
package types

import "encoding/json"

func ProductSMarshal(p *Product) (string, error) {
    data, err := json.Marshal(p)
    return string(data), err
}

func ProductSUnmarshal(data string) (*Product, error) {
    var p Product
    err := json.Unmarshal([]byte(data), &p)
    return &p, err
}
```

## Важные замечания

!!! warning "UpdateOps и сериализаторы"
    При изменении внутреннего состояния сериализованной структуры поле **не помечается** как изменённое автоматически.

    ```go
    // НЕ сработает — изменение внутренности
    settings := cfg.GetSettings()
    settings["theme"] = "light"
    cfg.Update(ctx)  // Поле не обновится в БД!

    // Правильно — полная замена
    settings := cfg.GetSettings()
    settings["theme"] = "light"
    cfg.SetSettings(settings)  // Помечает поле как изменённое
    cfg.Update(ctx)
    ```

!!! tip "Совет"
    Для атомарного обновления частей сериализованного поля используйте [пользовательские мутаторы](mutators.md).

## Полный пример

```go
package repository

import "github.com/myapp/types"

//ar:serverConf:mydb
//ar:namespace:configs
//ar:backend:postgres
type FieldsConfig struct {
    Id       int64  `ar:"primary_key;init_by_db"`
    Name     string `ar:"size:64"`

    // Встроенный Json
    Settings string `ar:"serializer:Json;size:4096"`

    // Printf формат
    Version string `ar:"serializer:Printf,%d.%d.%d;size:16"`

    // Пользовательский сериализатор
    Metadata string `ar:"serializer:Metadata;size:8192"`
}

type SerializersConfig struct {
    Settings map[string]interface{} `ar:""`
    Version  []int                  `ar:""`
    Metadata *types.Metadata        `ar:"pkg:github.com/myapp/types;object:MetadataS"`
}
```

## Следующие шаги

- [Мутаторы и флаги](mutators.md) — атомарные операции
- [Справочник тегов](../reference/tags.md) — все теги полей
