package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
)

// WebhookService handles webhook operations
type WebhookService struct {
	repo         *repository.WebhookRepository
	auditService *AuditService
	httpClient   *http.Client
}

// NewWebhookService creates a new webhook service
func NewWebhookService(repo *repository.WebhookRepository, auditService *AuditService) *WebhookService {
	return &WebhookService{
		repo:         repo,
		auditService: auditService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateWebhook creates a new webhook
func (s *WebhookService) CreateWebhook(ctx context.Context, req *models.CreateWebhookRequest, createdBy uuid.UUID) (*models.Webhook, string, error) {
	// Validate events
	availableEvents := models.GetAvailableEvents()
	for _, event := range req.Events {
		valid := false
		for _, ae := range availableEvents {
			if event == ae {
				valid = true
				break
			}
		}
		if !valid {
			return nil, "", fmt.Errorf("invalid event type: %s", event)
		}
	}

	// Generate secret key
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate secret key: %w", err)
	}
	secretKey := hex.EncodeToString(secretBytes)

	// Convert events and headers to JSON
	eventsJSON, _ := json.Marshal(req.Events)
	headersJSON, _ := json.Marshal(req.Headers)

	// Set default retry config if not provided
	retryConfig := models.DefaultRetryConfig()
	if req.RetryConfig != nil {
		retryConfig = *req.RetryConfig
	}
	retryConfigJSON, _ := json.Marshal(retryConfig)

	webhook := &models.Webhook{
		Name:        req.Name,
		URL:         req.URL,
		SecretKey:   secretKey,
		Events:      eventsJSON,
		Headers:     headersJSON,
		IsActive:    true,
		RetryConfig: retryConfigJSON,
		CreatedBy:   &createdBy,
	}

	if err := s.repo.CreateWebhook(ctx, webhook); err != nil {
		return nil, "", err
	}

	// Log audit
	s.auditService.Log(AuditLogParams{
		UserID: &createdBy,
		Action: models.ActionCreate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceWebhook,
			"webhook_id":    webhook.ID.String(),
			"webhook_name":  webhook.Name,
		},
	})

	return webhook, secretKey, nil
}

// GetWebhook retrieves a webhook by ID
func (s *WebhookService) GetWebhook(ctx context.Context, id uuid.UUID) (*models.Webhook, error) {
	return s.repo.GetWebhookByID(ctx, id)
}

