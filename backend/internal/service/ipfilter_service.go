package service

import (
	"context"
	"fmt"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/utils"

	"github.com/google/uuid"
)

// IPFilterService handles IP filtering business logic
type IPFilterService struct {
	ipFilterRepo *repository.IPFilterRepository
}

// NewIPFilterService creates a new IP filter service
func NewIPFilterService(ipFilterRepo *repository.IPFilterRepository) *IPFilterService {
	return &IPFilterService{
		ipFilterRepo: ipFilterRepo,
	}
}

// CreateIPFilter creates a new IP filter
func (s *IPFilterService) CreateIPFilter(ctx context.Context, req *models.CreateIPFilterRequest, createdBy uuid.UUID) (*models.IPFilter, error) {
	// Validate IP/CIDR
	if !utils.ValidateIPOrCIDR(req.IPCIDR) {
		return nil, fmt.Errorf("invalid IP address or CIDR range: %s", req.IPCIDR)
	}

	filter := &models.IPFilter{
		IPCIDR:     req.IPCIDR,
		FilterType: req.FilterType,
		Reason:     req.Reason,
		CreatedBy:  &createdBy,
		IsActive:   true,
		ExpiresAt:  req.ExpiresAt,
	}

	err := s.ipFilterRepo.CreateIPFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	return filter, nil
}

// GetIPFilter retrieves an IP filter by ID
func (s *IPFilterService) GetIPFilter(ctx context.Context, id uuid.UUID) (*models.IPFilter, error) {
	return s.ipFilterRepo.GetIPFilterByID(ctx, id)
}

// ListIPFilters retrieves all IP filters with pagination
func (s *IPFilterService) ListIPFilters(ctx context.Context, page, perPage int, filterType string) (*models.IPFilterListResponse, error) {
	filters, total, err := s.ipFilterRepo.ListIPFilters(ctx, page, perPage, filterType)
	if err != nil {
		return nil, err
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.IPFilterListResponse{
		Filters:    filters,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// UpdateIPFilter updates an IP filter
func (s *IPFilterService) UpdateIPFilter(ctx context.Context, id uuid.UUID, req *models.UpdateIPFilterRequest) error {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	return s.ipFilterRepo.UpdateIPFilter(ctx, id, req.Reason, isActive)
}

// DeleteIPFilter deletes an IP filter
func (s *IPFilterService) DeleteIPFilter(ctx context.Context, id uuid.UUID) error {
	return s.ipFilterRepo.DeleteIPFilter(ctx, id)
}

// CheckIPAllowed checks if an IP address is allowed based on filters
func (s *IPFilterService) CheckIPAllowed(ctx context.Context, ipAddress string) (*models.CheckIPResponse, error) {
	// Get all active filters
	filters, err := s.ipFilterRepo.GetActiveIPFilters(ctx)
	if err != nil {
		return nil, err
	}

	// If no filters, allow by default
	if len(filters) == 0 {
		return &models.CheckIPResponse{
			Allowed: true,
		}, nil
	}

	// Check whitelist first
	whitelisted := false
	for _, filter := range filters {
		if filter.FilterType == "whitelist" {
			matches, err := utils.IPMatchesCIDR(ipAddress, filter.IPCIDR)
			if err != nil {
				continue
			}
			if matches {
				whitelisted = true
				break
			}
		}
	}

	// If there are whitelists and IP is not whitelisted, deny
	hasWhitelists := false
	for _, filter := range filters {
		if filter.FilterType == "whitelist" {
			hasWhitelists = true
			break
		}
	}

	if hasWhitelists && !whitelisted {
		return &models.CheckIPResponse{
			Allowed:    false,
			Reason:     "IP address not in whitelist",
			FilterType: "whitelist",
		}, nil
	}

	// Check blacklist
	for _, filter := range filters {
		if filter.FilterType == "blacklist" {
			matches, err := utils.IPMatchesCIDR(ipAddress, filter.IPCIDR)
			if err != nil {
				continue
			}
			if matches {
				return &models.CheckIPResponse{
					Allowed:    false,
					Reason:     filter.Reason,
					FilterType: "blacklist",
				}, nil
			}
		}
	}

	// If not blacklisted and either whitelisted or no whitelists exist, allow
	return &models.CheckIPResponse{
		Allowed: true,
	}, nil
}
