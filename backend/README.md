# Auth Gateway

Централизованная система аутентификации и авторизации для экосистемы микросервисов.

## Возможности

- ✅ Традиционная аутентификация (email/пароль)
- ✅ JWT токены (access + refresh)
- ✅ Rate limiting с Redis
- ✅ Аудит логирование
- ✅ PostgreSQL для хранения данных
- ✅ Redis для кеширования
- ✅ Graceful shutdown
- ✅ Health checks
- ✅ CORS поддержка
- ✅ Structured logging
- ✅ gRPC API для микросервисов
- ✅ Постоянные API ключи для внешних сервисов
- ⏳ OAuth интеграция (Google, Yandex, GitHub, Instagram) - в разработке
- ⏳ Prometheus метрики - в разработке

## Технологический стек

- **Язык:** Go 1.23+
- **Web Framework:** Gin
- **RPC:** gRPC
- **Database:** PostgreSQL 14+
- **Cache:** Redis 7+
- **JWT:** golang-jwt/jwt
- **Password Hashing:** bcrypt
- **Containerization:** Docker & Docker Compose

## Быстрый старт

### Требования

- Go 1.23+
- Docker & Docker Compose
- Make (опционально)

### 1. Клонировать репозиторий

```bash
git clone https://github.com/smilemakc/auth-gateway.git
cd auth-gateway
```

### 2. Настроить environment переменные

```bash
cp .env.example .env
# Отредактируйте .env файл, установите JWT секреты
```

### 3. Запустить с Docker Compose

```bash
# Запустить все сервисы
docker-compose up -d

# Проверить логи
docker-compose logs -f auth-gateway
```

Или используя Makefile:

```bash
make docker-up
```

### 4. Проверить health

```bash
curl http://localhost:3000/auth/health
```

## API Endpoints

### Public Endpoints

| Method | Endpoint        | Description                     |
|--------|-----------------|---------------------------------|
| POST   | `/auth/signup`  | Регистрация нового пользователя |
| POST   | `/auth/signin`  | Вход в систему                  |
| POST   | `/auth/refresh` | Обновление access token         |
| GET    | `/auth/health`  | Health check                    |

### Protected Endpoints (требуют токен)

| Method | Endpoint                | Description      |
|--------|-------------------------|------------------|
| POST   | `/auth/logout`          | Выход из системы |
| GET    | `/auth/profile`         | Получить профиль |
| PUT    | `/auth/profile`         | Обновить профиль |
| POST   | `/auth/change-password` | Изменить пароль  |

### API Keys Management (требуют JWT токен)

| Method | Endpoint               | Description                      |
|--------|------------------------|----------------------------------|
| POST   | `/api-keys`            | Создать новый API ключ           |
| GET    | `/api-keys`            | Получить список своих API ключей |
| GET    | `/api-keys/:id`        | Получить информацию о ключе      |
| PUT    | `/api-keys/:id`        | Обновить API ключ                |
| POST   | `/api-keys/:id/revoke` | Отозвать API ключ                |
| DELETE | `/api-keys/:id`        | Удалить API ключ                 |

## Примеры использования

### Регистрация

```bash
curl -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "password": "securePassword123",
    "full_name": "John Doe"
  }'
```

### Вход

```bash
curl -X POST http://localhost:3000/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123"
  }'
```

### Получение профиля

```bash
curl -X GET http://localhost:3000/auth/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Выход

```bash
curl -X POST http://localhost:3000/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## API Ключи для внешних сервисов

API ключи предназначены для постоянной аутентификации внешних сервисов без необходимости обновления токенов. Ключи
поддерживают систему scopes для контроля доступа.

### Создание API ключа

```bash
curl -X POST http://localhost:3000/api-keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production API Key",
    "description": "Key for production service integration",
    "scopes": ["token:validate", "users:read"]
  }'
```

**Важно:** API ключ возвращается только один раз при создании. Сохраните его в безопасном месте!

Ответ:

