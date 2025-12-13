# Руководство: OAuth авторизация для нескольких сайтов

## Содержание
1. [Обзор архитектур](#обзор-архитектур)
2. [Подход 1: Auth Gateway как Consumer](#подход-1-auth-gateway-как-consumer)
3. [Подход 2: Auth Gateway как Provider](#подход-2-auth-gateway-как-provider)
4. [Подход 3: Гибридный подход](#подход-3-гибридный-подход)
5. [Практические примеры](#практические-примеры)
6. [Управление несколькими приложениями](#управление-несколькими-приложениями)

---

## Обзор архитектур

### Вариант 1: Auth Gateway как OAuth Consumer (текущая реализация)

```
┌──────────────────────────────────────────────────────────────┐
│                    Внешние OAuth Провайдеры                  │
│  Google, Yandex, GitHub, Instagram, Telegram                 │
└────────────┬─────────────────────────────────────────────────┘
             │
             ▼
┌──────────────────────────────────────────────────────────────┐
│                     Auth Gateway                              │
│  (OAuth Consumer + Identity Provider)                         │
└─┬──────────────────────────────────┬──────────────────────┬──┘
  │                                  │                      │
  ▼                                  ▼                      ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│  Website A       │  │  Website B       │  │  Mobile App      │
│  (React SPA)     │  │  (Vue.js SPA)    │  │  (React Native)  │
│  :3001           │  │  :3002           │  │  :3003           │
└──────────────────┘  └──────────────────┘  └──────────────────┘
```

**Как это работает:**
- Auth Gateway потребляет OAuth от внешних провайдеров (Google, GitHub и т.д.)
- Несколько фронтенд приложений используют Auth Gateway для авторизации
- Auth Gateway создает своих пользователей на основе OAuth данных
- Все приложения используют JWT токены от Auth Gateway

### Вариант 2: Auth Gateway как OAuth Provider

```
┌──────────────────────────────────────────────────────────────┐
│                     Auth Gateway                              │
│  (OAuth 2.0 Authorization Server)                             │
└─┬──────────────────────────────────┬──────────────────────┬──┘
  │                                  │                      │
  ▼                                  ▼                      ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│  Website A       │  │  Website B       │  │  Third-party App │
│  (Client App)    │  │  (Client App)    │  │  (External)      │
└──────────────────┘  └──────────────────┘  └──────────────────┘

    "Login with Auth Gateway" (SSO для ваших сервисов)
```

**Как это работает:**
- Auth Gateway работает как OAuth 2.0 Authorization Server
- Другие приложения/сайты используют Auth Gateway для входа
- Auth Gateway выдает JWT или OAuth токены этим приложениям
- Идеально для SSO (Single Sign-On) внутри вашей экосистемы

### Вариант 3: Гибридный подход

```
        Google, GitHub, etc.
              │
              ▼
    ┌──────────────────────┐
    │  Auth Gateway        │
    │  - Потребляет OAuth  │
    │  - Предоставляет API │
    │  - Работает как IdP  │
    └──────────────────────┘
              │
    ┌─────────┼─────────┐
    │         │         │
    ▼         ▼         ▼
  Web-1    Web-2    Web-3
  (JWT)    (OAuth)  (OAuth)
```

---

## Подход 1: Auth Gateway как Consumer

### Текущая реализация в проекте

Auth Gateway уже настроен как OAuth Consumer для этих провайдеров:

1. **Google**
2. **Yandex**
3. **GitHub**
4. **Instagram**
5. **Telegram**

### Конфигурация для нескольких сайтов

#### Шаг 1: Настройка Environment переменных

```bash
# .env файл

# ===== ОСНОВНЫЕ НАСТРОЙКИ =====
AUTH_GATEWAY_URL=https://auth.yourdomain.com
JWT_ACCESS_SECRET=your-secret-key-256-bit
JWT_REFRESH_SECRET=your-refresh-secret-key-256-bit

# ===== GOOGLE OAUTH =====
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_CALLBACK_URL=https://auth.yourdomain.com/auth/google/callback

# ===== YANDEX OAUTH =====
YANDEX_CLIENT_ID=your-yandex-client-id
YANDEX_CLIENT_SECRET=your-yandex-client-secret
YANDEX_CALLBACK_URL=https://auth.yourdomain.com/auth/yandex/callback

# ===== GITHUB OAUTH =====
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GITHUB_CALLBACK_URL=https://auth.yourdomain.com/auth/github/callback

# ===== INSTAGRAM OAUTH =====
INSTAGRAM_CLIENT_ID=your-instagram-client-id
INSTAGRAM_CLIENT_SECRET=your-instagram-client-secret
INSTAGRAM_CALLBACK_URL=https://auth.yourdomain.com/auth/instagram/callback

# ===== TELEGRAM =====
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
TELEGRAM_CALLBACK_URL=https://auth.yourdomain.com/auth/telegram/callback

# ===== CORS для нескольких сайтов =====
CORS_ALLOWED_ORIGINS=https://site1.yourdomain.com,https://site2.yourdomain.com,https://mobile.yourdomain.com,http://localhost:3001,http://localhost:3002
```

#### Шаг 2: Регистрация OAuth приложений

**Для Google:**
1. Перейти на https://console.cloud.google.com
2. Создать новый проект
3. Включить Google+ API
4. Создать OAuth 2.0 credentials (Web application)
5. Добавить Authorized redirect URIs:
   - `https://auth.yourdomain.com/auth/google/callback`

**Для GitHub:**
1. Settings → Developer settings → OAuth Apps
2. New OAuth App
3. Authorization callback URL: `https://auth.yourdomain.com/auth/github/callback`

**Для Yandex:**
1. https://oauth.yandex.com/
2. Создать приложение
3. Разрешить для веб-приложений
4. Добавить Callback URL

**Аналогично для других провайдеров...**

#### Шаг 3: Настройка приложений для использования Auth Gateway

**Website A (React SPA на :3001):**

```javascript
// src/auth/oauthClient.js

const AUTH_GATEWAY_URL = 'https://auth.yourdomain.com';

export const initiateOAuthLogin = (provider, redirectUri = null) => {
  const state = generateRandomState();
  localStorage.setItem('oauth_state', state);

  const callbackUrl = redirectUri || window.location.origin + '/auth/callback';

  window.location.href =
    `${AUTH_GATEWAY_URL}/auth/${provider}?` +
    `redirect_uri=${encodeURIComponent(callbackUrl)}&` +
    `state=${state}`;
};

export const handleOAuthCallback = async () => {
  const params = new URLSearchParams(window.location.search);
  const accessToken = params.get('access_token');
  const refreshToken = params.get('refresh_token');
  const isNewUser = params.get('is_new_user');
  const state = params.get('state');

  // Проверить state для CSRF защиты
  if (state !== localStorage.getItem('oauth_state')) {
    throw new Error('Invalid state parameter');
  }

  // Сохранить токены
  localStorage.setItem('access_token', accessToken);
  localStorage.setItem('refresh_token', refreshToken);
  localStorage.setItem('is_new_user', isNewUser);

  localStorage.removeItem('oauth_state');
};

function generateRandomState() {
  return Math.random().toString(36).substring(2, 15) +
         Math.random().toString(36).substring(2, 15);
}
```

**Компонент Login:**

```jsx
// src/components/LoginPage.jsx

import React from 'react';
import { initiateOAuthLogin } from '../auth/oauthClient';

export function LoginPage() {
  return (
    <div className="login-container">
      <h1>Sign In</h1>

      <button onClick={() => initiateOAuthLogin('google')}>
        Sign in with Google
      </button>

      <button onClick={() => initiateOAuthLogin('github')}>
        Sign in with GitHub
      </button>

      <button onClick={() => initiateOAuthLogin('yandex')}>
        Sign in with Yandex
      </button>
    </div>
  );
}
```

**Website B (Vue.js SPA на :3002):**

```vue
<!-- src/components/LoginForm.vue -->

<template>
  <div class="login-form">
    <h1>Login</h1>

    <button @click="loginWithProvider('google')">
      Login with Google
    </button>

    <button @click="loginWithProvider('github')">
      Login with GitHub
    </button>
  </div>
</template>

<script>
import { initiateOAuthLogin, handleOAuthCallback } from '@/auth/oauthClient';

export default {
  methods: {
    loginWithProvider(provider) {
      initiateOAuthLogin(provider);
    }
  },
  mounted() {
    // Если вернулись с OAuth callback
    if (window.location.search.includes('access_token')) {
      handleOAuthCallback().then(() => {
        this.$router.push('/dashboard');
      });
    }
  }
};
</script>
```

#### Шаг 4: Обработка OAuth Flow

**Поток для Website A:**

```
1. Пользователь кликает "Sign in with Google"
   └─> website-a.com → обращается к Auth Gateway

2. Браузер редирект на Auth Gateway
   /auth/google?state=xyz

3. Auth Gateway редирект на Google
   https://accounts.google.com/o/oauth2/v2/auth?...

4. Пользователь авторизуется в Google

5. Google редирект обратно на Auth Gateway
   /auth/google/callback?code=abc&state=xyz

6. Auth Gateway обменивает code на token у Google

7. Auth Gateway создает/обновляет пользователя

8. Auth Gateway редирект на Website A с JWT токенами
   website-a.com/auth/callback?access_token=jwt&refresh_token=jwt&is_new_user=false

9. Website A сохраняет токены в localStorage/sessionStorage

10. Пользователь залогинен в Website A
```

### Обработка повторного входа

Каждый раз когда пользователь входит через OAuth:

```go
// internal/service/oauth_service.go - HandleCallback

// 1. Проверяем, существует ли OAuthAccount
oauthAccount, err := s.oauthRepo.GetOAuthAccount(provider, providerUserID)

if oauthAccount == nil {
    // 2. Новый пользователь - создаем
    user, err := s.createUserFromOAuth(userInfo)
    // Сохраняем OAuthAccount связь
} else {
    // 3. Существующий пользователь - обновляем токены
    oauthAccount.AccessToken = newAccessToken
    oauthAccount.RefreshToken = newRefreshToken
    s.oauthRepo.UpdateOAuthAccount(oauthAccount)
}

// 4. Генерируем JWT токены нашего сервиса
accessToken := s.jwtService.GenerateAccessToken(user)
refreshToken := s.jwtService.GenerateRefreshToken(user)

// 5. Возвращаем токены
```

---

## Подход 2: Auth Gateway как Provider

### Настройка Auth Gateway как OAuth 2.0 сервера

#### Архитектура

```
Auth Gateway
├── Users (в БД)
├── OAuth Apps (регистрированные приложения)
│   ├── app_id: website-a
│   ├── app_id: website-b
│   └── app_id: mobile-app
└── OAuth Tokens (выданные токены)

Website A → Auth Gateway: "Hello, я website-a"
Website A ← Auth Gateway: "OK, используй этот access_token"
```

#### Шаг 1: Расширить модель данных

**Создать миграцию для OAuth Apps:**

```sql
-- migrations/020_create_oauth_apps_table.up.sql

CREATE TABLE oauth_apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_name VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) UNIQUE NOT NULL,
    app_secret VARCHAR(255) NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id),

    -- Настройки редиректа
    redirect_uris TEXT[] NOT NULL,  -- массив разрешенных URIs

    -- Скопы
    allowed_scopes TEXT[] NOT NULL DEFAULT '{}',

    -- Статус
    is_active BOOLEAN DEFAULT true,

    -- Логирование
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_app_id_length CHECK (length(app_id) >= 5)
);

CREATE INDEX idx_oauth_apps_app_id ON oauth_apps(app_id);
CREATE INDEX idx_oauth_apps_owner_id ON oauth_apps(owner_id);
```

#### Шаг 2: Модель OAuth App

```go
// internal/models/oauth_app.go

package models

import (
    "database/sql/driver"
    "encoding/json"
    "time"
    "github.com/google/uuid"
)

type OAuthApp struct {
    ID             uuid.UUID `json:"id"`
    AppName        string    `json:"app_name"`
    AppID          string    `json:"app_id"`       // публичный идентификатор
    AppSecret      string    `json:"-"`            // никогда не возвращаем
    OwnerID        uuid.UUID `json:"owner_id"`
    RedirectURIs   []string  `json:"redirect_uris"`
    AllowedScopes  []string  `json:"allowed_scopes"`
    IsActive       bool      `json:"is_active"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}

// OAuth Authorization Request
type OAuthAuthorizationRequest struct {
    ClientID    string `form:"client_id" binding:"required"`
    RedirectURI string `form:"redirect_uri" binding:"required"`
    ResponseType string `form:"response_type" binding:"required,oneof=code"`
    Scope       string `form:"scope"`
    State       string `form:"state"`
}

// OAuth Token Request
type OAuthTokenRequest struct {
    GrantType   string `form:"grant_type" binding:"required"`
    Code        string `form:"code"`
    ClientID    string `form:"client_id" binding:"required"`
    ClientSecret string `form:"client_secret" binding:"required"`
    RedirectURI string `form:"redirect_uri" binding:"required"`
}

// Authorization Code (временный код)
type AuthorizationCode struct {
    Code        string
    ClientID    string
    UserID      uuid.UUID
    RedirectURI string
    Scope       string
    ExpiresAt   time.Time
}
```

#### Шаг 3: OAuth Repository

```go
// internal/repository/oauth_app_repository.go

package repository

import (
    "database/sql"
    "time"
    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/smilemakc/auth-gateway/internal/models"
)

type OAuthAppRepository struct {
    db *sqlx.DB
}

func NewOAuthAppRepository(db *sqlx.DB) *OAuthAppRepository {
    return &OAuthAppRepository{db: db}
}

// Создать новое OAuth приложение
func (r *OAuthAppRepository) Create(app *models.OAuthApp) error {
    query := `
        INSERT INTO oauth_apps
        (app_name, app_id, app_secret, owner_id, redirect_uris, allowed_scopes)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at
    `
    return r.db.QueryRowx(query,
        app.AppName,
        app.AppID,
        app.AppSecret,
        app.OwnerID,
        pq.Array(app.RedirectURIs),
        pq.Array(app.AllowedScopes),
    ).Scan(&app.ID, &app.CreatedAt, &app.UpdatedAt)
}

// Получить приложение по ID
func (r *OAuthAppRepository) GetByAppID(appID string) (*models.OAuthApp, error) {
    app := &models.OAuthApp{}
    query := `
        SELECT id, app_name, app_id, app_secret, owner_id,
               redirect_uris, allowed_scopes, is_active, created_at, updated_at
        FROM oauth_apps
        WHERE app_id = $1 AND is_active = true
    `
    err := r.db.QueryRowx(query, appID).StructScan(app)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return app, err
}

// Список приложений пользователя
func (r *OAuthAppRepository) ListByOwnerID(ownerID uuid.UUID) ([]models.OAuthApp, error) {
    var apps []models.OAuthApp
    query := `
        SELECT id, app_name, app_id, app_secret, owner_id,
               redirect_uris, allowed_scopes, is_active, created_at, updated_at
        FROM oauth_apps
        WHERE owner_id = $1
        ORDER BY created_at DESC
    `
    err := r.db.Select(&apps, query, ownerID)
    return apps, err
}

// Удалить приложение
func (r *OAuthAppRepository) Delete(appID string, ownerID uuid.UUID) error {
    query := `DELETE FROM oauth_apps WHERE app_id = $1 AND owner_id = $2`
    _, err := r.db.Exec(query, appID, ownerID)
    return err
}
```

#### Шаг 4: OAuth Provider Service

```go
// internal/service/oauth_provider_service.go

package service

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "time"
    "github.com/google/uuid"
    "github.com/smilemakc/auth-gateway/internal/models"
    "github.com/smilemakc/auth-gateway/internal/repository"
)

type OAuthProviderService struct {
    appRepo       *repository.OAuthAppRepository
    jwtService    *JWTService
    tokenRepo     *repository.TokenRepository
}

// Генерировать authorization code
func (s *OAuthProviderService) GenerateAuthorizationCode(
    ctx context.Context,
    clientID string,
    userID uuid.UUID,
    redirectURI string,
    scope string,
) (string, error) {
    // Проверить приложение
    app, err := s.appRepo.GetByAppID(clientID)
    if err != nil || app == nil {
        return "", fmt.Errorf("invalid client_id")
    }

    // Проверить redirect_uri
    if !isValidRedirectURI(app.RedirectURIs, redirectURI) {
        return "", fmt.Errorf("invalid redirect_uri")
    }

    // Генерировать код
    code := generateRandomCode()

    // Сохранить в Redis с TTL (10 минут)
    // Или в отдельной таблице с TTL
    // authCodeKey := fmt.Sprintf("auth_code:%s", code)
    // redisClient.Set(authCodeKey, map{
    //     "client_id": clientID,
    //     "user_id": userID,
    //     "redirect_uri": redirectURI,
    //     "scope": scope,
    // }, 10*time.Minute)

    return code, nil
}

// Обменять authorization code на token
func (s *OAuthProviderService) ExchangeCodeForToken(
    ctx context.Context,
    clientID string,
    clientSecret string,
    code string,
    redirectURI string,
) (*OAuthTokenResponse, error) {
    // Проверить приложение по client_id и client_secret
    app, err := s.appRepo.GetByAppID(clientID)
    if err != nil || app == nil {
        return nil, fmt.Errorf("invalid client")
    }

    if app.AppSecret != clientSecret {
        return nil, fmt.Errorf("invalid client secret")
    }

    // Получить код из хранилища
    // authCodeData := redisClient.Get(fmt.Sprintf("auth_code:%s", code))
    //
    // Проверить код и redirect_uri
    // if authCodeData.RedirectURI != redirectURI {
    //     return nil, fmt.Errorf("invalid redirect_uri")
    // }

    // Для примера - создаем токен
    userID := uuid.New() // В реальности берем из сохраненного кода

    // Генерировать JWT токены
    accessToken := s.generateAccessToken(userID, app.ID, app.AllowedScopes)
    refreshToken := s.generateRefreshToken(userID, app.ID)

    return &OAuthTokenResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        TokenType:    "Bearer",
        ExpiresIn:    900, // 15 минут
    }, nil
}

// Генерировать access token для OAuth клиента
func (s *OAuthProviderService) generateAccessToken(
    userID uuid.UUID,
    appID uuid.UUID,
    scopes []string,
) string {
    claims := &models.CustomClaims{
        UserID:  userID,
        AppID:   appID,
        Scopes:  scopes,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
            IssuedAt:  time.Now().Unix(),
        },
    }
    token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
        SignedString([]byte(os.Getenv("JWT_ACCESS_SECRET")))
    return token
}

func generateRandomCode() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}

func isValidRedirectURI(allowedURIs []string, uri string) bool {
    for _, allowed := range allowedURIs {
        if allowed == uri {
            return true
        }
    }
    return false
}

type OAuthTokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token,omitempty"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"`
}
```

#### Шаг 5: Handler для OAuth Provider

```go
// internal/handler/oauth_provider_handler.go

package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/smilemakc/auth-gateway/internal/models"
    "github.com/smilemakc/auth-gateway/internal/service"
)

type OAuthProviderHandler struct {
    oauthProviderService *service.OAuthProviderService
    authService          *service.AuthService
}

// GET /oauth/authorize
// Инициировать OAuth flow
func (h *OAuthProviderHandler) Authorize(c *gin.Context) {
    var req models.OAuthAuthorizationRequest

    if err := c.ShouldBindQuery(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    // Получить текущего пользователя (должен быть залогинен)
    userID, exists := c.Get("user_id")
    if !exists {
        // Редирект на логин
        c.Redirect(http.StatusTemporaryRedirect,
            "/auth/signin?next=/oauth/authorize?...&client_id=...&...")
        return
    }

    // Генерировать authorization code
    code, err := h.oauthProviderService.GenerateAuthorizationCode(
        c.Request.Context(),
        req.ClientID,
        userID.(uuid.UUID),
        req.RedirectURI,
        req.Scope,
    )
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Редирект на приложение с кодом
    redirectURL := fmt.Sprintf("%s?code=%s&state=%s",
        req.RedirectURI, code, req.State)
    c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// POST /oauth/token
// Обменять code на token
func (h *OAuthProviderHandler) Token(c *gin.Context) {
    var req models.OAuthTokenRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    // Обменять code на token
    tokenResp, err := h.oauthProviderService.ExchangeCodeForToken(
        c.Request.Context(),
        req.ClientID,
        req.ClientSecret,
        req.Code,
        req.RedirectURI,
    )
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, tokenResp)
}

// GET /oauth/userinfo
// Получить информацию о текущем пользователе
func (h *OAuthProviderHandler) UserInfo(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Получить пользователя
    user, err := h.userService.GetByID(userID.(uuid.UUID))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":       user.ID,
        "email":    user.Email,
        "username": user.Username,
        "name":     user.FullName,
    })
}
```

#### Шаг 6: Регистрация OAuth приложения

**Endpoint для создания приложения:**

```go
// POST /admin/oauth-apps
func (h *OAuthProviderHandler) CreateApp(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var req struct {
        AppName      string   `json:"app_name" binding:"required"`
        RedirectURIs []string `json:"redirect_uris" binding:"required"`
        AllowedScopes []string `json:"allowed_scopes"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Генерировать app_id и app_secret
    appID := generateUniqueAppID()
    appSecret := generateRandomSecret()

    app := &models.OAuthApp{
        AppName:       req.AppName,
        AppID:         appID,
        AppSecret:     appSecret, // Хешируем перед сохранением
        OwnerID:       userID.(uuid.UUID),
        RedirectURIs: req.RedirectURIs,
        AllowedScopes: req.AllowedScopes,
        IsActive:      true,
    }

    if err := h.oauthAppRepo.Create(app); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Возвращаем secret только один раз!
    c.JSON(http.StatusCreated, gin.H{
        "app_id":     app.AppID,
        "app_secret": appSecret, // Только один раз!
        "client_credentials": gin.H{
            "client_id":     app.AppID,
            "client_secret": appSecret,
        },
    })
}
```

---

## Подход 3: Гибридный подход

Используйте **оба подхода одновременно**:

```
┌─────────────────────────────────────────────────┐
│              Auth Gateway                        │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │ OAuth Consumer (External Providers)       │  │
│  │ - Google, GitHub, Yandex                 │  │
│  │ - Потребляет их OAuth                    │  │
│  └──────────────────────────────────────────┘  │
│                    │                             │
│                    ▼                             │
│  ┌──────────────────────────────────────────┐  │
│  │ Internal User Database                    │  │
│  │ - Единая база для всех пользователей    │  │
│  └──────────────────────────────────────────┘  │
│                    │                             │
│                    ▼                             │
│  ┌──────────────────────────────────────────┐  │
│  │ OAuth Provider (для ваших приложений)    │  │
│  │ - Выдает токены вашим приложениям        │  │
│  │ - SSO для вашей экосистемы               │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
        │                         │
        ▼                         ▼
