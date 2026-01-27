# Рецепты

В этом документе представлены основные рецепты по использованию библиотеки.
Эта библиотека представляет из себя набор пакетов для подключения в приложение и утилиту для генерации.

## Декларативное описание

### Декларирование конфигурации хранилища

**Для Octopus:**
```go
//ar:serverConf:confKey
//ar:namespace:5
//ar:backend:octopus
type FieldsUser struct {
    Id   int32  `ar:"primary_key"`
    Name string `ar:"size:256"`
}
```

**Для PostgreSQL:**
```go
//ar:serverConf:pgConfKey
//ar:namespace:users
//ar:backend:postgres
type FieldsUser struct {
    Id   int32  `ar:"primary_key"`
    Name string `ar:"size:256"`
}
```

### Декларирование полей

**Основные теги для полей:**

```go
type FieldsProduct struct {
    // Первичный ключ
    Id int64 `ar:"primary_key"`

    // Простое поле с селектором
    Code string `ar:"size:64;selector:SelectByCode"`

    // Уникальный индекс
    Email string `ar:"unique;size:256;selector:SelectByEmail"`

    // Поле с сериализацией JSON
    Meta string `ar:"serializer:Json;size:1024"`

    // Поле с мутаторами для атомарных операций
    Counter uint32 `ar:"mutators:inc,dec"`
    Flags   uint32 `ar:""`  // мутаторы set_bit,clear_bit добавляются через FlagsProduct

    // Обычное поле
    CreatedAt uint32 `ar:""`
}

// Именованные флаги для поля Flags
type FlagsProduct struct {
    Flags string `ar:"flags:Active,Featured,OnSale,_,Archived"`
}
```

Структура `Flags*` генерирует константы для битовых флагов:
```go
const (
    FlagsActiveFlag   = 1 << 0  // 0x01
    FlagsFeaturedFlag = 1 << 1  // 0x02
    FlagsOnSaleFlag   = 1 << 2  // 0x04
    // позиция 3 пропущена
    FlagsArchivedFlag = 1 << 4  // 0x10
)
```

### Декларирование связанных сущностей

```go
type FieldsObjectUser struct {
    UserId int64 `ar:"key:Id;object:User;field:UserId"`
}
```

Позволяет автоматически подгружать связанные объекты через методы `GetUser()`.

### Декларирование индексов

**Простой индекс:**
```go
type IndexesProduct struct {
    // Уникальный составной индекс
    CodeType bool `ar:"fields:Code,Type;unique"`
}
```

**Индекс с условием (PostgreSQL):**
```go
type IndexesProduct struct {
    ActiveProducts bool `ar:"fields:Status,CreatedAt;condition:Status[=]1&&DeletedAt[is null]"`
}
```

**Индекс с сортировкой:**
```go
type IndexesProduct struct {
    RecentFirst bool `ar:"fields:Status,CreatedAt;order:Status asc,CreatedAt desc"`
}
```

#### Частичные индексы

Позволяют делать выборки по части составного индекса:

```go
type (
    IndexesProduct struct {
        StatusCreated bool `ar:"fields:Status,CreatedAt,Id"`
    }

    IndexPartsProduct struct {
        // Выборка только по Status (первое поле индекса)
        StatusPart bool `ar:"index:StatusCreated;fieldnum:1;selector:SelectByStatus"`

        // Выборка по Status и CreatedAt (первые 2 поля)
        StatusCreatedPart bool `ar:"index:StatusCreated;fieldnum:2;selector:SelectByStatusAndDate"`
    }
)
```

### Декларирование триггеров

Триггеры вызываются при исключительных ситуациях:

```go
type TriggersFoo struct {
    RepairTuple bool `ar:"pkg:github.com/myapp/model/repair;func:RepairTuple"`
}
```

Функция триггера для Octopus:
```go
package repair

import "github.com/Educentr/go-activerecord/v3/pkg/octopus"

func RepairTuple(tuple *octopus.TupleData) error {
    // Логика восстановления поврежденного tuple
    if len(tuple.Fields) < 5 {
        // Добавляем недостающие поля
        tuple.Fields = append(tuple.Fields, []byte("default_value"))
    }
    return nil
}
```

