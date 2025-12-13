package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protowire"
)

func init() {
	encoding.RegisterCodec(protoCodec{})
}

// protoCodec implements a custom codec for proper protobuf wire encoding
type protoCodec struct{}

func (protoCodec) Name() string { return "proto" }

func (protoCodec) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case *ValidateTokenRequest:
		return marshalValidateTokenRequest(m), nil
	case *ValidateTokenResponse:
		return marshalValidateTokenResponse(m), nil
	case *GetUserRequest:
		return marshalGetUserRequest(m), nil
	case *GetUserResponse:
		return marshalGetUserResponse(m), nil
	case *CheckPermissionRequest:
		return marshalCheckPermissionRequest(m), nil
	case *CheckPermissionResponse:
		return marshalCheckPermissionResponse(m), nil
	case *IntrospectTokenRequest:
		return marshalIntrospectTokenRequest(m), nil
	case *IntrospectTokenResponse:
		return marshalIntrospectTokenResponse(m), nil
	default:
		return nil, fmt.Errorf("unknown type: %T", v)
	}
}

func (protoCodec) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case *ValidateTokenRequest:
		return unmarshalValidateTokenRequest(data, m)
	case *ValidateTokenResponse:
		return unmarshalValidateTokenResponse(data, m)
	case *GetUserRequest:
		return unmarshalGetUserRequest(data, m)
	case *GetUserResponse:
		return unmarshalGetUserResponse(data, m)
	case *CheckPermissionRequest:
		return unmarshalCheckPermissionRequest(data, m)
	case *CheckPermissionResponse:
		return unmarshalCheckPermissionResponse(data, m)
	case *IntrospectTokenRequest:
		return unmarshalIntrospectTokenRequest(data, m)
	case *IntrospectTokenResponse:
		return unmarshalIntrospectTokenResponse(data, m)
	default:
		return fmt.Errorf("unknown type: %T", v)
	}
}

// Message types
type ValidateTokenRequest struct {
	AccessToken string
}

func (*ValidateTokenRequest) Reset()         {}
func (*ValidateTokenRequest) String() string { return "" }
func (*ValidateTokenRequest) ProtoMessage()  {}

type ValidateTokenResponse struct {
	Valid        bool
	UserId       string
	Email        string
	Username     string
	Roles        []string
	ErrorMessage string
	ExpiresAt    int64
	IsActive     bool
}

func (*ValidateTokenResponse) Reset()         {}
func (*ValidateTokenResponse) String() string { return "" }
func (*ValidateTokenResponse) ProtoMessage()  {}

type GetUserRequest struct {
	UserId string
}

func (*GetUserRequest) Reset()         {}
func (*GetUserRequest) String() string { return "" }
func (*GetUserRequest) ProtoMessage()  {}

type RoleInfo struct {
	Id          string
	Name        string
	DisplayName string
}

type User struct {
	Id                string
	Email             string
	Username          string
	FullName          string
	ProfilePictureUrl string
	Roles             []RoleInfo
	EmailVerified     bool
	IsActive          bool
	CreatedAt         int64
	UpdatedAt         int64
}

type GetUserResponse struct {
	User         *User
	ErrorMessage string
}

func (*GetUserResponse) Reset()         {}
func (*GetUserResponse) String() string { return "" }
func (*GetUserResponse) ProtoMessage()  {}

type CheckPermissionRequest struct {
	UserId   string
	Resource string
	Action   string
}

func (*CheckPermissionRequest) Reset()         {}
func (*CheckPermissionRequest) String() string { return "" }
func (*CheckPermissionRequest) ProtoMessage()  {}

type CheckPermissionResponse struct {
	Allowed      bool
	Roles        []string
	ErrorMessage string
}

func (*CheckPermissionResponse) Reset()         {}
func (*CheckPermissionResponse) String() string { return "" }
func (*CheckPermissionResponse) ProtoMessage()  {}

type IntrospectTokenRequest struct {
	AccessToken string
}

func (*IntrospectTokenRequest) Reset()         {}
func (*IntrospectTokenRequest) String() string { return "" }
func (*IntrospectTokenRequest) ProtoMessage()  {}

type IntrospectTokenResponse struct {
	Active       bool
	UserId       string
	Email        string
	Username     string
	Roles        []string
	IssuedAt     int64
	ExpiresAt    int64
	NotBefore    int64
	Subject      string
	Blacklisted  bool
	ErrorMessage string
}