┌─────────────────────────┐  ┌─────────────────────────┐
│ Website A               │  │ Website B               │
│ (потребляет OAuth от AG)│  │ (потребляет OAuth от AG)│
└─────────────────────────┘  └─────────────────────────┘
```

**Преимущества:**
- ✅ Единая база пользователей
- ✅ Поддержка множественных внешних OAuth провайдеров
- ✅ Своя OAuth система для внутренних приложений
- ✅ SSO для вашей экосистемы

---

## Практические примеры

### Пример 1: Две веб-приложения используют одного Auth Gateway

**Сценарий:** У вас есть:
- Website A: e-commerce платформа на React
- Website B: админ панель на Vue.js
- Обе используют один Auth Gateway

**Решение:**

```bash
# .env в Auth Gateway

GOOGLE_CLIENT_ID=your-id
GOOGLE_CLIENT_SECRET=your-secret
GOOGLE_CALLBACK_URL=https://auth.yourdomain.com/auth/google/callback

# CORS для обоих сайтов
CORS_ALLOWED_ORIGINS=https://shop.yourdomain.com,https://admin.yourdomain.com,http://localhost:3001,http://localhost:3002
```

**Website A (React):**

```javascript
// .env.local
VITE_AUTH_GATEWAY_URL=https://auth.yourdomain.com