```json
{
  "api_key": {
    "id": "uuid",
    "name": "Production API Key",
    "key_prefix": "agw_AbC1",
    "scopes": ["token:validate", "users:read"],
    "is_active": true,
    "created_at": 1234567890
  },
  "plain_key": "agw_AbC123XyZ...полный_ключ..."
}
```

### Использование API ключа

API ключи можно использовать двумя способами:

**1. В заголовке X-API-Key:**

```bash
curl -X GET http://localhost:3000/auth/profile \
  -H "X-API-Key: agw_YOUR_API_KEY"
```

**2. В заголовке Authorization:**

```bash
curl -X GET http://localhost:3000/auth/profile \
  -H "Authorization: Bearer agw_YOUR_API_KEY"
```

### Доступные scopes

**Базовые scopes:**
- `users:read` - чтение информации о пользователях
- `users:write` - изменение информации о пользователях
- `profile:read` - чтение профиля
- `profile:write` - изменение профиля
- `token:validate` - валидация токенов
- `token:introspect` - детальная информация о токенах
- `admin:all` - все административные права
- `all` - полный доступ ко всем операциям

**gRPC scopes:**
- `auth:login` - аутентификация через gRPC (Login)
- `auth:register` - регистрация через gRPC (CreateUser, RegisterWithOTP и др.)
- `auth:otp` - OTP операции через gRPC (SendOTP, VerifyOTP и др.)
- `email:send` - отправка email через gRPC
- `oauth:read` - чтение OAuth данных через gRPC
- `exchange:manage` - управление обменом токенов между приложениями
- `sync:users` - синхронизация пользователей между сервисами

### Управление API ключами

**Получить список ключей:**

```bash
curl -X GET http://localhost:3000/api-keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Отозвать ключ:**

```bash
curl -X POST http://localhost:3000/api-keys/{key_id}/revoke \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Удалить ключ:**

```bash
curl -X DELETE http://localhost:3000/api-keys/{key_id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## gRPC API для микросервисов

Auth Gateway предоставляет gRPC API для проверки токенов и авторизации между микросервисами.

**Важно:** Все gRPC методы требуют аутентификацию (кроме gRPC Reflection).

### Безопасность gRPC

- **Аутентификация:** Каждый запрос должен содержать учетные данные в metadata:
  - `x-api-key: agw_...` или `Authorization: Bearer agw_...` - API ключ (требует scopes)
  - `x-api-key: app_...` или `Authorization: Bearer app_...` - Application Secret (полный доступ)
- **Авторизация:**
  - **API ключи (`agw_`)**: Каждый метод требует определенный scope
  - **Application Secrets (`app_`)**: Полный доступ ко всем методам без проверки scopes
- **Application Context:** При использовании app secret, `application_id` автоматически извлекается из контекста — не нужно передавать в теле запроса
- **TLS:** Поддерживается TLS шифрование для production-окружений
- **Deny-by-default:** Методы не настроенные в системе scopes автоматически отклоняются
- **Reflection:** gRPC Reflection работает без аутентификации (для отладки, grpcurl)

### Переменные окружения gRPC

| Переменная | Описание | По умолчанию |
|-----------|----------|-------------|
| `GRPC_PORT` | Порт gRPC сервера | `50051` |
| `GRPC_TLS_ENABLED` | Включить TLS | `false` |
| `GRPC_TLS_CERT_FILE` | Путь к TLS сертификату | — |
| `GRPC_TLS_KEY_FILE` | Путь к TLS приватному ключу | — |

### gRPC Endpoints и Scopes

| Метод | Scope | Описание |
|-------|-------|----------|
| `ValidateToken` | `token:validate` | Проверка JWT токена или API ключа |
| `IntrospectToken` | `token:introspect` | Детальная информация о токене |
| `GetUser` | `users:read` | Получение пользователя по ID |
| `CheckPermission` | `users:read` | Проверка прав доступа (RBAC) |
| `GetApplicationAuthConfig` | `users:read` | Конфигурация аутентификации приложения |
| `GetUserApplicationProfile` | `profile:read` | Профиль пользователя в приложении |
| `GetUserTelegramBots` | `profile:read` | Telegram-боты пользователя |
| `Login` | `auth:login` | Аутентификация по email/паролю |
| `CreateUser` | `auth:register` | Создание пользователя |
| `RegisterWithOTP` | `auth:register` | Регистрация через OTP |
| `VerifyRegistrationOTP` | `auth:register` | Подтверждение регистрации OTP |
| `InitPasswordlessRegistration` | `auth:register` | Начало passwordless регистрации |
| `CompletePasswordlessRegistration` | `auth:register` | Завершение passwordless регистрации |
| `SendOTP` | `auth:otp` | Отправка OTP |
| `VerifyOTP` | `auth:otp` | Проверка OTP |
| `LoginWithOTP` | `auth:otp` | Вход через OTP |
| `VerifyLoginOTP` | `auth:otp` | Подтверждение входа через OTP |
| `SendEmail` | `email:send` | Отправка email |
| `IntrospectOAuthToken` | `oauth:read` | Интроспекция OAuth токена |
| `ValidateOAuthClient` | `oauth:read` | Валидация OAuth клиента |
| `GetOAuthClient` | `oauth:read` | Информация об OAuth клиенте |
| `CreateTokenExchange` | `exchange:manage` | Создание кода обмена токенов |
| `RedeemTokenExchange` | `exchange:manage` | Обмен кода на токены |
| `SyncUsers` | `sync:users` | Синхронизация пользователей |

### Адрес gRPC сервера

- **Локально:** `localhost:50051`
- **Docker:** `auth-gateway:50051`

### Пример использования gRPC

**С API ключом (требует scopes):**

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/smilemakc/auth-gateway/pkg/grpcclient"
)

func main() {
    // Подключение с API ключом
    client, err := grpcclient.NewClient(
        "localhost:50051",
        grpcclient.WithAPIKey("agw_YOUR_API_KEY"),
        grpcclient.WithTimeout(10*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Проверка токена
    resp, err := client.ValidateToken(
        context.Background(),
        "your-jwt-token",
    )
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Token valid: %t, User: %s", resp.Valid, resp.UserId)
}
```

