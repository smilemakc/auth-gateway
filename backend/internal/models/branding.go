package models

import (
	"time"

	"github.com/google/uuid"
)

// BrandingSettings represents the system branding configuration
type BrandingSettings struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	LogoURL         string     `json:"logo_url,omitempty" db:"logo_url"`
	FaviconURL      string     `json:"favicon_url,omitempty" db:"favicon_url"`
	PrimaryColor    string     `json:"primary_color" db:"primary_color"`       // Hex color
	SecondaryColor  string     `json:"secondary_color" db:"secondary_color"`   // Hex color
	BackgroundColor string     `json:"background_color" db:"background_color"` // Hex color
	CustomCSS       string     `json:"custom_css,omitempty" db:"custom_css"`
	CompanyName     string     `json:"company_name,omitempty" db:"company_name"`
	SupportEmail    string     `json:"support_email,omitempty" db:"support_email"`
	TermsURL        string     `json:"terms_url,omitempty" db:"terms_url"`
	PrivacyURL      string     `json:"privacy_url,omitempty" db:"privacy_url"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy       *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
}

// UpdateBrandingRequest is the request to update branding settings
type UpdateBrandingRequest struct {
	LogoURL         string `json:"logo_url" binding:"omitempty,url,max=500"`
	FaviconURL      string `json:"favicon_url" binding:"omitempty,url,max=500"`
	PrimaryColor    string `json:"primary_color" binding:"omitempty,hexcolor"`
	SecondaryColor  string `json:"secondary_color" binding:"omitempty,hexcolor"`
	BackgroundColor string `json:"background_color" binding:"omitempty,hexcolor"`
	CustomCSS       string `json:"custom_css"`
	CompanyName     string `json:"company_name" binding:"max=100"`
	SupportEmail    string `json:"support_email" binding:"omitempty,email"`
	TermsURL        string `json:"terms_url" binding:"omitempty,url,max=500"`
	PrivacyURL      string `json:"privacy_url" binding:"omitempty,url,max=500"`
}

// BrandingTheme contains the color scheme for frontend
type BrandingTheme struct {
	PrimaryColor    string `json:"primary_color"`
	SecondaryColor  string `json:"secondary_color"`
	BackgroundColor string `json:"background_color"`
}

// PublicBrandingResponse is the branding info exposed to public API
type PublicBrandingResponse struct {
	LogoURL      string        `json:"logo_url,omitempty"`
	FaviconURL   string        `json:"favicon_url,omitempty"`
	Theme        BrandingTheme `json:"theme"`
	CompanyName  string        `json:"company_name,omitempty"`
	SupportEmail string        `json:"support_email,omitempty"`
	TermsURL     string        `json:"terms_url,omitempty"`
	PrivacyURL   string        `json:"privacy_url,omitempty"`
}
