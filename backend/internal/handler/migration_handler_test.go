package handler

import (
	"encoding/json"
	"fmt"
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
// helpers
// ---------------------------------------------------------------------------

type migrationTestFixture struct {
	handler      *MigrationHandler
	migrationSvc *mockMigrationServicer
}

func setupMigrationTestFixture() *migrationTestFixture {
	svc := &mockMigrationServicer{}
	h := NewMigrationHandler(svc, testLogger())
	return &migrationTestFixture{handler: h, migrationSvc: svc}
}

func sampleImportResult() *models.ImportResult {
	return &models.ImportResult{
		Total:   1,
		Created: 1,
		Skipped: 0,
	}
}

// ===========================================================================
// ImportUsers
// ===========================================================================

func TestMigrationHandler_ImportUsers_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	appID := uuid.New()
	fix.migrationSvc.ImportUsersFunc = func(id uuid.UUID, entries []models.ImportUserEntry) (*models.ImportResult, error) {
		assert.Equal(t, appID, id)
		assert.Len(t, entries, 1)
		assert.Equal(t, "a@test.com", entries[0].Email)
		return sampleImportResult(), nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/import/users", fix.handler.ImportUsers)

	body := fmt.Sprintf(`{"application_id":"%s","users":[{"email":"a@test.com","password":"Test123!"}]}`, appID.String())
	req := httptest.NewRequest(http.MethodPost, "/import/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.ImportResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Created)
}

func TestMigrationHandler_ImportUsers_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing application_id", `{"users":[{"email":"a@test.com"}]}`},
		{"missing users", fmt.Sprintf(`{"application_id":"%s"}`, uuid.New().String())},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/import/users", fix.handler.ImportUsers)

			req := httptest.NewRequest(http.MethodPost, "/import/users", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestMigrationHandler_ImportUsers_ShouldReturn400_WhenInvalidAppID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/import/users", fix.handler.ImportUsers)

	body := `{"application_id":"not-a-uuid","users":[{"email":"a@test.com","password":"Test123!"}]}`
	req := httptest.NewRequest(http.MethodPost, "/import/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid application ID")
}

func TestMigrationHandler_ImportUsers_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	appID := uuid.New()
	fix.migrationSvc.ImportUsersFunc = func(id uuid.UUID, entries []models.ImportUserEntry) (*models.ImportResult, error) {
		return nil, models.NewAppError(http.StatusInternalServerError, "import failed")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/import/users", fix.handler.ImportUsers)

	body := fmt.Sprintf(`{"application_id":"%s","users":[{"email":"a@test.com","password":"Test123!"}]}`, appID.String())
	req := httptest.NewRequest(http.MethodPost, "/import/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMigrationHandler_ImportUsers_ShouldReturn200_WhenMultipleUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	appID := uuid.New()
	fix.migrationSvc.ImportUsersFunc = func(id uuid.UUID, entries []models.ImportUserEntry) (*models.ImportResult, error) {
		return &models.ImportResult{
			Total:   len(entries),
			Created: len(entries),
			Skipped: 0,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/import/users", fix.handler.ImportUsers)

	body := fmt.Sprintf(`{"application_id":"%s","users":[{"email":"a@test.com","password":"Test123!"},{"email":"b@test.com","password":"Test456!"}]}`, appID.String())
	req := httptest.NewRequest(http.MethodPost, "/import/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.ImportResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 2, result.Created)
}

// ===========================================================================
// ImportOAuthAccounts
// ===========================================================================

func TestMigrationHandler_ImportOAuthAccounts_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	fix.migrationSvc.ImportOAuthAccountsFunc = func(entries []models.ImportOAuthEntry) (*models.ImportResult, error) {
		assert.Len(t, entries, 1)
		assert.Equal(t, "google", entries[0].Provider)
		return sampleImportResult(), nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/import/oauth", fix.handler.ImportOAuthAccounts)

	body := `{"accounts":[{"email":"a@test.com","provider":"google","provider_user_id":"123"}]}`
	req := httptest.NewRequest(http.MethodPost, "/import/oauth", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.ImportResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Created)
}

func TestMigrationHandler_ImportOAuthAccounts_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing accounts", `{}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/import/oauth", fix.handler.ImportOAuthAccounts)

			req := httptest.NewRequest(http.MethodPost, "/import/oauth", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestMigrationHandler_ImportOAuthAccounts_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	fix.migrationSvc.ImportOAuthAccountsFunc = func(entries []models.ImportOAuthEntry) (*models.ImportResult, error) {
		return nil, models.NewAppError(http.StatusInternalServerError, "import failed")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/import/oauth", fix.handler.ImportOAuthAccounts)

	body := `{"accounts":[{"email":"a@test.com","provider":"google","provider_user_id":"123"}]}`
	req := httptest.NewRequest(http.MethodPost, "/import/oauth", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMigrationHandler_ImportOAuthAccounts_ShouldReturn200_WhenMultipleAccounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	fix.migrationSvc.ImportOAuthAccountsFunc = func(entries []models.ImportOAuthEntry) (*models.ImportResult, error) {
		return &models.ImportResult{
			Total:   len(entries),
			Created: len(entries),
			Skipped: 0,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/import/oauth", fix.handler.ImportOAuthAccounts)

	body := `{"accounts":[{"email":"a@test.com","provider":"google","provider_user_id":"123"},{"email":"b@test.com","provider":"github","provider_user_id":"456"}]}`
	req := httptest.NewRequest(http.MethodPost, "/import/oauth", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.ImportResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 2, result.Created)
}

// ===========================================================================
// ImportRoles
// ===========================================================================

func TestMigrationHandler_ImportRoles_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	appID := uuid.New()
	fix.migrationSvc.ImportRolesFunc = func(id uuid.UUID, entries []models.ImportRoleEntry) (*models.ImportResult, error) {
		assert.Equal(t, appID, id)
		assert.Len(t, entries, 1)
		assert.Equal(t, "editor", entries[0].Name)
		return sampleImportResult(), nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/import/roles", fix.handler.ImportRoles)

	body := fmt.Sprintf(`{"application_id":"%s","roles":[{"name":"editor"}]}`, appID.String())
	req := httptest.NewRequest(http.MethodPost, "/import/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.ImportResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Created)
}

func TestMigrationHandler_ImportRoles_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing application_id", `{"roles":[{"name":"editor"}]}`},
		{"missing roles", fmt.Sprintf(`{"application_id":"%s"}`, uuid.New().String())},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/import/roles", fix.handler.ImportRoles)

			req := httptest.NewRequest(http.MethodPost, "/import/roles", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestMigrationHandler_ImportRoles_ShouldReturn400_WhenInvalidAppID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/import/roles", fix.handler.ImportRoles)

	body := `{"application_id":"not-a-uuid","roles":[{"name":"editor"}]}`
	req := httptest.NewRequest(http.MethodPost, "/import/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid application ID")
}

func TestMigrationHandler_ImportRoles_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	appID := uuid.New()
	fix.migrationSvc.ImportRolesFunc = func(id uuid.UUID, entries []models.ImportRoleEntry) (*models.ImportResult, error) {
		return nil, models.NewAppError(http.StatusInternalServerError, "import failed")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/import/roles", fix.handler.ImportRoles)

	body := fmt.Sprintf(`{"application_id":"%s","roles":[{"name":"editor"}]}`, appID.String())
	req := httptest.NewRequest(http.MethodPost, "/import/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMigrationHandler_ImportRoles_ShouldReturn200_WhenMultipleRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupMigrationTestFixture()

	appID := uuid.New()
	fix.migrationSvc.ImportRolesFunc = func(id uuid.UUID, entries []models.ImportRoleEntry) (*models.ImportResult, error) {
		return &models.ImportResult{
			Total:   len(entries),
			Created: len(entries),
			Skipped: 0,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/import/roles", fix.handler.ImportRoles)

	body := fmt.Sprintf(`{"application_id":"%s","roles":[{"name":"editor"},{"name":"viewer"}]}`, appID.String())
	req := httptest.NewRequest(http.MethodPost, "/import/roles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.ImportResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 2, result.Created)
}
