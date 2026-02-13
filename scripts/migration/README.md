# Auth Gateway User Migration Tool

Python CLI-утилита для миграции пользовательских баз данных из внешних систем (CRM, легаси-приложения и т.д.) в Auth
Gateway. Поддерживает PostgreSQL и MySQL как источники данных.

## Структура

```
scripts/migration/
├── migrate.py              # CLI точка входа (click-команды)
├── config.yaml.example     # Пример конфигурации
├── requirements.txt        # Python-зависимости
├── tests/                  # Тесты
└── lib/
    ├── __init__.py          # Модели данных, enum'ы, конфигурация, загрузчик YAML
    ├── exporter.py          # Экспорт пользователей из source БД (server-side cursor)
    ├── deduplicator.py      # Дедупликация записей перед импортом
    ├── importer.py          # Импорт в Auth Gateway через REST API (async, батчевый)
    ├── verifier.py          # Верификация целостности после миграции
    ├── shadow.py            # Трансформация source-таблицы в shadow-таблицу
    ├── password.py          # Детекция алгоритма хеширования паролей
    └── report.py            # Генерация JSON-отчётов
```

## Быстрый старт

### Установка зависимостей

```bash
cd scripts/migration
pip install -r requirements.txt
```

### Подготовка конфигурации

```bash
cp config.yaml.example config.yaml
```

Отредактируйте `config.yaml`, указав параметры source БД и Auth Gateway. Секреты можно передать через переменные
окружения:

```bash
export SOURCE_DB_PASSWORD="your_db_password"
export AUTH_GATEWAY_API_KEY="agw_your_api_key"
```

### Запуск

```bash
# 1. Анализ source БД (без изменений)
python migrate.py analyze -c config.yaml

# 2. Пробная миграция (dry-run, по умолчанию)
python migrate.py migrate -c config.yaml

# 3. Боевая миграция
python migrate.py migrate -c config.yaml --no-dry-run

# 4. Проверка целостности
python migrate.py verify -c config.yaml

# 5. Генерация SQL для shadow-таблицы (для DBA-ревью)
python migrate.py generate-shadow-sql -c config.yaml -o shadow_migration.sql
```

## CLI-команды

| Команда               | Описание                                                                                            |
|-----------------------|-----------------------------------------------------------------------------------------------------|
| `analyze`             | Анализ source БД: подсчёт пользователей, тип ID, алгоритм паролей, поиск дубликатов                 |
| `migrate`             | Полная миграция (по умолчанию dry-run). Флаги: `--no-dry-run`, `--step import\|verify\|shadow\|all` |
| `verify`              | Проверка целостности: сравнение количества, рандомная выборка для проверки полей                    |
| `generate-shadow-sql` | Генерация SQL-файла для трансформации source-таблицы в shadow. Флаг: `-o output.sql`                |

## Пайплайн миграции

Команда `migrate --step all` выполняет 4 шага последовательно:

1. **Export & Dedup** — чтение пользователей из source БД через server-side cursor (экономия памяти), дедупликация в
   Python
2. **Import** — батчевый импорт в Auth Gateway через `POST /api/admin/users/import` (пакеты по 50 записей, параллельно
   через asyncio semaphore)
3. **Verify** — проверка целостности: сравнение общего количества, рандомная выборка для сверки полей
4. **Shadow** — трансформация исходной таблицы users: удаление чувствительных колонок, добавление служебных (только в
   live-режиме)

Каждый шаг можно запустить отдельно через `--step import`, `--step verify`, `--step shadow`.

После завершения создаётся JSON-отчёт `migration_report_YYYYMMDD_HHMMSS.json`.

## Конфигурация (`config.yaml`)

### `migration` — общие параметры

```yaml
migration:
  mode: "dry-run"       # "full" | "dry-run"
  batch_size: 100       # Размер пакета при чтении из source БД
  workers: 4            # Количество параллельных HTTP-запросов к Auth Gateway
```

| Параметр     | Тип    | По умолчанию | Описание                                                                                                                                |
|--------------|--------|--------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| `mode`       | string | `"dry-run"`  | Режим запуска. `dry-run` — только чтение и анализ. `full` — полная миграция. Также переопределяется CLI-флагом `--dry-run/--no-dry-run` |
| `batch_size` | int    | `100`        | Размер серверного курсора при чтении из PostgreSQL (`cursor.itersize`). Контролирует потребление памяти при больших объёмах данных      |
| `workers`    | int    | `4`          | Ограничение параллельности импорта через `asyncio.Semaphore`. Контролирует нагрузку на Auth Gateway API                                 |