**С Application Secret (полный доступ, application_id в контексте):**

```go
// Подключение с app secret
client, err := grpcclient.NewClient(
    "localhost:50051",
    grpcclient.WithAPIKey("app_YOUR_APPLICATION_SECRET"),
    grpcclient.WithTimeout(10*time.Second),
)
// application_id извлекается автоматически из app secret
// Не нужно передавать в теле запроса
```

**Примеры grpcurl:**

```bash
# С API ключом
grpcurl -H 'x-api-key: agw_YOUR_API_KEY' \
  -d '{"token": "your-jwt-token"}' \
  localhost:50051 auth.AuthService/ValidateToken

# С Application Secret
grpcurl -H 'Authorization: Bearer app_YOUR_APP_SECRET' \
  -d '{"token": "your-jwt-token"}' \
  localhost:50051 auth.AuthService/ValidateToken

# Reflection (без аутентификации)
grpcurl localhost:50051 list
grpcurl localhost:50051 describe auth.AuthService
```

### Интеграция с другими сервисами

Полные примеры интеграции и middleware для gRPC находятся в:

- `examples/grpc-client/` - пример клиента
- `proto/auth.proto` - proto определения

**Подробная документация:** [examples/grpc-client/README.md](examples/grpc-client/README.md)

## Разработка

### Локальный запуск (без Docker)

```bash
# 1. Запустить только PostgreSQL и Redis
make dev

# 2. Применить миграции (требуется migrate tool)
make migrate-up

# 3. Запустить приложение
make run
```

### Makefile команды

```bash
make help           # Показать все доступные команды
make build          # Собрать приложение
make run            # Запустить приложение
make test           # Запустить тесты
make test-coverage  # Запустить тесты с coverage
make clean          # Очистить build артефакты
make docker-up      # Запустить Docker Compose
make docker-down    # Остановить Docker Compose
make docker-logs    # Показать логи
make migrate-up     # Применить миграции
make migrate-down   # Откатить миграции
```