// src/utils/auth.js
export const getGoogleLoginURL = () => {
  const redirectUri = `${window.location.origin}/auth/callback`;
  return `${import.meta.env.VITE_AUTH_GATEWAY_URL}/auth/google?redirect_uri=${encodeURIComponent(redirectUri)}`;
};
```

**Website B (Vue):**

```javascript
// .env.local
VITE_AUTH_GATEWAY_URL=https://auth.yourdomain.com

// src/utils/auth.js
export const getGoogleLoginURL = () => {
  const redirectUri = `${window.location.origin}/auth/callback`;
  return `${process.env.VUE_APP_AUTH_GATEWAY_URL}/auth/google?redirect_uri=${encodeURIComponent(redirectUri)}`;
};
```

### Пример 2: Mobile App + Web App

**Сценарий:**
- Website: веб версия
- Mobile App: React Native или Flutter

**Mobile App (React Native):**

```javascript
// src/screens/LoginScreen.js

import { useEffect } from 'react';
import * as WebBrowser from 'expo-web-browser';
import * as AuthSession from 'expo-auth-session';

const useProxy = Platform.select({ web: false, default: true });

const discovery = {
  authorizationEndpoint: 'https://auth.yourdomain.com/auth/google',
  tokenEndpoint: 'https://auth.yourdomain.com/oauth/token',
  revocationEndpoint: 'https://auth.yourdomain.com/oauth/revoke',
};

