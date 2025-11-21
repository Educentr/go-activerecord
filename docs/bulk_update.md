# BulkUpdate для PostgreSQL

## Обзор

Функция `BulkUpdate` выполняет массовое обновление нескольких объектов и возвращает детальную информацию о результатах.

## Сигнатура

```go
func BulkUpdate(ctx context.Context, objs []*User) (updatedCount int, err error)
```

**Параметры:**

- `ctx` - контекст выполнения
- `objs` - слайс объектов для обновления

**Возвращаемые значения:**

- `updatedCount` - количество фактически обновленных строк в БД
- `err` - ошибка выполнения (если есть)

## Как это работает

### 1. Валидация

Для каждого объекта проверяется:

- Объект должен существовать (`Exists == true`)
- Объект должен иметь изменения (`len(UpdateOps) > 0`)

### 2. Выполнение обновления

Функция генерирует **один SQL запрос** используя паттерн `UPDATE ... FROM (VALUES ...)`:

1. **Сбор всех полей**: Собираются все уникальные поля из `UpdateOps` всех объектов
2. **Сортировка полей**: Поля сортируются для стабильности генерации запроса
3. **Генерация VALUES**: Для каждого объекта формируется строка значений (NULL для полей, которые не обновляются)
4. **Выполнение запроса**: Один SQL запрос обновляет все объекты

### 3. Обработка результатов

- Функция возвращает количество фактически обновленных строк из `CommandTag().RowsAffected()`
- При ошибке выполнения запроса возвращается ошибка, обновления не применяются

## Пример использования

### Базовое использование

```go
package main

import (
    "context"
    "fmt"
    "yourapp/model/repository/generated/user"
)

func updateUsers(ctx context.Context) error {
    // Получаем пользователей
    users, err := user.SelectByIds(ctx, []int64{1, 2, 3})
    if err != nil {
        return err
    }

    // Изменяем пользователей
    for _, u := range users {
        u.SetActive(true)
        u.SetLastLogin(time.Now().Unix())
    }

    // Массовое обновление
    updatedCount, err := user.BulkUpdate(ctx, users)
    if err != nil {
        return fmt.Errorf("bulk update failed: %w", err)
    }

    fmt.Printf("Updated: %d/%d\n", updatedCount, len(users))

    return nil
}
```

### Использование с транзакцией

```go
func updateWithTransaction(ctx context.Context, users []*user.User) error {
    // Начинаем транзакцию
    tx, err := db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // Массовое обновление в транзакции
    updatedCount, err := user.BulkUpdate(ctx, users)
    if err != nil {
        return fmt.Errorf("bulk update failed: %w", err)
    }

    logger.Info("Successfully updated %d users", updatedCount)

    // Коммитим транзакцию
    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit failed: %w", err)
    }

    return nil
}
```

### Фильтрация объектов без изменений

```go
func updateOnlyModified(ctx context.Context, users []*user.User) error {
    // Фильтруем только измененные объекты
    modified := make([]*user.User, 0, len(users))
    for _, u := range users {
        if len(u.BaseField.UpdateOps) > 0 {
            modified = append(modified, u)
        }
    }

    if len(modified) == 0 {
        return nil // Нечего обновлять
    }

    updatedCount, err := user.BulkUpdate(ctx, modified)
    if err != nil {
        return fmt.Errorf("bulk update failed: %w", err)
    }

    fmt.Printf("Updated %d objects\n", updatedCount)

    return nil
}
```

## Обработка ошибок

Функция возвращает единственную ошибку, которая может возникнуть:

- **Ошибка валидации**: Если все объекты не прошли валидацию (не существуют или нет изменений)
- **Ошибка генерации SQL**: Проблемы при формировании запроса
- **Ошибка выполнения БД**: Проблемы при выполнении UPDATE запроса

```go
updatedCount, err := user.BulkUpdate(ctx, users)
if err != nil {
    // Обработка ошибки
    log.Printf("Bulk update failed: %v", err)
    return err
}
```

**Важно:** При ошибке **ни один** объект не будет обновлен (атомарная операция в рамках одного SQL запроса).

## Метрики

Функция собирает следующие метрики:

- `bulk_update_request` - количество вызовов BulkUpdate
- `bulk_update_objects` - общее количество объектов для обновления
- `bulk_update_success` - количество успешно обновленных строк (из RowsAffected)
- `bulk_update_empty` - количество объектов без изменений (пропущенных при валидации)
- `bulk_update_notexists` - попытка обновить несуществующий объект
- `bulk_update_failed` - неудачное выполнение bulk update запроса

