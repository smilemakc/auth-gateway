package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// SAMLServiceProvider represents a SAML Service Provider configuration
type SAMLServiceProvider struct {
	bun.BaseModel `bun:"table:saml_service_providers"`

	ID          uuid.UUID `json:"id" bun:"id,pk,type:uuid" example:"123e4567-e89b-12d3-a456-426614174005"`
	Name        string    `json:"name" bun:"name,notnull,unique" example:"Salesforce"`
	EntityID    string    `json:"entity_id" bun:"entity_id,notnull,unique" example:"https://saml.salesforce.com"`
	ACSURL      string    `json:"acs_url" bun:"acs_url,notnull" example:"https://saml.salesforce.com/sp/ACS"`
	SLOURL      string    `json:"slo_url,omitempty" bun:"slo_url" example:"https://saml.salesforce.com/sp/SLO"`
	X509Cert    string    `json:"x509_cert,omitempty" bun:"x509_cert" example:"-----BEGIN CERTIFICATE-----..."`
	MetadataURL string    `json:"metadata_url,omitempty" bun:"metadata_url" example:"https://saml.salesforce.com/metadata"`
	IsActive    bool      `json:"is_active" bun:"is_active,default:true"`
	CreatedAt   time.Time `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt   time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel hook for automatic timestamp management
func (s *SAMLServiceProvider) BeforeAppendModel(ctx context.Context, query bun.QueryHook) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	s.UpdatedAt = time.Now()
	return nil
}

// CreateSAMLSPRequest defines the request body for creating a SAML SP
type CreateSAMLSPRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100" example:"Salesforce"`
	EntityID    string `json:"entity_id" validate:"required,url" example:"https://saml.salesforce.com"`
	ACSURL      string `json:"acs_url" validate:"required,url" example:"https://saml.salesforce.com/sp/ACS"`
	SLOURL      string `json:"slo_url,omitempty" validate:"omitempty,url" example:"https://saml.salesforce.com/sp/SLO"`
	X509Cert    string `json:"x509_cert,omitempty" example:"-----BEGIN CERTIFICATE-----..."`
	MetadataURL string `json:"metadata_url,omitempty" validate:"omitempty,url" example:"https://saml.salesforce.com/metadata"`
}

// UpdateSAMLSPRequest defines the request body for updating a SAML SP
type UpdateSAMLSPRequest struct {
	Name        *string `json:"name,omitempty" example:"Salesforce"`
	EntityID    *string `json:"entity_id,omitempty" validate:"omitempty,url" example:"https://saml.salesforce.com"`
	ACSURL      *string `json:"acs_url,omitempty" validate:"omitempty,url" example:"https://saml.salesforce.com/sp/ACS"`
	SLOURL      *string `json:"slo_url,omitempty" validate:"omitempty,url" example:"https://saml.salesforce.com/sp/SLO"`
	X509Cert    *string `json:"x509_cert,omitempty" example:"-----BEGIN CERTIFICATE-----..."`
	MetadataURL *string `json:"metadata_url,omitempty" validate:"omitempty,url" example:"https://saml.salesforce.com/metadata"`
	IsActive    *bool   `json:"is_active,omitempty" example:"true"`
}

// SAMLSPResponse defines the response structure for a SAML SP
type SAMLSPResponse struct {
	ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174005"`
	Name      string    `json:"name" example:"Salesforce"`
	EntityID  string    `json:"entity_id" example:"https://saml.salesforce.com"`
	ACSURL    string    `json:"acs_url" example:"https://saml.salesforce.com/sp/ACS"`
	SLOURL    string    `json:"slo_url,omitempty" example:"https://saml.salesforce.com/sp/SLO"`
	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// ListSAMLSPsResponse defines the response structure for listing SAML SPs
type ListSAMLSPsResponse struct {
	SPs      []SAMLSPResponse `json:"sps"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}
