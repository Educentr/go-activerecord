PWD = $(CURDIR)
# Название сервиса
SERVICE_NAME = argen
# 8 символов последнего коммита
LAST_COMMIT_HASH = $(shell git rev-parse HEAD | cut -c -8)
# Таймаут для тестов
TEST_TIMEOUT?=120s
# Тег golang-ci
GOLANGCI_TAG:=1.64.8
# Путь до бинарников
LOCAL_BIN:=$(CURDIR)/bin
# Путь до бинарника golang-ci
GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
# Тег mockery (генерация моков по директивам //go:generate). Не поднимать выше
# версии, у которой директива go в go.mod больше GO_MOD_VERSION ниже: go install
# откажется её собирать нужным репозиторию тулчейном (v2.53.6 -> go 1.24.0,
# v2.53.7 -> 1.25.0, v2.53.8 -> 1.26.0). Поднимать вместе с go.mod, не раньше.
MOCKERY_TAG:=2.53.6
# Путь до бинарника mockery
MOCKERY_BIN:=$(LOCAL_BIN)/mockery
# Минимальная верси гошки
MIN_GO_VERSION = 1.24.0
# Версия языка, на которую рассчитан репозиторий (директива go в go.mod)
GO_MOD_VERSION := $(shell awk '$$1 == "go" { print $$2; exit }' go.mod 2>/dev/null)
# Версии для сборки
RELEASE = $(shell git describe --tags --always)
# Время сборки
BUILD_DATE = $(shell TZ=UTC-3 date +%Y%m%d-%H%M)
# Операционка
OSNAME = $(shell uname)
# ld флаги
LD_FLAGS = "-X 'main.BuildCommit=$(LAST_COMMIT_HASH)' -X 'main.Version=$(RELEASE)' -X 'main.BuildTime=$(BUILD_DATE)' -X 'main.BuildOS=$(OSNAME)'"
# по дефолту просто make соберёт argen
default: build

# Добавляет флаг для тестирования на наличие гонок
ifdef GO_RACE_DETECTOR
    FLAGS += -race
endif

##################### Проверки для запуска golang-ci #####################
# GOLANGCI_TAG выше — единственный источник правды о версии линтера, а
# .golangci.yml написан в формате конфига этого мажора: бинарь другого мажора
# либо отвергает конфиг целиком (v2 про `output.formats`: «expected a map, got
# slice»), либо молча игнорирует большую часть настроек. Поэтому проверки ниже
# требуют *точного* совпадения версии, а иначе собирают пиннутую версию в
# ./bin (см. install-lint). Тот же паттерн, что в go-project-starter, — чтобы
# репозитории не расходились.
#
# Версия бинаря golangci-lint, пусто если определить не удалось. Поиск первого
# semver в выводе переваривает оба формата релизов:
#   v1: "golangci-lint has version 1.64.8 built from <commit> on <date>"
#   v2: "golangci-lint has version 2.5.0 built with go1.25.1 from (...) on (...)"
golangci_version = $(shell $(1) --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -n1)
# Go-тулчейн, которым собран бинарь ("<path>: go1.26.0" -> "1.26.0"), пусто если
# неизвестен. GOTOOLCHAIN=local не даёт пробе переключать тулчейн.
golangci_go_version = $(shell GOTOOLCHAIN=local go version -m $(1) 2>/dev/null | awk 'NR==1 { sub(/^go/, "", $$2); print $$2 }')
# Непусто, когда версия $(2) известна и меньше требуемой $(1); неизвестную $(2)
# оставляем в покое, чтобы сбойная проба не загоняла сборку в цикл.
version_older = $(if $(1),$(if $(2),$(filter-out $(1),$(word 1, $(sort $(1) $(2))))))
# Бинарь непригоден, если его версия не равна точно GOLANGCI_TAG либо он собран
# тулчейном старше той версии Go, на которую рассчитан репозиторий, — второе
# роняет линтер на тайпчеке ("file requires newer Go version goX.Y").
golangci_unusable = $(if $(filter $(GOLANGCI_TAG),$(call golangci_version,$(1))),,version)$(call version_older,$(GO_MOD_VERSION),$(call golangci_go_version,$(1)))

# Проверка локальной версии бинаря
ifneq ($(wildcard $(GOLANGCI_BIN)),)
ifneq ($(call golangci_unusable,$(GOLANGCI_BIN)),)
GOLANGCI_BIN:=
endif
endif

# Проверка глобальной версии бинаря
GOLANGCI_GLOBAL_BIN:=$(shell which golangci-lint)
ifneq ($(GOLANGCI_GLOBAL_BIN),)
ifeq ($(call golangci_unusable,$(GOLANGCI_GLOBAL_BIN)),)
GOLANGCI_BIN:=$(GOLANGCI_GLOBAL_BIN)
endif
endif
##################### Конец проверок golang-ci #####################