func (*IntrospectTokenResponse) Reset()         {}
func (*IntrospectTokenResponse) String() string { return "" }
func (*IntrospectTokenResponse) ProtoMessage()  {}

// Marshal functions
func marshalValidateTokenRequest(m *ValidateTokenRequest) []byte {
	var b []byte
	if m.AccessToken != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, m.AccessToken)
	}
	return b
}

func marshalValidateTokenResponse(m *ValidateTokenResponse) []byte {
	var b []byte
	if m.Valid {
		b = protowire.AppendTag(b, 1, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
	}
	if m.UserId != "" {
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendString(b, m.UserId)
	}
	if m.Email != "" {
		b = protowire.AppendTag(b, 3, protowire.BytesType)
		b = protowire.AppendString(b, m.Email)
	}
	if m.Username != "" {
		b = protowire.AppendTag(b, 4, protowire.BytesType)
		b = protowire.AppendString(b, m.Username)
	}
	for _, role := range m.Roles {
		b = protowire.AppendTag(b, 5, protowire.BytesType)
		b = protowire.AppendString(b, role)
	}
	if m.ErrorMessage != "" {
		b = protowire.AppendTag(b, 6, protowire.BytesType)
		b = protowire.AppendString(b, m.ErrorMessage)
	}
	if m.ExpiresAt != 0 {
		b = protowire.AppendTag(b, 7, protowire.VarintType)
		b = protowire.AppendVarint(b, uint64(m.ExpiresAt))
	}
	if m.IsActive {
		b = protowire.AppendTag(b, 8, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
	}
	return b
}

func marshalGetUserRequest(m *GetUserRequest) []byte {
	var b []byte
	if m.UserId != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, m.UserId)
	}
	return b
}

func marshalGetUserResponse(m *GetUserResponse) []byte {
	var b []byte
	if m.User != nil {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		userBytes := marshalUser(m.User)
		b = protowire.AppendBytes(b, userBytes)
	}
	if m.ErrorMessage != "" {
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendString(b, m.ErrorMessage)
	}
	return b
}

func marshalUser(u *User) []byte {
	var b []byte
	if u.Id != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, u.Id)
	}
	if u.Email != "" {
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendString(b, u.Email)
	}
	if u.Username != "" {
		b = protowire.AppendTag(b, 3, protowire.BytesType)
		b = protowire.AppendString(b, u.Username)
	}
	if u.FullName != "" {
		b = protowire.AppendTag(b, 4, protowire.BytesType)
		b = protowire.AppendString(b, u.FullName)
	}
	if u.ProfilePictureUrl != "" {
		b = protowire.AppendTag(b, 5, protowire.BytesType)
		b = protowire.AppendString(b, u.ProfilePictureUrl)
	}
	for _, role := range u.Roles {
		b = protowire.AppendTag(b, 6, protowire.BytesType)
		roleBytes := marshalRoleInfo(&role)
		b = protowire.AppendBytes(b, roleBytes)
	}
	if u.EmailVerified {
		b = protowire.AppendTag(b, 7, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
	}
	if u.IsActive {
		b = protowire.AppendTag(b, 8, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
	}
	if u.CreatedAt != 0 {
		b = protowire.AppendTag(b, 9, protowire.VarintType)
		b = protowire.AppendVarint(b, uint64(u.CreatedAt))
	}
	if u.UpdatedAt != 0 {
		b = protowire.AppendTag(b, 10, protowire.VarintType)
		b = protowire.AppendVarint(b, uint64(u.UpdatedAt))
	}
	return b
}

func marshalRoleInfo(r *RoleInfo) []byte {
	var b []byte
	if r.Id != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, r.Id)
	}
	if r.Name != "" {
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendString(b, r.Name)
	}
	if r.DisplayName != "" {
		b = protowire.AppendTag(b, 3, protowire.BytesType)
		b = protowire.AppendString(b, r.DisplayName)
	}
	return b
}

func marshalCheckPermissionRequest(m *CheckPermissionRequest) []byte {
	var b []byte
	if m.UserId != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, m.UserId)
	}
	if m.Resource != "" {
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendString(b, m.Resource)
	}
	if m.Action != "" {
		b = protowire.AppendTag(b, 3, protowire.BytesType)
		b = protowire.AppendString(b, m.Action)
	}
	return b
}

