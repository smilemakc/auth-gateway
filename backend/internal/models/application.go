package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Application struct {
	bun.BaseModel `bun:"table:applications,alias:app"`

	ID           uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name         string     `json:"name" bun:"name,notnull,unique" example:"my-app"`
	DisplayName  string     `json:"display_name" bun:"display_name,notnull" example:"My Application"`
	Description  string     `json:"description,omitempty" bun:"description" example:"My application description"`
	HomepageURL  string     `json:"homepage_url,omitempty" bun:"homepage_url" example:"https://example.com"`
	CallbackURLs []string   `json:"callback_urls,omitempty" bun:"callback_urls,type:jsonb,default:'[]'" example:"https://example.com/callback"`
	IsActive     bool       `json:"is_active" bun:"is_active,default:true" example:"true"`
	IsSystem     bool       `json:"is_system" bun:"is_system,default:false" example:"false"`
	OwnerID      *uuid.UUID `json:"owner_id,omitempty" bun:"owner_id,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	CreatedAt    time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
	UpdatedAt    time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

	Owner    *User                `json:"owner,omitempty" bun:"rel:belongs-to,join:owner_id=id"`
	Branding *ApplicationBranding `json:"branding,omitempty" bun:"rel:has-one,join:id=application_id"`
}

type ApplicationBranding struct {
	bun.BaseModel `bun:"table:application_branding,alias:ab"`

	ID              uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	ApplicationID   uuid.UUID `json:"application_id" bun:"application_id,type:uuid,notnull,unique" example:"123e4567-e89b-12d3-a456-426614174000"`
	LogoURL         string    `json:"logo_url,omitempty" bun:"logo_url" example:"https://example.com/logo.png"`
	FaviconURL      string    `json:"favicon_url,omitempty" bun:"favicon_url" example:"https://example.com/favicon.ico"`
	PrimaryColor    string    `json:"primary_color" bun:"primary_color,default:'#3B82F6'" example:"#3B82F6"`
	SecondaryColor  string    `json:"secondary_color" bun:"secondary_color,default:'#8B5CF6'" example:"#8B5CF6"`
	BackgroundColor string    `json:"background_color" bun:"background_color,default:'#FFFFFF'" example:"#FFFFFF"`
	CustomCSS       string    `json:"custom_css,omitempty" bun:"custom_css" example:".custom-class { color: red; }"`
	CompanyName     string    `json:"company_name,omitempty" bun:"company_name" example:"Acme Corporation"`
	SupportEmail    string    `json:"support_email,omitempty" bun:"support_email" example:"support@example.com"`
	TermsURL        string    `json:"terms_url,omitempty" bun:"terms_url" example:"https://example.com/terms"`
	PrivacyURL      string    `json:"privacy_url,omitempty" bun:"privacy_url" example:"https://example.com/privacy"`
	UpdatedAt       time.Time `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

	Application *Application `json:"application,omitempty" bun:"rel:belongs-to,join:application_id=id"`
}