### Декларирование сериализаторов

**Встроенные сериализаторы:**

```go
type FieldsConfig struct {
    // JSON сериализатор
    Settings string `ar:"serializer:Json;size:2048"`

    // Printf формат
    Version string `ar:"serializer:Printf,%d.%d.%d;size:16"`

    // Mapstructure для сложных структур
    Params string `ar:"serializer:Mapstructure;size:4096"`
}

type SerializersConfig struct {
    Settings map[string]interface{} `ar:""`
    Params   *MyParamsStruct        `ar:"pkg:github.com/myapp/types;object:ParamsObj"`
}
```

**Пользовательский сериализатор:**

```go
type SerializersProduct struct {
    Tags []string `ar:"pkg:github.com/myapp/serializers;object:TagsSerializer"`
}
```

В пакете serializers:
```go
package serializers

func TagsSerializerMarshal(tags []string) (string, error) {
    return strings.Join(tags, ","), nil
}

func TagsSerializerUnmarshal(data string) ([]string, error) {
    if data == "" {
        return []string{}, nil
    }
    return strings.Split(data, ","), nil
}
```

## Конфигурирование

### Интерфейс конфига

Интерфейс конфига очень схож с реализацией `onlineconf`.

Но на самом деле можно реализовать любую структуру, которая ему удовлетворяет.

Например, внутри проекта в котором вы используете AR, можно создать структуру `ARConfig`:

```golang
type ARConfig struct {
    updatedIn time.Time
}

func NewARConfig() *ARConfig {
    arcfg := &ARConfig{
        updatedIn: time.Now(),
    }

    return arcfg
}

func (dc *ARConfig) GetLastUpdateTime() time.Time {
    return dc.updatedIn
}

func (dc *ARConfig) GetBool(ctx context.Context, confPath string, dfl ...bool) bool {
    if len(dfl) != 0 {
        return dfl[0]
    }

    return false
}
func (dc *ARConfig) GetBoolIfExists(ctx context.Context, confPath string) (value bool, ok bool) {
    return false, false
}

func (dc *ARConfig) GetDurationIfExists(ctx context.Context, confPath string) (time.Duration, bool) {
    switch confPath {
    case "arcfg/Timeout":
        return time.Millisecond * 200, true
    default:
        return 0, false
    }
}
func (dc *ARConfig) GetDuration(ctx context.Context, confPath string, dfl ...time.Duration) time.Duration {
    ret, ok := dc.GetDurationIfExists(ctx, confPath)
    if !ok && len(dfl) != 0 {
        ret = dfl[0]
    }

    return ret
}

func (dc *ARConfig) GetIntIfExists(ctx context.Context, confPath string) (int, bool) {
    switch confPath {
    case "arcfg/PoolSize":
        return 10, true
    default:
        return 0, false
    }
}
func (dc *ARConfig) GetInt(ctx context.Context, confPath string, dfl ...int) int {
    ret, ok := dc.GetIntIfExists(ctx, confPath)
    if !ok && len(dfl) != 0 {
        ret = dfl[0]
    }

    return ret
}

func (dc *ARConfig) GetStringIfExists(ctx context.Context, confPath string) (string, bool) {
    switch confPath {
    case "arcfg/master":
        return "127.0.0.1:11011", true
    case "arcfg/replica":
        return "127.0.0.1:11011", true
    default:
        return "", false
    }
}
func (dc *ARConfig) GetString(ctx context.Context, confPath string, dfl ...string) string {
    ret, ok := dc.GetStringIfExists(ctx, confPath)
    if !ok && len(dfl) != 0 {
        ret = dfl[0]
    }

    return ret
}

func (dc *ARConfig) GetStrings(ctx context.Context, confPath string, dfl []string) []string {
    return []string{}
}
func (dc *ARConfig) GetStruct(ctx context.Context, confPath string, valuePtr interface{}) (bool, error) {
    return false, nil
}
```

