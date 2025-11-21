# BulkInsertReplace для PostgreSQL

## Обзор

Функция `BulkInsertReplace` реализует массовую вставку записей с автоматическим сопоставлением автоинкрементных ключей через уникальный индекс.

## Как это работает

### 1. Генерация функции

Функция **всегда генерируется** для всех моделей PostgreSQL. Однако для корректной работы требуется **уникальный индекс без полей InitByDB** (автоинкрементных полей), который используется для сопоставления результатов из БД с исходными объектами.

Если такой индекс отсутствует, функция возвращает ошибку при вызове.

### 2. Выбор индекса для сопоставления

Функция автоматически выбирает индекс для сопоставления результатов с приоритетом:

1. **Первый приоритет**: `conflictIndex` (переданный как параметр)
   - Если он уникальный и не содержит InitByDB полей
   - Это наиболее логичный выбор, так как конфликт происходит по этому индексу

2. **Второй приоритет**: Любой другой уникальный индекс без InitByDB полей
   - Выбирается первый найденный подходящий индекс

3. **Ошибка**: Если ни один индекс не подходит
   - Возвращается runtime ошибка с пояснением

### 3. Процесс выполнения

1. **Выбор индекса**: Определяется индекс для сопоставления (см. выше)
2. **Сбор данных**: Собираются все значения из объектов для bulk insert
3. **RETURNING клауза**: Формируется список полей для RETURNING:
   - Все поля с флагом `InitByDB` (управляемые БД)
   - Все поля выбранного индекса (для сопоставления)
4. **Выполнение запроса**: Выполняется `INSERT ... ON CONFLICT ... DO UPDATE RETURNING ...`
5. **Сопоставление**: Результаты из RETURNING сопоставляются с исходными объектами через выбранный индекс
6. **Обновление объектов**: Автоинкрементные поля обновляются в объектах

## Пример использования

### Декларация модели

```go
//ar:serverConf:postgres_conf
//ar:table:users
//ar:backend:postgres
type FieldsUser struct {
    Id       int64  `ar:"primary_key"`           // InitByDB: auto-increment
    Email    string `ar:"size:256;unique"`       // Уникальный индекс без InitByDB
    Name     string `ar:"size:128"`
    Created  int64  `ar:""`
}
```

В этом примере:

- `Id` - автоинкрементное поле (InitByDB)
- `Email` - уникальный индекс без InitByDB (используется для сопоставления)

### Использование в коде

```go
package main

import (
    "context"
    "yourapp/model/repository/generated/user"
)

func bulkCreateUsers(ctx context.Context, emails []string) error {
    // Создаем объекты
    users := make([]*user.User, len(emails))
    for i, email := range emails {
        u := user.New(ctx)
        u.SetEmail(email)
        u.SetName("User " + email)
        users[i] = u
    }

    // Выполняем bulk insert с конфликтом по первичному ключу
    err := user.BulkInsertReplace(ctx, users, user.OnConflictId)
    if err != nil {
        return err
    }

    // После вызова все объекты в users получат свои Id из БД
    for _, u := range users {
        fmt.Printf("User %s got ID: %d\n", u.GetEmail(), u.GetId())
    }

    return nil
}
```

## SQL запрос

Генерируемый SQL выглядит примерно так:

```sql
INSERT INTO users (id, email, name, created)
VALUES
    (DEFAULT, 'user1@example.com', 'User 1', 1234567890),
    (DEFAULT, 'user2@example.com', 'User 2', 1234567891),
    (DEFAULT, 'user3@example.com', 'User 3', 1234567892)
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    name = EXCLUDED.name,
    created = EXCLUDED.created
RETURNING id, email;
```

Где:

- `id` - возвращается как InitByDB поле
- `email` - возвращается для сопоставления с исходными объектами

## Ограничения

1. **Требуется уникальный индекс без InitByDB полей**
   - Если такого индекса нет, функция вернет ошибку при вызове:

     ```go
     err := user.BulkInsertReplace(ctx, users, OnConflictId)
     // Вернет: "BulkInsertReplace is not available for User: no unique index
     // without InitByDB fields found. Please add a unique index on
     // non-auto-generated fields to enable this functionality"
     ```

   - Функция всегда присутствует в API, но проверка происходит в runtime

2. **Конфликт только по уникальным индексам**
   - Параметр `conflictIndex` должен быть уникальным индексом

3. **Не поддерживается IgnoreDuplicate**
   - Bulk операции не поддерживают `ON CONFLICT DO NOTHING`

## Производительность

### Преимущества

- Один SQL запрос вместо N запросов
- Автоматическое сопоставление результатов
- Поддержка транзакций

### Метрики

Функция собирает следующие метрики:

- `bulk_insertorreplace_request` - количество запросов
- `bulk_insert_gen` - ошибки генерации запроса
- `bulk_insert_preparedb` - ошибки подготовки соединения
- `bulk_insert_db` - ошибки выполнения запроса
- `bulk_insert_scan` - ошибки чтения результатов
- `bulk_insert_match` - ошибки сопоставления
- `bulk_insertorreplace_success` - количество успешно вставленных строк

## Альтернативы

Если уникального индекса без InitByDB полей нет, используйте:

1. **Цикл с Insert/InsertOrReplace**

   ```go
   for _, obj := range objs {
       err := obj.InsertOrReplace(ctx, OnConflictId)
       if err != nil {
           return err
       }
   }
   ```

2. **Создание подходящего индекса**
   - Добавьте уникальный индекс без автоинкрементных полей
   - Перегенерируйте код с помощью `argen`

## Технические детали

### Сопоставление через map

Для быстрого поиска используется map с составным ключом:

```go
key := fmt.Sprintf("%v|%v|%v", field1, field2, field3)
objMap[key] = obj
```

### Порядок RETURNING

Поля в RETURNING следуют строго в порядке:

1. Все InitByDB поля (в порядке объявления в модели)
2. Все поля уникального индекса (в порядке полей индекса)

Это обеспечивает корректное чтение через `rows.Scan()`.