### `source` — подключение к исходной БД

```yaml
source:
  type: "postgresql"              # postgresql | mysql
  host: "localhost"
  port: 5432
  database: "crm_production"
  user: "readonly_user"
  password: "${SOURCE_DB_PASSWORD}"
  ssl: false
  users_table: "users"
  id_strategy: "preserve_uuid"    # preserve_uuid | generate_new
```

| Параметр      | Тип    | По умолчанию      | Описание                                                                                         |
|---------------|--------|-------------------|--------------------------------------------------------------------------------------------------|
| `type`        | string | `"postgresql"`    | Тип СУБД. Определяет драйвер: `psycopg2` для PostgreSQL, `pymysql` для MySQL                     |
| `host`        | string | `"localhost"`     | Хост БД                                                                                          |
| `port`        | int    | `5432`            | Порт БД                                                                                          |
| `database`    | string | —                 | Имя базы данных                                                                                  |
| `user`        | string | —                 | Пользователь БД (рекомендуется readonly)                                                         |
| `password`    | string | —                 | Пароль. Поддерживает синтаксис `${ENV_VAR}` для подстановки из переменных окружения              |
| `ssl`         | bool   | `false`           | Включает SSL (`sslmode=require` для PostgreSQL, `ssl={"require": True}` для MySQL)               |
| `users_table` | string | `"users"`         | Имя таблицы пользователей в source БД                                                            |
| `id_strategy` | string | `"preserve_uuid"` | `preserve_uuid` — UUID сохраняются как есть. `generate_new` — Auth Gateway генерирует новые UUID |

#### `source.columns` — маппинг колонок

```yaml
  columns:
    id: "id"
    email: "email"
    username: "username"
    full_name: "full_name"
    phone: "phone"
    is_active: "is_active"
    email_verified: "email_verified"
    created_at: "created_at"
    # password_hash: "password_hash"  # Optional — omit if source uses OTP/Telegram-only auth
```

Маппинг внутренних полей на колонки source БД. **Ключ** — внутреннее имя поля, **значение** — имя колонки в таблице.
Используется для формирования SQL-запроса. Указывайте только те колонки, которые реально существуют в source-таблице.

**Поддерживаемые поля:**

| Ключ | Обязательный | Описание |
|------|-------------|----------|
| `id` | да | Идентификатор пользователя |
| `email` | нет | Email пользователя (не обязателен, если есть phone или username) |
| `username` | нет | Имя пользователя (если нет — генерируется из email) |
| `full_name` | нет | Полное имя |
| `phone` | нет | Телефон |
| `is_active` | нет | Активен ли аккаунт (default: `true`) |
| `email_verified` | нет | Подтверждён ли email (default: `false`) |
| `phone_verified` | нет | Подтверждён ли телефон (default: `false`) |
| `created_at` | нет | Дата создания (используется для дедупликации `keep_latest`) |
| `password_hash` | нет | Хеш пароля. Не указывайте, если source использует OTP/Telegram-авторизацию |

Пример для сервиса без паролей (OTP-only):

```yaml
  columns:
    id: "id"
    email: "email"
    phone: "phone_number"
    email_verified: "is_email_confirmed"
    created_at: "registered_at"
```

Пример для нестандартных имён колонок с паролями:

```yaml
  columns:
    id: "user_id"
    email: "email"
    password_hash: "pwd_bcrypt"
    full_name: "display_name"
    is_active: "active"
```

Сгенерирует: `SELECT pwd_bcrypt AS password_hash, display_name AS full_name, active AS is_active FROM users`.

### `target` — подключение к Auth Gateway

```yaml
target:
  base_url: "http://auth-gateway:3000"
  api_key: "${AUTH_GATEWAY_API_KEY}"
  application_id: "uuid-of-application"
```

