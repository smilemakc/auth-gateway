# Примеры интеграции для многосайтовых приложений

## Содержание
1. [Пример 1: React SPA с множественными сайтами](#пример-1-react-spa)
2. [Пример 2: Vue.js приложение](#пример-2-vuejs-приложение)
3. [Пример 3: Next.js с SSR](#пример-3-nextjs-с-ssr)
4. [Пример 4: Mobile интеграция (React Native)](#пример-4-react-native)
5. [Пример 5: Backend интеграция (Node.js)](#пример-5-nodejs-backend)

---

## Пример 1: React SPA

### Структура проекта

```
website-a/
├── src/
│   ├── hooks/
│   │   └── useAuth.js
│   ├── services/
│   │   ├── authService.js
│   │   └── oauthClient.js
│   ├── components/
│   │   ├── LoginPage.jsx
│   │   ├── OAuthCallback.jsx
│   │   └── ProtectedRoute.jsx
│   ├── pages/
│   │   ├── auth/
│   │   │   ├── signin.jsx
│   │   │   └── callback.jsx
│   │   ├── dashboard.jsx
│   │   └── profile.jsx
│   └── App.jsx
├── .env.local
└── vite.config.js
```

### Конфигурация

**`.env.local`:**
```
VITE_AUTH_GATEWAY_URL=https://auth.yourdomain.com
VITE_API_URL=http://localhost:3000
VITE_APP_NAME=Website A
```

### OAuth Client

**`src/services/oauthClient.js`:**

```javascript
/**
 * OAuth Client для Auth Gateway
 * Поддерживает множество провайдеров
 */

const AUTH_GATEWAY_URL = import.meta.env.VITE_AUTH_GATEWAY_URL;

// Хранить state для CSRF защиты
const OAUTH_STATE_KEY = 'oauth_state_';
const OAUTH_PROVIDER_KEY = 'oauth_provider_';

/**
 * Генерировать случайный state для CSRF защиты
 */
export function generateOAuthState() {
  return Math.random().toString(36).substring(2, 15) +
         Math.random().toString(36).substring(2, 15) +
         Math.random().toString(36).substring(2, 15);
}

/**
 * Начать OAuth поток
 * @param {string} provider - OAuth провайдер (google, github, yandex и т.д.)
 * @param {object} options - дополнительные опции
 */
export function initiateOAuthLogin(provider, options = {}) {
  // Генерировать state
  const state = generateOAuthState();

  // Сохранить в sessionStorage для проверки
  sessionStorage.setItem(`${OAUTH_STATE_KEY}${provider}`, state);
  sessionStorage.setItem(`${OAUTH_PROVIDER_KEY}${provider}`, provider);

  // Дополнительные параметры
  const params = new URLSearchParams({
    state,
    provider,
    ...options,
  });

  // Редирект на Auth Gateway
  window.location.href = `${AUTH_GATEWAY_URL}/auth/${provider}?${params.toString()}`;
}

/**
 * Получить доступный список провайдеров
 */
export async function getAvailableProviders() {
  try {
    const response = await fetch(
      `${AUTH_GATEWAY_URL}/auth/providers`,
      {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      throw new Error('Failed to fetch providers');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching providers:', error);
    return [];
  }
}

/**
 * Обработать OAuth callback
 * Вызывается после редиректа от Auth Gateway
 */
export function handleOAuthCallback() {
  const params = new URLSearchParams(window.location.search);

  const accessToken = params.get('access_token');
  const refreshToken = params.get('refresh_token');
  const isNewUser = params.get('is_new_user') === 'true';
  const state = params.get('state');
  const provider = params.get('provider') ||
                  sessionStorage.getItem(`${OAUTH_PROVIDER_KEY}${provider}`);

  // Проверить state для CSRF защиты
  const savedState = sessionStorage.getItem(`${OAUTH_STATE_KEY}${provider}`);

  if (!savedState || savedState !== state) {
    throw new Error('Invalid state parameter - possible CSRF attack');
  }

  // Очистить сохраненные значения
  sessionStorage.removeItem(`${OAUTH_STATE_KEY}${provider}`);
  sessionStorage.removeItem(`${OAUTH_PROVIDER_KEY}${provider}`);

  if (!accessToken) {
    throw new Error('No access token in callback');
  }

  // Сохранить токены
  const tokenData = {
    accessToken,
    refreshToken,
    provider,
    isNewUser,
    expiresAt: Date.now() + (15 * 60 * 1000), // 15 минут
  };

  // Используем sessionStorage для безопасности
  sessionStorage.setItem('auth_token_data', JSON.stringify(tokenData));

  // Также можно сохранить в более долгосрочное хранилище с refresh логикой
  if (refreshToken) {
    localStorage.setItem('refresh_token', refreshToken);
  }

  return tokenData;
}

/**
 * Получить текущий access token
 */
export function getAccessToken() {
  const data = sessionStorage.getItem('auth_token_data');
  if (!data) return null;

  const tokenData = JSON.parse(data);

  // Проверить, не истек ли токен
  if (tokenData.expiresAt < Date.now()) {
    sessionStorage.removeItem('auth_token_data');
    return null;
  }

  return tokenData.accessToken;
}

/**
 * Обновить access token используя refresh token
 */
export async function refreshAccessToken() {
  const refreshToken = localStorage.getItem('refresh_token');

  if (!refreshToken) {
    throw new Error('No refresh token available');
  }

  try {
    const response = await fetch(
      `${AUTH_GATEWAY_URL}/auth/refresh`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          refresh_token: refreshToken,
        }),
      }
    );

    if (!response.ok) {
      // Refresh token истек, нужна переавторизация
      logout();
      throw new Error('Refresh token expired');
    }

    const data = await response.json();

    // Обновить токены
    const tokenData = {
      accessToken: data.access_token,
      refreshToken: data.refresh_token || refreshToken,
      provider: JSON.parse(sessionStorage.getItem('auth_token_data') || '{}').provider,
      expiresAt: Date.now() + (15 * 60 * 1000),
    };

    sessionStorage.setItem('auth_token_data', JSON.stringify(tokenData));

    if (data.refresh_token) {
      localStorage.setItem('refresh_token', data.refresh_token);
    }

    return tokenData.accessToken;
  } catch (error) {
    console.error('Token refresh failed:', error);
    throw error;
  }
}

/**
 * Логаут
 */
export function logout() {
  sessionStorage.removeItem('auth_token_data');
  localStorage.removeItem('refresh_token');
}

/**
 * Проверить, авторизован ли пользователь
 */
export function isAuthenticated() {
  return getAccessToken() !== null;
}
```

### Auth Service

**`src/services/authService.js`:**

```javascript
/**
 * Сервис для работы с профилем и авторизацией
 */

import { getAccessToken, refreshAccessToken, logout } from './oauthClient';

const API_URL = import.meta.env.VITE_API_URL;

/**
 * Получить профиль текущего пользователя
 */
export async function getProfile() {
  let accessToken = getAccessToken();

  if (!accessToken) {
    throw new Error('Not authenticated');
  }

  try {
    const response = await fetch(
      `${API_URL}/auth/profile`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${accessToken}`,
          'Content-Type': 'application/json',
        },
      }
    );

    if (response.status === 401) {
      // Токен истек, попробовать обновить
      accessToken = await refreshAccessToken();
      return getProfile(); // Retry с новым токеном
    }

    if (!response.ok) {
      throw new Error(`Failed to fetch profile: ${response.statusText}`);
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching profile:', error);
    throw error;
  }
}

/**
 * Обновить профиль
 */
export async function updateProfile(data) {
  const accessToken = getAccessToken();

  if (!accessToken) {
    throw new Error('Not authenticated');
  }

  const response = await fetch(
    `${API_URL}/auth/profile`,
    {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    }
  );

  if (response.status === 401) {
    const newToken = await refreshAccessToken();
    return updateProfile(data); // Retry
  }

  if (!response.ok) {
    throw new Error('Failed to update profile');
  }

  return await response.json();
}

/**
 * Выполнить логаут на сервере
 */
export async function performLogout() {
  const accessToken = getAccessToken();

  if (!accessToken) {
    logout();
    return;
  }

  try {
    await fetch(
      `${API_URL}/auth/logout`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${accessToken}`,
        },
      }
    );
  } catch (error) {
    console.error('Server logout failed:', error);
  }

  // Локальный логаут в любом случае
  logout();
}
```

### Custom Hook

**`src/hooks/useAuth.js`:**

```javascript
import { useEffect, useState } from 'react';
import { getProfile, performLogout } from '../services/authService';
import { getAccessToken, isAuthenticated, refreshAccessToken } from '../services/oauthClient';

export function useAuth() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const loadUser = async () => {
      try {
        if (!isAuthenticated()) {
          setUser(null);
          return;
        }

        const profile = await getProfile();
        setUser(profile);
      } catch (err) {
        console.error('Failed to load user:', err);
        setError(err.message);
        setUser(null);
      } finally {
        setLoading(false);
      }
    };

    loadUser();

    // Обновлять токен каждые 10 минут
    const refreshInterval = setInterval(() => {
      if (isAuthenticated()) {
        refreshAccessToken().catch(() => {
          // Если refresh не удался, очистить
          setUser(null);
        });
      }
    }, 10 * 60 * 1000);

    return () => clearInterval(refreshInterval);
  }, []);

  const logout = async () => {
    await performLogout();
    setUser(null);
  };

  return {
    user,
    loading,
    error,
    isAuthenticated: !!user,
    logout,
  };
}
```

### Login Page

**`src/pages/auth/signin.jsx`:**

```jsx
import { useEffect, useState } from 'react';
import { getAvailableProviders, initiateOAuthLogin } from '../../services/oauthClient';

export function SignInPage() {
  const [providers, setProviders] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadProviders = async () => {
      try {
        const availableProviders = await getAvailableProviders();
        setProviders(availableProviders.filter(p => p.enabled));
      } catch (error) {
        console.error('Failed to load providers:', error);
      } finally {
        setLoading(false);
      }
    };

    loadProviders();
  }, []);

  const handleOAuthLogin = (provider) => {
    initiateOAuthLogin(provider);
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="signin-container">
      <h1>Sign In</h1>

      <div className="oauth-providers">
        {providers.map(provider => (
          <button
            key={provider.name}
            onClick={() => handleOAuthLogin(provider.name)}
            className="oauth-button"
          >
            <img src={provider.iconURL} alt={provider.displayName} />
            Sign in with {provider.displayName}
          </button>
        ))}
      </div>
    </div>
  );
}
```

### OAuth Callback Page

**`src/pages/auth/callback.jsx`:**

```jsx
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { handleOAuthCallback } from '../../services/oauthClient';

export function OAuthCallbackPage() {
  const navigate = useNavigate();
  const [error, setError] = useState(null);

  useEffect(() => {
    try {
      const tokenData = handleOAuthCallback();
      console.log('Successfully authenticated:', tokenData);

      // Перенаправить на dashboard
      navigate('/dashboard');
    } catch (err) {
      console.error('OAuth callback failed:', err);
      setError(err.message);

      // Перенаправить обратно на логин через 3 секунды
      setTimeout(() => {
        navigate('/auth/signin');
      }, 3000);
    }
  }, [navigate]);

  if (error) {
    return (
      <div className="error-container">
        <h1>Authentication Failed</h1>
        <p>{error}</p>
        <p>Redirecting to login...</p>
      </div>
    );
  }

  return (
    <div className="loading-container">
      <h1>Processing Authentication...</h1>
      <p>Please wait while we complete your sign in.</p>
    </div>
  );
}
```

### Protected Route

**`src/components/ProtectedRoute.jsx`:**

```jsx
import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export function ProtectedRoute({ children }) {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/auth/signin" replace />;
  }

  return children;
}
```

---

## Пример 2: Vue.js приложение

### Структура проекта

```
website-b/
├── src/
│   ├── composables/
│   │   └── useAuth.js
│   ├── services/
│   │   ├── authService.js
│   │   └── oauthClient.js
│   ├── components/
│   │   ├── LoginForm.vue
│   │   └── UserProfile.vue
│   ├── pages/
│   │   └── AuthCallback.vue
│   ├── App.vue
│   └── main.js
├── .env.local
└── vite.config.js
```

### OAuth Client для Vue

**`src/services/oauthClient.js`:**

```javascript
// Аналогично React версии, но с небольшой адаптацией

const AUTH_GATEWAY_URL = import.meta.env.VITE_AUTH_GATEWAY_URL;

export function initiateOAuthLogin(provider) {
  const state = generateOAuthState();
  sessionStorage.setItem(`oauth_state_${provider}`, state);

  window.location.href =
    `${AUTH_GATEWAY_URL}/auth/${provider}?state=${state}`;
}

export function handleOAuthCallback() {
  const params = new URLSearchParams(window.location.search);

  const accessToken = params.get('access_token');
  const refreshToken = params.get('refresh_token');
  const state = params.get('state');

  // Валидировать state
  const savedState = sessionStorage.getItem('oauth_state');
  if (savedState !== state) {
    throw new Error('Invalid state');
  }

  if (!accessToken) {
    throw new Error('No access token');
  }

  // Сохранить
  const tokenData = {
    accessToken,
    refreshToken,
    expiresAt: Date.now() + (15 * 60 * 1000),
  };

  localStorage.setItem('token_data', JSON.stringify(tokenData));
  sessionStorage.removeItem('oauth_state');

  return tokenData;
}

// ... остальные функции
```

### Composable Hook

**`src/composables/useAuth.js`:**

```javascript
import { ref, computed, onMounted } from 'vue';
import { getProfile, performLogout } from '../services/authService';
import { isAuthenticated, refreshAccessToken } from '../services/oauthClient';

export function useAuth() {
  const user = ref(null);
  const loading = ref(true);
  const error = ref(null);

  const isAuth = computed(() => !!user.value);

  const loadUser = async () => {
    try {
      if (!isAuthenticated()) {
        user.value = null;
        return;
      }

      user.value = await getProfile();
      error.value = null;
    } catch (err) {
      console.error('Failed to load user:', err);
      error.value = err.message;
      user.value = null;
    } finally {
      loading.value = false;
    }
  };

  const logout = async () => {
    await performLogout();
    user.value = null;
  };

  onMounted(() => {
    loadUser();

    // Refresh token every 10 minutes
    const interval = setInterval(() => {
      if (isAuth.value) {
        refreshAccessToken().catch(() => {
          user.value = null;
        });
      }
    }, 10 * 60 * 1000);

    return () => clearInterval(interval);
  });

  return {
    user,
    loading,
    error,
    isAuth,
    loadUser,
    logout,
  };
}
```

### Login Component

**`src/components/LoginForm.vue`:**

```vue
<template>
  <div class="login-container">
    <h1>{{ appName }} - Sign In</h1>

    <div v-if="loading" class="loading">
      Loading available providers...
    </div>

    <div v-else class="oauth-buttons">
      <button
        v-for="provider in providers"
        :key="provider.name"
        @click="loginWith(provider.name)"
        class="oauth-btn"
      >
        <img :src="provider.iconURL" :alt="provider.displayName" />
        {{ provider.displayName }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { initiateOAuthLogin, getAvailableProviders } from '../services/oauthClient';

const appName = import.meta.env.VITE_APP_NAME;
const providers = ref([]);
const loading = ref(true);

const loginWith = (provider) => {
  initiateOAuthLogin(provider);
};

onMounted(async () => {
  try {
    const available = await getAvailableProviders();
    providers.value = available.filter(p => p.enabled);
  } catch (error) {
    console.error('Failed to load providers:', error);
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.login-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.oauth-buttons {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.oauth-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 20px;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
}

.oauth-btn img {
  width: 20px;
  height: 20px;
}
</style>
```

### Callback Page

**`src/pages/AuthCallback.vue`:**

```vue
<template>
  <div class="callback-container">
    <div v-if="loading" class="loading">
      <h1>Processing authentication...</h1>
      <p>Please wait...</p>
    </div>

    <div v-else-if="error" class="error">
      <h1>Authentication Failed</h1>
      <p>{{ error }}</p>
      <p>Redirecting to login...</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { handleOAuthCallback } from '../services/oauthClient';

const router = useRouter();
const loading = ref(true);
const error = ref(null);

onMounted(async () => {
  try {
    handleOAuthCallback();
    // Успешно аутентифицированы
    await router.push('/dashboard');
  } catch (err) {
    error.value = err.message;
    // Редирект на логин через 3 сек
    setTimeout(() => {
      router.push('/auth/signin');
    }, 3000);
  } finally {
    loading.value = false;
  }
});
</script>
```

---

## Пример 3: Next.js с SSR

### Структура проекта

```
website-c/
├── pages/
│   ├── auth/
│   │   ├── signin.js
│   │   └── callback.js
│   ├── dashboard.js
│   └── api/
│       ├── auth/
│       │   ├── login.js
│       │   └── refresh.js
│       └── profile.js
├── lib/
│   ├── authClient.js
│   ├── oauthClient.js
│   └── withAuth.js
├── .env.local
└── next.config.js
```

### OAuth Client

**`lib/oauthClient.js`:**

```javascript
/**
 * OAuth Client для Next.js (работает на фронтенде)
 */

const AUTH_GATEWAY_URL = process.env.NEXT_PUBLIC_AUTH_GATEWAY_URL;

export function getOAuthURL(provider, options = {}) {
  const state = generateState();

  // Сохранить state в cookie
  setCookie('oauth_state', state, 600);

  const params = new URLSearchParams({
    state,
    ...options,
  });

  return `${AUTH_GATEWAY_URL}/auth/${provider}?${params.toString()}`;
}

export function handleOAuthCallback(query) {
  const { access_token, refresh_token, state } = query;

  // Проверить state
  const savedState = getCookie('oauth_state');
  if (savedState !== state) {
    throw new Error('Invalid state');
  }

  // Сохранить токены
  if (access_token) {
    setCookie('access_token', access_token, 900); // 15 минут
  }

  if (refresh_token) {
    setCookie('refresh_token', refresh_token, 7 * 24 * 60 * 60); // 7 дней
  }

  deleteCookie('oauth_state');
}

function generateState() {
  return Math.random().toString(36).substring(2, 15) +
         Math.random().toString(36).substring(2, 15);
}

function setCookie(name, value, maxAge) {
  const cookie = `${name}=${value}; path=/; max-age=${maxAge};`;

  if (typeof window !== 'undefined') {
    // Клиентская сторона
    document.cookie = cookie;
  }
}

function getCookie(name) {
  if (typeof window === 'undefined') return null;
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop().split(';').shift();
  return null;
}

function deleteCookie(name) {
  setCookie(name, '', -1);
}
```

### API Route для OAuth

**`pages/api/auth/oauth.js`:**

```javascript
/**
 * API route для обработки OAuth callback на серверной стороне
 * Более безопасно - токены не передаются через URL
 */

import { parse } from 'cookie';

export default async function handler(req, res) {
  if (req.method !== 'POST') {
    return res.status(405).json({ error: 'Method not allowed' });
  }

  const { code, provider, state } = req.body;

  // Проверить state
  const cookies = parse(req.headers.cookie || '');
  if (cookies.oauth_state !== state) {
    return res.status(400).json({ error: 'Invalid state' });
  }

  try {
    // Обменять code на token (на сервере, безопасно)
    const tokenResponse = await fetch(
      `${process.env.AUTH_GATEWAY_URL}/auth/${provider}/callback`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ code }),
      }
    );

    if (!tokenResponse.ok) {
      throw new Error('Token exchange failed');
    }

    const data = await tokenResponse.json();

    // Установить secure httpOnly cookies
    res.setHeader('Set-Cookie', [
      `access_token=${data.access_token}; Path=/; HttpOnly; Secure; Max-Age=900`,
      `refresh_token=${data.refresh_token}; Path=/; HttpOnly; Secure; Max-Age=${7 * 24 * 60 * 60}`,
      `oauth_state=; Path=/; Max-Age=-1`,
    ]);

    return res.status(200).json({
      success: true,
      isNewUser: data.is_new_user,
    });
  } catch (error) {
    console.error('OAuth error:', error);
    return res.status(500).json({ error: error.message });
  }
}
```

### Страница Callback

**`pages/auth/callback.js`:**

```javascript
import { useEffect, useState } from 'react';
import { useRouter } from 'next/router';

export default function OAuthCallback() {
  const router = useRouter();
  const { code, provider, state } = router.query;
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!code || !provider || !state) {
      return;
    }

    const handleCallback = async () => {
      try {
        // Отправить на backend API route
        const response = await fetch('/api/auth/oauth', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ code, provider, state }),
        });

        if (!response.ok) {
          throw new Error('Authentication failed');
        }

        const data = await response.json();

        // Успешно аутентифицированы
        if (data.isNewUser) {
          router.push('/onboarding');
        } else {
          router.push('/dashboard');
        }
      } catch (err) {
        setError(err.message);
        setLoading(false);

        // Редирект на логин через 3 сек
        setTimeout(() => {
          router.push('/auth/signin');
        }, 3000);
      }
    };

    handleCallback();
  }, [code, provider, state, router]);

  if (error) {
    return (
      <div className="error-container">
        <h1>Authentication Failed</h1>
        <p>{error}</p>
        <p>Redirecting to login...</p>
      </div>
    );
  }

  return (
    <div className="loading-container">
      <h1>Processing authentication...</h1>
      <p>Please wait while we complete your sign in.</p>
    </div>
  );
}
```

---

## Пример 4: React Native

### Конфигурация

**`App.js`:**

```javascript
import React, { useEffect } from 'react';
import * as AuthSession from 'expo-auth-session';
import * as WebBrowser from 'expo-web-browser';
import { NavigationContainer } from '@react-navigation/native';

WebBrowser.maybeCompleteAuthSession();

const discovery = {
  authorizationEndpoint: 'https://auth.yourdomain.com/oauth/authorize',
  tokenEndpoint: 'https://auth.yourdomain.com/oauth/token',
};

export default function App() {
  const [request, response, promptAsync] = AuthSession.useAuthRequest(
    {
      clientId: 'your-client-id',
      scopes: ['openid', 'profile', 'email'],
      redirectUri: AuthSession.getRedirectUrl(),
      usePKCE: true, // PKCE для дополнительной безопасности
    },
    discovery
  );

  useEffect(() => {
    if (response?.type === 'success') {
      const { code } = response.params;
      exchangeCodeForToken(code);
    }
  }, [response]);

  async function exchangeCodeForToken(code) {
    try {
      const response = await fetch(
        'https://auth.yourdomain.com/oauth/token',
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            grant_type: 'authorization_code',
            code,
            client_id: 'your-client-id',
            redirect_uri: AuthSession.getRedirectUrl(),
          }),
        }
      );

      const data = await response.json();

      // Сохранить токены
      await AsyncStorage.setItem('access_token', data.access_token);
      await AsyncStorage.setItem('refresh_token', data.refresh_token);

      // Перейти на главную
      // navigation.navigate('Home');
    } catch (error) {
      console.error('Token exchange failed:', error);
    }
  }

  return (
    <NavigationContainer>
      {/* Your app navigation */}
      <button
        disabled={!request}
        onPress={() => promptAsync()}
        title="Sign in with Auth Gateway"
      />
    </NavigationContainer>
  );
}
```

---

## Пример 5: Node.js Backend

### Интеграция для Backend приложения

**`auth-client.js`:**

```javascript
/**
 * Клиент для интеграции с Auth Gateway
 * Для использования в Backend приложениях
 */

const axios = require('axios');

class AuthGatewayClient {
  constructor(options = {}) {
    this.baseURL = options.baseURL || process.env.AUTH_GATEWAY_URL;
    this.clientID = options.clientID || process.env.AUTH_CLIENT_ID;
    this.clientSecret = options.clientSecret || process.env.AUTH_CLIENT_SECRET;
    this.cache = new Map();
  }

  /**
   * Валидировать JWT token (через gRPC или REST)
   */
  async validateToken(token) {
    const cacheKey = `token:${token}`;

    // Проверить кеш
    const cached = this.cache.get(cacheKey);
    if (cached && cached.expiresAt > Date.now()) {
      return cached.data;
    }

    try {
      const response = await axios.post(
        `${this.baseURL}/grpc/ValidateToken`,
        { token },
        {
          headers: {
            'X-API-Key': this.clientSecret,
          },
        }
      );

      // Кешировать на 5 минут
      this.cache.set(cacheKey, {
        data: response.data,
        expiresAt: Date.now() + (5 * 60 * 1000),
      });

      return response.data;
    } catch (error) {
      console.error('Token validation failed:', error);
      throw error;
    }
  }

  /**
   * Получить информацию о пользователе
   */
  async getUser(userID) {
    try {
      const response = await axios.get(
        `${this.baseURL}/users/${userID}`,
        {
          headers: {
            'X-API-Key': this.clientSecret,
          },
        }
      );

      return response.data;
    } catch (error) {
      console.error('Failed to get user:', error);
      throw error;
    }
  }

  /**
   * Проверить право доступа
   */
  async checkPermission(userID, permission) {
    try {
      const response = await axios.post(
        `${this.baseURL}/grpc/CheckPermission`,
        {
          user_id: userID,
          permission,
        },
        {
          headers: {
            'X-API-Key': this.clientSecret,
          },
        }
      );

      return response.data.allowed;
    } catch (error) {
      console.error('Permission check failed:', error);
      return false;
    }
  }

  /**
   * Создать API ключ для сервиса
   */
  async createServiceKey(name, scopes = []) {
    try {
      const response = await axios.post(
        `${this.baseURL}/admin/api-keys`,
        {
          name,
          scopes,
          type: 'service',
        },
        {
          headers: {
            'X-API-Key': this.clientSecret,
          },
        }
      );

      return response.data.api_key;
    } catch (error) {
      console.error('Failed to create service key:', error);
      throw error;
    }
  }
}

module.exports = AuthGatewayClient;
```

### Middleware для Express

**`auth-middleware.js`:**

```javascript
const AuthGatewayClient = require('./auth-client');

const authClient = new AuthGatewayClient();

/**
 * Middleware для проверки JWT токена
 */
function authMiddleware(req, res, next) {
  const token = extractToken(req);

  if (!token) {
    return res.status(401).json({ error: 'No token provided' });
  }

  authClient.validateToken(token)
    .then(validationResult => {
      if (!validationResult.valid) {
        return res.status(401).json({ error: 'Invalid token' });
      }

      // Сохранить информацию о пользователе в req
      req.user = {
        id: validationResult.user_id,
        email: validationResult.email,
        roles: validationResult.roles || [],
        scopes: validationResult.scopes || [],
      };

      next();
    })
    .catch(error => {
      console.error('Auth middleware error:', error);
      res.status(500).json({ error: 'Authorization failed' });
    });
}

/**
 * Middleware для проверки прав доступа
 */
function requirePermission(permission) {
  return async (req, res, next) => {
    const hasPermission = await authClient.checkPermission(
      req.user.id,
      permission
    );

    if (!hasPermission) {
      return res.status(403).json({ error: 'Insufficient permissions' });
    }

    next();
  };
}

function extractToken(req) {
  const authHeader = req.headers.authorization;

  if (!authHeader) {
    return null;
  }

  const parts = authHeader.split(' ');
  if (parts.length !== 2 || parts[0] !== 'Bearer') {
    return null;
  }

  return parts[1];
}

module.exports = {
  authMiddleware,
  requirePermission,
};
```

### Использование в маршрутах

**`routes/api.js`:**

```javascript
const express = require('express');
const { authMiddleware, requirePermission } = require('../auth-middleware');

const router = express.Router();

// Защищенный маршрут
router.get('/profile', authMiddleware, (req, res) => {
  res.json({
    id: req.user.id,
    email: req.user.email,
    roles: req.user.roles,
  });
});

// Маршрут требующий определенного разрешения
router.post(
  '/users',
  authMiddleware,
  requirePermission('users:write'),
  (req, res) => {
    // Создать пользователя
    res.json({ success: true });
  }
);

module.exports = router;
```

---

## Типичные ошибки и решения

### 1. CORS ошибки

**Проблема:** `Access to XMLHttpRequest blocked by CORS policy`

**Решение:**
```bash
# В Auth Gateway .env
CORS_ALLOWED_ORIGINS=https://site1.com,https://site2.com,http://localhost:3000

# Убедитесь что callback URL указан в OAuth провайдере
```

### 2. Истекший token

**Проблема:** 401 Unauthorized при каждом запросе

**Решение:**
```javascript
// Реализовать логику refresh token
async function apiCall(url, options) {
  let response = await fetch(url, {
    ...options,
    headers: {
      'Authorization': `Bearer ${getAccessToken()}`,
      ...options.headers,
    },
  });

  if (response.status === 401) {
    // Попробовать обновить токен
    await refreshAccessToken();

    // Retry с новым токеном
    response = await fetch(url, {
      ...options,
      headers: {
        'Authorization': `Bearer ${getAccessToken()}`,
        ...options.headers,
      },
    });
  }

  return response;
}
```

### 3. CSRF атаки

**Проблема:** Неавторизованное перенаправление от других сайтов

**Решение:**
```javascript
// Всегда проверять state
const savedState = sessionStorage.getItem('oauth_state');
if (savedState !== state) {
  throw new Error('Invalid state - possible CSRF attack');
}
```

---

## Чек-лист для интеграции

- [ ] Настроены environment переменные
- [ ] Auth Gateway доступен по HTTPS
- [ ] CORS origins добавлены в конфиг
- [ ] OAuth приложения зарегистрированы у провайдеров
- [ ] Callback URLs указаны в провайдерах
- [ ] Реализована обработка OAuth callback
- [ ] Реализована логика refresh token
- [ ] Проверяется state для CSRF защиты
- [ ] Токены хранятся безопасно (localStorage для длительных, sessionStorage для временных)
- [ ] Реализована выход из аккаунта
- [ ] Тестирована работа с разными браузерами/платформами
- [ ] Логирование и мониторинг включены
