package authgateway

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type defaultLogger struct{}

func (l *defaultLogger) Info(msg string, args ...interface{})  { log.Printf("[INFO] "+msg, args...) }
func (l *defaultLogger) Error(msg string, args ...interface{}) { log.Printf("[ERROR] "+msg, args...) }

const upsertUserSQL = `INSERT INTO users (id, email, username, full_name, is_active, synced_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    username = EXCLUDED.username,
    full_name = EXCLUDED.full_name,
    is_active = EXCLUDED.is_active,
    synced_at = NOW()`

const selectMaxSyncedAtSQL = `SELECT COALESCE(MAX(synced_at), '1970-01-01T00:00:00Z') FROM users`

const syncPageLimit int32 = 100

// --- WebhookEvent ---

type WebhookEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// --- WebhookSyncer ---

type WebhookSyncer struct {
	db     *sql.DB
	secret string
	logger Logger
}

func (ws *WebhookSyncer) GinWebhookHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		signature := c.GetHeader("X-Webhook-Signature")
		if !ws.verifySignature(body, signature) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}

		var event WebhookEvent
		if err := json.Unmarshal(body, &event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event payload"})
			return
		}

		ws.dispatchEvent(c.Request.Context(), event)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func (ws *WebhookSyncer) verifySignature(body []byte, signature string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	providedHex := strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(ws.secret))
	mac.Write(body)
	expectedHex := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedHex), []byte(providedHex))
}

func (ws *WebhookSyncer) dispatchEvent(ctx context.Context, event WebhookEvent) {
	switch event.Type {
	case "user.created":
		ws.handleUserCreated(ctx, event.Data)
	case "user.updated":
		ws.handleUserUpdated(ctx, event.Data)
	case "user.blocked":
		ws.handleUserDeactivated(ctx, event.Data)
	case "user.deleted":
		ws.handleUserDeactivated(ctx, event.Data)
	default:
		ws.logger.Info("unhandled webhook event type: %s", event.Type)
	}
}

func (ws *WebhookSyncer) handleUserCreated(ctx context.Context, data map[string]interface{}) {
	id := getStringField(data, "id")
	if id == "" {
		ws.logger.Error("webhook user.created: missing user id")
		return
	}

	_, err := ws.db.ExecContext(ctx, upsertUserSQL,
		id,
		getStringField(data, "email"),
		getStringField(data, "username"),
		getStringField(data, "full_name"),
		getBoolField(data, "is_active"),
	)
	if err != nil {
		ws.logger.Error("webhook user.created: upsert failed: %v", err)
		return
	}

	ws.logger.Info("webhook user.created: synced user %s", id)
}

func (ws *WebhookSyncer) handleUserUpdated(ctx context.Context, data map[string]interface{}) {
	id := getStringField(data, "id")
	if id == "" {
		ws.logger.Error("webhook user.updated: missing user id")
		return
	}

	_, err := ws.db.ExecContext(ctx, upsertUserSQL,
		id,
		getStringField(data, "email"),
		getStringField(data, "username"),
		getStringField(data, "full_name"),
		getBoolField(data, "is_active"),
	)
	if err != nil {
		ws.logger.Error("webhook user.updated: upsert failed: %v", err)
		return
	}

	ws.logger.Info("webhook user.updated: synced user %s", id)
}

func (ws *WebhookSyncer) handleUserDeactivated(ctx context.Context, data map[string]interface{}) {
	id := getStringField(data, "id")
	if id == "" {
		ws.logger.Error("webhook user.deactivated: missing user id")
		return
	}

	_, err := ws.db.ExecContext(ctx,
		`UPDATE users SET is_active = false, synced_at = NOW() WHERE id = $1`, id)
	if err != nil {
		ws.logger.Error("webhook user.deactivated: update failed: %v", err)
		return
	}

	ws.logger.Info("webhook user.deactivated: deactivated user %s", id)
}

// --- PeriodicSyncer ---

type PeriodicSyncer struct {
	client   *GRPCClient
	db       *sql.DB
	interval time.Duration
	logger   Logger
}