## Особенности реализации

### Текущая версия

**Bulk SQL через UPDATE FROM:**

```sql
UPDATE users AS t
SET
    name = v.name,
    email = v.email,
    active = v.active
FROM (VALUES
    ($1, $2, $3, $4),  -- (pk, name, email, active)
    ($5, $6, $7, $8)
) AS v(id, name, email, active)
WHERE t.id = v.id
```

**Преимущества:**

- Один SQL запрос для всех объектов
- Значительно быстрее для больших объемов
- Атомарная операция (все или ничего)
- Эффективное использование сетевых ресурсов

**Как решаются сложности:**

- **Разные поля у объектов**: Собираются все уникальные поля, для отсутствующих используется NULL
- **NULL значения**: PostgreSQL корректно обрабатывает NULL в UPDATE (не изменяет поле)
- **Стабильность**: Поля сортируются для одинакового порядка при повторных вызовах

## Производительность

### Текущая реализация

Для N объектов:

- **Запросов к БД:** 1 (один SQL запрос)
- **Время:** O(1) относительно количества объектов (с точки зрения сетевых roundtrips)
- **Рекомендуется:** до 1000 объектов в одном вызове

### Рекомендации

1. **Малые и средние пакеты (< 1000 объектов):**

   ```go
   updatedCount, err := user.BulkUpdate(ctx, users)
   if err != nil {
       return err
   }
   ```

2. **Большие пакеты (> 1000 объектов):**

   ```go
   // Разбивайте на батчи для контроля памяти и размера SQL
   batchSize := 500
   totalUpdated := 0

   for i := 0; i < len(users); i += batchSize {
       end := i + batchSize
       if end > len(users) {
           end = len(users)
       }

       batch := users[i:end]
       updatedCount, err := user.BulkUpdate(ctx, batch)
       if err != nil {
           return fmt.Errorf("batch %d-%d failed: %w", i, end, err)
       }

       totalUpdated += updatedCount
   }

   fmt.Printf("Total updated: %d\n", totalUpdated)
   ```

3. **С транзакцией (для атомарности всех батчей):**

   ```go
   tx, _ := db.Begin(ctx)
   defer tx.Rollback(ctx)

   updatedCount, err := user.BulkUpdate(ctx, users)
   if err != nil {
       return err
   }

   return tx.Commit(ctx)
   ```

## Сравнение с альтернативами

### vs Цикл с Update()

```go
// Вариант 1: Явный цикл
for _, u := range users {
    if err := u.Update(ctx); err != nil {
        return err // Первая ошибка прерывает процесс
    }
}
// N запросов к БД

// Вариант 2: BulkUpdate
updatedCount, err := user.BulkUpdate(ctx, users)
if err != nil {
    return err
}
// 1 запрос к БД
```

**BulkUpdate лучше когда:**

- Обновляется много объектов (> 10)
- Важна производительность
- Нужна атомарность всех обновлений в одном SQL
- Нужны метрики bulk операций

**Явный цикл лучше когда:**

- Обновляется 1-5 объектов (overhead минимален)
- Нужна специфическая логика между обновлениями
- Важен контроль над каждым отдельным обновлением

## Ограничения

1. **Один SQL запрос для всех объектов**
   - При ошибке **ни один** объект не обновится
   - Нет возможности получить информацию о том, какой конкретно объект вызвал ошибку
   - Все объекты должны быть валидными для успешного выполнения

2. **NULL для необновляемых полей**
   - Если объекты обновляют разные наборы полей, для отсутствующих используется NULL
   - PostgreSQL корректно обрабатывает это, но стоит помнить об этой особенности

3. **Размер SQL запроса**
   - Для очень большого количества объектов (> 1000) SQL может стать слишком большим
   - Рекомендуется разбивать на батчи по 500-1000 объектов

4. **Не поддерживаются idempotency keys**
   - Bulk update не работает с ключами идемпотентности
   - Для этого используйте обычный Update() в цикле

## Roadmap

### Возможные улучшения

**Автоматическое разбиение на батчи:**

- Автоматическое определение оптимального размера батча
- Прозрачная обработка больших объемов без ручного разбиения
- Настраиваемые параметры батчинга

**Детальная информация об ошибках:**

- При ошибке предоставлять информацию о проблемном объекте
- Возможность частичного успеха с возвратом списка ошибок
- Режим "best effort" vs "all or nothing"

**RETURNING поддержка:**

- Возврат обновленных значений полей с триггерами и вычисляемыми значениями
- Синхронизация объектов с актуальным состоянием БД после обновления