export function LoginScreen() {
  const [request, response, promptAsync] = AuthSession.useAuthRequest(
    {
      clientId: 'your-client-id',
      scopes: ['openid', 'profile', 'email'],
      redirectUri: AuthSession.getRedirectUrl(),
    },
    discovery,
  );

  useEffect(() => {
    if (response?.type === 'success') {
      const { code } = response.params;
      // Обменять code на token
      exchangeCodeForToken(code);
    }
  }, [response]);

  return (
    <button
      disabled={!request}
      onPress={() => promptAsync()}
      title="Login with Google"
    />
  );
}
```

### Пример 3: Third-party интеграция

**Сценарий:** Вы хотите, чтобы внешние разработчики интегрировали "Login with Your Platform"

**Для внешних разработчиков:**

1. **Получить credentials:**
```bash
POST https://auth.yourdomain.com/admin/oauth-apps
{
  "app_name": "Third Party App",
  "redirect_uris": ["https://thirdparty.com/callback"],
  "allowed_scopes": ["users:read"]
}

Response:
{
  "app_id": "client_abc123",
  "app_secret": "secret_xyz789"
}
```

2. **Документация для разработчика:**

```markdown
# "Login with YourBrand" OAuth Integration

## Endpoints

- **Authorization:** `https://auth.yourdomain.com/oauth/authorize`
- **Token:** `https://auth.yourdomain.com/oauth/token`
- **User Info:** `https://auth.yourdomain.com/oauth/userinfo`

