package grpcclient

import (
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/encoding/protowire"
)

func init() {
	encoding.RegisterCodec(AuthCodec{})
}

// AuthCodec is a custom codec for auth messages
type AuthCodec struct{}

func (AuthCodec) Name() string {
	return "proto"
}

func (AuthCodec) Marshal(v interface{}) ([]byte, error) {
	switch msg := v.(type) {
	case *ValidateTokenRequest:
		return msg.MarshalBinary()
	case *GetUserRequest:
		return msg.MarshalBinary()
	case *CheckPermissionRequest:
		return msg.MarshalBinary()
	case *IntrospectTokenRequest:
		return msg.MarshalBinary()
	default:
		// For unknown types, return empty
		return nil, nil
	}
}

func (AuthCodec) Unmarshal(data []byte, v interface{}) error {
	switch msg := v.(type) {
	case *ValidateTokenResponse:
		return msg.UnmarshalBinary(data)
	case *GetUserResponse:
		return msg.UnmarshalBinary(data)
	case *CheckPermissionResponse:
		return msg.UnmarshalBinary(data)
	case *IntrospectTokenResponse:
		return msg.UnmarshalBinary(data)
	default:
		return nil
	}
}

// marshalValidateTokenRequest marshals ValidateTokenRequest to protobuf wire format
func marshalValidateTokenRequest(req *ValidateTokenRequest) []byte {
	var b []byte
	if req.AccessToken != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, req.AccessToken)
	}
	return b
}

// marshalGetUserRequest marshals GetUserRequest to protobuf wire format
func marshalGetUserRequest(req *GetUserRequest) []byte {
	var b []byte
	if req.UserId != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, req.UserId)
	}
	return b
}

// marshalCheckPermissionRequest marshals CheckPermissionRequest to protobuf wire format
func marshalCheckPermissionRequest(req *CheckPermissionRequest) []byte {
	var b []byte
	if req.UserId != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, req.UserId)
	}
	if req.Resource != "" {
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendString(b, req.Resource)
	}
	if req.Action != "" {
		b = protowire.AppendTag(b, 3, protowire.BytesType)
		b = protowire.AppendString(b, req.Action)
	}
	return b
}

// marshalIntrospectTokenRequest marshals IntrospectTokenRequest to protobuf wire format
func marshalIntrospectTokenRequest(req *IntrospectTokenRequest) []byte {
	var b []byte
	if req.AccessToken != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, req.AccessToken)
	}
	return b
}
