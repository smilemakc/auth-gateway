# Auth Gateway - Подробное описание проекта

## Содержание

1. [Обзор проекта](#обзор-проекта)
2. [Цели и задачи](#цели-и-задачи)
3. [Ключевые особенности](#ключевые-особенности)
4. [Архитектура системы](#архитектура-системы)
5. [Технологический стек](#технологический-стек)
6. [Структура проекта](#структура-проекта)
7. [Компоненты системы](#компоненты-системы)
8. [API документация](#api-документация)
9. [Безопасность](#безопасность)
10. [Развертывание](#развертывание)
11. [Мониторинг и метрики](#мониторинг-и-метрики)

---

## Обзор проекта

**Auth Gateway** — это централизованная система аутентификации и авторизации, разработанная на языке программирования Go
для использования в экосистеме микросервисов. Проект представляет собой полнофункциональный authentication hub,
обеспечивающий безопасное управление пользовательскими данными, токенами, сессиями и правами доступа.

### Назначение

Auth Gateway служит единой точкой входа для управления аутентификацией и авторизацией во всей микросервисной
архитектуре. Он предоставляет:

- **REST API** для веб-приложений и мобильных клиентов
- **gRPC API** для взаимодействия между микросервисами
- **Централизованное управление** пользователями, правами и сессиями
- **Интеграцию** с различными провайдерами (OAuth, SMS, Email)

### Контекст использования

Проект предназначен для:

- Организаций, требующих централизованного управления аутентификацией
- Микросервисных архитектур, нуждающихся в единой системе авторизации
- Систем с высокими требованиями к безопасности и аудиту
- Приложений, требующих поддержки множественных способов аутентификации

---

## Цели и задачи

### Основные цели

1. **Централизованное управление аутентификацией** — единая система для управления пользователями и их доступом во всех
   микросервисах
2. **Высокая безопасность** — защита от атак, шифрование данных, secure token management
3. **Масштабируемость** — поддержка большого количества пользователей и запросов
4. **Гибкость** — поддержка различных способов аутентификации и авторизации
5. **Аудит и мониторинг** — полное логирование всех действий в системе

### Решаемые задачи

- Управление жизненным циклом JWT токенов (создание, валидация, отзыв)
- Управление постоянными API ключами для интеграций
- Управление сессиями пользователей и отслеживание устройств
- Организация RBAC (Role-Based Access Control) с гибкой системой прав
- Защита от несанкционированного доступа через IP-фильтры
- Интеграция с внешними OAuth провайдерами
- Отправка OTP через SMS и Email для верификации
- Двухфакторная аутентификация (2FA) с TOTP

---

## Ключевые особенности

### Методы аутентификации

#### 1. Традиционная аутентификация

- Email/пароль — классический метод входа
- Телефон/пароль — альтернативный способ с использованием номера телефона
- Безопасное хранение пароля через bcrypt с cost фактором 10

#### 2. OAuth 2.0 интеграция

Поддержка популярных OAuth провайдеров:

- **Google** — интеграция с Google аккаунтом
- **Yandex** — русскоязычный провайдер
- **GitHub** — для разработчиков
- **Instagram** — социальная сеть
- **Telegram** — мессенджер с высокой безопасностью

#### 3. Безпарольная аутентификация

- **Email OTP** — одноразовые коды через email
- **SMS OTP** — коды доставляются по SMS
- **TOTP (Time-based One-Time Password)** — приложения типа Google Authenticator

#### 4. API ключи

- Постоянные ключи для service-to-service интеграций
- Поддержка scopes для гранулярного контроля доступа
- Хеширование с SHA-256
- Отслеживание последнего использования

### Управление доступом (RBAC)

- **Ролевая система** — определение ролей (admin, user, moderator и т.д.)
- **Система прав** — гранулярные разрешения для каждой операции
- **Группирование прав** — привязка прав к ролям
- **Динамическая проверка** — валидация прав в реальном времени

### Управление сессиями

- **Отслеживание устройств** — информация о каждом сеансе пользователя
- **IP адреса** — логирование IP для анализа доступа
- **User Agent** — информация о браузере/приложении
- **Последняя активность** — отслеживание времени последнего использования
- **Отзыв сессий** — возможность закрыть конкретный сеанс

### Безопасность

- **JWT токены** с ограниченным временем жизни
- **Refresh токены** для безопасного продления сеансов
- **Rate limiting** на разных уровнях (регистрация, вход, API)
- **IP фильтры** — whitelist и blacklist IP адресов
- **Токен blacklist** — ведение списка отозванных токенов
- **Email/Phone верификация** — OTP-based подтверждение
- **Audit logging** — полная история всех действий

### Функции администратора

- **Управление пользователями** — CRUD операции с профилями
- **Управление ролями и правами** — редактирование RBAC системы
- **Статистика системы** — метрики использования и производительности
- **Логирование аудита** — просмотр истории действий
- **Управление SMS** — конфигурация провайдеров, логирование отправок
- **Кастомизация брендинга** — изменение цветов, логотипа, текстов
- **Режим обслуживания** — временное отключение сервиса
- **Интеграции** — управление webhooks и внешними сервисами

---

## Архитектура системы

### Архитектура по слоям (Layered Architecture)

```
┌─────────────────────────────────────────────────┐
│           HTTP & gRPC Interfaces                │
│     (REST API, gRPC Services)                   │
└──────────────┬──────────────────────────────────┘
               │
┌──────────────▼──────────────────────────────────┐
│          Middleware Layer                       │
│ (Auth, CORS, Rate Limiting, IP Filter, RBAC)    │
└──────────────┬──────────────────────────────────┘
               │
┌──────────────▼──────────────────────────────────┐
│         Handler/API Layer                       │
│    (Request handling, validation)               │
└──────────────┬──────────────────────────────────┘
               │
┌──────────────▼──────────────────────────────────┐
│         Service/Business Logic Layer            │
│  (Auth, User, RBAC, Session, SMS, Email)        │
└──────────────┬──────────────────────────────────┘
               │
┌──────────────▼──────────────────────────────────┐
│       Repository/Data Access Layer              │
│     (Database queries, transactions)            │
└──────────────┬──────────────────────────────────┘
               │
┌──────────────▼──────────────────────────────────┐
│    PostgreSQL Database & Redis Cache            │
│          (Data Persistence)                     │
└─────────────────────────────────────────────────┘
```

### Компоненты архитектуры

#### 1. Handler Layer (HTTP Handlers)

Отвечает за обработку HTTP запросов и формирование ответов. Основные группы:

- `auth_handler.go` — регистрация, вход, логаут
- `user_handler.go` — управление профилем пользователя
- `apikey_handler.go` — управление API ключами
- `oauth_handler.go` — OAuth интеграции
- `admin_handler.go` — административные операции
- `health_handler.go` — health checks

#### 2. Service Layer (Business Logic)

Содержит основную бизнес-логику приложения:

- `auth_service.go` — логика аутентификации
- `user_service.go` — управление пользователями
- `jwt_service.go` — создание и валидация JWT
- `apikey_service.go` — работа с API ключами
- `oauth_service.go` — OAuth логика
- `twofa_service.go` — двухфакторная аутентификация
- `rbac_service.go` — управление правами доступа
- `session_service.go` — управление сессиями
- `email_service.go` — отправка email
- `sms_service.go` — отправка SMS

#### 3. Repository Layer (Data Access)

Слой доступа к данным, обеспечивающий CRUD операции:

- `user_repository.go` — работа с таблицей users
- `token_repository.go` — управление токенами
- `apikey_repository.go` — хранение API ключей
- `session_repository.go` — информация о сессиях
- `audit_repository.go` — логирование действий
- Другие специализированные repositories

#### 4. Middleware Layer

Промежуточный слой для обработки запросов:

- **Auth Middleware** — проверка JWT токенов и API ключей
- **Rate Limiting** — ограничение количества запросов
- **CORS Middleware** — разрешение кросс-доменных запросов
- **IP Filter** — проверка IP адреса против whitelist/blacklist
- **RBAC Middleware** — проверка прав доступа
- **Recovery** — обработка паник
- **Logging** — логирование всех запросов

### Поток данных

```
Клиент (веб-приложение, мобильное приложение, микросервис)
    │
    ▼
HTTP Запрос / gRPC Call
    │
    ▼
Middleware (валидация, rate limiting, RBAC)
    │
    ▼
Handler (парсинг и валидация данных)
    │
    ▼
Service (бизнес-логика)
    │
    ▼
Repository (доступ к базе)
    │
    ▼
PostgreSQL / Redis
    │
    ▼
Ответ (JSON / gRPC message)
    │
    ▼
Клиент
```

### Интеграция с внешними системами

```
                    Auth Gateway
                         │
        ┌────────────────┼────────────────┐
        │                │                │
        ▼                ▼                ▼
    PostgreSQL         Redis           AWS SNS
    (основные)      (кеширование,    (SMS)
    (данные)        rate limit)
        │                │
        └────────────────┴────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
        ▼                ▼                ▼
    Google OAuth    Yandex OAuth   Email Service
    GitHub OAuth    Instagram      Twilio
    Telegram OAuth
```

---

## Технологический стек

### Языки и фреймворки

| Компонент            | Технология | Версия  |
|----------------------|------------|---------|
| **Язык**             | Go         | 1.24.0+ |
| **Web Framework**    | Gin        | Latest  |
| **RPC**              | gRPC       | Latest  |
| **Protocol Buffers** | protobuf   | v3      |

### Базы данных и кеширование

| Сервис             | Назначение                                | Версия |
|--------------------|-------------------------------------------|--------|
| **PostgreSQL**     | Основное хранилище данных                 | 14+    |
| **Redis**          | Кеширование, rate limiting, session store | 7+     |
| **UUID Extension** | Для работы с UUID в PostgreSQL            | -      |

### Безопасность

| Компонент               | Библиотека                 | Назначение                   |
|-------------------------|----------------------------|------------------------------|
| **JWT**                 | golang-jwt/jwt v5          | Создание и валидация токенов |
| **Хеширование паролей** | golang.org/x/crypto/bcrypt | Безопасное хранение паролей  |
| **2FA (TOTP)**          | pquerna/otp                | Двухфакторная аутентификация |
| **CORS**                | gin-contrib/cors           | Кросс-доменные запросы       |

### Интеграции

| Сервис               | Назначение                                  | Тип         |
|----------------------|---------------------------------------------|-------------|
| **AWS SDK v2**       | Отправка SMS через SNS                      | Cloud       |
| **Twilio**           | Альтернативный SMS провайдер                | Third-party |
| **OAuth Провайдеры** | Google, Yandex, GitHub, Instagram, Telegram | Third-party |
| **Email Service**    | SMTP для отправки писем                     | Third-party |

### Утилиты и инструменты

| Инструмент                  | Назначение                    |
|-----------------------------|-------------------------------|
| **Docker & Docker Compose** | Контейнеризация и оркестрация |
| **Make**                    | Автоматизация сборки          |
| **golang-migrate**          | Управление миграциями БД      |
| **sqlx**                    | Работа с SQL запросами        |

### Зависимости (основные)

```go
// Веб-фреймворк
github.com/gin-gonic/gin v1.x.x

// Протоколы
google.golang.org/grpc v1.x.x
google.golang.org/protobuf v1.x.x

// Безопасность
github.com/golang-jwt/jwt/v5 v5.x.x
golang.org/x/crypto v0.x.x
github.com/pquerna/otp v1.x.x

// Базы данных
github.com/jmoiron/sqlx v1.x.x
github.com/lib/pq v1.x.x
github.com/redis/go -redis/v9 v9.x.x

// CORS и сеть
github.com/gin-contrib/cors v1.x.x

// AWS
github.com/aws/aws-sdk-go -v2 v1.x.x
github.com/aws/aws-sdk-go -v2/service/sns v1.x.x

// Логирование
go.uber.org/zap v1.x.x
```

---

## Структура проекта

### Дерево каталогов

```
auth-gateway/
├── cmd/
│   └── server/
│       └── main.go                    # Точка входа приложения
│
├── internal/
│   ├── config/
│   │   ├── config.go                  # Парсинг конфигурации
│   │   └── database.go                # Конфиг БД
│   │
│   ├── models/                        # Data models (19 файлов)
│   │   ├── user.go
│   │   ├── token.go
│   │   ├── api_key.go
│   │   ├── session.go
│   │   ├── permission.go
│   │   ├── otp.go
│   │   ├── oauth.go
│   │   ├── audit_log.go
│   │   └── ...
│   │
│   ├── repository/                    # Data access layer (18 файлов)
│   │   ├── user_repository.go
│   │   ├── token_repository.go
│   │   ├── apikey_repository.go
│   │   ├── session_repository.go
│   │   ├── role_repository.go
│   │   ├── permission_repository.go
│   │   ├── audit_repository.go
│   │   └── ...
│   │
│   ├── service/                       # Business logic (15 файлов)
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── jwt_service.go
│   │   ├── apikey_service.go
│   │   ├── session_service.go
│   │   ├── oauth_service.go
│   │   ├── twofa_service.go
│   │   ├── email_service.go
│   │   ├── sms_service.go
│   │   ├── rbac_service.go
│   │   └── ...
│   │
│   ├── handler/                       # HTTP handlers (10 файлов)
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── apikey_handler.go
│   │   ├── oauth_handler.go
│   │   ├── admin_handler.go
│   │   ├── health_handler.go
│   │   └── ...
│   │
│   ├── middleware/
│   │   ├── auth.go                    # JWT + API Key аутентификация
│   │   ├── rate_limit.go              # Rate limiting
│   │   ├── cors.go                    # CORS поддержка
│   │   ├── ip_filter.go               # IP whitelist/blacklist
│   │   ├── rbac.go                    # RBAC проверка прав
│   │   ├── recovery.go                # Error recovery
│   │   └── logger.go                  # Request logging
│   │
│   ├── grpc/
│   │   └── server.go                  # gRPC сервер
│   │
│   ├── sms/
│   │   ├── provider.go                # Интерфейс SMS провайдера
│   │   ├── aws_sns.go                 # AWS SNS реализация
│   │   ├── twilio.go                  # Twilio реализация
│   │   └── mock.go                    # Mock для тестирования
│   │
│   └── utils/
│       ├── helpers.go                 # Вспомогательные функции
│       ├── validators.go              # Валидация данных
│       └── errors.go                  # Кастомные ошибки
│
├── pkg/
│   ├── jwt/
│   │   ├── jwt.go                     # JWT сервис
│   │   └── claims.go                  # Custom claims
│   │
│   └── logger/
│       └── logger.go                  # Логирование (zap)
│
├── migrations/                        # SQL миграции (15 файлов)
│   ├── 001_create_users_table.up.sql
│   ├── 002_create_tokens_table.up.sql
│   ├── 003_create_api_keys_table.up.sql
│   ├── 004_create_sessions_table.up.sql
│   ├── 005_create_roles_table.up.sql
│   ├── 006_create_permissions_table.up.sql
│   ├── 007_create_oauth_accounts_table.up.sql
│   └── ...
│
├── proto/
│   └── auth.proto                     # gRPC сервис определения
│
├── examples/
│   └── grpc-client/
│       ├── main.go                    # Пример gRPC клиента
│       └── README.md
│
├── docs/
│   ├── README.md
│   ├── API.md
│   ├── DEPLOYMENT.md
│   ├── ARCHITECTURE.md
│   └── SMS_PROVIDER.md
│
├── docker/
│   ├── Dockerfile                     # Multi-stage Docker build
│   └── docker-compose.yml             # Оркестрация сервисов
│
├── Makefile                           # Автоматизация задач
├── .env.example                       # Пример конфигурации
├── go.mod & go.sum                    # Go зависимости
├── go.mod
├── go.sum
├── README.md                          # Основная документация
└── PROJECT_DESCRIPTION.md             # Этот файл
```

### Количество файлов

- **Go файлы:** 93 в total (89 в internal/)
- **SQL миграции:** 15 файлов
- **Proto определения:** 1 файл
- **Docker конфигурация:** 2 файла
- **Примеры и документация:** множество вспомогательных файлов

---

## Компоненты системы

### 1. Аутентификация (Authentication Service)

**Файл:** `internal/service/auth_service.go`

#### Функциональность

- Регистрация новых пользователей
- Вход с email/пароль или телефон/пароль
- Валидация учетных данных
- Создание JWT токенов
- Обновление токенов через refresh token
- Логаут и отзыв токенов

#### Процесс входа

```
1. Получить email/пароль от клиента
2. Найти пользователя в БД
3. Сравнить пароль с bcrypt хешем
4. Проверить статус пользователя (активен/заблокирован)
5. Создать access token (TTL: 15 минут)
6. Создать refresh token (TTL: 7 дней)
7. Сохранить refresh token в БД
8. Вернуть оба токена клиенту
```

#### Безопасность

- Пароли никогда не возвращаются
- Хеширование bcrypt с cost 10
- Rate limiting на попытки входа
- Логирование всех попыток в аудит

### 2. Управление пользователями (User Service)

**Файл:** `internal/service/user_service.go`

#### Операции

- Создание профиля пользователя
- Получение информации о пользователе
- Обновление профиля
- Удаление пользователя (soft delete)
- Смена пароля
- Верификация email/телефона

#### Данные пользователя

```go
type User struct {
ID            uuid.UUID
Email         string
Phone         string
Username      string
FullName      string
ProfilePicture string
PasswordHash  string
IsVerified    bool
IsTwoFAEnabled bool
Roles         []Role
Status        string // active, suspended, deleted
CreatedAt     time.Time
UpdatedAt     time.Time
LastLogin     time.Time
}
```

### 3. JWT и токены (JWT Service)

**Файл:** `pkg/jwt/jwt.go`

#### Типы токенов

**Access Token:**

- TTL: 15 минут
- Содержит: User ID, email, roles, permissions
- Используется для каждого запроса к API
- Хранится в памяти клиента

**Refresh Token:**

- TTL: 7 дней
- Используется для получения нового access token
- Хранится в защищенной БД
- Может быть отозван

#### Custom Claims

```go
type Claims struct {
UserID   uuid.UUID
Email    string
Roles    []string
Scopes   []string
jwt.RegisteredClaims
}
```

#### Процесс валидации

```
1. Получить токен из заголовка Authorization или X-API-Key
2. Парсить токен и проверить подпись
3. Проверить expiration time
4. Проверить, не в ли токен в blacklist'е
5. Вернуть claims для использования в handler'е
```

### 4. API Ключи (API Key Service)

**Файл:** `internal/service/apikey_service.go`

#### Функциональность

- Генерация постоянных API ключей
- Хеширование ключей перед сохранением
- Валидация ключей при каждом запросе
- Управление scopes для гранулярного доступа
- Отслеживание последнего использования
- Отзыв и удаление ключей

#### Формат ключа

```
agw_<base64_random_32_bytes>
Пример: agw_AbC1XyZ9DefGhIjKlMnOpQrStUvWxYz
```

#### Scopes

- `users:read` — чтение данных пользователей
- `users:write` — изменение пользователей
- `profile:read` — чтение собственного профиля
- `profile:write` — изменение собственного профиля
- `token:validate` — валидация токенов (gRPC)
- `token:introspect` — детальная информация о токенах
- `admin:all` — все административные права
- `all` — полный доступ

### 5. OAuth интеграция

**Файл:** `internal/service/oauth_service.go`

#### Поддерживаемые провайдеры

1. **Google OAuth**
    - URL: https://accounts.google.com
    - Использует: email, профильное фото

2. **Yandex OAuth**
    - URL: https://oauth.yandex.com
    - Локализовано для русскоязычных пользователей

3. **GitHub OAuth**
    - Для разработчиков
    - Использует: email, username, аватар

4. **Instagram OAuth**
    - Социальная сеть
    - Интеграция профиля

5. **Telegram OAuth**
    - Высокоуровневая безопасность
    - Использует: Telegram ID

#### Процесс OAuth

```
1. Пользователь кликает на "Sign in with Google"
2. Редирект на Google с client_id и redirect_uri
3. Пользователь авторизуется в Google
4. Google редирект обратно с authorization code
5. Backend обменивает code на access_token
6. Получаем профиль пользователя
7. Создаем или обновляем OAuthAccount запись
8. Создаем JWT токены
9. Редирект на фронтенд с токеном
```

### 6. Двухфакторная аутентификация (2FA Service)

**Файл:** `internal/service/twofa_service.go`

#### Методы 2FA

1. **TOTP (Time-based One-Time Password)**
    - Использует пакет `pquerna/otp`
    - Совместим с Google Authenticator, Authy
    - QR код для быстрого добавления в приложение

2. **Backup Codes**
    - Генерируются при включении 2FA
    - 8 кодов по 8 символов
    - Используются когда нет доступа к TOTP приложению

#### Процесс включения 2FA

```
1. Пользователь запрашивает setup 2FA
2. Сервер генерирует secret ключ
3. Возвращаем QR код
4. Пользователь сканирует QR в приложение (Google Auth)
5. Пользователь подтверждает 6-значный код из приложения
6. Сохраняем secret в БД
7. Генерируем и сохраняем 8 backup кодов
8. 2FA включена
```

#### Использование 2FA при входе

```
1. Пользователь вводит email/пароль
2. Проверяем пароль
3. Если 2FA включена, требуем TOTP код
4. Пользователь вводит 6-значный код из приложения
5. Верифицируем код
6. Выдаем JWT токены
```

### 7. OTP верификация (Email/SMS)

**Файл:** `internal/service/otp_service.go`

#### Применение

- Email верификация при регистрации
- SMS верификация телефона
- Восстановление пароля через email
- Двухфакторная аутентификация

#### Процесс OTP

```
1. Пользователь запрашивает OTP
2. Генерируем 6-значный код
3. Сохраняем в БД с TTL (обычно 10 минут)
4. Отправляем код пользователю (email/SMS)
5. Пользователь вводит код
6. Проверяем код и TTL
7. Если верен, выполняем нужное действие
8. Удаляем использованный код
```

### 8. SMS интеграция

**Файл:** `internal/sms/`

#### Поддерживаемые провайдеры

**1. AWS SNS (Amazon Simple Notification Service)**

```go
// Конфиг
SMS_PROVIDER= aws
AWS_REGION = us-east-1
AWS_ACCESS_KEY_ID = ***
AWS_SECRET_ACCESS_KEY = ***
```

**2. Twilio**

```go
SMS_PROVIDER= twilio
TWILIO_ACCOUNT_SID = ***
TWILIO_AUTH_TOKEN =***
TWILIO_PHONE_NUMBER = +1234567890
```

**3. Mock (для тестирования)**

```go
SMS_PROVIDER = mock
```

#### Rate Limiting SMS

- Per phone: max 3 попытки
- Per hour: max 5 SMS
- Per day: max 10 SMS

#### SMS логирование

- Все отправленные SMS записываются
- Статус доставки
- Время отправки и доставки
- Номер телефона получателя

### 9. Email сервис

**Файл:** `internal/service/email_service.go`

#### Использование

- Регистрационные письма
- OTP коды
- Восстановление пароля
- Уведомления о действиях аккаунта
- Admin уведомления

#### Интеграция

- SMTP для отправки
- HTML шаблоны писем
- Поддержка кастомных шаблонов

### 10. RBAC система (Role-Based Access Control)

**Файлы:** `internal/service/rbac_service.go`

#### Компоненты

1. **Роли** — наборы прав (admin, user, moderator)
2. **Права** — атомарные разрешения (user:read, user:write)
3. **Привязка** — связь между ролями и правами

#### Таблицы

```sql
-- Определяем роли
roles
    (id, name, description, created_at)

-- Определяем права
    permissions
    (id, name, resource, action, description)

-- Привязываем права к ролям
    role_permissions
    (role_id, permission_id)

-- Пользователи имеют роли
    user_roles
    (user_id, role_id)
```

#### Процесс проверки прав

```
1. Получить пользователя из JWT токена
2. Получить все роли пользователя
3. Получить все права для этих ролей
4. Проверить наличие требуемого права
5. Разрешить или запретить действие
```

#### Примеры ролей

```
admin:
  - users:read, users:write, users:delete
  - roles:read, roles:write, roles:delete
  - audit:read
  - system:manage

user:
  - profile:read, profile:write
  - api-keys:create, api-keys:read, api-keys:delete

moderator:
  - users:read, users:suspend
  - audit:read
  - reports:read, reports:resolve
```

### 11. Управление сессиями

**Файл:** `internal/service/session_service.go`

#### Информация о сессии

```go
type Session struct {
ID           uuid.UUID
UserID       uuid.UUID
RefreshToken string
IPAddress    string
UserAgent    string
CreatedAt    time.Time
LastActiveAt time.Time
ExpiresAt    time.Time
DeviceName   string
IsActive     bool
}
```

#### Функциональность

- Отслеживание каждой сессии пользователя
- IP адрес и User Agent для идентификации устройства
- Время последней активности
- Возможность закрыть конкретную сессию
- Просмотр всех активных сессий

#### Использование

```
GET /auth/sessions - список всех сессий
POST /auth/sessions/{id}/revoke - закрыть сессию
DELETE /auth/sessions/{id} - удалить сессию
```

### 12. Аудит логирование

**Файл:** `internal/service/audit_service.go`

#### Логируемые события

- Вход в систему (успешный/неудачный)
- Регистрация
- Смена пароля
- Изменение профиля
- Включение/отключение 2FA
- Создание/удаление API ключей
- Административные действия
- Изменение прав доступа
- Изменение ролей

#### Информация в логе

```go
type AuditLog struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Action    string // "user.login", "user.password_change"
    Resource  string // "user", "api_key", "role"
    Detail    string // JSON с деталями
    IPAddress string
    Status    string // "success", "failure"
    Timestamp time.Time
}

```

#### Просмотр логов

```
GET /admin/audit-logs - список логов (фильтрация по пользователю, действию)
GET /admin/audit-logs/{id} - детали логов
```

### 13. IP фильтры

**Файл:** `internal/middleware/ip_filter.go`

#### Функциональность

- Whitelist (только разрешенные IP)
- Blacklist (блокировка определенных IP)
- CIDR диапазоны (например, 192.168.1.0/24)
- Динамическое обновление без перезагрузки
- Логирование блокированных IP

#### Применение

```go
// Проверяется в middleware для всех запросов
middleware.IPFilter(allowedIPs, blockedIPs)

// Результат: заблокированные IP получают 403 Forbidden
```

### 14. Rate Limiting

**Файл:** `internal/middleware/rate_limit.go`

#### Стратегии

**1. Регистрация (signup)**

- Max: 5 попыток за 1 час с одного IP
- Применяется на POST /auth/signup

**2. Вход (signin)**

- Max: 10 попыток за 15 минут с одного IP
- Применяется на POST /auth/signin

**3. API ключи (OTP)**

- Max: 3 попытки per phone
- Max: 5 попыток за 1 час per phone
- Max: 10 попыток за 1 день per phone

**4. General API**

- Max: 100 запросов в минуту для authenticated пользователей
- Max: 30 запросов в минуту для unauthenticated

#### Реализация

- Использует Redis для хранения счетчиков
- Key: `rate_limit:<ip>:<endpoint>`
- TTL: соответствующий временной интервал

---

## API документация

### HTTP REST API (Port 3000)

#### Base URL

```
http://localhost:3000
```

#### Authentication методы

1. **JWT Bearer Token**
   ```
   Authorization: Bearer <access_token>
   ```

2. **API Key (заголовок)**
   ```
   X-API-Key: <api_key>
   ```

3. **API Key (Bearer)**
   ```
   Authorization: Bearer <api_key>
   ```

### Endpoints по категориям

#### 1. Аутентификация (/auth)

**POST /auth/signup** — Регистрация

```
Request:
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "SecurePass123",
  "full_name": "John Doe"
}

Response:
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "johndoe",
    "full_name": "John Doe",
    "is_verified": false
  },
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc..."
}
```

**POST /auth/signin** — Вход

```
Request:
{
  "email": "user@example.com",
  "password": "SecurePass123"
}

Response:
{
  "user": {...},
  "access_token": "...",
  "refresh_token": "..."
}
```

**POST /auth/refresh** — Обновление токена

```
Request:
{
  "refresh_token": "eyJhbGc..."
}

Response:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc..."
}
```

**POST /auth/logout** — Логаут (требует авторизацию)

```
Request: (пусто)

Response:
{
  "message": "Logged out successfully"
}
```

#### 2. OTP верификация

**POST /otp/send** — Отправить OTP

```
Request:
{
  "type": "email" | "sms",
  "destination": "user@example.com" | "+1234567890"
}

Response:
{
  "message": "OTP sent successfully",
  "expires_in": 600
}
```

**POST /otp/verify** — Верифицировать OTP

```
Request:
{
  "destination": "user@example.com",
  "code": "123456"
}

Response:
{
  "verified": true,
  "message": "OTP verified"
}
```

#### 3. Управление профилем (требует авторизацию)

**GET /auth/profile** — Получить профиль

```
Response:
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "username": "johndoe",
  "full_name": "John Doe",
  "is_verified": true,
  "is_two_fa_enabled": false,
  "roles": ["user"],
  "created_at": "2024-01-01T00:00:00Z"
}
```

**PUT /auth/profile** — Обновить профиль

```
Request:
{
  "full_name": "John Doe Updated",
  "profile_picture": "https://example.com/pic.jpg"
}

Response: (обновленный профиль)
```

**POST /auth/change-password** — Смена пароля

```
Request:
{
  "old_password": "OldPass123",
  "new_password": "NewPass456"
}

Response:
{
  "message": "Password changed successfully"
}
```

#### 4. API Ключи (требует авторизацию)

**POST /api-keys** — Создать API ключ

```
Request:
{
  "name": "Production API Key",
  "description": "For backend service",
  "scopes": ["users:read", "users:write"],
  "expires_at": "2025-12-31T23:59:59Z"  // опционально
}

Response:
{
  "api_key": {
    "id": "...",
    "name": "Production API Key",
    "key_prefix": "agw_AbC1",
    "scopes": ["users:read", "users:write"],
    "is_active": true,
    "created_at": 1234567890
  },
  "plain_key": "agw_AbC123XyZ...полный_ключ..."  // только при создании!
}
```

**GET /api-keys** — Список API ключей

```
Response:
{
  "keys": [
    {
      "id": "...",
      "name": "Production API Key",
      "key_prefix": "agw_AbC1",
      "scopes": ["users:read"],
      "is_active": true,
      "last_used_at": "2024-01-15T10:30:00Z",
      "created_at": 1234567890
    }
  ]
}
```

**PUT /api-keys/{id}** — Обновить API ключ

```
Request:
{
  "name": "Updated Name",
  "description": "Updated description",
  "scopes": ["users:read"]  // опционально
}
```

**POST /api-keys/{id}/revoke** — Отозвать ключ

```
Response:
{
  "message": "API key revoked"
}
```

**DELETE /api-keys/{id}** — Удалить ключ

```
Response:
{
  "message": "API key deleted"
}
```

#### 5. OAuth (в разработке)

**GET /auth/{provider}** — Инициировать OAuth

```
Parameters: provider = google | yandex | github | instagram | telegram
Query params: redirect_uri, state
```

**GET /auth/{provider}/callback** — OAuth callback

```
Parameters: code, state
```

#### 6. 2FA (требует авторизацию)

**POST /auth/2fa/setup** — Настройка 2FA

```
Response:
{
  "secret": "JBSWY3DPEBLW64TMMQ======",
  "qr_code": "data:image/png;base64,...",
  "backup_codes": [
    "ABCD1234",
    "EFGH5678",
    ...
  ]
}
```

**POST /auth/2fa/verify** — Верификация 2FA

```
Request:
{
  "code": "123456"
}

Response:
{
  "verified": true,
  "message": "2FA enabled"
}
```

**POST /auth/2fa/disable** — Отключение 2FA

```
Request:
{
  "password": "YourPassword"
}

Response:
{
  "message": "2FA disabled"
}
```

#### 7. Администратор (/admin)

**GET /admin/stats** — Статистика системы

```
Response:
{
  "total_users": 1234,
  "active_users": 567,
  "new_users_today": 12,
  "total_api_keys": 89,
  "active_sessions": 456,
  "failed_logins_today": 3
}
```

**GET /admin/users** — Список пользователей

```
Query params: page=1, limit=20, search=john, role=admin
```

**GET /admin/users/{id}** — Информация о пользователе

**PUT /admin/users/{id}** — Обновить пользователя

```
Request:
{
  "full_name": "New Name",
  "roles": ["user", "moderator"],
  "status": "active" | "suspended"
}
```

**DELETE /admin/users/{id}** — Удалить пользователя

**GET /admin/audit-logs** — Логи аудита

```
Query params: page=1, limit=20, user_id=..., action=...
```

**GET /admin/rbac/roles** — Управление ролями

**GET /admin/rbac/permissions** — Управление правами

**POST /admin/ip-filters** — Управление IP фильтрами

**PUT /admin/branding** — Кастомизация брендинга

```
Request:
{
  "primary_color": "#FF0000",
  "logo_url": "https://...",
  "app_name": "My App"
}
```

**PUT /admin/system/maintenance** — Режим обслуживания

```
Request:
{
  "is_maintenance": true,
  "message": "System is under maintenance"
}
```

### gRPC API (Port 50051)

#### Service Definition

```protobuf
service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CheckPermission(CheckPermissionRequest) returns (CheckPermissionResponse);
  rpc IntrospectToken(IntrospectTokenRequest) returns (IntrospectTokenResponse);
}
```

#### Методы

**ValidateToken** — Валидировать JWT/API key

```protobuf
request {
string token = 1;
    string token_type = 2;  // "jwt" или "api_key"
    }

response {
bool valid = 1;
    string user_id = 2;
string email = 3;
    repeated string roles = 4;
    repeated string scopes = 5;
    }
```

**GetUser** — Получить пользователя по ID

```protobuf
request {
string user_id = 1;
    }

response {
string id = 1;
    string email = 2;
string username = 3;
    string full_name = 4;
repeated string roles = 5;
    int64 created_at = 6;
    }
```

**CheckPermission** — Проверить право доступа

```protobuf
request {
string user_id = 1;
    string permission = 2;  // e.g., "users:write"
    }

response {
bool allowed = 1;
    }
```

**IntrospectToken** — Детальная информация о токене

```protobuf
request {
string token = 1;
    }

response {
string user_id = 1;
    string email = 2;
int64 issued_at = 3;
    int64 expires_at = 4;
repeated string scopes = 5;
    bool is_valid = 6;
    }
```

### Health Checks

**GET /auth/health** — Полная проверка

```
Response:
{
  "status": "healthy",
  "database": "ok",
  "redis": "ok",
  "uptime": 3600,
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**GET /auth/ready** — Готовность к работе

```
Response: 200 OK (если готов) или 503 Service Unavailable
```

**GET /auth/live** — Liveness (использует Kubernetes)

```
Response: 200 OK (если жив)
```

---

## Безопасность

### 1. Защита от атак

#### CSRF (Cross-Site Request Forgery)

- JWT токены содержат уникальный ID пользователя
- Tokens имеют ограниченное время жизни
- Используется SameSite cookies где возможно

#### XSS (Cross-Site Scripting)

- Все пользовательские данные экранируются
- Используется Content-Security-Policy заголовок
- API возвращает JSON, не HTML

#### SQL Injection

- Используется параметризованные запросы через sqlx
- Никогда не используется string concatenation в SQL

#### Brute Force

- Rate limiting на регистрацию (5 в час)
- Rate limiting на вход (10 в 15 минут)
- Логирование всех неудачных попыток

#### Token Theft

- Refresh tokens хранятся в защищенной БД
- Access tokens имеют короткий TTL (15 минут)
- Token blacklist для отозванных токенов
- HTTPS для передачи токенов

#### Man-in-the-Middle (MITM)

- Обязательное использование HTTPS
- Certificate pinning для критичных операций

### 2. Шифрование и хеширование

#### Пароли

```go
// Bcrypt с cost factor 10
bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
// Cost 10 = ~100ms на современном оборудовании
```

#### API Ключи

```go
// SHA-256 хеширование
sha256.Sum256([]byte(apiKey))
```

#### Токены

```go
// HMAC-SHA256 подпись для JWT
// Алгоритм: HS256
```

### 3. Управление секретами

#### Требуемые секреты

```
JWT_ACCESS_SECRET       # Подпись access токенов (256-бит)
JWT_REFRESH_SECRET      # Подпись refresh токенов (256-бит)
DB_PASSWORD             # Пароль PostgreSQL
REDIS_PASSWORD          # Пароль Redis
OAUTH_GOOGLE_SECRET     # Google OAuth secret
OAUTH_GITHUB_SECRET     # GitHub OAuth secret
...
```

#### Хранение

- В `.env` файле для разработки
- В `docker-compose.override.yml` для локального docker
- В Kubernetes Secrets для production
- В AWS Secrets Manager или аналоге для cloud

#### Ротация

- Рекомендуется менять JWT секреты каждые 3 месяца
- Старые секреты должны быть доступны для валидации старых токенов

### 4. Data Privacy

#### GDPR соответствие

- Право на удаление данных (right to be forgotten)
- Право на доступ к данным
- Право на исправление данных
- Экспорт данных в машиночитаемом формате

#### Логирование

- Логируются только необходимые данные
- Чувствительные данные (пароли) никогда не логируются
- Логи хранятся с ограниченный период

#### Резервные копии

- Регулярные резервные копии БД
- Зашифрованное хранилище резервов
- Тестирование восстановления из резервов

### 5. Transport Security

#### HTTPS

- Обязательно для production
- Минимум TLS 1.2
- Сильные cipher suites

#### CORS

```
Allowed Origins: https://yourdomain.com
Allowed Methods: GET, POST, PUT, DELETE
Allowed Headers: Content-Type, Authorization
Allow Credentials: true
```

#### Security Headers

```
Strict-Transport-Security: max-age=31536000
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
```

---

## Развертывание

### Локальное развертывание (Docker)

#### Требования

- Docker 20.10+
- Docker Compose 2.0+
- 2GB свободной памяти
- 1GB свободного диска

#### Команды

```bash
# 1. Подготовка
git clone https://github.com/smilemakc/auth-gateway.git
cd auth-gateway
cp .env.example .env

# 2. Запуск сервисов
docker-compose up -d

# 3. Проверка статуса
docker-compose ps
docker-compose logs -f auth-gateway

# 4. Тестирование
curl http://localhost:3000/auth/health
```

### Production развертывание

#### Архитектура

```
Internet
    │
    ▼
Nginx (Reverse Proxy, HTTPS termination)
    │
    ▼
Auth Gateway Instances (Load Balanced)
    │
┌───┴───┬─────────┬──────────┐
│       │         │          │
▼       ▼         ▼          ▼
PG      PG     Redis      AWS SNS
(Master) (Replica) (Cluster)
```

#### Infrastructure as Code

**Kubernetes Deployment (YAML)**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: auth-gateway
  template:
    metadata:
      labels:
        app: auth-gateway
    spec:
      containers:
        - name: auth-gateway
          image: auth-gateway:latest
          ports:
            - containerPort: 3000
            - containerPort: 50051
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: auth-gateway-secrets
                  key: database-url
          livenessProbe:
            httpGet:
              path: /auth/live
              port: 3000
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /auth/ready
              port: 3000
            initialDelaySeconds: 10
            periodSeconds: 5
```

#### Шаги развертывания

1. **Подготовка инфраструктуры**
    - Настроить PostgreSQL 14+
    - Настроить Redis 7+
    - Настроить AWS SNS для SMS

2. **Конфигурация**
    - Установить production JWT секреты
    - Настроить CORS для вашего домена
    - Установить database credentials
    - Настроить SMTP для email

3. **SSL/TLS**
    - Получить SSL сертификат (Let's Encrypt)
    - Настроить Nginx с HTTPS
    - Установить правильные headers

4. **Миграции БД**
   ```bash
   migrate -path migrations -database "postgres://..." up
   ```

5. **Запуск приложения**
    - Запустить Docker контейнер или бинарник
    - Проверить логи
    - Тестировать endpoints

6. **Мониторинг**
    - Настроить Prometheus для метрик
    - Настроить Grafana для dashboard
    - Настроить alerting (PagerDuty, Slack)

### Масштабирование

#### Горизонтальное масштабирование

- Запустить несколько инстансов Auth Gateway
- Использовать Load Balancer (Nginx, AWS ELB)
- Общая PostgreSQL БД для всех инстансов
- Общий Redis кластер для rate limiting

#### Вертикальное масштабирование

- Увеличить CPU/память для инстанса
- Увеличить connection pool к БД
- Увеличить Redis memory

#### Оптимизация

- Кеширование в Redis (session data)
- Connection pooling к БД
- Индексы в PostgreSQL
- Compression для больших ответов

---

## Мониторинг и метрики

### Health Checks

**Endpoints:**

- `GET /auth/health` — Полная проверка (DB + Redis)
- `GET /auth/ready` — Readiness probe (Kubernetes)
- `GET /auth/live` — Liveness probe (Kubernetes)

### Логирование

**Уровни:**

- DEBUG — Подробная информация
- INFO — Информационные сообщения
- WARN — Предупреждения
- ERROR — Ошибки
- FATAL — Критические ошибки

**Логируемые события:**

- Аутентификация (вход, выход, ошибки)
- Database операции
- Rate limiting триггеры
- IP filter триггеры
- API ключ использование
- OAuth интеграции
- SMS отправки
- Все ошибки

### Prometheus метрики (в разработке)

**Плануемые метрики:**

- `auth_gateway_login_total` — Общее количество входов
- `auth_gateway_login_failures` — Неудачные входы
- `auth_gateway_signup_total` — Новые регистрации
- `auth_gateway_api_requests_duration` — Длительность запросов
- `auth_gateway_database_queries_duration` — Время БД запросов
- `auth_gateway_active_sessions` — Активные сессии
- `auth_gateway_rate_limit_triggers` — Срабатывания rate limit
- `auth_gateway_token_validations` — Валидация токенов

### Алерты

**Критичные:**

- Database connection lost
- Redis connection lost
- High rate of failed logins (brute force)
- High rate of 5xx errors
- Service down (health check failed)

**Важные:**

- High memory usage
- High CPU usage
- Rate limit frequently triggered
- Slow database queries
- High authentication latency

**Информационные:**

- JWT secret rotation reminder
- Certificate expiration reminder
- Backup verification
- Disk space running low

---

## Заключение

**Auth Gateway** — это полнофункциональная, производительная и безопасная система аутентификации и авторизации для
микросервисных архитектур. Проект разработан с учетом лучших практик безопасности, масштабируемости и удобства
использования.

### Ключевые преимущества

- ✅ Централизованное управление аутентификацией
- ✅ Поддержка множественных методов входа
- ✅ Гибкая RBAC система
- ✅ Высокоуровневая безопасность
- ✅ Простота интеграции через REST и gRPC
- ✅ Полный аудит и логирование
- ✅ Production-ready с Docker и K8s поддержкой

### Дальнейшее развитие

- Завершение OAuth интеграции
- Добавление Prometheus метрик
- Улучшение тестового покрытия
- OpenAPI/Swagger документация
- Kubernetes Helm charts
- Улучшенное управление email шаблонами

---

**Версия документации:** 1.0
**Последнее обновление:** 2025-12-12
**Язык:** Русский
**Статус проекта:** Active Development
