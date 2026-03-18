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
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// TwoFactorHandler test helpers
// ---------------------------------------------------------------------------

func setupTwoFactorHandler() (*TwoFactorHandler, *mockTwoFactorServicer, *mockUserServicer) {
	twoFASvc := &mockTwoFactorServicer{}
	userSvc := &mockUserServicer{}
	h := NewTwoFactorHandler(twoFASvc, userSvc, nil, testLogger())
	return h, twoFASvc, userSvc
}

// ---------------------------------------------------------------------------
// Setup Tests
// ---------------------------------------------------------------------------

func TestTwoFactorHandler_Setup_ShouldReturn200_WhenRequestValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, twoFASvc, _ := setupTwoFactorHandler()

	userID := uuid.New()
	twoFASvc.SetupTOTPFunc = func(uid uuid.UUID, password string) (*models.TwoFactorSetupResponse, error) {
		return &models.TwoFactorSetupResponse{
			Secret:      "JBSWY3DPEHPK3PXP",
			QRCodeURL:   "otpauth://totp/test",
			BackupCodes: []string{"111111", "222222"},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/setup", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Setup(c)
	})

	body := `{"password":"SecurePass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.TwoFactorSetupResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "JBSWY3DPEHPK3PXP", resp.Secret)
	assert.Len(t, resp.BackupCodes, 2)
}

func TestTwoFactorHandler_Setup_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupTwoFactorHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/setup", h.Setup)

	body := `{"password":"SecurePass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTwoFactorHandler_Setup_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupTwoFactorHandler()

	userID := uuid.New()
	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/setup", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Setup(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/2fa/setup", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTwoFactorHandler_Setup_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, twoFASvc, _ := setupTwoFactorHandler()

	userID := uuid.New()
	twoFASvc.SetupTOTPFunc = func(uid uuid.UUID, password string) (*models.TwoFactorSetupResponse, error) {
		return nil, models.NewAppError(http.StatusForbidden, "Incorrect password")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/setup", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Setup(c)
	})

	body := `{"password":"WrongPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---------------------------------------------------------------------------
// Verify Tests
// ---------------------------------------------------------------------------

func TestTwoFactorHandler_Verify_ShouldReturn200_WhenCodeValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, twoFASvc, _ := setupTwoFactorHandler()

	userID := uuid.New()
	twoFASvc.VerifyTOTPSetupFunc = func(uid uuid.UUID, code string) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/verify", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Verify(c)
	})

	body := `{"code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "2FA enabled successfully")
}

