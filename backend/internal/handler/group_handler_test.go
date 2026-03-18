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
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

type groupTestFixture struct {
	handler  *GroupHandler
	groupSvc *mockGroupServicer
}

func setupGroupTestFixture() *groupTestFixture {
	svc := &mockGroupServicer{}
	h := NewGroupHandler(svc, testLogger())
	return &groupTestFixture{handler: h, groupSvc: svc}
}

func sampleGroup() *models.Group {
	return &models.Group{
		ID:          uuid.New(),
		Name:        "dev-team",
		DisplayName: "Dev Team",
		Description: "Development team",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ===========================================================================
// CreateGroup
// ===========================================================================

func TestGroupHandler_CreateGroup_ShouldReturn201_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	grp := sampleGroup()

	fix.groupSvc.CreateGroupFunc = func(req *models.CreateGroupRequest) (*models.Group, error) {
		assert.Equal(t, "dev-team", req.Name)
		assert.Equal(t, "Dev Team", req.DisplayName)
		return grp, nil
	}
	fix.groupSvc.GetGroupMemberCountFunc = func(groupID uuid.UUID) (int, error) {
		return 0, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/groups", fix.handler.CreateGroup)

	body := `{"name":"dev-team","display_name":"Dev Team"}`
	req := httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.GroupResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "dev-team", resp.Name)
	assert.Equal(t, "Dev Team", resp.DisplayName)
	assert.Equal(t, 0, resp.MemberCount)
}

func TestGroupHandler_CreateGroup_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing name", `{"display_name":"Dev Team"}`},
		{"missing display_name", `{"name":"dev-team"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/groups", fix.handler.CreateGroup)

			req := httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestGroupHandler_CreateGroup_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	fix.groupSvc.CreateGroupFunc = func(req *models.CreateGroupRequest) (*models.Group, error) {
		return nil, models.NewAppError(http.StatusConflict, "Group name already exists")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/groups", fix.handler.CreateGroup)

	body := `{"name":"dev-team","display_name":"Dev Team"}`
	req := httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestGroupHandler_CreateGroup_ShouldReturn201_WithMemberCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	grp := sampleGroup()

	fix.groupSvc.CreateGroupFunc = func(req *models.CreateGroupRequest) (*models.Group, error) {
		return grp, nil
	}
	fix.groupSvc.GetGroupMemberCountFunc = func(groupID uuid.UUID) (int, error) {
		return 5, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/groups", fix.handler.CreateGroup)

	body := `{"name":"dev-team","display_name":"Dev Team"}`
	req := httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.GroupResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 5, resp.MemberCount)
}

// ===========================================================================
// GetGroup
// ===========================================================================

func TestGroupHandler_GetGroup_ShouldReturn200_WhenFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	grp := sampleGroup()

	fix.groupSvc.GetGroupFunc = func(id uuid.UUID) (*models.Group, error) {
		assert.Equal(t, grp.ID, id)
		return grp, nil
	}
	fix.groupSvc.GetGroupMemberCountFunc = func(groupID uuid.UUID) (int, error) {
		return 3, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/groups/:id", fix.handler.GetGroup)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+grp.ID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.GroupResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, grp.ID, resp.ID)
	assert.Equal(t, "dev-team", resp.Name)
	assert.Equal(t, 3, resp.MemberCount)
}

func TestGroupHandler_GetGroup_ShouldReturnError_WhenNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	fix.groupSvc.GetGroupFunc = func(id uuid.UUID) (*models.Group, error) {
		return nil, models.NewAppError(http.StatusNotFound, "Group not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/groups/:id", fix.handler.GetGroup)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGroupHandler_GetGroup_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/groups/:id", fix.handler.GetGroup)

	req := httptest.NewRequest(http.MethodGet, "/groups/bad-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===========================================================================
// ListGroups
// ===========================================================================

func TestGroupHandler_ListGroups_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	fix.groupSvc.ListGroupsFunc = func(page, pageSize int) (*models.GroupListResponse, error) {
		return &models.GroupListResponse{
			Groups: []models.GroupResponse{
				{
					ID:          uuid.New(),
					Name:        "dev-team",
					DisplayName: "Dev Team",
					MemberCount: 5,
				},
			},
			Total:      1,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: 1,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/groups", fix.handler.ListGroups)

	req := httptest.NewRequest(http.MethodGet, "/groups?page=1&page_size=20", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.GroupListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Groups, 1)
}

func TestGroupHandler_ListGroups_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	fix.groupSvc.ListGroupsFunc = func(page, pageSize int) (*models.GroupListResponse, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/groups", fix.handler.ListGroups)

	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGroupHandler_ListGroups_ShouldUseDefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	var capturedPage, capturedPageSize int
	fix.groupSvc.ListGroupsFunc = func(page, pageSize int) (*models.GroupListResponse, error) {
		capturedPage = page
		capturedPageSize = pageSize
		return &models.GroupListResponse{
			Groups:     []models.GroupResponse{},
			Total:      0,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: 0,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/groups", fix.handler.ListGroups)

	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, capturedPage)
	assert.Equal(t, 20, capturedPageSize)
}

// ===========================================================================
// UpdateGroup
// ===========================================================================

func TestGroupHandler_UpdateGroup_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	grp := sampleGroup()
	grp.DisplayName = "Updated Team"

	fix.groupSvc.UpdateGroupFunc = func(id uuid.UUID, req *models.UpdateGroupRequest) (*models.Group, error) {
		assert.Equal(t, grp.ID, id)
		assert.Equal(t, "Updated Team", *req.DisplayName)
		return grp, nil
	}
	fix.groupSvc.GetGroupMemberCountFunc = func(groupID uuid.UUID) (int, error) {
		return 2, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/groups/:id", fix.handler.UpdateGroup)

	body := `{"display_name":"Updated Team"}`
	req := httptest.NewRequest(http.MethodPut, "/groups/"+grp.ID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.GroupResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Updated Team", resp.DisplayName)
	assert.Equal(t, 2, resp.MemberCount)
}

func TestGroupHandler_UpdateGroup_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/groups/:id", fix.handler.UpdateGroup)

	req := httptest.NewRequest(http.MethodPut, "/groups/"+uuid.New().String(), strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGroupHandler_UpdateGroup_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/groups/:id", fix.handler.UpdateGroup)

	body := `{"display_name":"Updated Team"}`
	req := httptest.NewRequest(http.MethodPut, "/groups/bad-uuid", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGroupHandler_UpdateGroup_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	fix.groupSvc.UpdateGroupFunc = func(id uuid.UUID, req *models.UpdateGroupRequest) (*models.Group, error) {
		return nil, models.NewAppError(http.StatusNotFound, "Group not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/groups/:id", fix.handler.UpdateGroup)

	body := `{"display_name":"Updated Team"}`
	req := httptest.NewRequest(http.MethodPut, "/groups/"+uuid.New().String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ===========================================================================
// DeleteGroup
// ===========================================================================

func TestGroupHandler_DeleteGroup_ShouldReturn204_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	groupID := uuid.New()
	fix.groupSvc.DeleteGroupFunc = func(id uuid.UUID) error {
		assert.Equal(t, groupID, id)
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/groups/:id", fix.handler.DeleteGroup)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestGroupHandler_DeleteGroup_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	fix.groupSvc.DeleteGroupFunc = func(id uuid.UUID) error {
		return models.NewAppError(http.StatusBadRequest, "Cannot delete group with members")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/groups/:id", fix.handler.DeleteGroup)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGroupHandler_DeleteGroup_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/groups/:id", fix.handler.DeleteGroup)

	req := httptest.NewRequest(http.MethodDelete, "/groups/bad-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGroupHandler_DeleteGroup_ShouldReturnNotFound_WhenGroupDoesNotExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	fix.groupSvc.DeleteGroupFunc = func(id uuid.UUID) error {
		return models.NewAppError(http.StatusNotFound, "Group not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/groups/:id", fix.handler.DeleteGroup)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ===========================================================================
// AddGroupMembers
// ===========================================================================

func TestGroupHandler_AddGroupMembers_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	groupID := uuid.New()
	userID := uuid.New()

	fix.groupSvc.AddGroupMembersFunc = func(gID uuid.UUID, userIDs []uuid.UUID) error {
		assert.Equal(t, groupID, gID)
		assert.Len(t, userIDs, 1)
		assert.Equal(t, userID, userIDs[0])
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/groups/:id/members", fix.handler.AddGroupMembers)

	body := fmt.Sprintf(`{"user_ids":["%s"]}`, userID.String())
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID.String()+"/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Users added to group successfully", resp.Message)
}

func TestGroupHandler_AddGroupMembers_ShouldReturn400_WhenInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid JSON", `{invalid}`},
		{"missing user_ids", `{}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/groups/:id/members", fix.handler.AddGroupMembers)

			req := httptest.NewRequest(http.MethodPost, "/groups/"+uuid.New().String()+"/members", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestGroupHandler_AddGroupMembers_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/groups/:id/members", fix.handler.AddGroupMembers)

	body := fmt.Sprintf(`{"user_ids":["%s"]}`, uuid.New().String())
	req := httptest.NewRequest(http.MethodPost, "/groups/bad-uuid/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGroupHandler_AddGroupMembers_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	fix.groupSvc.AddGroupMembersFunc = func(gID uuid.UUID, userIDs []uuid.UUID) error {
		return models.NewAppError(http.StatusNotFound, "Group not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/groups/:id/members", fix.handler.AddGroupMembers)

	body := fmt.Sprintf(`{"user_ids":["%s"]}`, uuid.New().String())
	req := httptest.NewRequest(http.MethodPost, "/groups/"+uuid.New().String()+"/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ===========================================================================
// RemoveGroupMember
// ===========================================================================

func TestGroupHandler_RemoveGroupMember_ShouldReturn204_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	groupID := uuid.New()
	userID := uuid.New()

	fix.groupSvc.RemoveGroupMemberFunc = func(gID, uID uuid.UUID) error {
		assert.Equal(t, groupID, gID)
		assert.Equal(t, userID, uID)
		return nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/groups/:id/members/:user_id", fix.handler.RemoveGroupMember)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/groups/%s/members/%s", groupID.String(), userID.String()), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestGroupHandler_RemoveGroupMember_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	fix.groupSvc.RemoveGroupMemberFunc = func(gID, uID uuid.UUID) error {
		return models.NewAppError(http.StatusNotFound, "Member not found in group")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/groups/:id/members/:user_id", fix.handler.RemoveGroupMember)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/groups/%s/members/%s", uuid.New().String(), uuid.New().String()), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGroupHandler_RemoveGroupMember_ShouldReturn400_WhenInvalidGroupUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/groups/:id/members/:user_id", fix.handler.RemoveGroupMember)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/groups/bad-uuid/members/%s", uuid.New().String()), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGroupHandler_RemoveGroupMember_ShouldReturn400_WhenInvalidUserUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/groups/:id/members/:user_id", fix.handler.RemoveGroupMember)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/groups/%s/members/bad-uuid", uuid.New().String()), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===========================================================================
// GetGroupMembers
// ===========================================================================

func TestGroupHandler_GetGroupMembers_ShouldReturn200_WhenSuccessful(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	groupID := uuid.New()
	userID := uuid.New()

	fix.groupSvc.GetGroupMembersFunc = func(gID uuid.UUID, page, pageSize int) ([]*models.User, int, error) {
		assert.Equal(t, groupID, gID)
		return []*models.User{
			{
				ID:       userID,
				Email:    "user@test.com",
				Username: "testuser",
				IsActive: true,
			},
		}, 1, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/groups/:id/members", fix.handler.GetGroupMembers)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+groupID.String()+"/members?page=1&page_size=20", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.AdminUserListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Users, 1)
	assert.Equal(t, "user@test.com", resp.Users[0].Email)
}

func TestGroupHandler_GetGroupMembers_ShouldReturnError_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	fix.groupSvc.GetGroupMembersFunc = func(gID uuid.UUID, page, pageSize int) ([]*models.User, int, error) {
		return nil, 0, models.NewAppError(http.StatusNotFound, "Group not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/groups/:id/members", fix.handler.GetGroupMembers)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+uuid.New().String()+"/members", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGroupHandler_GetGroupMembers_ShouldReturn400_WhenInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/groups/:id/members", fix.handler.GetGroupMembers)

	req := httptest.NewRequest(http.MethodGet, "/groups/bad-uuid/members", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGroupHandler_GetGroupMembers_ShouldReturnEmptyList_WhenNoMembers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupGroupTestFixture()

	fix.groupSvc.GetGroupMembersFunc = func(gID uuid.UUID, page, pageSize int) ([]*models.User, int, error) {
		return []*models.User{}, 0, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/groups/:id/members", fix.handler.GetGroupMembers)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+uuid.New().String()+"/members", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.AdminUserListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Len(t, resp.Users, 0)
}
