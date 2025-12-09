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
- ⏳ OAuth интеграция (Google, Yandex, GitHub, Instagram) - в разработке
- ⏳ Prometheus метрики - в разработке

## Технологический стек

- **Язык:** Go 1.23+
- **Web Framework:** Gin
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