// ListWebhooks lists all webhooks with pagination
func (s *WebhookService) ListWebhooks(ctx context.Context, page, perPage int) (*models.WebhookListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	webhooks, total, err := s.repo.ListWebhooks(ctx, page, perPage)
	if err != nil {
		return nil, err
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.WebhookListResponse{
		Webhooks:   webhooks,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// UpdateWebhook updates a webhook
func (s *WebhookService) UpdateWebhook(ctx context.Context, id uuid.UUID, req *models.UpdateWebhookRequest, updatedBy uuid.UUID) error {
	// Get existing webhook
	webhook, err := s.repo.GetWebhookByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply updates
	name := webhook.Name
	if req.Name != "" {
		name = req.Name
	}

	url := webhook.URL
	if req.URL != "" {
		url = req.URL
	}

	events := webhook.Events
	if len(req.Events) > 0 {
		// Validate events
		availableEvents := models.GetAvailableEvents()
		for _, event := range req.Events {
			valid := false
			for _, ae := range availableEvents {
				if event == ae {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid event type: %s", event)
			}
		}
		events, _ = json.Marshal(req.Events)
	}

	headers := webhook.Headers
	if req.Headers != nil {
		headers, _ = json.Marshal(req.Headers)
	}

	isActive := webhook.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	if err := s.repo.UpdateWebhook(ctx, id, name, url, events, headers, isActive); err != nil {
		return err
	}

	// Log audit
	s.auditService.Log(AuditLogParams{
		UserID: &updatedBy,
		Action: models.ActionUpdate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceWebhook,
			"webhook_id":    id.String(),
			"webhook_name":  name,
		},
	})

	return nil
}

// DeleteWebhook deletes a webhook
func (s *WebhookService) DeleteWebhook(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	webhook, err := s.repo.GetWebhookByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteWebhook(ctx, id); err != nil {
		return err
	}

	// Log audit
	s.auditService.Log(AuditLogParams{
		UserID: &deletedBy,
		Action: models.ActionDelete,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceWebhook,
			"webhook_id":    id.String(),
			"webhook_name":  webhook.Name,
		},
	})

	return nil
}

// TriggerWebhook sends a webhook event to all subscribed endpoints
func (s *WebhookService) TriggerWebhook(ctx context.Context, eventType string, data map[string]interface{}) error {
	webhooks, err := s.repo.GetActiveWebhooksByEvent(ctx, eventType)
	if err != nil {
		return err
	}

	event := models.WebhookEvent{
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}

	for _, webhook := range webhooks {
		// Use context with timeout for webhook delivery
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		go func(w models.Webhook) {
			defer cancel()
			s.deliverWebhook(ctx, w, event)
		}(webhook)
	}

	return nil
}

// deliverWebhook delivers a webhook to a single endpoint
func (s *WebhookService) deliverWebhook(ctx context.Context, webhook models.Webhook, event models.WebhookEvent) {
	payload, _ := json.Marshal(event)

	// Create delivery record
	delivery := &models.WebhookDelivery{
		WebhookID: webhook.ID,
		EventType: event.EventType,
		Payload:   payload,
		Status:    "pending",
		Attempts:  0,
	}

	if err := s.repo.CreateWebhookDelivery(ctx, delivery); err != nil {
		return
	}

	// Create signature
	signature := s.createSignature(payload, webhook.SecretKey)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(payload))
	if err != nil {
		s.updateDeliveryFailed(ctx, delivery.ID, 0, err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", event.EventType)

	// Add custom headers
	var headers map[string]string
	if err := json.Unmarshal(webhook.Headers, &headers); err == nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.updateDeliveryFailed(ctx, delivery.ID, 0, err.Error())
		return
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.repo.UpdateDeliveryStatus(ctx, delivery.ID, "success", &resp.StatusCode, "", nil)
		s.repo.UpdateWebhookLastTriggered(ctx, webhook.ID)
	} else {
		s.updateDeliveryFailed(ctx, delivery.ID, resp.StatusCode, "non-2xx response")
	}
}

// updateDeliveryFailed updates a delivery as failed
func (s *WebhookService) updateDeliveryFailed(ctx context.Context, id uuid.UUID, httpStatus int, responseBody string) {
	var statusPtr *int
	if httpStatus > 0 {
		statusPtr = &httpStatus
	}
	s.repo.UpdateDeliveryStatus(ctx, id, "failed", statusPtr, responseBody, nil)
}

// createSignature creates HMAC-SHA256 signature
func (s *WebhookService) createSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// ListWebhookDeliveries lists deliveries for a webhook
func (s *WebhookService) ListWebhookDeliveries(ctx context.Context, webhookID uuid.UUID, page, perPage int) (*models.WebhookDeliveryListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	deliveries, total, err := s.repo.ListWebhookDeliveries(ctx, webhookID, page, perPage)
	if err != nil {
		return nil, err
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.WebhookDeliveryListResponse{
		Deliveries: deliveries,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// TestWebhook sends a test webhook
func (s *WebhookService) TestWebhook(ctx context.Context, id uuid.UUID, req *models.TestWebhookRequest) error {
	webhook, err := s.repo.GetWebhookByID(ctx, id)
	if err != nil {
		return err
	}

	event := models.WebhookEvent{
		EventType: req.EventType,
		Timestamp: time.Now().UTC(),
		Data:      req.Payload,
	}

	// Use context with timeout for test webhook delivery
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	go func() {
		defer cancel()
		s.deliverWebhook(ctx, *webhook, event)
	}()

	return nil
}

// GetAvailableEvents returns all available webhook events
func (s *WebhookService) GetAvailableEvents() []string {
	return models.GetAvailableEvents()
}

// ListWebhooksByApp retrieves webhooks for a specific application
func (s *WebhookService) ListWebhooksByApp(ctx context.Context, appID uuid.UUID) ([]*models.Webhook, error) {
	return s.repo.ListByApp(ctx, appID)
}