Это статический конфиг, и в таком виде он кажется избыточным, но можно передать при инициализации такого пакета конфиг приложения, который может
изменять свои параметры в течении времени. Тогда необходимо в методе `GetLastUpdateTime` отдавать время последнего обновления конфига.
Это позволит перечитывать параметры подключения на лету и пере-подключаться к базе.

## Работа с моделями в коде

### Создание новых записей

```go
package main

import (
    "context"
    "github.com/myapp/model/repository/generated/user"
)

func CreateUser(ctx context.Context, name, email string) error {
    // Создаем новый объект
    u := user.New(ctx)

    // Устанавливаем значения полей
    u.SetName(name)
    u.SetEmail(email)
    u.SetCreatedAt(uint32(time.Now().Unix()))

    // Сохраняем в БД
    if err := u.Insert(ctx); err != nil {
        return fmt.Errorf("failed to insert user: %w", err)
    }

    return nil
}
```

### Поиск записей

**По уникальному ключу:**
```go
// Возвращает один объект или ошибку
user, err := user.SelectById(ctx, 123)
if err != nil {
    return err
}

fmt.Println(user.GetName())
```

**По неуникальному индексу:**
```go
// Требуется лимитер для защиты от выборки всей БД
limiter := activerecord.NewLimiter(100)
users, err := user.SelectByStatus(ctx, "active", limiter)
if err != nil {
    return err
}

for _, u := range users {
    fmt.Printf("User: %s (%s)\n", u.GetName(), u.GetEmail())
}
```

**По составному индексу:**
```go
// Создаем ключ для поиска
key := user.EmailStatusIndexType{
    Email:  "test@example.com",
    Status: "active",
}

u, err := user.SelectByEmailStatus(ctx, key)
if err != nil {
    return err
}
```

**Множественная выборка по ключам:**
```go
ids := []int64{1, 2, 3, 4, 5}
users, err := user.SelectByIds(ctx, ids)
if err != nil {
    return err
}
```

### Обновление записей

**Простое обновление:**
```go
user, err := user.SelectById(ctx, 123)
if err != nil {
    return err
}

// Изменяем поля
user.SetName("New Name")
user.SetUpdatedAt(uint32(time.Now().Unix()))

// Обновляем в БД (обновятся только измененные поля)
if err := user.Update(ctx); err != nil {
    return err
}
```

**Полная перезапись:**
```go
// Replace обновляет ВСЕ поля, не только измененные
if err := user.Replace(ctx); err != nil {
    return err
}
```

**Insert or Replace:**
```go
u := user.New(ctx)
u.SetId(123)
u.SetName("Name")
u.SetEmail("email@example.com")

// Добавит новую запись или перезапишет существующую
if err := u.InsertOrReplace(ctx); err != nil {
    return err
}
```

### Удаление записей

```go
user, err := user.SelectById(ctx, 123)
if err != nil {
    return err
}

if err := user.Delete(ctx); err != nil {
    return err
}
```

## Атомарность на уровне БД

### Мутаторы

Мутаторы позволяют выполнять атомарные операции на уровне БД:

**Инкремент/декремент:**
```go
user, err := user.SelectById(ctx, 123)
if err != nil {
    return err
}

// Увеличиваем счетчик на 10 (атомарно в БД)
user.IncLoginCount(10)

// Применяем изменения
if err := user.Update(ctx); err != nil {
    return err
}

// После Update значение синхронизируется обратно в объект
fmt.Println("New count:", user.GetLoginCount())
```

**Битовые операции с именованными флагами:**

Декларируем именованные флаги:
```go
type FlagsUser struct {
    Flags string `ar:"flags:Active,Verified,Premium,_,Admin"`
}
```

