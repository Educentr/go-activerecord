# Использование

Этот раздел описывает работу со сгенерированным кодом.

## Содержание раздела

- [CRUD операции](crud.md) — создание, чтение, обновление, удаление
- [Селекторы](selectors.md) — поиск записей
- [Атомарные операции](atomic-ops.md) — мутаторы и битовые операции
- [Конфигурация](configuration.md) — настройка подключений

## Обзор

После генерации для каждой модели доступны:

```mermaid
graph TB
    subgraph "Сгенерированный пакет"
        A[New(ctx)] --> B[Объект]
        B --> C[Аксессоры<br/>Get*/Set*]
        B --> D[CRUD<br/>Insert/Update/Delete]
        B --> E[Мутаторы<br/>Inc*/SetBit*]
    end

    subgraph "Селекторы"
        F[SelectById]
        G[SelectByEmail]
        H[SelectByIndex]
    end
```

| Категория | Методы |
|-----------|--------|
| Создание | `New(ctx)` |
| CRUD | `Insert`, `Update`, `Replace`, `InsertOrReplace`, `Delete` |
| Аксессоры | `Get{Field}`, `Set{Field}` |
| Селекторы | `SelectBy{Index}`, `SelectBy{Index}s` |
| Мутаторы | `Inc{Field}`, `Dec{Field}`, `SetBit{Field}`, `ClearBit{Field}` |

## Базовый пример

```go
package main

import (
    "context"
    "github.com/Educentr/go-activerecord/v3/pkg/activerecord"
    "yourapp/model/repository/generated/user"
)

func main() {
    ctx := context.Background()

    // Инициализация
    activerecord.RegisterConfig(config)

    // Создание
    u := user.New(ctx)
    u.SetEmail("john@example.com")
    u.SetName("John")
    u.Insert(ctx)

    // Чтение
    found, _ := user.SelectByEmail(ctx, "john@example.com")

    // Обновление
    found.SetName("John Doe")
    found.IncLoginCount(1)
    found.Update(ctx)

    // Удаление
    found.Delete(ctx)
}
```

!!! tip "Совет"
    Начните с раздела [CRUD операции](crud.md) для понимания жизненного цикла объекта.

## Следующие шаги

- [CRUD операции](crud.md) — полный цикл работы с данными
- [Конфигурация](configuration.md) — настройка подключения к БД