type UserApplicationProfile struct {
	bun.BaseModel `bun:"table:user_application_profiles,alias:uap"`

	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID        uuid.UUID  `json:"user_id" bun:"user_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`
	ApplicationID uuid.UUID  `json:"application_id" bun:"application_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`
	DisplayName   *string    `json:"display_name,omitempty" bun:"display_name" example:"John Doe"`
	AvatarURL     *string    `json:"avatar_url,omitempty" bun:"avatar_url" example:"https://example.com/avatar.jpg"`
	Nickname      *string    `json:"nickname,omitempty" bun:"nickname" example:"johnd"`
	Metadata      []byte     `json:"metadata,omitempty" bun:"metadata,type:jsonb,default:'{}'" example:"{\"level\":10}"`
	AppRoles      []string   `json:"app_roles,omitempty" bun:"app_roles,type:jsonb,default:'[]'" example:"admin,editor"`
	IsActive      bool       `json:"is_active" bun:"is_active,default:true" example:"true"`
	IsBanned      bool       `json:"is_banned" bun:"is_banned,default:false" example:"false"`
	BanReason     *string    `json:"ban_reason,omitempty" bun:"ban_reason" example:"Violation of terms"`
	BannedAt      *time.Time `json:"banned_at,omitempty" bun:"banned_at" example:"2024-01-15T10:30:00Z"`
	BannedBy      *uuid.UUID `json:"banned_by,omitempty" bun:"banned_by,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	LastAccessAt  *time.Time `json:"last_access_at,omitempty" bun:"last_access_at" example:"2024-01-15T10:30:00Z"`
	CreatedAt     time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
	UpdatedAt     time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp" example:"2024-01-15T10:30:00Z"`

	User        *User        `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Application *Application `json:"application,omitempty" bun:"rel:belongs-to,join:application_id=id"`
}

type CreateApplicationRequest struct {
	Name         string   `json:"name" binding:"required,min=3,max=100" example:"my-app"`
	DisplayName  string   `json:"display_name" binding:"required,min=3,max=100" example:"My Application"`
	Description  string   `json:"description,omitempty" binding:"max=500" example:"My application description"`
	HomepageURL  string   `json:"homepage_url,omitempty" binding:"omitempty,url,max=500" example:"https://example.com"`
	CallbackURLs []string `json:"callback_urls,omitempty" binding:"omitempty,dive,url" example:"https://example.com/callback"`
	IsActive     *bool    `json:"is_active,omitempty" example:"true"`
}

type UpdateApplicationRequest struct {
	DisplayName  string   `json:"display_name,omitempty" binding:"omitempty,min=3,max=100" example:"My Updated Application"`
	Description  string   `json:"description,omitempty" binding:"max=500" example:"Updated description"`
	HomepageURL  string   `json:"homepage_url,omitempty" binding:"omitempty,url,max=500" example:"https://example.com"`
	CallbackURLs []string `json:"callback_urls,omitempty" binding:"omitempty,dive,url" example:"https://example.com/callback"`
	IsActive     *bool    `json:"is_active,omitempty" example:"true"`
}

type UpdateApplicationBrandingRequest struct {
	LogoURL         string `json:"logo_url,omitempty" binding:"omitempty,url,max=500" example:"https://example.com/logo.png"`
	FaviconURL      string `json:"favicon_url,omitempty" binding:"omitempty,url,max=500" example:"https://example.com/favicon.ico"`
	PrimaryColor    string `json:"primary_color,omitempty" binding:"omitempty,hexcolor" example:"#3B82F6"`
	SecondaryColor  string `json:"secondary_color,omitempty" binding:"omitempty,hexcolor" example:"#8B5CF6"`
	BackgroundColor string `json:"background_color,omitempty" binding:"omitempty,hexcolor" example:"#FFFFFF"`
	CustomCSS       string `json:"custom_css,omitempty" example:".custom-class { color: red; }"`
	CompanyName     string `json:"company_name,omitempty" binding:"max=100" example:"Acme Corporation"`
	SupportEmail    string `json:"support_email,omitempty" binding:"omitempty,email" example:"support@example.com"`
	TermsURL        string `json:"terms_url,omitempty" binding:"omitempty,url,max=500" example:"https://example.com/terms"`
	PrivacyURL      string `json:"privacy_url,omitempty" binding:"omitempty,url,max=500" example:"https://example.com/privacy"`
}

type CreateUserAppProfileRequest struct {
	UserID        uuid.UUID `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	ApplicationID uuid.UUID `json:"application_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	DisplayName   *string   `json:"display_name,omitempty" binding:"omitempty,max=100" example:"John Doe"`
	AvatarURL     *string   `json:"avatar_url,omitempty" binding:"omitempty,url,max=500" example:"https://example.com/avatar.jpg"`
	Nickname      *string   `json:"nickname,omitempty" binding:"omitempty,max=50" example:"johnd"`
	Metadata      []byte    `json:"metadata,omitempty" example:"{\"level\":10}"`
	AppRoles      []string  `json:"app_roles,omitempty" example:"admin,editor"`
	IsActive      *bool     `json:"is_active,omitempty" example:"true"`
}

type UpdateUserAppProfileRequest struct {
	DisplayName *string  `json:"display_name,omitempty" binding:"omitempty,max=100" example:"John Doe"`
	AvatarURL   *string  `json:"avatar_url,omitempty" binding:"omitempty,url,max=500" example:"https://example.com/avatar.jpg"`
	Nickname    *string  `json:"nickname,omitempty" binding:"omitempty,max=50" example:"johnd"`
	Metadata    []byte   `json:"metadata,omitempty" example:"{\"level\":10}"`
	AppRoles    []string `json:"app_roles,omitempty" example:"admin,editor"`
	IsActive    *bool    `json:"is_active,omitempty" example:"true"`
	IsBanned    *bool    `json:"is_banned,omitempty" example:"false"`
	BanReason   *string  `json:"ban_reason,omitempty" binding:"max=500" example:"Violation of terms"`
}

type ApplicationListResponse struct {
	Applications []Application `json:"applications"`
	Total        int           `json:"total" example:"100"`
	Page         int           `json:"page" example:"1"`
	PageSize     int           `json:"page_size" example:"20"`
	TotalPages   int           `json:"total_pages" example:"5"`
}

type UserAppProfileListResponse struct {
	Profiles   []UserApplicationProfile `json:"profiles"`
	Total      int                      `json:"total" example:"100"`
	Page       int                      `json:"page" example:"1"`
	PageSize   int                      `json:"page_size" example:"20"`
	TotalPages int                      `json:"total_pages" example:"5"`
}

func (a *Application) PublicApplication() *Application {
	return &Application{
		ID:           a.ID,
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Description:  a.Description,
		HomepageURL:  a.HomepageURL,
		CallbackURLs: a.CallbackURLs,
		IsActive:     a.IsActive,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
		Branding:     a.Branding,
	}
}

func (b *ApplicationBranding) ToPublicResponse() PublicApplicationBrandingResponse {
	return PublicApplicationBrandingResponse{
		LogoURL:    b.LogoURL,
		FaviconURL: b.FaviconURL,
		Theme: ApplicationBrandingTheme{
			PrimaryColor:    b.PrimaryColor,
			SecondaryColor:  b.SecondaryColor,
			BackgroundColor: b.BackgroundColor,
		},
		CompanyName:  b.CompanyName,
		SupportEmail: b.SupportEmail,
		TermsURL:     b.TermsURL,
		PrivacyURL:   b.PrivacyURL,
	}
}

func (p *UserApplicationProfile) GetMetadataMap() (map[string]interface{}, error) {
	if len(p.Metadata) == 0 || string(p.Metadata) == "{}" {
		return make(map[string]interface{}), nil
	}

	var metadataMap map[string]interface{}
	if err := json.Unmarshal(p.Metadata, &metadataMap); err != nil {
		return nil, err
	}

	return metadataMap, nil
}

type ApplicationBrandingTheme struct {
	PrimaryColor    string `json:"primary_color" example:"#3B82F6"`
	SecondaryColor  string `json:"secondary_color" example:"#8B5CF6"`
	BackgroundColor string `json:"background_color" example:"#FFFFFF"`
}

type PublicApplicationBrandingResponse struct {
	LogoURL      string                   `json:"logo_url,omitempty" example:"https://example.com/logo.png"`
	FaviconURL   string                   `json:"favicon_url,omitempty" example:"https://example.com/favicon.ico"`
	Theme        ApplicationBrandingTheme `json:"theme"`
	CompanyName  string                   `json:"company_name,omitempty" example:"Acme Corporation"`
	SupportEmail string                   `json:"support_email,omitempty" example:"support@example.com"`
	TermsURL     string                   `json:"terms_url,omitempty" example:"https://example.com/terms"`
	PrivacyURL   string                   `json:"privacy_url,omitempty" example:"https://example.com/privacy"`
}