##################### Проверки для запуска mockery #####################
# Директивы //go:generate вызывают mockery без указания пути, поэтому моки
# собираются тем бинарём, который нашёлся в PATH, — и его версия попадает в
# шапку сгенерированного файла. Чтобы генерация была воспроизводимой, mockery
# пиннится в MOCKERY_TAG, ставится в ./bin и подкладывается в PATH целью
# generate. Версия должна оставаться в пределах v2: директивы используют
# CLI-флаги v2 (--name, --structname, --inpackage), в v3 их нет — там нужен
# .mockery.yml, и это отдельная работа.
#
# Версия бинаря mockery ("v2.53.8" -> "2.53.8"), пусто если не определилась.
mockery_version = $(shell $(1) --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -n1)
ifneq ($(wildcard $(MOCKERY_BIN)),)
MOCKERY_BIN_VERSION:=$(call mockery_version,$(MOCKERY_BIN))
endif
##################### Конец проверок mockery #####################

# Устанавливает линтер
.PHONY: install-lint
install-lint:
ifeq ($(wildcard $(GOLANGCI_BIN)),)
	$(info #Downloading golangci-lint v$(GOLANGCI_TAG))
	tmp=$$(mktemp -d) && cd $$tmp && pwd && go mod init temp && go get -d github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_TAG) && \
		go build -ldflags "-X 'main.version=$(GOLANGCI_TAG)' -X 'main.commit=test' -X 'main.date=test'" -o $(LOCAL_BIN)/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint && \
		rm -rf $$tmp
GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
endif

# Линтер проверяет лишь отличия от мастера
.PHONY: lint
lint: install-lint
	$(GOLANGCI_BIN) run --config=.golangci.yml ./... --new-from-rev=origin/main --build-tags=activerecord

# Проверяет .golangci.yml на соответствие схеме конфига пиннутой версии линтера.
# `golangci-lint run` игнорирует незнакомые ключи, поэтому конфиг, оставшийся от
# другого мажора, продолжает «работать», молча потеряв часть настроек, — заметит
# это только эта цель. Запускать при каждой смене GOLANGCI_TAG.
# В CI-гейт не входит: схема тянется с golangci-lint.run, и недоступность сайта
# не должна красить все pull request'ы.
.PHONY: lint-config-verify
lint-config-verify: install-lint
	$(GOLANGCI_BIN) config verify --config=.golangci.yml

# Линтер проходится по всему коду
.PHONY: full-lint
full-lint: install-lint
	$(GOLANGCI_BIN) run --config=.golangci.yml ./... --build-tags=activerecord

# Linter will check only diffs with main branch (default)
.PHONY: lint-fix
lint-fix: lint
	$(GOLANGCI_BIN) run --fix --config=.golangci.yml ./... --build-tags=activerecord

# создание отчета о покрытии тестами
.PHONY: cover
cover:
	go test -timeout=$(TEST_TIMEOUT) -v -coverprofile=coverage.out ./...  && go tool cover -html=coverage.out

# Генерация моков по директивам //go:generate. Зависит от install-tool, поэтому
# работает «из коробки»: нужный mockery доставится в ./bin сам. LOCAL_BIN
# впереди PATH — чтобы глобально установленный mockery другой версии не
# переписал моки своей шапкой.
.PHONY: generate
generate: install-tool
	PATH="$(LOCAL_BIN):$$PATH" go generate ./...

# Запустить unit тесты
.PHONY: test
test:
	echo "Start testing activerecord \n"
	go test -parallel=10 $(PWD)/... -coverprofile=cover.out -timeout=$(TEST_TIMEOUT)

.PHONY: install
install:
	go install -ldflags=$(LD_FLAGS) ./...

# Ставит пиннутый mockery в ./bin, если там его нет или версия другая.
# Именно ./bin, а не $GOPATH/bin: тот не обязан быть в PATH, из-за чего
# `make install-tool && make generate` падал с «executable file not found».
.PHONY: install-tool
install-tool:
ifneq ($(MOCKERY_TAG),$(MOCKERY_BIN_VERSION))
	$(info #Installing mockery v$(MOCKERY_TAG))
	GOBIN=$(LOCAL_BIN) go install github.com/vektra/mockery/v2@v$(MOCKERY_TAG)
endif

# Сборка сервиса
.PHONY: build
build:
	./scripts/goversioncheck.sh $(MIN_GO_VERSION) && go build -o bin/$(SERVICE_NAME) -ldflags=$(LD_FLAGS) $(PWD)/cmd/$(SERVICE_NAME)

# Устанавливает в локальный проект хук, который проверяет запускает линтеры
.PHONY: pre-commit-hook
pre-commit-hook:
	touch ./.git/hooks/pre-commit
	echo '#!/bin/sh' > ./.git/hooks/pre-commit
	echo 'make generate' >> ./.git/hooks/pre-commit
	echo 'make lint' >> ./.git/hooks/pre-commit
	chmod +x ./.git/hooks/pre-commit

# Устанавливает в локальный проект хук, который проверяет запускает линтеры
.PHONY: pre-push-hook
pre-push-hook:
	touch ./.git/hooks/pre-push
	echo '#!/bin/sh' > ./.git/hooks/pre-push
	echo 'make cover' >> ./.git/hooks/pre-push
	chmod +x ./.git/hooks/pre-push

# Запуск MkDocs сервера для документации
.PHONY: run-mkdoc
run-mkdoc: ## Run MkDocs dev server for documentation
	docker run --rm -p 8000:8000 -v $(PWD):/docs squidfunk/mkdocs-material serve --dev-addr=0.0.0.0:8000 --watch-theme

# Сборка документации
.PHONY: build-docs
build-docs: ## Build documentation
	docker run --rm -v $(PWD):/docs squidfunk/mkdocs-material build --strict

