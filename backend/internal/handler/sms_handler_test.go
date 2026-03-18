package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

type smsTestFixture struct {
	handler *SMSHandler
	smsSvc  *mockSMSServicer
}

func setupSMSTestFixture() *smsTestFixture {
	svc := &mockSMSServicer{}
	h := NewSMSHandler(svc, testLogger())
	return &smsTestFixture{handler: h, smsSvc: svc}
}

// ===========================================================================
// SendSMS
// ===========================================================================

func TestSMSHandler_SendSMS_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupSMSTestFixture()

	fix.smsSvc.SendOTPFunc = func(req *models.SendSMSRequest, ipAddress string) (*models.SendSMSResponse, error) {
		assert.Equal(t, "+1234567890", req.Phone)
		msgID := "SM123"
		return &models.SendSMSResponse{
			Success:   true,
			MessageID: &msgID,
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/sms/send", fix.handler.SendSMS)

	body := `{"phone":"+1234567890","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/sms/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.SendSMSResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.MessageID)
}

func TestSMSHandler_SendSMS_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupSMSTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing phone", `{"type":"verification"}`},
		{"missing type", `{"phone":"+1234567890"}`},
		{"invalid type", `{"phone":"+1234567890","type":"invalid"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/sms/send", fix.handler.SendSMS)

			req := httptest.NewRequest(http.MethodPost, "/sms/send", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestSMSHandler_SendSMS_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupSMSTestFixture()

	fix.smsSvc.SendOTPFunc = func(req *models.SendSMSRequest, ipAddress string) (*models.SendSMSResponse, error) {
		return nil, models.NewAppError(http.StatusTooManyRequests, "Rate limit exceeded")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/sms/send", fix.handler.SendSMS)

	body := `{"phone":"+1234567890","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/sms/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

// ===========================================================================
// VerifySMS
// ===========================================================================

func TestSMSHandler_VerifySMS_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupSMSTestFixture()

	fix.smsSvc.VerifyOTPFunc = func(req *models.VerifySMSOTPRequest) (*models.VerifySMSOTPResponse, error) {
		assert.Equal(t, "+1234567890", req.Phone)
		assert.Equal(t, "123456", req.Code)
		return &models.VerifySMSOTPResponse{
			Valid:   true,
			Message: "OTP verified successfully",
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/sms/verify", fix.handler.VerifySMS)

	body := `{"phone":"+1234567890","code":"123456","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/sms/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.VerifySMSOTPResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Valid)
}

func TestSMSHandler_VerifySMS_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupSMSTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing phone", `{"code":"123456","type":"verification"}`},
		{"missing code", `{"phone":"+1234567890","type":"verification"}`},
		{"code wrong length", `{"phone":"+1234567890","code":"12345","type":"verification"}`},
		{"missing type", `{"phone":"+1234567890","code":"123456"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/sms/verify", fix.handler.VerifySMS)

			req := httptest.NewRequest(http.MethodPost, "/sms/verify", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestSMSHandler_VerifySMS_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupSMSTestFixture()

	fix.smsSvc.VerifyOTPFunc = func(req *models.VerifySMSOTPRequest) (*models.VerifySMSOTPResponse, error) {
		return nil, models.NewAppError(http.StatusBadRequest, "Invalid OTP code")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/sms/verify", fix.handler.VerifySMS)

	body := `{"phone":"+1234567890","code":"123456","type":"verification"}`
	req := httptest.NewRequest(http.MethodPost, "/sms/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===========================================================================
// GetStats
// ===========================================================================

func TestSMSHandler_GetStats_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupSMSTestFixture()

	fix.smsSvc.GetStatsFunc = func() (*models.SMSStatsResponse, error) {
		return &models.SMSStatsResponse{
			TotalSent:   100,
			TotalFailed: 5,
			SentToday:   10,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/sms/stats", fix.handler.GetStats)

	req := httptest.NewRequest(http.MethodGet, "/sms/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.SMSStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, int64(100), resp.TotalSent)
	assert.Equal(t, int64(5), resp.TotalFailed)
	assert.Equal(t, int64(10), resp.SentToday)
}

func TestSMSHandler_GetStats_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupSMSTestFixture()

	fix.smsSvc.GetStatsFunc = func() (*models.SMSStatsResponse, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/sms/stats", fix.handler.GetStats)

	req := httptest.NewRequest(http.MethodGet, "/sms/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSMSHandler_GetStats_ShouldReturnZeros_WhenNoData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupSMSTestFixture()

	fix.smsSvc.GetStatsFunc = func() (*models.SMSStatsResponse, error) {
		return &models.SMSStatsResponse{
			TotalSent:   0,
			TotalFailed: 0,
			SentToday:   0,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/sms/stats", fix.handler.GetStats)

	req := httptest.NewRequest(http.MethodGet, "/sms/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.SMSStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, int64(0), resp.TotalSent)
}
