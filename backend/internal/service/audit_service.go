package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
)

type AuditService struct {
	auditRepo  *repository.AuditRepository
	geoService *GeoService
}

func NewAuditService(auditRepo *repository.AuditRepository, geoService *GeoService) *AuditService {
	return &AuditService{
		auditRepo:  auditRepo,
		geoService: geoService,
	}
}

type AuditLogParams struct {
	UserID    *uuid.UUID
	Action    models.AuditAction
	Status    models.AuditStatus
	IP        string
	UserAgent string
	Details   map[string]interface{}
}

func (s *AuditService) Log(params AuditLogParams) {
	go s.logAsync(params)
}

func (s *AuditService) logAsync(params AuditLogParams) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLog := s.buildAuditLog(params)

	if s.geoService != nil && params.IP != "" {
		location := s.geoService.GetLocation(ctx, params.IP)
		s.enrichWithGeoData(auditLog, location)
	}

	_ = s.auditRepo.Create(ctx, auditLog)
}

func (s *AuditService) LogSync(ctx context.Context, params AuditLogParams) error {
	auditLog := s.buildAuditLog(params)

	if s.geoService != nil && params.IP != "" {
		location := s.geoService.GetLocation(ctx, params.IP)
		s.enrichWithGeoData(auditLog, location)
	}

	return s.auditRepo.Create(ctx, auditLog)
}

func (s *AuditService) buildAuditLog(params AuditLogParams) *models.AuditLog {
	var detailsJSON []byte
	if params.Details != nil {
		detailsJSON, _ = json.Marshal(params.Details)
	}

	return &models.AuditLog{
		ID:        uuid.New(),
		UserID:    params.UserID,
		Action:    string(params.Action),
		IPAddress: params.IP,
		UserAgent: params.UserAgent,
		Status:    string(params.Status),
		Details:   detailsJSON,
		CreatedAt: time.Now(),
	}
}

func (s *AuditService) enrichWithGeoData(auditLog *models.AuditLog, location *models.GeoLocation) {
	if location == nil {
		return
	}

	auditLog.CountryCode = location.CountryCode
	auditLog.CountryName = location.CountryName
	auditLog.City = location.City
	auditLog.Latitude = location.Latitude
	auditLog.Longitude = location.Longitude
}

func (s *AuditService) LogWithAction(userID *uuid.UUID, action string, status string, ip, userAgent string, details map[string]interface{}) {
	params := AuditLogParams{
		UserID:    userID,
		Action:    models.AuditAction(action),
		Status:    models.AuditStatus(status),
		IP:        ip,
		UserAgent: userAgent,
		Details:   details,
	}
	s.Log(params)
}

func (s *AuditService) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	return s.auditRepo.GetByUserID(ctx, userID, limit, offset)
}

func (s *AuditService) List(ctx context.Context, limit, offset int) ([]*models.AuditLog, error) {
	return s.auditRepo.List(ctx, limit, offset)
}

func (s *AuditService) Count(ctx context.Context) (int, error) {
	return s.auditRepo.Count(ctx)
}

func (s *AuditService) CountByActionSince(ctx context.Context, action models.AuditAction, since time.Time) (int, error) {
	return s.auditRepo.CountByActionSince(ctx, action, since)
}

func (s *AuditService) DeleteOlderThan(ctx context.Context, days int) error {
	return s.auditRepo.DeleteOlderThan(ctx, days)
}
