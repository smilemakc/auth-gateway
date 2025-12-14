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
	// URL to company logo image
	LogoURL string `json:"logo_url" binding:"omitempty,url,max=500" example:"https://example.com/logo.png"`
	// URL to favicon image
	FaviconURL string `json:"favicon_url" binding:"omitempty,url,max=500" example:"https://example.com/favicon.ico"`
	// Primary brand color (hex format)
	PrimaryColor string `json:"primary_color" binding:"omitempty,hexcolor" example:"#3B82F6"`
	// Secondary brand color (hex format)
	SecondaryColor string `json:"secondary_color" binding:"omitempty,hexcolor" example:"#8B5CF6"`
	// Background color (hex format)
	BackgroundColor string `json:"background_color" binding:"omitempty,hexcolor" example:"#FFFFFF"`
	// Custom CSS to inject
	CustomCSS string `json:"custom_css" example:".custom-class { color: red; }"`
	// Company name (max 100 characters)
	CompanyName string `json:"company_name" binding:"max=100" example:"Acme Corporation"`
	// Support email address
	SupportEmail string `json:"support_email" binding:"omitempty,email" example:"support@example.com"`
	// URL to terms of service
	TermsURL string `json:"terms_url" binding:"omitempty,url,max=500" example:"https://example.com/terms"`
	// URL to privacy policy
	PrivacyURL string `json:"privacy_url" binding:"omitempty,url,max=500" example:"https://example.com/privacy"`
}

// BrandingTheme contains the color scheme for frontend
type BrandingTheme struct {
	// Primary brand color
	PrimaryColor string `json:"primary_color" example:"#3B82F6"`
	// Secondary brand color
	SecondaryColor string `json:"secondary_color" example:"#8B5CF6"`
	// Background color
	BackgroundColor string `json:"background_color" example:"#FFFFFF"`
}

// PublicBrandingResponse is the branding info exposed to public API
type PublicBrandingResponse struct {
	// URL to company logo
	LogoURL string `json:"logo_url,omitempty" example:"https://example.com/logo.png"`
	// URL to favicon
	FaviconURL string `json:"favicon_url,omitempty" example:"https://example.com/favicon.ico"`
	// Color theme
	Theme BrandingTheme `json:"theme"`
	// Company name
	CompanyName string `json:"company_name,omitempty" example:"Acme Corporation"`
	// Support email
	SupportEmail string `json:"support_email,omitempty" example:"support@example.com"`
	// Terms of service URL
	TermsURL string `json:"terms_url,omitempty" example:"https://example.com/terms"`
	// Privacy policy URL
	PrivacyURL string `json:"privacy_url,omitempty" example:"https://example.com/privacy"`
}
