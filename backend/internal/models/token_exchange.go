package models

import (
	"time"

	"github.com/google/uuid"
)

type TokenExchangeCode struct {
	Code        string    `json:"code"`
	UserID      uuid.UUID `json:"user_id"`
	SourceAppID uuid.UUID `json:"source_app_id"`
	TargetAppID uuid.UUID `json:"target_app_id"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type CreateTokenExchangeRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
	TargetAppID string `json:"target_application_id" binding:"required,uuid"`
}

type CreateTokenExchangeResponse struct {
	ExchangeCode string    `json:"exchange_code"`
	ExpiresAt    time.Time `json:"expires_at"`
	RedirectURL  string    `json:"redirect_url,omitempty"`
}

type RedeemTokenExchangeRequest struct {
	ExchangeCode string `json:"exchange_code" binding:"required"`
}

type RedeemTokenExchangeResponse struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int64  `json:"expires_in"`
	User          *User  `json:"user"`
	ApplicationID string `json:"application_id"`
}
