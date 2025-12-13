package models

import (
	"time"

	"github.com/google/uuid"
)

// BrandingSettings represents the system branding configuration
type BrandingSettings struct {
	ID              uuid.UUID  `json:"id" bun:"id"`
	LogoURL         string     `json:"logo_url,omitempty" bun:"logo_url"`
	FaviconURL      string     `json:"favicon_url,omitempty" bun:"favicon_url"`
	PrimaryColor    string     `json:"primary_color" bun:"primary_color"`       // Hex color
	SecondaryColor  string     `json:"secondary_color" bun:"secondary_color"`   // Hex color
	BackgroundColor string     `json:"background_color" bun:"background_color"` // Hex color
	CustomCSS       string     `json:"custom_css,omitempty" bun:"custom_css"`
	CompanyName     string     `json:"company_name,omitempty" bun:"company_name"`
	SupportEmail    string     `json:"support_email,omitempty" bun:"support_email"`
	TermsURL        string     `json:"terms_url,omitempty" bun:"terms_url"`
	PrivacyURL      string     `json:"privacy_url,omitempty" bun:"privacy_url"`
	UpdatedAt       time.Time  `json:"updated_at" bun:"updated_at"`
	UpdatedBy       *uuid.UUID `json:"updated_by,omitempty" bun:"updated_by"`
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
