# GophKeeper

[![Tests & Coverage](https://github.com/PolyakovEvg/gophkeeper/actions/workflows/coverage.yml/badge.svg)](https://github.com/PolyakovEvg/gophkeeper/actions/workflows/coverage.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/newmersedez/gophkeeper/develop/.github/badges/coverage.json)](https://github.com/PolyakovEvg/gophkeeper/actions/workflows/coverage.yml)
[![golangci-lint](https://github.com/PolyakovEvg/gophkeeper/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/PolyakovEvg/gophkeeper/actions/workflows/golangci-lint.yml)

Клиент-серверный менеджер паролей. Финальный проект курса "Продвинутый Go-разработчик" (Яндекс Практикум).

## Возможности

- Регистрация и аутентификация (JWT)
- Хранение паролей, текста, файлов, банковских карт и OTP-кодов
- Клиентское шифрование (AES-256-GCM)
- Синхронизация между устройствами
- CLI и TUI интерфейсы

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
```

Сервер использует PostgreSQL. Клиент хранит данные локально в `~/.gophkeeper/vault.db`.

## Быстрый старт

### 1. Установка зависимостей

```bash
make deps
```

### 2. Настройка окружения

```bash
cp .env.example .env
# Отредактируйте .env под свою среду
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

./bin/gophkeeper register -login alice -password secret
./bin/gophkeeper add text -title note -content "hello"
./bin/gophkeeper sync
./bin/gophkeeper list
./bin/gophkeeper tui
```

## Переменные окружения

| Переменная      | Описание              | По умолчанию      |
|-----------------|-----------------------|-------------------|
| RUN_ADDRESS     | Адрес сервера         | localhost:8080    |
| DATABASE_URI    | PostgreSQL DSN        | обязателен        |
| JWT_SECRET      | Секрет для JWT        | обязателен в prod |

## Тесты

```bash
make test
make coverage
make lint
```

## Безопасность

- TLS обязателен в production - используйте reverse proxy (Nginx, Caddy)
- JWT_SECRET обязателен в production
- Пароли хешируются bcrypt (cost 12)
- Данные шифруются AES-256-GCM на клиенте
- PBKDF2-SHA256 с 600,000 итераций