func (ps *PeriodicSyncer) Start(ctx context.Context) {
	ps.logger.Info("periodic sync started with interval %s", ps.interval)

	if err := ps.sync(ctx); err != nil {
		ps.logger.Error("initial periodic sync failed: %v", err)
	}

	ticker := time.NewTicker(ps.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ps.logger.Info("periodic sync stopped")
			return
		case <-ticker.C:
			if err := ps.sync(ctx); err != nil {
				ps.logger.Error("periodic sync failed: %v", err)
			}
		}
	}
}

func (ps *PeriodicSyncer) RunOnce(ctx context.Context) error {
	return ps.sync(ctx)
}

func (ps *PeriodicSyncer) sync(ctx context.Context) error {
	updatedAfter, err := ps.getLastSyncTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("get last sync timestamp: %w", err)
	}

	var totalSynced int
	var offset int32

	for {
		resp, err := ps.client.SyncUsers(ctx, updatedAfter, "", syncPageLimit, offset)
		if err != nil {
			return fmt.Errorf("sync users page (offset=%d): %w", offset, err)
		}

		for _, user := range resp.GetUsers() {
			if err := ps.upsertSyncUser(ctx, user); err != nil {
				ps.logger.Error("periodic sync: upsert user %s failed: %v", user.GetId(), err)
				continue
			}
			totalSynced++
		}

		if !resp.GetHasMore() {
			break
		}
		offset += syncPageLimit
	}

	ps.logger.Info("periodic sync completed: %d users synced", totalSynced)
	return nil
}

func (ps *PeriodicSyncer) getLastSyncTimestamp(ctx context.Context) (string, error) {
	var syncedAt string
	err := ps.db.QueryRowContext(ctx, selectMaxSyncedAtSQL).Scan(&syncedAt)
	if err != nil {
		return "1970-01-01T00:00:00Z", nil
	}
	return syncedAt, nil
}

type syncableUser interface {
	GetId() string
	GetEmail() string
	GetUsername() string
	GetFullName() string
	GetIsActive() bool
}

func (ps *PeriodicSyncer) upsertSyncUser(ctx context.Context, user syncableUser) error {
	_, err := ps.db.ExecContext(ctx, upsertUserSQL,
		user.GetId(),
		user.GetEmail(),
		user.GetUsername(),
		user.GetFullName(),
		user.GetIsActive(),
	)
	return err
}

// --- SyncManager ---

type SyncManager struct {
	LoginSyncer    *LoginSyncer
	WebhookSyncer  *WebhookSyncer
	PeriodicSyncer *PeriodicSyncer
}

type SyncConfig struct {
	WebhookSecret    string
	PeriodicInterval time.Duration
	SyncOnLogin      bool
	Logger           Logger
}

func NewSyncManager(client *GRPCClient, db *sql.DB, config SyncConfig) *SyncManager {
	logger := config.Logger
	if logger == nil {
		logger = &defaultLogger{}
	}

	manager := &SyncManager{}

	if config.SyncOnLogin {
		manager.LoginSyncer = &LoginSyncer{
			db:     db,
			logger: logger,
		}
	}

	if config.WebhookSecret != "" {
		manager.WebhookSyncer = &WebhookSyncer{
			db:     db,
			secret: config.WebhookSecret,
			logger: logger,
		}
	}

	if config.PeriodicInterval > 0 {
		manager.PeriodicSyncer = &PeriodicSyncer{
			client:   client,
			db:       db,
			interval: config.PeriodicInterval,
			logger:   logger,
		}
	}

	return manager
}

func (sm *SyncManager) WebhookHandler() gin.HandlerFunc {
	if sm.WebhookSyncer == nil {
		return func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "webhook sync not configured"})
		}
	}
	return sm.WebhookSyncer.GinWebhookHandler()
}

func (sm *SyncManager) Start(ctx context.Context) {
	if sm.PeriodicSyncer != nil {
		sm.PeriodicSyncer.Start(ctx)
	}
}

// --- Helpers ---

func getStringField(data map[string]interface{}, key string) string {
	val, ok := data[key]
	if !ok {
		return ""
	}
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}

func getBoolField(data map[string]interface{}, key string) bool {
	val, ok := data[key]
	if !ok {
		return false
	}
	b, ok := val.(bool)
	if !ok {
		return false
	}
	return b
}