## Flow

1. Редирект на:
```
https://auth.yourdomain.com/oauth/authorize?
  client_id=YOUR_CLIENT_ID&
  redirect_uri=YOUR_REDIRECT_URI&
  response_type=code&
  scope=openid+profile+email
```

2. Обменять code на token:
```
POST https://auth.yourdomain.com/oauth/token
{
  "grant_type": "authorization_code",
  "code": "AUTH_CODE",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "redirect_uri": "YOUR_REDIRECT_URI"
}
```

3. Получить user info:
```
GET https://auth.yourdomain.com/oauth/userinfo
Authorization: Bearer ACCESS_TOKEN
```
```

---

## Управление несколькими приложениями

### Таблица совместимости методов

| Компонент | Метод | Приложение 1 | Приложение 2 | Mobile | Third-party |
|-----------|-------|--------------|--------------|--------|-------------|
| **Authentication** | Email/Pass | ✅ | ✅ | ✅ | ❌ |
| | Google OAuth | ✅ | ✅ | ✅ | ✅ |
| | GitHub OAuth | ✅ | ✅ | ✅ | ✅ |
| **Tokens** | JWT | ✅ | ✅ | ✅ | ✅ |
| | API Keys | ✅ | ✅ | ❌ | ✅ |
| **Sessions** | Tracking | ✅ | ✅ | ✅ | ✅ |
| | Per-device | ✅ | ✅ | ✅ | ✅ |

### Dashboard для управления приложениями

```go
// GET /admin/applications
// Получить список всех регистрированных приложений

