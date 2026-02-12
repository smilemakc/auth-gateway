package grpcclient

import (
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

// ValidateTokenRequest contains the token to validate
type ValidateTokenRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AccessToken string `protobuf:"bytes,1,opt,name=access_token,json=accessToken,proto3" json:"access_token,omitempty"`
}

func (x *ValidateTokenRequest) Reset()         { *x = ValidateTokenRequest{} }
func (x *ValidateTokenRequest) String() string { return x.AccessToken }
func (*ValidateTokenRequest) ProtoMessage()    {}

func (x *ValidateTokenRequest) ProtoReflect() protoreflect.Message {
	return nil
}

func (m *ValidateTokenRequest) MarshalBinary() ([]byte, error) {
	var b []byte
	if m.AccessToken != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, m.AccessToken)
	}
	return b, nil
}

// ValidateTokenResponse contains validation result and user info
type ValidateTokenResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Valid        bool     `protobuf:"varint,1,opt,name=valid,proto3" json:"valid,omitempty"`
	UserId       string   `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Email        string   `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	Username     string   `protobuf:"bytes,4,opt,name=username,proto3" json:"username,omitempty"`
	Roles        []string `protobuf:"bytes,5,rep,name=roles,proto3" json:"roles,omitempty"`
	ErrorMessage string   `protobuf:"bytes,6,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	ExpiresAt    int64    `protobuf:"varint,7,opt,name=expires_at,json=expiresAt,proto3" json:"expires_at,omitempty"`
	IsActive     bool     `protobuf:"varint,8,opt,name=is_active,json=isActive,proto3" json:"is_active,omitempty"`
}

func (x *ValidateTokenResponse) Reset()         { *x = ValidateTokenResponse{} }
func (x *ValidateTokenResponse) String() string { return "" }
func (*ValidateTokenResponse) ProtoMessage()    {}

func (x *ValidateTokenResponse) ProtoReflect() protoreflect.Message {
	return nil
}

func (m *ValidateTokenResponse) UnmarshalBinary(b []byte) error {
	for len(b) > 0 {
		fieldNum, wireType, n := protowire.ConsumeTag(b)
		if n < 0 {
			return protowire.ParseError(n)
		}
		b = b[n:]

		switch fieldNum {
		case 1: // valid (bool)
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Valid = v != 0
			b = b[n:]
		case 2: // user_id (string)
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.UserId = v
			b = b[n:]
		case 3: // email (string)
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Email = v
			b = b[n:]
		case 4: // username (string)
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Username = v
			b = b[n:]
		case 5: // roles (repeated string)
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Roles = append(m.Roles, v)
			b = b[n:]
		case 6: // error_message (string)
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.ErrorMessage = v
			b = b[n:]
		case 7: // expires_at (int64)
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.ExpiresAt = int64(v)
			b = b[n:]
		case 8: // is_active (bool)
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.IsActive = v != 0
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(fieldNum, wireType, b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			b = b[n:]
		}
	}
	return nil
}

// GetUserRequest contains the user ID to retrieve
type GetUserRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

func (x *GetUserRequest) Reset()         { *x = GetUserRequest{} }
func (x *GetUserRequest) String() string { return x.UserId }
func (*GetUserRequest) ProtoMessage()    {}

func (x *GetUserRequest) ProtoReflect() protoreflect.Message {
	return nil
}

func (m *GetUserRequest) MarshalBinary() ([]byte, error) {
	var b []byte
	if m.UserId != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, m.UserId)
	}
	return b, nil
}

// RoleInfo represents basic role information
type RoleInfo struct {
	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	DisplayName string `protobuf:"bytes,3,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
}