| Параметр         | Тип    | Описание                                                                                                                         |
|------------------|--------|----------------------------------------------------------------------------------------------------------------------------------|
| `base_url`       | string | Адрес Auth Gateway REST API                                                                                                      |
| `api_key`        | string | API-ключ для аутентификации (поддерживает `${ENV_VAR}`). Передаётся в заголовке `X-API-Key`. Должен иметь scope для bulk-импорта |
| `application_id` | string | UUID приложения в Auth Gateway, к которому привязываются мигрируемые пользователи. Передаётся в заголовке `X-Application-ID`     |

### `password` — стратегия миграции паролей

```yaml
# Для сервисов с паролями:
password:
  strategy: "transfer"          # transfer | force_reset | none
  source_algorithm: "bcrypt"    # bcrypt | argon2 | scrypt | md5 | sha256

# Для сервисов без паролей (OTP, Telegram-бот и т.д.):
password:
  strategy: "none"
```

| Параметр           | Тип    | По умолчанию | Описание                                                                                                                                                                                        |
|--------------------|--------|--------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `strategy`         | string | `"transfer"` | `transfer` — хеши паролей передаются как есть через поле `password_hash_import`. `force_reset` — пароли не мигрируются, пользователи получат принудительный сброс. `none` — пароли игнорируются |
| `source_algorithm` | string | `"bcrypt"`   | Формат хешей в source БД. Автодетекция также выполняется при `analyze` по паттернам: `$2b$` → bcrypt, `$argon2` → argon2, hex длина 64 → sha256 и т.д. Не используется при `strategy: "none"`   |

> **Важно:** Если source-сервис использует только OTP или Telegram-авторизацию (без паролей), установите `strategy: "none"` и не включайте `password_hash` в секцию `source.columns`.

### `conflicts` — стратегия при коллизиях в Auth Gateway

```yaml
conflicts:
  strategy: "skip"    # skip | update | error
```

Поведение при обнаружении уже существующего пользователя **в Auth Gateway** (передаётся как `on_conflict` в payload к
API):

| Значение | Описание                                        |
|----------|-------------------------------------------------|
| `skip`   | Пропустить, оставить существующего пользователя |
| `update` | Обновить данные существующего пользователя      |
| `error`  | Завершить миграцию с ошибкой                    |

### `deduplication` — дедупликация source-данных

```yaml
deduplication:
  key: "email"              # email | phone | email_or_phone
  strategy: "keep_latest"   # keep_latest | keep_first | error
```

Дедупликация выполняется **до импорта**, на стороне Python, в памяти.

| Параметр   | Тип    | По умолчанию    | Описание                                                                                                                                                  |
|------------|--------|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| `key`      | string | `"email"`       | Поле для поиска дубликатов: `email`, `phone`, `username`, `email_or_phone`, `username_or_email`                                                           |
| `strategy` | string | `"keep_latest"` | `none` — пропустить дедупликацию. `keep_latest` — оставляет запись с более поздним `created_at`. `keep_first` — оставляет первую встреченную. `error` — бросает ошибку при первом дубликате |

### `validation` — настройки валидации

```yaml
validation:
  skip_email_validation: false
  skip_phone_validation: false
```

Позволяет отключить валидацию формата email или телефона для «грязных» данных из source-БД. Полезно, если в исходной
системе email-адреса не проходили строгую валидацию.

| Параметр                | Тип  | По умолчанию | Описание                                                          |
|-------------------------|------|--------------|-------------------------------------------------------------------|
| `skip_email_validation` | bool | `false`      | Пропускать валидацию формата email (позволяет импорт невалидных)  |
| `skip_phone_validation` | bool | `false`      | Пропускать валидацию формата телефона (E.164)                     |

### `roles` — миграция ролей

```yaml
roles:
  enabled: false
  source_column: "role_id"      # Колонка роли в таблице пользователей
  # source_table: "user_roles"  # M2M таблица (альтернативный вариант)
  # source_user_id_column: "user_id"
  # source_role_id_column: "role_id"
  mapping:
    1: "admin"
    2: "user"
    3: "manager"
```

Маппинг числовых ролей из source-системы в строковые имена ролей Auth Gateway. Поддерживает два варианта хранения ролей:

**Вариант 1: Колонка в таблице пользователей** (e.g. `users.role_id INT`):
- `source_column` — имя колонки с ID роли

**Вариант 2: M2M таблица** (e.g. `user_roles(user_id, role_id)`):
- `source_table` — имя таблицы связи
- `source_user_id_column` — колонка с ID пользователя
- `source_role_id_column` — колонка с ID роли

