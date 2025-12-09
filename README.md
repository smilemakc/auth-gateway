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

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/signup` | Регистрация нового пользователя |
| POST | `/auth/signin` | Вход в систему |
| POST | `/auth/refresh` | Обновление access token |
| GET | `/auth/health` | Health check |

### Protected Endpoints (требуют токен)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/logout` | Выход из системы |
| GET | `/auth/profile` | Получить профиль |
| PUT | `/auth/profile` | Обновить профиль |
| POST | `/auth/change-password` | Изменить пароль |

### API Keys Management (требуют JWT токен)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api-keys` | Создать новый API ключ |
| GET | `/api-keys` | Получить список своих API ключей |
| GET | `/api-keys/:id` | Получить информацию о ключе |
| PUT | `/api-keys/:id` | Обновить API ключ |
| POST | `/api-keys/:id/revoke` | Отозвать API ключ |
| DELETE | `/api-keys/:id` | Удалить API ключ |

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

API ключи предназначены для постоянной аутентификации внешних сервисов без необходимости обновления токенов. Ключи поддерживают систему scopes для контроля доступа.

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

- `users:read` - чтение информации о пользователях
- `users:write` - изменение информации о пользователях
- `profile:read` - чтение профиля
- `profile:write` - изменение профиля
- `token:validate` - валидация токенов (для gRPC)
- `token:introspect` - детальная информация о токенах
- `admin:all` - все административные права
- `all` - полный доступ ко всем операциям

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

Auth Gateway предоставляет gRPC API для проверки токенов и авторизации между микросервисами. Поддерживает как JWT токены, так и API ключи.

### gRPC Endpoints

| Метод | Описание |
|-------|----------|
| `ValidateToken` | Проверка JWT токена или API ключа и получение user info |
| `GetUser` | Получение информации о пользователе по ID |
| `CheckPermission` | Проверка прав доступа (RBAC) |
| `IntrospectToken` | Детальная информация о JWT токене |

### Адрес gRPC сервера

- **Локально:** `localhost:50051`
- **Docker:** `auth-gateway:50051`

### Пример использования gRPC

```go
package main

import (
    "context"
    "log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // Подключение к auth gateway
    conn, err := grpc.NewClient(
        "localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Проверка токена
    // См. examples/grpc-client для полного примера
}
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
