# Быстрый старт

Создадим простую модель пользователя за 5 минут.

## 1. Создание декларации

Создайте файл `model/repository/declaration/user.go`:

=== "PostgreSQL"

    ```go
    package repository

    //ar:serverConf:mydb
    //ar:namespace:users
    //ar:backend:postgres
    type FieldsUser struct {
        Id        int64  `ar:"primary_key"`
        Email     string `ar:"unique;size:256;selector:SelectByEmail"`
        Name      string `ar:"size:256"`
        CreatedAt uint32 `ar:""`
    }
    ```

=== "Octopus"

    ```go
    package repository

    //ar:serverConf:mydb
    //ar:namespace:5
    //ar:backend:octopus
    type FieldsUser struct {
        Id        int64  `ar:"primary_key"`
        Email     string `ar:"size:256;selector:SelectByEmail"`
        Name      string `ar:"size:256"`
        CreatedAt uint32 `ar:""`
    }
    ```

!!! note "Примечание"
    Для Octopus namespace — это числовой ID спейса. Для PostgreSQL — имя таблицы.

## 2. Генерация кода

```bash
argen --path "model/repository" \
      --declaration "declaration" \
      --destination "generated"
```

После выполнения появится структура:

```
model/repository/
├── declaration/
│   └── user.go
└── generated/
    ├── .argen           # Маркер-файл (не удалять!)
    └── user/
        └── postgres.go  # или octopus.go
```

## 3. Использование модели

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/Educentr/go-activerecord/v3/pkg/activerecord"
    "yourapp/model/repository/generated/user"
)

func main() {
    ctx := context.Background()

    // Регистрация конфигурации (см. раздел Конфигурация)
    activerecord.RegisterConfig(yourConfig)

    // Создание пользователя
    u := user.New(ctx)
    u.SetEmail("john@example.com")
    u.SetName("John Doe")
    u.SetCreatedAt(uint32(time.Now().Unix()))

    if err := u.Insert(ctx); err != nil {
        panic(err)
    }

    fmt.Printf("Created user with ID: %d\n", u.GetId())

    // Поиск по email
    found, err := user.SelectByEmail(ctx, "john@example.com")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found user: %s\n", found.GetName())
}
```

## 4. Добавление индексов

Расширим модель составным индексом:

```go
package repository

//ar:serverConf:mydb
//ar:namespace:users
//ar:backend:postgres
type FieldsUser struct {
    Id        int64  `ar:"primary_key"`
    Email     string `ar:"unique;size:256;selector:SelectByEmail"`
    Name      string `ar:"size:256"`
    Status    string `ar:"size:32"`
    CreatedAt uint32 `ar:""`
}

type IndexesUser struct {
    StatusCreated bool `ar:"fields:Status,CreatedAt;order:Status asc,CreatedAt desc"`
}
```

После перегенерации появится селектор `SelectByStatusCreated`:

```go
limiter := activerecord.NewLimiter(100)
users, err := user.SelectByStatusCreated(ctx,
    user.StatusCreatedIndexType{
        Status:    "active",
        CreatedAt: 1234567890,
    },
    limiter,
)
```

## 5. Добавление мутаторов

Добавим счётчик входов и битовые флаги:

```go
type FieldsUser struct {
    Id         int64  `ar:"primary_key"`
    Email      string `ar:"unique;size:256;selector:SelectByEmail"`
    Name       string `ar:"size:256"`
    LoginCount uint32 `ar:"mutators:inc,dec"`
    Flags      uint32 `ar:""`
    CreatedAt  uint32 `ar:""`
}

type FlagsUser struct {
    Flags string `ar:"flags:Active,Verified,Premium,_,Admin"`
}
```

Использование:

```go
// Атомарный инкремент счётчика
user.IncLoginCount(1)

// Установка флага (атомарная битовая операция)
user.SetBitFlags(user.FlagsVerifiedFlag)

// Применение изменений
user.Update(ctx)

// Проверка флага
if user.GetFlags() & user.FlagsPremiumFlag != 0 {
    fmt.Println("Premium пользователь")
}
```

!!! tip "Совет"
    Мутаторы выполняются атомарно в БД, что исключает race conditions.

## Следующие шаги

- [CRUD операции](../usage/crud.md) — полный обзор операций
- [Декларации полей](../declaration/fields.md) — все доступные теги
- [Конфигурация](../usage/configuration.md) — настройка подключения к БД
