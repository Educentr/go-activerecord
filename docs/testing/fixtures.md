# Фикстуры

Генерация и использование тестовых фикстур.

## Генерация

```bash
argen --path "model/repository" \
      --declaration "declaration" \
      --destination "generated" \
      --fixture_path "testutil/fixture"
```

## Структура

```
testutil/fixture/
├── fixture/
│   └── data/
│       ├── user.yaml              # Данные Select
│       ├── user_insert_replace.yaml    # Данные Insert/Replace
│       └── user_update.yaml       # Данные Update
└── user/
    └── fixture.go                 # Сгенерированный код
```

## YAML формат

### Данные для Select

`fixture/data/user.yaml`:

```yaml
- id: 1
  email: admin@example.com
  name: Admin User
  status: active
  flags: 3  # Active | Verified

- id: 2
  email: user@example.com
  name: Regular User
  status: active
  flags: 1  # Active

- id: 3
  email: inactive@example.com
  name: Inactive User
  status: inactive
  flags: 0
```

!!! note "snake_case"
    Имена полей в YAML используют snake_case.

### Данные для Insert/Replace

`fixture/data/user_insert_replace.yaml`:

```yaml
- code: new-user
  email: new@example.com
  name: New User
  status: pending
  flags: 0

- code: premium-user
  email: premium@example.com
  name: Premium User
  status: active
  flags: 7  # Active | Verified | Premium
```

### Данные для Update

`fixture/data/user_update.yaml`:

```yaml
- code: activate-user
  update_options:
    - status:
        set_value: active
    - flags:
        set_bit: 1  # Active flag

- code: deactivate-user
  update_options:
    - status:
        set_value: inactive
    - flags:
        clear_bit: 1  # Clear Active flag
```

## Использование

### Select фикстуры

```go
import "yourapp/testutil/fixture"

func TestSelectUser(t *testing.T) {
    ctx := context.Background()

    // Получение фикстуры по первичному ключу
    fix, used := fixture.GetUserFixtureById(ctx, 1)

    // Проверка данных
    assert.Equal(t, "admin@example.com", fix.GetEmail())
    assert.Equal(t, "Admin User", fix.GetName())

    // Проверка использования
    assert.True(t, used(), "fixture should be used")
}
```

### Insert фикстуры

```go
func TestInsertUser(t *testing.T) {
    ctx := context.Background()

    triggerFunc := func(fixtures []octopus.FixtureType) []octopus.FixtureType {
        // Модификация фикстур перед использованием
        return fixtures
    }

    // Получение фикстуры для Insert
    fix, used := fixture.GetInsertUserFixtureByCode(ctx, "new-user", triggerFunc)

    // Использование в тесте
    // ...

    assert.True(t, used(), "fixture must be applied")
}
```

### Update фикстуры

```go
func TestUpdateUser(t *testing.T) {
    ctx := context.Background()

    triggerFunc := func(fixtures []octopus.FixtureType) []octopus.FixtureType {
        return fixtures
    }

    // Из YAML
    fix, used := fixture.GetUpdateUserFixtureByCode(ctx, "activate-user", triggerFunc)

    // Или ручное создание
    fix, used := fixture.UpdateUserFixture("user-code").
        WithUpdatedStatus("active").
        WithUpdatedFlags(1).
        OnUpdate(triggerFunc).
        Build(ctx)

    assert.True(t, used())
}
```

### Hand-made фикстуры

```go
func TestCustomFixture(t *testing.T) {
    ctx := context.Background()

    // Создание модели вручную
    model := user.New(ctx)
    model.SetEmail("test@example.com")
    model.SetName("Test User")
    model.SetStatus("active")

    triggerFunc := func(fixtures []octopus.FixtureType) []octopus.FixtureType {
        return fixtures
    }

    // Получение фикстуры из модели
    fix, used := fixture.GetInsertUserFixtureByModel(ctx, model, triggerFunc)

    // ...

    assert.True(t, used())
}
```

## Trigger функции

Trigger позволяет модифицировать фикстуры перед использованием:

```go
triggerFunc := func(fixtures []octopus.FixtureType) []octopus.FixtureType {
    for i := range fixtures {
        // Модификация каждой фикстуры
        fixtures[i].Data["timestamp"] = time.Now().Unix()
    }
    return fixtures
}
```

## Проверка использования

Функция `used()` возвращает `true`, если фикстура была использована:

```go
fix, used := fixture.GetUserFixtureById(ctx, 1)

// Некоторый код...

// В конце теста проверяем
if !used() {
    t.Error("fixture was not used")
}
```

## Генерируемые методы

Для каждой модели генерируются:

### Select фикстуры

```go
func GetUserFixtureById(ctx context.Context, id int64) (*user.User, func() bool)
func GetUserFixtureByEmail(ctx context.Context, email string) (*user.User, func() bool)
```

### Insert фикстуры

```go
func GetInsertUserFixtureByCode(ctx context.Context, code string, trigger TriggerFunc) (*user.User, func() bool)
func GetInsertUserFixtureByModel(ctx context.Context, model *user.User, trigger TriggerFunc) (*user.User, func() bool)
```

### Replace фикстуры

```go
func GetReplaceUserFixtureByCode(ctx context.Context, code string, trigger TriggerFunc) (*user.User, func() bool)
func GetReplaceUserFixtureByModel(ctx context.Context, model *user.User, trigger TriggerFunc) (*user.User, func() bool)
```

### InsertOrReplace фикстуры

```go
func GetInsertOrReplaceUserFixtureByCode(ctx context.Context, code string, trigger TriggerFunc) (*user.User, func() bool)
func GetInsertOrReplaceUserFixtureByModel(ctx context.Context, model *user.User, trigger TriggerFunc) (*user.User, func() bool)
```

### Update фикстуры

```go
func GetUpdateUserFixtureByCode(ctx context.Context, code string, trigger TriggerFunc) (*user.User, func() bool)
func UpdateUserFixture(code string) *UserFixtureBuilder
```

## Builder для Update

```go
fix, used := fixture.UpdateUserFixture("user-code").
    WithUpdatedName("New Name").
    WithUpdatedStatus("active").
    WithUpdatedFlags(user.FlagsActiveFlag | user.FlagsVerifiedFlag).
    OnUpdate(triggerFunc).
    Build(ctx)
```

## Best Practices

### 1. Используйте понятные коды

```yaml
# ✅ Хорошо
- code: admin-user
- code: premium-expired

# ❌ Плохо
- code: user1
- code: test
```

### 2. Проверяйте использование

```go
fix, used := fixture.GetUserFixtureById(ctx, 1)
defer func() {
    if !used() {
        t.Error("fixture not used")
    }
}()
```

### 3. Изолируйте данные

```go
func TestSomething(t *testing.T) {
    // Используйте уникальные идентификаторы
    code := fmt.Sprintf("test-%s", t.Name())
    // ...
}
```

## Следующие шаги

- [CRUD операции](../usage/crud.md) — основные операции
- [Конфигурация](../usage/configuration.md) — настройка тестового окружения