Response:
{
  "applications": [
    {
      "id": "app1",
      "name": "Website A",
      "type": "web",
      "oauth_consumers": ["google", "github"],
      "users_count": 1234,
      "last_active": "2024-01-15T10:00:00Z"
    },
    {
      "id": "app2",
      "name": "Website B",
      "type": "web",
      "oauth_consumers": ["google", "yandex"],
      "users_count": 567,
      "last_active": "2024-01-15T09:00:00Z"
    }
  ]
}
```

### Синхронизация данных между приложениями

Если пользователь залогинен в приложении A и хочет перейти в приложение B:

```javascript
// Website A
const user = JSON.parse(localStorage.getItem('user'));
const accessToken = localStorage.getItem('access_token');

// Отправить на Website B
window.open(
  `https://app-b.com/auth/sso?token=${accessToken}&app=website-a`,
  '_blank'
);

// Website B
const token = new URLSearchParams(window.location.search).get('token');
const app = new URLSearchParams(window.location.search).get('app');

// Проверить токен у Auth Gateway
fetch('https://auth.yourdomain.com/auth/profile', {
  headers: { Authorization: `Bearer ${token}` }
})
  .then(r => r.json())
  .then(user => {
    // Пользователь уже авторизован!
    localStorage.setItem('access_token', token);
  });
