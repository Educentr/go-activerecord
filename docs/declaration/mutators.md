# Мутаторы и флаги

Мутаторы обеспечивают атомарные операции в БД. Структура `Flags*` добавляет именованные битовые константы.

## Встроенные мутаторы

### Объявление

```go
type FieldsUser struct {
    Id         int64  `ar:"primary_key"`
    Counter    uint32 `ar:"mutators:inc,dec"`
    Points     int64  `ar:"mutators:inc,dec"`
    Flags      uint32 `ar:"mutators:and,or,xor,set_bit,clear_bit"`
}
```

### Типы мутаторов

| Мутатор | Генерируемый метод | Операция в БД |
|---------|-------------------|---------------|
| `inc` | `Inc{Field}(delta)` | `field = field + delta` |
| `dec` | `Dec{Field}(delta)` | `field = field - delta` |
| `and` | `And{Field}(mask)` | `field = field & mask` |
| `or` | `Or{Field}(mask)` | `field = field \| mask` |
| `xor` | `Xor{Field}(mask)` | `field = field ^ mask` |
| `set_bit` | `SetBit{Field}(mask)` | `field = field \| mask` |
| `clear_bit` | `ClearBit{Field}(mask)` | `field = field & ~mask` |

### Использование

```go
user, _ := user.SelectById(ctx, 123)

// Атомарный инкремент (выполняется в БД)
user.IncCounter(10)
user.DecPoints(5)

// Битовые операции
user.SetBitFlags(0x04)    // Установить бит
user.ClearBitFlags(0x01)  // Очистить бит
user.OrFlags(0x10)        // ИЛИ
user.XorFlags(0x08)       // XOR (переключить)
user.AndFlags(0xFF)       // И (маска)

// Применение всех операций одним запросом
user.Update(ctx)
```

!!! tip "Атомарность"
    Все мутаторы выполняются атомарно в БД. Это исключает race conditions при конкурентном доступе.

## Именованные флаги (Flags*)

### Объявление

```go
type FieldsUser struct {
    Id    int64  `ar:"primary_key"`
    Flags uint32 `ar:""`  // Мутаторы добавятся автоматически
}

type FlagsUser struct {
    Flags string `ar:"flags:Active,Verified,Premium,_,Admin"`
}
```

### Генерируемые константы

```go
const (
    FlagsActiveFlag   = 1 << 0  // 0x01 (1)
    FlagsVerifiedFlag = 1 << 1  // 0x02 (2)
    FlagsPremiumFlag  = 1 << 2  // 0x04 (4)
    // позиция 3 пропущена (underscore)
    FlagsAdminFlag    = 1 << 4  // 0x10 (16)
)
```

### Автоматические мутаторы

При объявлении `Flags*` автоматически добавляются мутаторы `set_bit` и `clear_bit`:

```go
// Эти методы генерируются автоматически
user.SetBitFlags(mask uint32)
user.ClearBitFlags(mask uint32)
```

### Использование констант

```go
user := user.New(ctx)

// Установка флагов при создании
user.SetFlags(user.FlagsActiveFlag | user.FlagsVerifiedFlag)
user.Insert(ctx)

// Атомарная установка флага
user.SetBitFlags(user.FlagsPremiumFlag)
user.Update(ctx)

// Атомарная очистка флага
user.ClearBitFlags(user.FlagsActiveFlag)
user.Update(ctx)

// Проверка флага
if user.GetFlags() & user.FlagsVerifiedFlag != 0 {
    fmt.Println("Пользователь верифицирован")
}

// Проверка нескольких флагов
if user.GetFlags() & (user.FlagsPremiumFlag | user.FlagsAdminFlag) != 0 {
    fmt.Println("Premium или Admin")
}
```

### Синтаксис флагов

```
flags:Name1,Name2,Name3,_,Name5
```

- Имена разделяются запятыми
- `_` (underscore) пропускает позицию бита
- Позиция в списке = номер бита (начиная с 0)

## Пользовательские мутаторы (Mutators*)

Для сложной логики модификации используйте `Mutators*`:

### Объявление

```go
type FieldsUser struct {
    Profile string `ar:"serializer:Profile;mutators:ProfilePart"`
}

type SerializersUser struct {
    Profile *types.Profile `ar:"pkg:github.com/myapp/types;object:ProfileS"`
}

type MutatorsUser struct {
    ProfilePart bool `ar:"update:updateProfile;pkg:github.com/myapp/conv"`
}
```

### Теги Mutators*

| Тег | Описание |
|-----|----------|
| `update` | Имя функции/процедуры БД для UPDATE |
| `replace` | Имя функции/процедуры БД для REPLACE |
| `pkg` | Пакет с функцией преобразования |

### Функция преобразования

```go
package conv

import "github.com/myapp/types"

// Для сериализованного поля Profile
// Принимает: текущее значение и map с изменёнными полями
func UserProfilePart(profile *types.Profile, parts map[string]any) ([]string, error) {
    // Формируем параметры для вызова процедуры БД
    values := make([]string, 0)

    if name, ok := parts["Name"]; ok {
        values = append(values, fmt.Sprintf("%v", name))
    }
    if age, ok := parts["Age"]; ok {
        values = append(values, fmt.Sprintf("%v", age))
    }

    return values, nil
}
```

### Использование

```go
user, _ := user.SelectById(ctx, 123)

// Для каждого поля сериализуемой структуры генерируется сеттер
user.SetProfileName("New Name")
user.SetProfileAge(25)

// При Update вызывается процедура БД через функцию преобразования
user.Update(ctx)
```

## Порядок применения

При вызове `Update()` операции применяются в порядке:

1. Обычные SET операции
2. Мутаторы (inc, dec, and, or, xor, set_bit, clear_bit)
3. Пользовательские мутаторы (вызов процедур)

## Полный пример

```go
package repository

//ar:serverConf:mydb
//ar:namespace:users
//ar:backend:postgres
type FieldsUser struct {
    Id int64 `ar:"primary_key;init_by_db"`

    // Счётчики
    LoginCount   uint32 `ar:"mutators:inc"`
    FailedLogins uint32 `ar:"mutators:inc,dec"`
    Points       int64  `ar:"mutators:inc,dec"`

    // Битовые флаги
    Flags uint32 `ar:""`

    // Сериализованное поле с кастомным мутатором
    Profile string `ar:"serializer:Profile;mutators:ProfilePart"`
}

// Именованные флаги
type FlagsUser struct {
    Flags string `ar:"flags:Active,Verified,Premium,_,Admin,Banned"`
}

// Сериализатор
type SerializersUser struct {
    Profile *types.UserProfile `ar:"pkg:github.com/myapp/types;object:ProfileS"`
}

// Кастомный мутатор
type MutatorsUser struct {
    ProfilePart bool `ar:"update:updateUserProfile;pkg:github.com/myapp/conv"`
}
```

```go
// Использование
user, _ := user.SelectById(ctx, 123)

// Инкремент счётчика
user.IncLoginCount(1)

// Установка флага Premium
user.SetBitFlags(user.FlagsPremiumFlag)

// Очистка флага Banned
user.ClearBitFlags(user.FlagsBannedFlag)

// Обновление части профиля
user.SetProfileName("John Doe")

// Все операции в одном запросе
user.Update(ctx)
```

## Следующие шаги

- [Триггеры](triggers.md) — обработка исключений
- [Атомарные операции](../usage/atomic-ops.md) — подробнее об атомарности
