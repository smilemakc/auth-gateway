package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

type AdminAuditService struct {
	auditRepo AuditStore
}

func (s *AdminAuditService) ListAuditLogs(ctx context.Context, page, pageSize int, userID *uuid.UUID) (*models.AuditLogListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	var logs []*models.AuditLog
	var total int
	var err error

	if userID != nil {
		logs, err = s.auditRepo.GetByUserID(ctx, *userID, pageSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list audit logs: %w", err)
		}
		total = len(logs)
		if len(logs) == pageSize {
			total = page * pageSize
		}
	} else {
		logs, err = s.auditRepo.List(ctx, pageSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list audit logs: %w", err)
		}
		total, err = s.auditRepo.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count audit logs: %w", err)
		}
	}

	adminLogs := make([]*models.AdminAuditLogResponse, 0, len(logs))
	for _, log := range logs {
		var details map[string]interface{}
		if log.Details != nil {
			json.Unmarshal(log.Details, &details)
		}

		resp := &models.AdminAuditLogResponse{
			ID:        log.ID,
			UserID:    log.UserID,
			Action:    string(log.Action),
			Status:    string(log.Status),
			IP:        log.IPAddress,
			UserAgent: log.UserAgent,
			Details:   details,
			CreatedAt: log.CreatedAt,
		}

		if log.User != nil {
			resp.UserEmail = log.User.Email
		}

		adminLogs = append(adminLogs, resp)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &models.AuditLogListResponse{
		Logs:       adminLogs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