```

---

## Безопасность при управлении несколькими приложениями

### CORS политика

```go
// config/cors.go

func GetCORSConfig() cors.Config {
    return cors.Config{
        AllowOrigins:     []string{
            "https://app-a.yourdomain.com",
            "https://app-b.yourdomain.com",
            "https://mobile.yourdomain.com",
        },
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Authorization", "Content-Type"},
        ExposeHeaders:    []string{"X-Total-Count"},
        AllowCredentials: true,
    }
}
```

### Изоляция данных пользователя

```go
// middleware/auth.go

func (m *AuthMiddleware) ensureUserDataIsolation(c *gin.Context) {
    userID := c.GetString("user_id")

    // Если пытаемся получить чужого пользователя
    if paramID := c.Param("user_id"); paramID != "" && paramID != userID {
        // Разрешить только если пользователь админ
        roles := c.GetStringSlice("roles")
        if !contains(roles, "admin") {
            c.JSON(403, gin.H{"error": "Forbidden"})
            c.Abort()
            return
        }
    }
}
```

### Rate limiting per app

```go
// middleware/rate_limit.go

func RateLimitPerApp(c *gin.Context) {
    appID := c.GetString("app_id") // Из JWT claim
    clientIP := c.ClientIP()
    key := fmt.Sprintf("rate_limit:%s:%s", appID, clientIP)

    // Redis rate limiting
    limit := 100 // 100 запросов в минуту на приложение
    // ...
}
```

---

## Заключение

### Когда использовать какой подход?

| Подход | Когда использовать |
|--------|-------------------|
| **Auth Gateway Consumer** | Несколько своих приложений используют внешние OAuth провайдеры (Google, GitHub). Уже реализовано. |
| **Auth Gateway Provider** | Вы хотите, чтобы другие приложения/сайты логинились через вас. Как "Login with Facebook". |
| **Гибридный** | Комбинация обоих. Ваши приложения используют внешние OAuth, и одновременно вы провайдер для третьих лиц. |

### Чек-лист для многосайтовой архитектуры

- [ ] Настроены все необходимые CORS origins
- [ ] Каждый OAuth провайдер имеет callback URL для Auth Gateway
- [ ] Все приложения настроены на использование JWT токенов
- [ ] Реализована синхронизация сессий между приложениями
- [ ] Проверена изоляция данных пользователей
- [ ] Настроены rate limits per application
- [ ] Документирован процесс регистрации нового приложения
- [ ] Есть мониторинг cross-site attacks (CSRF, XSS)
- [ ] Все токены передаются по HTTPS
- [ ] Настроено логирование всех cross-site операций