Использование сгенерированных констант:
```go
u, err := user.SelectById(ctx, 123)
if err != nil {
    return err
}

// Устанавливаем биты флагов используя константы (атомарно)
u.SetBitFlags(user.FlagsPremiumFlag)   // Установить Premium
u.ClearBitFlags(user.FlagsActiveFlag)  // Очистить Active

if err := u.Update(ctx); err != nil {
    return err
}

// Проверка флагов
if u.GetFlags() & user.FlagsVerifiedFlag != 0 {
    fmt.Println("Пользователь верифицирован")
}
```

**Другие битовые операции:**
```go
// Побитовое ИЛИ (установить несколько флагов)
u.OrFlags(user.FlagsActiveFlag | user.FlagsVerifiedFlag)

// Побитовое И (оставить только указанные флаги)
u.AndFlags(user.FlagsActiveFlag | user.FlagsPremiumFlag)

// Побитовое XOR (переключить флаг)
u.XorFlags(user.FlagsAdminFlag)

// Все операции применяются при вызове Update
u.Update(ctx)
```

## Архитектурное построение

### Разделение на слои

Рекомендуемая структура проекта:

```
myapp/
├── cmd/
│   └── myapp/
│       └── main.go
├── internal/
│   ├── service/          # Бизнес-логика
│   │   └── user.go
│   └── handler/          # HTTP/gRPC обработчики
│       └── user.go
└── model/
    └── repository/
        ├── declaration/  # Декларативные описания (decl/)
        │   ├── user.go
        │   └── product.go
        └── generated/    # Сгенерированный код (cmpl/)
            ├── user/
            └── product/
```

### Инициализация подключений

```go
package main

import (
    "context"
    "github.com/Educentr/go-activerecord/v3/pkg/activerecord"
    "github.com/myapp/internal/config"
    "github.com/myapp/model/repository/generated/user"
)

func main() {
    ctx := context.Background()

    // Инициализируем конфиг
    cfg := config.New()

    // Регистрируем конфиг в ActiveRecord
    activerecord.RegisterConfig(cfg)

    // Опционально: регистрируем метрики
    metrics := NewPrometheusMetrics()
    activerecord.RegisterMetrics(metrics)

    // Опционально: регистрируем логгер
    logger := zerolog.New(os.Stdout)
    activerecord.RegisterLogger(&logger)

    // Теперь можно работать с моделями
    u, err := user.SelectById(ctx, 1)
    // ...
}
```

## Best practices

### 1. Всегда используйте лимитеры для неуникальных индексов

❌ **Плохо:**
```go
// Может вернуть миллионы записей!
users, _ := user.SelectByStatus(ctx, "active", nil)
```

✅ **Хорошо:**
```go
limiter := activerecord.NewLimiter(1000)
users, err := user.SelectByStatus(ctx, "active", limiter)
if err != nil {
    return err
}
```

### 2. Проверяйте ошибки выборки

```go
user, err := user.SelectById(ctx, id)
if err != nil {
    if errors.Is(err, activerecord.ErrNotFound) {
        return fmt.Errorf("user not found")
    }
    return fmt.Errorf("database error: %w", err)
}
```

### 3. Используйте мутаторы для счетчиков

❌ **Плохо (race condition):**
```go
user, _ := user.SelectById(ctx, id)
user.SetCounter(user.GetCounter() + 1)
user.Update(ctx)
```

✅ **Хорошо (атомарно):**
```go
user, _ := user.SelectById(ctx, id)
user.IncCounter(1)
user.Update(ctx)
```

### 4. Не изменяйте первичные ключи

```go
user, _ := user.SelectById(ctx, 123)
// user.SetId(456) // ОШИБКА! Первичный ключ защищен от изменения
```

### 5. Используйте контекст для таймаутов

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := user.SelectById(ctx, id)
```

### 6. Порядок полей для Octopus критичен

Для Octopus порядок полей в `Fields*` должен **точно** совпадать с порядком в tuple!

### 7. Всегда запускайте `make generate` перед коммитом

```bash
make generate  # Генерирует моки и проверяет код
make lint      # Проверяет качество кода
make test      # Запускает тесты
```