// User represents user information
type User struct {
	Id                string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Email             string     `protobuf:"bytes,2,opt,name=email,proto3" json:"email,omitempty"`
	Username          string     `protobuf:"bytes,3,opt,name=username,proto3" json:"username,omitempty"`
	FullName          string     `protobuf:"bytes,4,opt,name=full_name,json=fullName,proto3" json:"full_name,omitempty"`
	ProfilePictureUrl string     `protobuf:"bytes,5,opt,name=profile_picture_url,json=profilePictureUrl,proto3" json:"profile_picture_url,omitempty"`
	Roles             []RoleInfo `protobuf:"bytes,6,rep,name=roles,proto3" json:"roles,omitempty"`
	EmailVerified     bool       `protobuf:"varint,7,opt,name=email_verified,json=emailVerified,proto3" json:"email_verified,omitempty"`
	IsActive          bool       `protobuf:"varint,8,opt,name=is_active,json=isActive,proto3" json:"is_active,omitempty"`
	CreatedAt         int64      `protobuf:"varint,9,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt         int64      `protobuf:"varint,10,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

// GetUserResponse contains user information
type GetUserResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	User         *User  `protobuf:"bytes,1,opt,name=user,proto3" json:"user,omitempty"`
	ErrorMessage string `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

func (x *GetUserResponse) Reset()         { *x = GetUserResponse{} }
func (x *GetUserResponse) String() string { return "" }
func (*GetUserResponse) ProtoMessage()    {}

func (x *GetUserResponse) ProtoReflect() protoreflect.Message {
	return nil
}

func (m *GetUserResponse) UnmarshalBinary(b []byte) error {
	for len(b) > 0 {
		fieldNum, wireType, n := protowire.ConsumeTag(b)
		if n < 0 {
			return protowire.ParseError(n)
		}
		b = b[n:]

		switch fieldNum {
		case 1: // user (message)
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			msgBytes, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.User = &User{}
			if err := unmarshalUser(msgBytes, m.User); err != nil {
				return err
			}
			b = b[n:]
		case 2: // error_message (string)
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.ErrorMessage = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(fieldNum, wireType, b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			b = b[n:]
		}
	}
	return nil
}

func unmarshalUser(b []byte, u *User) error {
	for len(b) > 0 {
		fieldNum, wireType, n := protowire.ConsumeTag(b)
		if n < 0 {
			return protowire.ParseError(n)
		}
		b = b[n:]

		switch fieldNum {
		case 1: // id
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			u.Id = v
			b = b[n:]
		case 2: // email
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			u.Email = v
			b = b[n:]
		case 3: // username
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			u.Username = v
			b = b[n:]
		case 4: // full_name
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			u.FullName = v
			b = b[n:]
		case 5: // profile_picture_url
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			u.ProfilePictureUrl = v
			b = b[n:]
		case 6: // roles (repeated RoleInfo)
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			msgBytes, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			role := RoleInfo{}
			if err := unmarshalRoleInfo(msgBytes, &role); err != nil {
				return err
			}
			u.Roles = append(u.Roles, role)
			b = b[n:]
		case 7: // email_verified
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			u.EmailVerified = v != 0
			b = b[n:]
		case 8: // is_active
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			u.IsActive = v != 0
			b = b[n:]
		case 9: // created_at
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			u.CreatedAt = int64(v)
			b = b[n:]
		case 10: // updated_at
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			u.UpdatedAt = int64(v)
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(fieldNum, wireType, b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			b = b[n:]
		}
	}
	return nil
}

func unmarshalRoleInfo(b []byte, r *RoleInfo) error {
	for len(b) > 0 {
		fieldNum, wireType, n := protowire.ConsumeTag(b)
		if n < 0 {
			return protowire.ParseError(n)
		}
		b = b[n:]

		switch fieldNum {
		case 1: // id
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			r.Id = v
			b = b[n:]
		case 2: // name
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			r.Name = v
			b = b[n:]
		case 3: // display_name
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			r.DisplayName = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(fieldNum, wireType, b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			b = b[n:]
		}
	}
	return nil
}

// CheckPermissionRequest contains permission check data
type CheckPermissionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId   string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Resource string `protobuf:"bytes,2,opt,name=resource,proto3" json:"resource,omitempty"`
	Action   string `protobuf:"bytes,3,opt,name=action,proto3" json:"action,omitempty"`
}

func (x *CheckPermissionRequest) Reset()         { *x = CheckPermissionRequest{} }
func (x *CheckPermissionRequest) String() string { return "" }
func (*CheckPermissionRequest) ProtoMessage()    {}

func (x *CheckPermissionRequest) ProtoReflect() protoreflect.Message {
	return nil
}

func (m *CheckPermissionRequest) MarshalBinary() ([]byte, error) {
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
	return b, nil
}

// CheckPermissionResponse contains permission check result
type CheckPermissionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Allowed      bool     `protobuf:"varint,1,opt,name=allowed,proto3" json:"allowed,omitempty"`
	Roles        []string `protobuf:"bytes,2,rep,name=roles,proto3" json:"roles,omitempty"`
	ErrorMessage string   `protobuf:"bytes,3,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

func (x *CheckPermissionResponse) Reset()         { *x = CheckPermissionResponse{} }
func (x *CheckPermissionResponse) String() string { return "" }
func (*CheckPermissionResponse) ProtoMessage()    {}

func (x *CheckPermissionResponse) ProtoReflect() protoreflect.Message {
	return nil
}

func (m *CheckPermissionResponse) UnmarshalBinary(b []byte) error {
	for len(b) > 0 {
		fieldNum, wireType, n := protowire.ConsumeTag(b)
		if n < 0 {
			return protowire.ParseError(n)
		}
		b = b[n:]

		switch fieldNum {
		case 1: // allowed
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Allowed = v != 0
			b = b[n:]
		case 2: // roles (repeated string)
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Roles = append(m.Roles, v)
			b = b[n:]
		case 3: // error_message
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.ErrorMessage = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(fieldNum, wireType, b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			b = b[n:]
		}
	}
	return nil
}

// IntrospectTokenRequest contains the token to introspect
type IntrospectTokenRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AccessToken string `protobuf:"bytes,1,opt,name=access_token,json=accessToken,proto3" json:"access_token,omitempty"`
}

func (x *IntrospectTokenRequest) Reset()         { *x = IntrospectTokenRequest{} }
func (x *IntrospectTokenRequest) String() string { return x.AccessToken }
func (*IntrospectTokenRequest) ProtoMessage()    {}

func (x *IntrospectTokenRequest) ProtoReflect() protoreflect.Message {
	return nil
}

func (m *IntrospectTokenRequest) MarshalBinary() ([]byte, error) {
	var b []byte
	if m.AccessToken != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, m.AccessToken)
	}
	return b, nil
}