## Структура проекта

```
auth-gateway/
├── cmd/
│   └── server/          # Точка входа приложения
├── internal/
│   ├── config/          # Конфигурация
│   ├── models/          # Модели данных
│   ├── repository/      # Database слой
│   ├── service/         # Бизнес-логика
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # Middleware
│   └── utils/           # Утилиты
├── pkg/
│   ├── jwt/             # JWT сервис
│   └── logger/          # Логгер
├── migrations/          # SQL миграции
├── docker/              # Docker конфигурация
├── docs/                # Документация
├── tests/               # Тесты
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## База данных

### Таблицы

- `users` - пользователи
- `refresh_tokens` - refresh токены
- `api_keys` - постоянные API ключи для внешних сервисов
- `oauth_accounts` - OAuth аккаунты
- `token_blacklist` - черный список токенов
- `audit_logs` - логи аудита

### Миграции

Миграции находятся в папке `migrations/`. Для применения миграций:

```bash
# Требуется установить migrate tool:
# https://github.com/golang-migrate/migrate

make migrate-up    # Применить
make migrate-down  # Откатить
```

## Безопасность

### JWT Токены

- **Access Token:** TTL 15 минут
- **Refresh Token:** TTL 7 дней
- Алгоритм: HMAC-SHA256
- Токены хранятся в базе данных (refresh) и Redis (blacklist)

### API Ключи

- Формат: `agw_<base64_random_32_bytes>`
- Хешируются с помощью SHA-256 перед сохранением
- Полный ключ возвращается только один раз при создании
- Поддержка опциональной даты истечения (NULL = никогда не истекает)
- Система scopes для гранулярного контроля доступа
- Автоматическое отслеживание последнего использования
- Можно отозвать или удалить в любой момент

### Пароли

- Хешируются с помощью bcrypt (cost: 10)
- Минимальная длина: 8 символов
- Никогда не возвращаются в API

### Rate Limiting

- Регистрация: max 5 за час с одного IP
- Вход: max 10 за 15 минут с одного IP
- API: max 100 запросов в минуту

### CORS

- Настраивается через environment переменные
- Белый список origin'ов
- Поддержка credentials

## Production Deployment

### Важно перед деплоем:

1. **Изменить JWT секреты** в `.env`:
   ```
   JWT_ACCESS_SECRET=your-production-secret-256-bit
   JWT_REFRESH_SECRET=your-production-refresh-secret-256-bit
   ```

2. **Настроить CORS**:
   ```
   CORS_ALLOWED_ORIGINS=https://yourdomain.com
   ```

3. **Настроить database credentials**

4. **Включить HTTPS** (используйте reverse proxy как Nginx)

5. **Настроить мониторинг** (Prometheus, Grafana)

6. **Настроить gRPC TLS** (для production):
   ```
   GRPC_TLS_ENABLED=true
   GRPC_TLS_CERT_FILE=/path/to/cert.pem
   GRPC_TLS_KEY_FILE=/path/to/key.pem
   ```

## Тестирование

```bash
# Запустить все тесты
make test

# Тесты с coverage
make test-coverage
```

## Мониторинг

### Health Checks

- `/auth/health` - полная проверка (DB + Redis)
- `/auth/ready` - готовность к работе
- `/auth/live` - liveness probe

## Лицензия

MIT

## Контакты

- **Автор:** Auth Gateway Team
- **Email:** support@example.com
- **GitHub:** https://github.com/smilemakc/auth-gateway

## TODO

- [ ] OAuth интеграция (Google, Yandex, GitHub, Instagram)
- [ ] Prometheus метрики
- [ ] Email верификация
- [ ] Восстановление пароля
- [ ] 2FA (Two-Factor Authentication)
- [ ] API документация (Swagger)
- [ ] Kubernetes manifests
- [ ] CI/CD pipeline
- [ ] Больше unit тестов
- [ ] Integration тесты
- [ ] E2E тесты
