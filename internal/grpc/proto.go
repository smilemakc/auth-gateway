package grpc

// This file contains manually created protobuf message and service definitions
// In production, these should be generated from .proto files using protoc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidateTokenRequest contains the token to validate
type ValidateTokenRequest struct {
	AccessToken string `json:"access_token"`
}

// ValidateTokenResponse contains validation result and user info
type ValidateTokenResponse struct {
	Valid        bool   `json:"valid"`
	UserId       string `json:"user_id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	ErrorMessage string `json:"error_message,omitempty"`
	ExpiresAt    int64  `json:"expires_at"`
}

// GetUserRequest contains the user ID to retrieve
type GetUserRequest struct {
	UserId string `json:"user_id"`
}

// User represents user information
type User struct {
	Id                string `json:"id"`
	Email             string `json:"email"`
	Username          string `json:"username"`
	FullName          string `json:"full_name"`
	ProfilePictureUrl string `json:"profile_picture_url"`
	Role              string `json:"role"`
	EmailVerified     bool   `json:"email_verified"`
	IsActive          bool   `json:"is_active"`
	CreatedAt         int64  `json:"created_at"`
	UpdatedAt         int64  `json:"updated_at"`
}

// GetUserResponse contains user information
type GetUserResponse struct {
	User         *User  `json:"user"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// CheckPermissionRequest contains permission check data
type CheckPermissionRequest struct {
	UserId   string `json:"user_id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// CheckPermissionResponse contains permission check result
type CheckPermissionResponse struct {
	Allowed      bool   `json:"allowed"`
	Role         string `json:"role"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// IntrospectTokenRequest contains the token to introspect
type IntrospectTokenRequest struct {
	AccessToken string `json:"access_token"`
}

// IntrospectTokenResponse contains detailed token information
type IntrospectTokenResponse struct {
	Active       bool   `json:"active"`
	UserId       string `json:"user_id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	IssuedAt     int64  `json:"issued_at"`
	ExpiresAt    int64  `json:"expires_at"`
	NotBefore    int64  `json:"not_before"`
	Subject      string `json:"subject"`
	Blacklisted  bool   `json:"blacklisted"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// AuthServiceServer is the server API for AuthService
type AuthServiceServer interface {
	ValidateToken(context.Context, *ValidateTokenRequest) (*ValidateTokenResponse, error)
	GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
	CheckPermission(context.Context, *CheckPermissionRequest) (*CheckPermissionResponse, error)
	IntrospectToken(context.Context, *IntrospectTokenRequest) (*IntrospectTokenResponse, error)
	mustEmbedUnimplementedAuthServiceServer()
}

// UnimplementedAuthServiceServer must be embedded to have forward compatible implementations.
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

// UnsafeAuthServiceServer may be embedded to opt out of forward compatibility for this service.
type UnsafeAuthServiceServer interface {
	mustEmbedUnimplementedAuthServiceServer()
}

// RegisterAuthServiceServer registers the service
func RegisterAuthServiceServer(s grpc.ServiceRegistrar, srv AuthServiceServer) {
	s.RegisterService(&AuthService_ServiceDesc, srv)
}

// AuthService_ServiceDesc is the grpc.ServiceDesc for AuthService service.
var AuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "auth.AuthService",
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ValidateToken",
			Handler:    _AuthService_ValidateToken_Handler,
		},
		{
			MethodName: "GetUser",
			Handler:    _AuthService_GetUser_Handler,
		},
		{
			MethodName: "CheckPermission",
			Handler:    _AuthService_CheckPermission_Handler,
		},
		{
			MethodName: "IntrospectToken",
			Handler:    _AuthService_IntrospectToken_Handler,
		},
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
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.AuthService/ValidateToken",
	}
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
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.AuthService/GetUser",
	}
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
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.AuthService/CheckPermission",
	}
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
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.AuthService/IntrospectToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).IntrospectToken(ctx, req.(*IntrospectTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}
