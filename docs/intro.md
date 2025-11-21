# Вступление

Простой способ организовать модель в своём приложении:

- Скачайте и установите `argen` (git clone <https://github.com/Educentr/go-activerecord> && cd go-activerecord && make install)
- Добавьте зависимость в своём пакете `go get github.com/Educentr/go-activerecord`
- Создайте каталог `model/repository/decl`
- Создайте файлы декларации, например: `model/repository/decl/foo.go`
- Запустите генерацию `argen --path "model/repository/" --declaration "decl" --destination "cmpl"`
- Подключайте `import "..../model/repository/cmpl/foo"`
- Используйте `foo.SelectBy...()`
- Запускайте генерацию в любой момент, когда вам необходимо

Профит!

## Пример

Подсмотреть на пример можно в [activerecord-cookbook](https://github.com/mailru/activerecord-cookbook)

## Поддерживаемые базы данных

### Octopus / Tarantool 1.5

Используется для подключения к базам `octopus` и `tarantool` версии 1.5

**Особенности:**

- Использует бинарный протокол `iproto` ([описание протокола](https://github.com/Vespertinus/octopus/blob/master/doc/silverbox-protocol.txt))
- Спейсы идентифицируются числовым ID
- Порядок полей критичен и должен совпадать с порядком в tuple
- Поддержка хранимых процедур через ProcFields
- Дополнительные поля сохраняются в extraFields

**Установка Tarantool 1.5:**

```bash
# Debian/Ubuntu
apt-get install tarantool-lts
```

Пакет: <https://packages.debian.org/ru/buster/tarantool-lts>

### PostgreSQL

Полная поддержка PostgreSQL через драйвер `pgx/v5`

**Особенности:**

- Имя таблицы указывается через `//ar:namespace:table_name`
- Поддержка условных индексов с WHERE-клаузами
- Полная поддержка SQL с prepared statements
- Поддержка транзакций
- Встроенные сериализаторы для JSON полей
