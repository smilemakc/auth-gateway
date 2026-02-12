package authgateway

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/packages/go-sdk/proto"
)

type ExchangeResponse struct {
	ExchangeCode string `json:"exchange_code"`
	ExpiresAt    string `json:"expires_at"`
	RedirectURL  string `json:"redirect_url,omitempty"`
}

type ExchangeAuthResponse struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	ApplicationID string `json:"application_id"`
}

func (c *GRPCClient) CreateExchange(ctx context.Context, accessToken, targetAppID string) (*ExchangeResponse, error) {
	resp, err := c.client.CreateTokenExchange(c.withMetadata(ctx), &proto.CreateTokenExchangeGrpcRequest{
		AccessToken:          accessToken,
		TargetApplicationId:  targetAppID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create token exchange: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeBadRequest,
			Message: resp.ErrorMessage,
		}
	}

	return &ExchangeResponse{
		ExchangeCode: resp.ExchangeCode,
		ExpiresAt:    resp.ExpiresAt,
		RedirectURL:  resp.RedirectUrl,
	}, nil
}

func (c *GRPCClient) RedeemExchange(ctx context.Context, code string) (*ExchangeAuthResponse, error) {
	resp, err := c.client.RedeemTokenExchange(c.withMetadata(ctx), &proto.RedeemTokenExchangeGrpcRequest{
		ExchangeCode: code,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to redeem token exchange: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeUnauthorized,
			Message: resp.ErrorMessage,
		}
	}

	return &ExchangeAuthResponse{
		AccessToken:   resp.AccessToken,
		RefreshToken:  resp.RefreshToken,
		UserID:        resp.UserId,
		Email:         resp.Email,
		ApplicationID: resp.ApplicationId,
	}, nil
}

func (c *GRPCClient) GinSSOCallback(onSuccess func(*gin.Context, *ExchangeAuthResponse)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		code := ctx.Query("code")
		if code == "" {
			ctx.AbortWithStatusJSON(400, gin.H{"error": "Missing exchange code"})
			return
		}
		resp, err := c.RedeemExchange(ctx, code)
		if err != nil {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Invalid or expired exchange code"})
			return
		}
		onSuccess(ctx, resp)
	}
}
