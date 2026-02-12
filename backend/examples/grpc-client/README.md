# gRPC Client Example

Пример клиента для взаимодействия с Auth Gateway через gRPC.

## Использование

### 1. Запустить Auth Gateway

```bash
# В корне проекта
docker-compose up -d
# или
make docker-up
```

### 2. Получить JWT токен через REST API

```bash
# Регистрация
curl -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123",
    "full_name": "Test User"
  }'

# Сохраните access_token из ответа
```

### 3. Запустить пример клиента

```bash
cd examples/grpc-client
go run main.go
```

## Интеграция в другие сервисы

### Шаг 1: Скопировать proto файл

```bash
# Из корня auth-gateway
cp proto/auth.proto your-service/proto/
```

### Шаг 2: Сгенерировать код

```bash
# В вашем сервисе
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/auth.proto
```

### Шаг 3: Использовать в коде

```go
package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "your-service/api/proto"  // generated proto
)

func main() {
    // Подключиться к auth gateway
    conn, err := grpc.NewClient(
        "auth-gateway:50051",  // в Docker
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewAuthServiceClient(conn)

    // Проверить токен
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := client.ValidateToken(ctx, &pb.ValidateTokenRequest{
        AccessToken: "your-jwt-token",
    })
    if err != nil {
        log.Fatalf("Failed to validate token: %v", err)
    }

    if !resp.Valid {
        log.Printf("Token is invalid: %s", resp.ErrorMessage)
        return
    }

    log.Printf("Token is valid!")
    log.Printf("User ID: %s", resp.UserId)
    log.Printf("Email: %s", resp.Email)
    log.Printf("Role: %s", resp.Role)
}
```

## Middleware для gRPC сервисов

Пример middleware для автоматической проверки токенов:

```go
package middleware

import (
    "context"
    "strings"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"

    pb "your-service/api/proto"
)

type AuthInterceptor struct {
    authClient pb.AuthServiceClient
}

func NewAuthInterceptor(authClient pb.AuthServiceClient) *AuthInterceptor {
    return &AuthInterceptor{authClient: authClient}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        // Получить токен из metadata
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return nil, status.Error(codes.Unauthenticated, "missing metadata")
        }

        values := md["authorization"]
        if len(values) == 0 {
            return nil, status.Error(codes.Unauthenticated, "missing token")
        }

        token := strings.TrimPrefix(values[0], "Bearer ")

        // Проверить токен через auth gateway
        resp, err := a.authClient.ValidateToken(ctx, &pb.ValidateTokenRequest{
            AccessToken: token,
        })
        if err != nil {
            return nil, status.Error(codes.Internal, "auth service error")
        }

        if !resp.Valid {
            return nil, status.Error(codes.Unauthenticated, "invalid token")
        }

        // Добавить user info в context
        ctx = context.WithValue(ctx, "user_id", resp.UserId)
        ctx = context.WithValue(ctx, "user_role", resp.Role)
        ctx = context.WithValue(ctx, "user_email", resp.Email)

        return handler(ctx, req)
    }
}
```

## Методы gRPC API

### ValidateToken

Проверяет JWT токен и возвращает информацию о пользователе.

**Request:**
```json
{
  "access_token": "eyJhbGc..."
}
```

**Response:**
```json
{
  "valid": true,
  "user_id": "uuid",
  "email": "user@example.com",
  "username": "johndoe",
  "role": "user",
  "expires_at": 1234567890
}
```

### GetUser

Получить информацию о пользователе по ID.

**Request:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "johndoe",
    "full_name": "John Doe",
    "role": "user",
    "email_verified": true,
    "is_active": true
  }
}
```

### CheckPermission

Проверить права доступа пользователя.

**Request:**
```json
{
  "user_id": "uuid",
  "resource": "orders",
  "action": "read"
}
```

**Response:**
```json
{
  "allowed": true,
  "role": "user"
}
```

### IntrospectToken

Получить детальную информацию о токене.

**Request:**
```json
{
  "access_token": "eyJhbGc..."
}
```

**Response:**
```json
{
  "active": true,
  "user_id": "uuid",
  "email": "user@example.com",
  "username": "johndoe",
  "role": "user",
  "issued_at": 1234567890,
  "expires_at": 1234567890,
  "not_before": 1234567890,
  "blacklisted": false
}
```

## Production рекомендации

1. **TLS**: Используйте TLS для production
```go
creds, err := credentials.NewClientTLSFromFile("cert.pem", "")
conn, err := grpc.NewClient(
    "auth-gateway:50051",
    grpc.WithTransportCredentials(creds),
)
```

2. **Connection Pooling**: Переиспользуйте соединения
```go
// Создайте один раз при старте сервиса
var authClient pb.AuthServiceClient

func init() {
    conn, _ := grpc.NewClient(...)
    authClient = pb.NewAuthServiceClient(conn)
}
```

3. **Retry Policy**: Настройте повторные попытки
```go
conn, err := grpc.NewClient(
    "auth-gateway:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithDefaultServiceConfig(`{
        "methodConfig": [{
            "name": [{"service": "auth.AuthService"}],
            "retryPolicy": {
                "MaxAttempts": 3,
                "InitialBackoff": "0.1s",
                "MaxBackoff": "1s",
                "BackoffMultiplier": 2.0,
                "RetryableStatusCodes": ["UNAVAILABLE"]
            }
        }]
    }`),
)
```

4. **Health Checks**: Проверяйте доступность сервиса
```go
import "google.golang.org/grpc/health/grpc_health_v1"

healthClient := grpc_health_v1.NewHealthClient(conn)
resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
```

5. **Metrics**: Добавьте метрики для мониторинга
```go
import "github.com/grpc-ecosystem/go-grpc-prometheus"

conn, err := grpc.NewClient(
    "auth-gateway:50051",
    grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
)
```

## Troubleshooting

### Connection refused

Убедитесь что Auth Gateway запущен и gRPC порт открыт:
```bash
docker-compose ps
netstat -an | grep 50051
```

### Invalid token

Проверьте что:
- Токен получен из `/auth/signin` или `/auth/signup`
- Токен не истек (15 минут для access token)
- Токен не был revoked через `/auth/logout`

### Permission denied

Убедитесь что пользователь имеет необходимую роль (user/moderator/admin).