// IntrospectTokenResponse contains detailed token information
type IntrospectTokenResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Active       bool     `protobuf:"varint,1,opt,name=active,proto3" json:"active,omitempty"`
	UserId       string   `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Email        string   `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	Username     string   `protobuf:"bytes,4,opt,name=username,proto3" json:"username,omitempty"`
	Roles        []string `protobuf:"bytes,5,rep,name=roles,proto3" json:"roles,omitempty"`
	IssuedAt     int64    `protobuf:"varint,6,opt,name=issued_at,json=issuedAt,proto3" json:"issued_at,omitempty"`
	ExpiresAt    int64    `protobuf:"varint,7,opt,name=expires_at,json=expiresAt,proto3" json:"expires_at,omitempty"`
	NotBefore    int64    `protobuf:"varint,8,opt,name=not_before,json=notBefore,proto3" json:"not_before,omitempty"`
	Subject      string   `protobuf:"bytes,9,opt,name=subject,proto3" json:"subject,omitempty"`
	Blacklisted  bool     `protobuf:"varint,10,opt,name=blacklisted,proto3" json:"blacklisted,omitempty"`
	ErrorMessage string   `protobuf:"bytes,11,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

func (x *IntrospectTokenResponse) Reset()         { *x = IntrospectTokenResponse{} }
func (x *IntrospectTokenResponse) String() string { return "" }
func (*IntrospectTokenResponse) ProtoMessage()    {}

func (x *IntrospectTokenResponse) ProtoReflect() protoreflect.Message {
	return nil
}

func (m *IntrospectTokenResponse) UnmarshalBinary(b []byte) error {
	for len(b) > 0 {
		fieldNum, wireType, n := protowire.ConsumeTag(b)
		if n < 0 {
			return protowire.ParseError(n)
		}
		b = b[n:]

		switch fieldNum {
		case 1: // active
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Active = v != 0
			b = b[n:]
		case 2: // user_id
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.UserId = v
			b = b[n:]
		case 3: // email
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Email = v
			b = b[n:]
		case 4: // username
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Username = v
			b = b[n:]
		case 5: // roles (repeated string)
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Roles = append(m.Roles, v)
			b = b[n:]
		case 6: // issued_at
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.IssuedAt = int64(v)
			b = b[n:]
		case 7: // expires_at
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.ExpiresAt = int64(v)
			b = b[n:]
		case 8: // not_before
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.NotBefore = int64(v)
			b = b[n:]
		case 9: // subject
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Subject = v
			b = b[n:]
		case 10: // blacklisted
			if wireType != protowire.VarintType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.Blacklisted = v != 0
			b = b[n:]
		case 11: // error_message
			if wireType != protowire.BytesType {
				return protowire.ParseError(-1)
			}
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			m.ErrorMessage = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(fieldNum, wireType, b)
			if n < 0 {
				return protowire.ParseError(n)
			}
			b = b[n:]
		}
	}
	return nil
}
