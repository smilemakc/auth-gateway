package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// OTPHandler test helpers
// ---------------------------------------------------------------------------

func setupOTPHandler() (*OTPHandler, *mockOTPServicer, *mockAuthServicer) {
	otpSvc := &mockOTPServicer{}
	authSvc := &mockAuthServicer{}
	h := NewOTPHandler(otpSvc, authSvc, testLogger())
	return h, otpSvc, authSvc
}

func strPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// ResendVerification Tests
// ---------------------------------------------------------------------------

func TestOTPHandler_ResendVerification_ShouldReturn200_WhenRequestValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.SendOTPFunc = func(req *models.SendOTPRequest) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/verify/resend", h.ResendVerification)

	body := `{"email":"test@example.com","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/verify/resend", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Verification code sent")
}

func TestOTPHandler_ResendVerification_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupOTPHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/verify/resend", h.ResendVerification)

	req := httptest.NewRequest(http.MethodPost, "/verify/resend", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOTPHandler_ResendVerification_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.SendOTPFunc = func(req *models.SendOTPRequest) error {
		return models.NewAppError(http.StatusTooManyRequests, "Rate limit exceeded")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/verify/resend", h.ResendVerification)

	body := `{"email":"test@example.com","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/verify/resend", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestOTPHandler_ResendVerification_ShouldReturn400_WhenEmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupOTPHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/verify/resend", h.ResendVerification)

	req := httptest.NewRequest(http.MethodPost, "/verify/resend", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// VerifyEmailOTP Tests
// ---------------------------------------------------------------------------

func TestOTPHandler_VerifyEmailOTP_ShouldReturn200_WhenCodeValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	userID := uuid.New()
	otpSvc.VerifyOTPFunc = func(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
		return &models.VerifyOTPResponse{
			Valid: true,
			User:  &models.User{ID: userID, Email: "test@example.com"},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/verify/email", h.VerifyEmailOTP)

	body := `{"email":"test@example.com","code":"123456","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/verify/email", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Verification successful")
}

func TestOTPHandler_VerifyEmailOTP_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupOTPHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/verify/email", h.VerifyEmailOTP)

	req := httptest.NewRequest(http.MethodPost, "/verify/email", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOTPHandler_VerifyEmailOTP_ShouldReturn401_WhenCodeInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.VerifyOTPFunc = func(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
		return &models.VerifyOTPResponse{Valid: false}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/verify/email", h.VerifyEmailOTP)

	body := `{"email":"test@example.com","code":"000000","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/verify/email", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired verification code")
}

func TestOTPHandler_VerifyEmailOTP_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.VerifyOTPFunc = func(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
		return nil, errors.New("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/verify/email", h.VerifyEmailOTP)

	body := `{"email":"test@example.com","code":"123456","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/verify/email", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// SendOTP Tests
// ---------------------------------------------------------------------------

func TestOTPHandler_SendOTP_ShouldReturn200_WhenRequestValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.SendOTPFunc = func(req *models.SendOTPRequest) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/otp/send", h.SendOTP)

	body := `{"email":"test@example.com","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/otp/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "OTP code sent successfully")
}

func TestOTPHandler_SendOTP_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupOTPHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/otp/send", h.SendOTP)

	req := httptest.NewRequest(http.MethodPost, "/otp/send", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOTPHandler_SendOTP_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.SendOTPFunc = func(req *models.SendOTPRequest) error {
		return models.NewAppError(http.StatusTooManyRequests, "Too many OTP requests")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/otp/send", h.SendOTP)

	body := `{"email":"test@example.com","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/otp/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestOTPHandler_SendOTP_ShouldReturn400_WhenEmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupOTPHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/otp/send", h.SendOTP)

	req := httptest.NewRequest(http.MethodPost, "/otp/send", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// VerifyOTP Tests
// ---------------------------------------------------------------------------

func TestOTPHandler_VerifyOTP_ShouldReturn200_WhenCodeValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.VerifyOTPFunc = func(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
		return &models.VerifyOTPResponse{Valid: true}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/otp/verify", h.VerifyOTP)

	body := `{"email":"test@example.com","code":"123456","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/otp/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.VerifyOTPResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Valid)
}

func TestOTPHandler_VerifyOTP_ShouldReturn401_WhenCodeInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.VerifyOTPFunc = func(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
		return &models.VerifyOTPResponse{Valid: false}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/otp/verify", h.VerifyOTP)

	body := `{"email":"test@example.com","code":"000000","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/otp/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired OTP code")
}

func TestOTPHandler_VerifyOTP_ShouldReturn200WithTokens_WhenLoginType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, authSvc := setupOTPHandler()

	userID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com", IsActive: true}

	otpSvc.VerifyOTPFunc = func(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
		return &models.VerifyOTPResponse{Valid: true, User: user}, nil
	}
	authSvc.GenerateTokensForUserFunc = func(u *models.User, ip, userAgent string) (*models.AuthResponse, error) {
		return &models.AuthResponse{
			AccessToken:  "access-token-123",
			RefreshToken: "refresh-token-456",
			User:         u,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/otp/verify", h.VerifyOTP)

	body := `{"email":"test@example.com","code":"123456","type":"login"}`
	req := httptest.NewRequest(http.MethodPost, "/otp/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.VerifyOTPResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, "access-token-123", resp.AccessToken)
	assert.Equal(t, "refresh-token-456", resp.RefreshToken)
}

func TestOTPHandler_VerifyOTP_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.VerifyOTPFunc = func(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
		return nil, errors.New("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/otp/verify", h.VerifyOTP)

	body := `{"email":"test@example.com","code":"123456","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/otp/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// RequestPasswordlessLogin Tests
// ---------------------------------------------------------------------------

func TestOTPHandler_RequestPasswordlessLogin_ShouldReturn200_WhenRequestValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.SendOTPFunc = func(req *models.SendOTPRequest) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/passwordless/request", h.RequestPasswordlessLogin)

	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/passwordless/request", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Login code sent to your email")
}

func TestOTPHandler_RequestPasswordlessLogin_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupOTPHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/passwordless/request", h.RequestPasswordlessLogin)

	req := httptest.NewRequest(http.MethodPost, "/auth/passwordless/request", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOTPHandler_RequestPasswordlessLogin_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.SendOTPFunc = func(req *models.SendOTPRequest) error {
		return models.NewAppError(http.StatusTooManyRequests, "Rate limit exceeded")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/passwordless/request", h.RequestPasswordlessLogin)

	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/passwordless/request", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestOTPHandler_RequestPasswordlessLogin_ShouldReturn400_WhenInvalidEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupOTPHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/passwordless/request", h.RequestPasswordlessLogin)

	body := `{"email":"not-an-email"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/passwordless/request", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// VerifyPasswordlessLogin Tests
// ---------------------------------------------------------------------------

func TestOTPHandler_VerifyPasswordlessLogin_ShouldReturn200_WhenCodeValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, authSvc := setupOTPHandler()

	userID := uuid.New()
	user := &models.User{ID: userID, Email: "test@example.com", IsActive: true}

	otpSvc.VerifyOTPFunc = func(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
		return &models.VerifyOTPResponse{Valid: true, User: user}, nil
	}
	authSvc.GenerateTokensForUserFunc = func(u *models.User, ip, userAgent string) (*models.AuthResponse, error) {
		return &models.AuthResponse{
			AccessToken:  "access-token-123",
			RefreshToken: "refresh-token-456",
			User:         u,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/passwordless/verify", h.VerifyPasswordlessLogin)

	body := `{"email":"test@example.com","code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/passwordless/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "access-token-123")
	assert.Contains(t, w.Body.String(), "refresh-token-456")
}

func TestOTPHandler_VerifyPasswordlessLogin_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupOTPHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/passwordless/verify", h.VerifyPasswordlessLogin)

	req := httptest.NewRequest(http.MethodPost, "/auth/passwordless/verify", strings.NewReader("{bad}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOTPHandler_VerifyPasswordlessLogin_ShouldReturn401_WhenCodeInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.VerifyOTPFunc = func(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
		return &models.VerifyOTPResponse{Valid: false}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/passwordless/verify", h.VerifyPasswordlessLogin)

	body := `{"email":"test@example.com","code":"000000"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/passwordless/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOTPHandler_VerifyPasswordlessLogin_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, otpSvc, _ := setupOTPHandler()

	otpSvc.VerifyOTPFunc = func(req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
		return nil, errors.New("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/passwordless/verify", h.VerifyPasswordlessLogin)

	body := `{"email":"test@example.com","code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/passwordless/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