func marshalCheckPermissionResponse(m *CheckPermissionResponse) []byte {
	var b []byte
	if m.Allowed {
		b = protowire.AppendTag(b, 1, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
	}
	for _, role := range m.Roles {
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendString(b, role)
	}
	if m.ErrorMessage != "" {
		b = protowire.AppendTag(b, 3, protowire.BytesType)
		b = protowire.AppendString(b, m.ErrorMessage)
	}
	return b
}

func marshalIntrospectTokenRequest(m *IntrospectTokenRequest) []byte {
	var b []byte
	if m.AccessToken != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, m.AccessToken)
	}
	return b
}

func marshalIntrospectTokenResponse(m *IntrospectTokenResponse) []byte {
	var b []byte
	if m.Active {
		b = protowire.AppendTag(b, 1, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
	}
	if m.UserId != "" {
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendString(b, m.UserId)
	}
	if m.Email != "" {
		b = protowire.AppendTag(b, 3, protowire.BytesType)
		b = protowire.AppendString(b, m.Email)
	}
	if m.Username != "" {
		b = protowire.AppendTag(b, 4, protowire.BytesType)
		b = protowire.AppendString(b, m.Username)
	}
	for _, role := range m.Roles {
		b = protowire.AppendTag(b, 5, protowire.BytesType)
		b = protowire.AppendString(b, role)
	}
	if m.IssuedAt != 0 {
		b = protowire.AppendTag(b, 6, protowire.VarintType)
		b = protowire.AppendVarint(b, uint64(m.IssuedAt))
	}
	if m.ExpiresAt != 0 {
		b = protowire.AppendTag(b, 7, protowire.VarintType)
		b = protowire.AppendVarint(b, uint64(m.ExpiresAt))
	}
	if m.NotBefore != 0 {
		b = protowire.AppendTag(b, 8, protowire.VarintType)
		b = protowire.AppendVarint(b, uint64(m.NotBefore))
	}
	if m.Subject != "" {
		b = protowire.AppendTag(b, 9, protowire.BytesType)
		b = protowire.AppendString(b, m.Subject)
	}
	if m.Blacklisted {
		b = protowire.AppendTag(b, 10, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
	}
	if m.ErrorMessage != "" {
		b = protowire.AppendTag(b, 11, protowire.BytesType)
		b = protowire.AppendString(b, m.ErrorMessage)
	}
	return b
}

// Unmarshal functions
func unmarshalValidateTokenRequest(data []byte, m *ValidateTokenRequest) error {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch num {
		case 1:
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.AccessToken = v
			data = data[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				return fmt.Errorf("invalid field")
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalValidateTokenResponse(data []byte, m *ValidateTokenResponse) error {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch num {
		case 1:
			v, n := protowire.ConsumeVarint(data)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			m.Valid = v != 0
			data = data[n:]
		case 2:
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.UserId = v
			data = data[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				return fmt.Errorf("invalid field")
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalGetUserRequest(data []byte, m *GetUserRequest) error {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch num {
		case 1:
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.UserId = v
			data = data[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				return fmt.Errorf("invalid field")
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalGetUserResponse(data []byte, m *GetUserResponse) error {
	return nil // Simplified - not needed for server
}

func unmarshalCheckPermissionRequest(data []byte, m *CheckPermissionRequest) error {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch num {
		case 1:
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.UserId = v
			data = data[n:]
		case 2:
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.Resource = v
			data = data[n:]
		case 3:
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.Action = v
			data = data[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				return fmt.Errorf("invalid field")
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalCheckPermissionResponse(data []byte, m *CheckPermissionResponse) error {
	return nil // Simplified - not needed for server
}

func unmarshalIntrospectTokenRequest(data []byte, m *IntrospectTokenRequest) error {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		data = data[n:]
		switch num {
		case 1:
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.AccessToken = v
			data = data[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				return fmt.Errorf("invalid field")
			}
			data = data[n:]
		}
	}
	return nil
}

func unmarshalIntrospectTokenResponse(data []byte, m *IntrospectTokenResponse) error {
	return nil // Simplified - not needed for server
}

// AuthServiceServer interface
type AuthServiceServer interface {
	ValidateToken(context.Context, *ValidateTokenRequest) (*ValidateTokenResponse, error)
	GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
	CheckPermission(context.Context, *CheckPermissionRequest) (*CheckPermissionResponse, error)
	IntrospectToken(context.Context, *IntrospectTokenRequest) (*IntrospectTokenResponse, error)
	mustEmbedUnimplementedAuthServiceServer()
}

type UnimplementedAuthServiceServer struct{}

func (UnimplementedAuthServiceServer) ValidateToken(context.Context, *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateToken not implemented")
}

func (UnimplementedAuthServiceServer) GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}

func (UnimplementedAuthServiceServer) CheckPermission(context.Context, *CheckPermissionRequest) (*CheckPermissionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckPermission not implemented")
}

func (UnimplementedAuthServiceServer) IntrospectToken(context.Context, *IntrospectTokenRequest) (*IntrospectTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IntrospectToken not implemented")
}

func (UnimplementedAuthServiceServer) mustEmbedUnimplementedAuthServiceServer() {}

// MockAuthHandler is a mock implementation for testing
type MockAuthHandler struct {
	UnimplementedAuthServiceServer
	users map[string]*User
}

func NewMockAuthHandler() *MockAuthHandler {
	now := time.Now().Unix()
	mockUserID := uuid.New().String()
	adminUserID := uuid.New().String()

	return &MockAuthHandler{
		users: map[string]*User{
			mockUserID: {
				Id:            mockUserID,
				Email:         "testuser@example.com",
				Username:      "testuser",
				FullName:      "Test User",
				EmailVerified: true,
				IsActive:      true,
				Roles: []RoleInfo{
					{Id: uuid.New().String(), Name: "user", DisplayName: "User"},
				},
				CreatedAt: now - 86400,
				UpdatedAt: now,
			},
			adminUserID: {
				Id:            adminUserID,
				Email:         "admin@example.com",
				Username:      "admin",
				FullName:      "Admin User",
				EmailVerified: true,
				IsActive:      true,
				Roles: []RoleInfo{
					{Id: uuid.New().String(), Name: "admin", DisplayName: "Administrator"},
					{Id: uuid.New().String(), Name: "user", DisplayName: "User"},
				},
				CreatedAt: now - 86400*7,
				UpdatedAt: now,
			},
		},
	}
}

func (h *MockAuthHandler) ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	log.Printf("[ValidateToken] Received request for token: %s", truncateToken(req.AccessToken))

	if req.AccessToken == "" {
		return &ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "access_token is required",
		}, nil
	}

	// Check if it's an API key
	if len(req.AccessToken) > 4 && req.AccessToken[:4] == "agw_" {
		if req.AccessToken == "agw_valid_test_key" {
			user := h.getFirstUser()
			return &ValidateTokenResponse{
				Valid:    true,
				UserId:   user.Id,
				Email:    user.Email,
				Username: user.Username,
				Roles:    extractRoleNames(user.Roles),
				IsActive: true,
			}, nil
		}
		return &ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "invalid API key",
		}, nil
	}

	// Mock JWT validation - accept tokens starting with "valid_"
	if len(req.AccessToken) > 6 && req.AccessToken[:6] == "valid_" {
		user := h.getFirstUser()
		return &ValidateTokenResponse{
			Valid:     true,
			UserId:    user.Id,
			Email:     user.Email,
			Username:  user.Username,
			Roles:     extractRoleNames(user.Roles),
			ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
			IsActive:  true,
		}, nil
	}

	return &ValidateTokenResponse{
		Valid:        false,
		ErrorMessage: "invalid or expired token",
	}, nil
}

func (h *MockAuthHandler) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	log.Printf("[GetUser] Received request for user ID: %s", req.UserId)

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if _, err := uuid.Parse(req.UserId); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	user, exists := h.users[req.UserId]
	if !exists {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &GetUserResponse{User: user}, nil
}

func (h *MockAuthHandler) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (*CheckPermissionResponse, error) {
	log.Printf("[CheckPermission] User: %s, Resource: %s, Action: %s", req.UserId, req.Resource, req.Action)

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if _, err := uuid.Parse(req.UserId); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	user, exists := h.users[req.UserId]
	if !exists {
		return &CheckPermissionResponse{
			Allowed:      false,
			Roles:        []string{},
			ErrorMessage: "user not found",
		}, nil
	}

	roles := extractRoleNames(user.Roles)

	for _, role := range user.Roles {
		if role.Name == "admin" {
			return &CheckPermissionResponse{
				Allowed: true,
				Roles:   roles,
			}, nil
		}
	}

	if req.Action == "read" {
		return &CheckPermissionResponse{
			Allowed: true,
			Roles:   roles,
		}, nil
	}

	return &CheckPermissionResponse{
		Allowed: false,
		Roles:   roles,
	}, nil
}

func (h *MockAuthHandler) IntrospectToken(ctx context.Context, req *IntrospectTokenRequest) (*IntrospectTokenResponse, error) {
	log.Printf("[IntrospectToken] Received request for token: %s", truncateToken(req.AccessToken))

	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	if len(req.AccessToken) > 6 && req.AccessToken[:6] == "valid_" {
		user := h.getFirstUser()
		now := time.Now()
		return &IntrospectTokenResponse{
			Active:      true,
			UserId:      user.Id,
			Email:       user.Email,
			Username:    user.Username,
			Roles:       extractRoleNames(user.Roles),
			IssuedAt:    now.Add(-5 * time.Minute).Unix(),
			ExpiresAt:   now.Add(10 * time.Minute).Unix(),
			NotBefore:   now.Add(-5 * time.Minute).Unix(),
			Subject:     user.Id,
			Blacklisted: false,
		}, nil
	}

	return &IntrospectTokenResponse{
		Active:       false,
		ErrorMessage: "invalid or expired token",
	}, nil
}

func (h *MockAuthHandler) getFirstUser() *User {
	for _, user := range h.users {
		return user
	}
	return nil
}

func extractRoleNames(roles []RoleInfo) []string {
	names := make([]string, len(roles))
	for i, role := range roles {
		names[i] = role.Name
	}
	return names
}

func truncateToken(token string) string {
	if len(token) <= 20 {
		return token
	}
	return token[:20] + "..."
}

// Service registration
func RegisterAuthServiceServer(s grpc.ServiceRegistrar, srv AuthServiceServer) {
	s.RegisterService(&AuthService_ServiceDesc, srv)
}

var AuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "auth.AuthService",
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "ValidateToken", Handler: _AuthService_ValidateToken_Handler},
		{MethodName: "GetUser", Handler: _AuthService_GetUser_Handler},
		{MethodName: "CheckPermission", Handler: _AuthService_CheckPermission_Handler},
		{MethodName: "IntrospectToken", Handler: _AuthService_IntrospectToken_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/auth.proto",
}

func _AuthService_ValidateToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).ValidateToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/auth.AuthService/ValidateToken"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).ValidateToken(ctx, req.(*ValidateTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_GetUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).GetUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/auth.AuthService/GetUser"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).GetUser(ctx, req.(*GetUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_CheckPermission_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckPermissionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).CheckPermission(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/auth.AuthService/CheckPermission"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).CheckPermission(ctx, req.(*CheckPermissionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_IntrospectToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IntrospectTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).IntrospectToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/auth.AuthService/IntrospectToken"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).IntrospectToken(ctx, req.(*IntrospectTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func main() {
	port := flag.String("port", "50051", "gRPC server port")
	flag.Parse()

	fmt.Println("==========================================")
	fmt.Println("  Auth Gateway Mock gRPC Server")
	fmt.Println("==========================================")
	fmt.Printf("Starting server on port %s...\n\n", *port)

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	handler := NewMockAuthHandler()
	RegisterAuthServiceServer(grpcServer, handler)

	reflection.Register(grpcServer)

	fmt.Println("Mock Users Available:")
	fmt.Println("----------------------")
	for id, user := range handler.users {
		fmt.Printf("  ID: %s\n", id)
		fmt.Printf("  Email: %s\n", user.Email)
		fmt.Printf("  Username: %s\n", user.Username)
		fmt.Printf("  Roles: %v\n", extractRoleNames(user.Roles))
		fmt.Println()
	}

	fmt.Println("Valid Test Tokens:")
	fmt.Println("------------------")
	fmt.Println("  JWT:     valid_test_token")
	fmt.Println("  API Key: agw_valid_test_key")
	fmt.Println()

	fmt.Printf("gRPC Server listening on :%s\n", *port)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")
	grpcServer.GracefulStop()
	fmt.Println("Server stopped")
}