| Параметр                | Тип    | По умолчанию | Описание                                           |
|-------------------------|--------|--------------|-----------------------------------------------------|
| `enabled`               | bool   | `false`      | Включить миграцию ролей                             |
| `source_column`         | string | `"role_id"`  | Колонка роли в таблице пользователей                |
| `source_table`          | string | `""`         | M2M таблица (если указана, `source_column` игнорируется) |
| `source_user_id_column` | string | `"user_id"`  | Колонка user_id в M2M таблице                       |
| `source_role_id_column` | string | `"role_id"`  | Колонка role_id в M2M таблице                       |
| `mapping`               | map    | `{}`         | Маппинг `int → string` (source ID → AG role name)  |

### `shadow` — трансформация source-таблицы

```yaml
shadow:
  enabled: true
  drop_columns:
    - "password_hash"
    - "totp_secret"
    - "totp_enabled"
  add_columns:
    - name: "synced_at"
      type: "TIMESTAMP"
      default: "NOW()"
    - name: "display_name"
      type: "VARCHAR(255)"
      default: "NULL"
    - name: "avatar_url"
      type: "TEXT"
      default: "NULL"
```

Shadow-режим трансформирует **исходную** таблицу users после миграции. Идея: приложение-источник продолжает читать из
своей таблицы users (JOIN'ы и FK не ломаются), но аутентификация теперь полностью через Auth Gateway.

| Параметр       | Тип  | По умолчанию        | Описание                                                                                  |
|----------------|------|---------------------|-------------------------------------------------------------------------------------------|
| `enabled`      | bool | `true`              | Включение shadow-режима                                                                   |
| `drop_columns` | list | `["password_hash"]` | Колонки для удаления (чувствительные данные, которые теперь хранятся в Auth Gateway)      |
| `add_columns`  | list | `[]`                | Колонки для добавления: `name` — имя, `type` — SQL-тип, `default` — значение по умолчанию |

Генерируемый SQL:

- `ALTER TABLE ... DROP COLUMN IF EXISTS ...` для каждой drop-колонки
- `ALTER TABLE ... ADD COLUMN IF NOT EXISTS ...` для каждой add-колонки
- `UPDATE ... SET synced_at = NOW()` если добавлена колонка `synced_at`
- `COMMENT ON TABLE ...` с пометкой «Shadow table, READ-ONLY for auth data»

При экспорте в файл (`generate-shadow-sql`) SQL оборачивается в `BEGIN/COMMIT` транзакцию.

## Архитектурные решения

### Разделение ответственности: dedup vs conflicts

- **Deduplication** (`deduplicator.py`) — чистка данных **до** отправки в Auth Gateway. Работает в памяти Python,
  убирает дубликаты внутри source-данных
- **Conflicts** (`importer.py → on_conflict`) — разрешение коллизий **при** импорте. Обрабатывается на стороне Auth
  Gateway API, когда пользователь уже существует в целевой системе

### Экономия памяти

- PostgreSQL: server-side cursor (`cursor.itersize = batch_size`) — записи передаются порциями
- MySQL: `SSCursor` — аналогичный streaming-подход
- Импорт: батчи по 50 записей с ограничением параллельности через semaphore

### Безопасность

- Секреты (пароли БД, API-ключи) подставляются из переменных окружения через синтаксис `${ENV_VAR}`
- Рекомендуется `readonly_user` для подключения к source БД
- Dry-run по умолчанию — первый запуск никогда не модифицирует данные
- Live-режим запрашивает подтверждение через интерактивный `click.confirm`

## Зависимости

| Пакет             | Версия   | Назначение                                                        |
|-------------------|----------|-------------------------------------------------------------------|
| `click`           | >=8.1.0  | CLI-фреймворк                                                     |
| `pyyaml`          | >=6.0    | Парсинг YAML-конфигурации                                         |
| `psycopg2-binary` | >=2.9.0  | PostgreSQL драйвер                                                |
| `pymysql`         | >=1.1.0  | MySQL драйвер                                                     |
| `httpx`           | >=0.27.0 | Async HTTP-клиент для Auth Gateway API                            |
| `rich`            | >=13.0.0 | Форматированный вывод в терминал (прогресс-бары, таблицы, панели) |
