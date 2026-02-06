# Установка

## Установка argen

### Из исходников (рекомендуется)

```bash
# Клонируем репозиторий
git clone https://github.com/Educentr/go-activerecord
cd go-activerecord

# Устанавливаем в $GOPATH/bin
make install
```

### Сборка без установки

```bash
# Сборка в bin/argen
make build

# Проверка версии
./bin/argen --version
```

## Добавление зависимости в проект

```bash
go get github.com/Educentr/go-activerecord/v3
```

## Проверка установки

```bash
# Проверка версии argen
argen --version

# Вывод справки
argen --help
```

Ожидаемый вывод:

```
Usage of argen:
  --path string          путь к модели
  --declaration string   папка с декларациями (default "declaration")
  --destination string   папка для сгенерированного кода (default "generated")
  --module string        имя модуля (default из go.mod)
  --fixture_path string  путь для тестовых фикстур
```

## Структура проекта

Рекомендуемая структура для нового проекта:

```
myapp/
├── go.mod
├── main.go
└── model/
    └── repository/
        ├── declaration/     # Декларации моделей
        │   ├── user.go
        │   └── product.go
        └── generated/       # Сгенерированный код (создаётся argen)
            ├── user/
            └── product/
```

!!! warning "Важно"
    Файлы в папке `generated/` автоматически удаляются при перегенерации. Не редактируйте их вручную!

## База данных

### PostgreSQL

```bash
# Ubuntu/Debian
apt-get install postgresql

# macOS
brew install postgresql
```

### Octopus/Tarantool 1.5

```bash
# Ubuntu/Debian
apt-get install tarantool-lts
```

Пакет: [packages.debian.org/tarantool-lts](https://packages.debian.org/ru/buster/tarantool-lts)

## Следующие шаги

- [Быстрый старт](quickstart.md) — создайте первую модель
- [Архитектура](../architecture/index.md) — понимание работы генератора
