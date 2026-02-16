package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// TokenExchangeHandler Tests
// ---------------------------------------------------------------------------

func setupTokenExchangeHandler() (*TokenExchangeHandler, *mockTokenExchangeServicer) {
	svc := &mockTokenExchangeServicer{}
	h := NewTokenExchangeHandler(svc)
	return h, svc
}

// ---------------------------------------------------------------------------
// CreateExchange Tests
// ---------------------------------------------------------------------------

func TestTokenExchangeHandler_CreateExchange_ShouldReturn200_WhenRequestValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupTokenExchangeHandler()

	targetAppID := uuid.New()
	svc.CreateExchangeFunc = func(req *models.CreateTokenExchangeRequest, sourceAppID *uuid.UUID) (*models.CreateTokenExchangeResponse, error) {
		return &models.CreateTokenExchangeResponse{
			ExchangeCode: "exchange-code-123",
			ExpiresAt:    time.Now().Add(5 * time.Minute),
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/exchange/create", h.CreateExchange)

	body := `{"access_token":"test-token","target_application_id":"` + targetAppID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/exchange/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.CreateTokenExchangeResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "exchange-code-123", resp.ExchangeCode)
}

func TestTokenExchangeHandler_CreateExchange_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupTokenExchangeHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/exchange/create", h.CreateExchange)

	req := httptest.NewRequest(http.MethodPost, "/exchange/create", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTokenExchangeHandler_CreateExchange_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupTokenExchangeHandler()

	svc.CreateExchangeFunc = func(req *models.CreateTokenExchangeRequest, sourceAppID *uuid.UUID) (*models.CreateTokenExchangeResponse, error) {
		return nil, models.NewAppError(http.StatusForbidden, "Exchange not allowed")
	}

	targetAppID := uuid.New()
	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/exchange/create", h.CreateExchange)

	body := `{"access_token":"test-token","target_application_id":"` + targetAppID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/exchange/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTokenExchangeHandler_CreateExchange_ShouldReturn400_WhenEmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupTokenExchangeHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/exchange/create", h.CreateExchange)

	req := httptest.NewRequest(http.MethodPost, "/exchange/create", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// RedeemExchange Tests
// ---------------------------------------------------------------------------

func TestTokenExchangeHandler_RedeemExchange_ShouldReturn200_WhenRequestValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupTokenExchangeHandler()

	svc.RedeemExchangeFunc = func(req *models.RedeemTokenExchangeRequest, redeemingAppID *uuid.UUID) (*models.RedeemTokenExchangeResponse, error) {
		return &models.RedeemTokenExchangeResponse{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    900,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/exchange/redeem", h.RedeemExchange)

	body := `{"exchange_code":"exchange-code-123"}`
	req := httptest.NewRequest(http.MethodPost, "/exchange/redeem", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.RedeemTokenExchangeResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "new-access-token", resp.AccessToken)
	assert.Equal(t, "new-refresh-token", resp.RefreshToken)
}

func TestTokenExchangeHandler_RedeemExchange_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupTokenExchangeHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/exchange/redeem", h.RedeemExchange)

	req := httptest.NewRequest(http.MethodPost, "/exchange/redeem", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTokenExchangeHandler_RedeemExchange_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, svc := setupTokenExchangeHandler()

	svc.RedeemExchangeFunc = func(req *models.RedeemTokenExchangeRequest, redeemingAppID *uuid.UUID) (*models.RedeemTokenExchangeResponse, error) {
		return nil, errors.New("exchange code expired")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/exchange/redeem", h.RedeemExchange)

	body := `{"exchange_code":"expired-code"}`
	req := httptest.NewRequest(http.MethodPost, "/exchange/redeem", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTokenExchangeHandler_RedeemExchange_ShouldReturn400_WhenEmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _ := setupTokenExchangeHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/exchange/redeem", h.RedeemExchange)

	req := httptest.NewRequest(http.MethodPost, "/exchange/redeem", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
