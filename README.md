# GophKeeper

[![Tests & Coverage](https://github.com/PolyakovEvg/gophkeeper/actions/workflows/coverage.yml/badge.svg)](https://github.com/PolyakovEvg/gophkeeper/actions/workflows/coverage.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/newmersedez/gophkeeper/develop/.github/badges/coverage.json)](https://github.com/PolyakovEvg/gophkeeper/actions/workflows/coverage.yml)
[![golangci-lint](https://github.com/PolyakovEvg/gophkeeper/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/PolyakovEvg/gophkeeper/actions/workflows/golangci-lint.yml)

Клиент-серверный менеджер паролей. Финальный проект курса "Продвинутый Go-разработчик" (Яндекс Практикум).

## Возможности

- Регистрация и аутентификация (JWT)
- Хранение паролей, текста, файлов, банковских карт и OTP-кодов
- Клиентское шифрование (AES-256-GCM) - сервер видит только зашифрованные данные
- Синхронизация между устройствами
- CLI и TUI интерфейсы
- OpenAPI документация: `api/swagger.yaml`

## Структура проекта

```
cmd/server          - HTTP-сервер
cmd/client          - CLI-клиент
internal/models      - модели данных
internal/crypto      - шифрование
internal/auth        - bcrypt + JWT
internal/otp         - TOTP генерация
internal/protocol    - бинарный протокол
internal/server/*    - хранилище, хендлеры, middleware, конфиг
internal/client/*    - API-клиент, локальное хранилище, CLI, TUI
api/swagger.yaml     - REST API спецификация
```

Сервер использует PostgreSQL. Клиент хранит данные локально в `~/.gophkeeper/vault.db`.

## Быстрый старт

### 1. Установка зависимостей

```bash
make deps
```

### 2. Настройка окружения

Скопируйте пример конфигурации:

```bash
cp .env.example .env
```

Отредактируйте `.env` под свою среду:

```bash
# .env
DATABASE_URI=postgres://localhost:5432/gophkeeper?sslmode=disable
TEST_DATABASE_URI=postgres://localhost:5432/gophkeeper_test?sslmode=disable
```

Создайте базы данных:

```bash
createdb gophkeeper
createdb gophkeeper_test
```

### 3. Запуск сервера

```bash
source .env
make run-server
```

Сервер запустится на `localhost:8080`.

### 4. Использование клиента

```bash
make build-client

# Регистрация
./bin/gophkeeper register -login alice -password secret

# Добавление записи
./bin/gophkeeper add text -title note -content "hello"

# Синхронизация с сервером
./bin/gophkeeper sync

# Список записей
./bin/gophkeeper list

# TUI интерфейс
./bin/gophkeeper tui
```

## Переменные окружения

### Сервер

| Переменная      | Описание              | По умолчанию      |
|-----------------|-----------------------|-------------------|
| RUN_ADDRESS     | Адрес сервера         | localhost:8080    |
| DATABASE_URI    | PostgreSQL DSN        | обязателен        |
| JWT_SECRET      | Секрет для JWT        | dev-значение      |

### Клиент

| Переменная        | Описание              |
|-------------------|-----------------------|
| GOPHKEEPER_SERVER | Адрес сервера         |

## Сборка

```bash
# Текущая платформа
make build-client

# Все платформы
make build-client-all
```

Результат в `bin/`: `gophkeeper-linux-amd64`, `gophkeeper-darwin-arm64`, `gophkeeper-windows-amd64.exe`

## Тесты

```bash
make test              # Запуск тестов
make coverage          # Покрытие кода
make coverage-html     # HTML отчет
make lint              # golangci-lint
```

## Безопасность

1. Пароли хешируются bcrypt на сервере
2. Данные шифруются на клиенте перед отправкой
3. Используйте HTTPS в production
4. Сессия хранится в `~/.gophkeeper/vault.db`