func TestTwoFactorHandler_Verify_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupTwoFactorHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/verify", h.Verify)

	body := `{"code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTwoFactorHandler_Verify_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupTwoFactorHandler()

	userID := uuid.New()
	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/verify", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Verify(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/2fa/verify", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTwoFactorHandler_Verify_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, twoFASvc, _ := setupTwoFactorHandler()

	userID := uuid.New()
	twoFASvc.VerifyTOTPSetupFunc = func(uid uuid.UUID, code string) error {
		return models.NewAppError(http.StatusBadRequest, "Invalid TOTP code")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/verify", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Verify(c)
	})

	body := `{"code":"000000"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Disable Tests
// ---------------------------------------------------------------------------

func TestTwoFactorHandler_Disable_ShouldReturn200_WhenRequestValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, twoFASvc, _ := setupTwoFactorHandler()

	userID := uuid.New()
	twoFASvc.DisableTOTPFunc = func(uid uuid.UUID, password, code string) error {
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/disable", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Disable(c)
	})

	body := `{"password":"SecurePass123!","code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/disable", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "2FA disabled successfully")
}

func TestTwoFactorHandler_Disable_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupTwoFactorHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/disable", h.Disable)

	body := `{"password":"SecurePass123!","code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/disable", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTwoFactorHandler_Disable_ShouldReturn400_WhenMissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupTwoFactorHandler()

	userID := uuid.New()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"missing code", `{"password":"SecurePass123!"}`},
		{"missing password", `{"code":"123456"}`},
		{"code wrong length", `{"password":"SecurePass123!","code":"12345"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/2fa/disable", func(c *gin.Context) {
				c.Set(utils.UserIDKey, userID)
				h.Disable(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/2fa/disable", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestTwoFactorHandler_Disable_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, twoFASvc, _ := setupTwoFactorHandler()

	userID := uuid.New()
	twoFASvc.DisableTOTPFunc = func(uid uuid.UUID, password, code string) error {
		return errors.New("incorrect password")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/disable", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.Disable(c)
	})

	body := `{"password":"WrongPass123!","code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/disable", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// GetStatus Tests
// ---------------------------------------------------------------------------

func TestTwoFactorHandler_GetStatus_ShouldReturn200_WhenSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, twoFASvc, _ := setupTwoFactorHandler()

	userID := uuid.New()
	twoFASvc.GetStatusFunc = func(uid uuid.UUID) (*models.TwoFactorStatusResponse, error) {
		return &models.TwoFactorStatusResponse{
			Enabled:     true,
			BackupCodes: 5,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/2fa/status", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.GetStatus(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/2fa/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.TwoFactorStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Enabled)
	assert.Equal(t, 5, resp.BackupCodes)
}

func TestTwoFactorHandler_GetStatus_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupTwoFactorHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/2fa/status", h.GetStatus)

	req := httptest.NewRequest(http.MethodGet, "/2fa/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTwoFactorHandler_GetStatus_ShouldReturn200Disabled_WhenNot2FA(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, twoFASvc, _ := setupTwoFactorHandler()

	userID := uuid.New()
	twoFASvc.GetStatusFunc = func(uid uuid.UUID) (*models.TwoFactorStatusResponse, error) {
		return &models.TwoFactorStatusResponse{
			Enabled:     false,
			BackupCodes: 0,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/2fa/status", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.GetStatus(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/2fa/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.TwoFactorStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Enabled)
}

func TestTwoFactorHandler_GetStatus_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, twoFASvc, _ := setupTwoFactorHandler()

	userID := uuid.New()
	twoFASvc.GetStatusFunc = func(uid uuid.UUID) (*models.TwoFactorStatusResponse, error) {
		return nil, errors.New("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/2fa/status", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.GetStatus(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/2fa/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// RegenerateBackupCodes Tests
// ---------------------------------------------------------------------------

func TestTwoFactorHandler_RegenerateBackupCodes_ShouldReturn200_WhenRequestValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, twoFASvc, _ := setupTwoFactorHandler()

	userID := uuid.New()
	twoFASvc.RegenerateBackupCodesFunc = func(uid uuid.UUID, password string) ([]string, error) {
		return []string{"111111", "222222", "333333"}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/backup-codes/regenerate", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.RegenerateBackupCodes(c)
	})

	body := `{"password":"SecurePass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/backup-codes/regenerate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "backup_codes")
	assert.Contains(t, w.Body.String(), "111111")
}

func TestTwoFactorHandler_RegenerateBackupCodes_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupTwoFactorHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/backup-codes/regenerate", h.RegenerateBackupCodes)

	body := `{"password":"SecurePass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/backup-codes/regenerate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTwoFactorHandler_RegenerateBackupCodes_ShouldReturn400_WhenMissingPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupTwoFactorHandler()

	userID := uuid.New()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"missing password", `{}`},
		{"invalid JSON", `{invalid}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/2fa/backup-codes/regenerate", func(c *gin.Context) {
				c.Set(utils.UserIDKey, userID)
				h.RegenerateBackupCodes(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/2fa/backup-codes/regenerate", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestTwoFactorHandler_RegenerateBackupCodes_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, twoFASvc, _ := setupTwoFactorHandler()

	userID := uuid.New()
	twoFASvc.RegenerateBackupCodesFunc = func(uid uuid.UUID, password string) ([]string, error) {
		return nil, models.NewAppError(http.StatusForbidden, "Incorrect password")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/2fa/backup-codes/regenerate", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		h.RegenerateBackupCodes(c)
	})

	body := `{"password":"WrongPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/2fa/backup-codes/regenerate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
