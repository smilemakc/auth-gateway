package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protowire"
)

// Request/Response types with protobuf wire encoding

type ValidateTokenRequest struct {
	AccessToken string
}

func (m *ValidateTokenRequest) Reset()         {}
func (m *ValidateTokenRequest) String() string { return m.AccessToken }
func (m *ValidateTokenRequest) ProtoMessage()  {}

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

func (m *ValidateTokenResponse) Reset()         {}
func (m *ValidateTokenResponse) String() string { return "" }
func (m *ValidateTokenResponse) ProtoMessage()  {}

type GetUserRequest struct {
	UserId string
}

func (m *GetUserRequest) Reset()         {}
func (m *GetUserRequest) String() string { return m.UserId }
func (m *GetUserRequest) ProtoMessage()  {}

type GetUserResponse struct {
	User         *User
	ErrorMessage string
}

func (m *GetUserResponse) Reset()         {}
func (m *GetUserResponse) String() string { return "" }
func (m *GetUserResponse) ProtoMessage()  {}

type User struct {
	Id            string
	Email         string
	Username      string
	FullName      string
	EmailVerified bool
	IsActive      bool
	Roles         []RoleInfo
	CreatedAt     int64
	UpdatedAt     int64
}

type RoleInfo struct {
	Id          string
	Name        string
	DisplayName string
}

type CheckPermissionRequest struct {
	UserId   string
	Resource string
	Action   string
}

func (m *CheckPermissionRequest) Reset()         {}
func (m *CheckPermissionRequest) String() string { return "" }
func (m *CheckPermissionRequest) ProtoMessage()  {}

type CheckPermissionResponse struct {
	Allowed      bool
	Roles        []string
	ErrorMessage string
}

func (m *CheckPermissionResponse) Reset()         {}
func (m *CheckPermissionResponse) String() string { return "" }
func (m *CheckPermissionResponse) ProtoMessage()  {}

type IntrospectTokenRequest struct {
	AccessToken string
}

func (m *IntrospectTokenRequest) Reset()         {}
func (m *IntrospectTokenRequest) String() string { return m.AccessToken }
func (m *IntrospectTokenRequest) ProtoMessage()  {}

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

func (m *IntrospectTokenResponse) Reset()         {}
func (m *IntrospectTokenResponse) String() string { return "" }
func (m *IntrospectTokenResponse) ProtoMessage()  {}

// Custom codec that encodes/decodes using protobuf wire format
type authCodec struct{}

func (authCodec) Name() string { return "proto" }

func (authCodec) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case *ValidateTokenRequest:
		var b []byte
		if m.AccessToken != "" {
			b = protowire.AppendTag(b, 1, protowire.BytesType)
			b = protowire.AppendString(b, m.AccessToken)
		}
		return b, nil
	case *GetUserRequest:
		var b []byte
		if m.UserId != "" {
			b = protowire.AppendTag(b, 1, protowire.BytesType)
			b = protowire.AppendString(b, m.UserId)
		}
		return b, nil
	case *CheckPermissionRequest:
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
	case *IntrospectTokenRequest:
		var b []byte
		if m.AccessToken != "" {
			b = protowire.AppendTag(b, 1, protowire.BytesType)
			b = protowire.AppendString(b, m.AccessToken)
		}
		return b, nil
	default:
		return nil, fmt.Errorf("unknown type: %T", v)
	}
}

func (authCodec) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case *ValidateTokenResponse:
		return unmarshalValidateTokenResponse(data, m)
	case *GetUserResponse:
		return unmarshalGetUserResponse(data, m)
	case *CheckPermissionResponse:
		return unmarshalCheckPermissionResponse(data, m)
	case *IntrospectTokenResponse:
		return unmarshalIntrospectTokenResponse(data, m)
	default:
		return fmt.Errorf("unknown type: %T", v)
	}
}

func unmarshalValidateTokenResponse(b []byte, m *ValidateTokenResponse) error {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		b = b[n:]

		switch num {
		case 1:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			m.Valid = v != 0
			b = b[n:]
		case 2:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.UserId = v
			b = b[n:]
		case 3:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.Email = v
			b = b[n:]
		case 4:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.Username = v
			b = b[n:]
		case 5:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.Roles = append(m.Roles, v)
			b = b[n:]
		case 6:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.ErrorMessage = v
			b = b[n:]
		case 7:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			m.ExpiresAt = int64(v)
			b = b[n:]
		case 8:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			m.IsActive = v != 0
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return fmt.Errorf("invalid field")
			}
			b = b[n:]
		}
	}
	return nil
}

func unmarshalGetUserResponse(b []byte, m *GetUserResponse) error {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		b = b[n:]

		switch num {
		case 1:
			msgBytes, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return fmt.Errorf("invalid bytes")
			}
			m.User = &User{}
			if err := unmarshalUser(msgBytes, m.User); err != nil {
				return err
			}
			b = b[n:]
		case 2:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.ErrorMessage = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return fmt.Errorf("invalid field")
			}
			b = b[n:]
		}
	}
	return nil
}

func unmarshalUser(b []byte, u *User) error {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		b = b[n:]

		switch num {
		case 1:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			u.Id = v
			b = b[n:]
		case 2:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			u.Email = v
			b = b[n:]
		case 3:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			u.Username = v
			b = b[n:]
		case 4:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			u.FullName = v
			b = b[n:]
		case 6:
			msgBytes, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return fmt.Errorf("invalid bytes")
			}
			role := RoleInfo{}
			if err := unmarshalRoleInfo(msgBytes, &role); err != nil {
				return err
			}
			u.Roles = append(u.Roles, role)
			b = b[n:]
		case 7:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			u.EmailVerified = v != 0
			b = b[n:]
		case 8:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			u.IsActive = v != 0
			b = b[n:]
		case 9:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			u.CreatedAt = int64(v)
			b = b[n:]
		case 10:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			u.UpdatedAt = int64(v)
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return fmt.Errorf("invalid field")
			}
			b = b[n:]
		}
	}
	return nil
}

func unmarshalRoleInfo(b []byte, r *RoleInfo) error {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		b = b[n:]

		switch num {
		case 1:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			r.Id = v
			b = b[n:]
		case 2:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			r.Name = v
			b = b[n:]
		case 3:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			r.DisplayName = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return fmt.Errorf("invalid field")
			}
			b = b[n:]
		}
	}
	return nil
}

func unmarshalCheckPermissionResponse(b []byte, m *CheckPermissionResponse) error {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		b = b[n:]

		switch num {
		case 1:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			m.Allowed = v != 0
			b = b[n:]
		case 2:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.Roles = append(m.Roles, v)
			b = b[n:]
		case 3:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.ErrorMessage = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return fmt.Errorf("invalid field")
			}
			b = b[n:]
		}
	}
	return nil
}

func unmarshalIntrospectTokenResponse(b []byte, m *IntrospectTokenResponse) error {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return fmt.Errorf("invalid tag")
		}
		b = b[n:]

		switch num {
		case 1:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			m.Active = v != 0
			b = b[n:]
		case 2:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.UserId = v
			b = b[n:]
		case 3:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.Email = v
			b = b[n:]
		case 4:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.Username = v
			b = b[n:]
		case 5:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.Roles = append(m.Roles, v)
			b = b[n:]
		case 6:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			m.IssuedAt = int64(v)
			b = b[n:]
		case 7:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			m.ExpiresAt = int64(v)
			b = b[n:]
		case 8:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			m.NotBefore = int64(v)
			b = b[n:]
		case 9:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.Subject = v
			b = b[n:]
		case 10:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint")
			}
			m.Blacklisted = v != 0
			b = b[n:]
		case 11:
			v, n := protowire.ConsumeString(b)
			if n < 0 {
				return fmt.Errorf("invalid string")
			}
			m.ErrorMessage = v
			b = b[n:]
		default:
			n := protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return fmt.Errorf("invalid field")
			}
			b = b[n:]
		}
	}
	return nil
}

func main() {
	fmt.Println("==========================================")
	fmt.Println("  Auth Gateway gRPC Integration Test")
	fmt.Println("==========================================")
	fmt.Println()

	// Connect to server
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(authCodec{})),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to gRPC server at localhost:50051")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test 1: ValidateToken with valid test token
	fmt.Println("Test 1: ValidateToken (valid_test_token)")
	fmt.Println("-----------------------------------------")
	resp1 := &ValidateTokenResponse{}
	err = conn.Invoke(ctx, "/auth.AuthService/ValidateToken", &ValidateTokenRequest{AccessToken: "valid_test_token"}, resp1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printJSON(resp1)
		if resp1.Valid {
			fmt.Println("PASS: Token validated successfully!")
		}
	}
	fmt.Println()

	// Test 2: ValidateToken with valid API key
	fmt.Println("Test 2: ValidateToken (agw_valid_test_key)")
	fmt.Println("-------------------------------------------")
	resp2 := &ValidateTokenResponse{}
	err = conn.Invoke(ctx, "/auth.AuthService/ValidateToken", &ValidateTokenRequest{AccessToken: "agw_valid_test_key"}, resp2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printJSON(resp2)
		if resp2.Valid {
			fmt.Println("PASS: API key validated successfully!")
		}
	}
	fmt.Println()

	// Test 3: ValidateToken with invalid token
	fmt.Println("Test 3: ValidateToken (invalid token)")
	fmt.Println("--------------------------------------")
	resp3 := &ValidateTokenResponse{}
	err = conn.Invoke(ctx, "/auth.AuthService/ValidateToken", &ValidateTokenRequest{AccessToken: "invalid-token"}, resp3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printJSON(resp3)
		if !resp3.Valid && resp3.ErrorMessage != "" {
			fmt.Println("PASS: Invalid token correctly rejected!")
		}
	}
	fmt.Println()

	// Get a user ID from the ValidateToken response
	userID := resp1.UserId
	if userID == "" {
		userID = resp2.UserId
	}

	if userID != "" {
		// Test 4: GetUser with valid user ID
		fmt.Println("Test 4: GetUser (valid user)")
		fmt.Println("-----------------------------")
		resp4 := &GetUserResponse{}
		err = conn.Invoke(ctx, "/auth.AuthService/GetUser", &GetUserRequest{UserId: userID}, resp4)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			printJSON(resp4)
			if resp4.User != nil && resp4.User.Email != "" {
				fmt.Println("PASS: User retrieved successfully!")
			}
		}
		fmt.Println()

		// Test 5: CheckPermission (read access)
		fmt.Println("Test 5: CheckPermission (users:read)")
		fmt.Println("-------------------------------------")
		resp5 := &CheckPermissionResponse{}
		err = conn.Invoke(ctx, "/auth.AuthService/CheckPermission", &CheckPermissionRequest{
			UserId:   userID,
			Resource: "users",
			Action:   "read",
		}, resp5)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			printJSON(resp5)
			fmt.Println("PASS: Permission check completed!")
		}
		fmt.Println()
	}

	// Test 6: IntrospectToken
	fmt.Println("Test 6: IntrospectToken (valid_test_token)")
	fmt.Println("-------------------------------------------")
	resp6 := &IntrospectTokenResponse{}
	err = conn.Invoke(ctx, "/auth.AuthService/IntrospectToken", &IntrospectTokenRequest{AccessToken: "valid_test_token"}, resp6)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printJSON(resp6)
		if resp6.Active {
			fmt.Println("PASS: Token introspected successfully!")
		}
	}
	fmt.Println()

	// Test 7: GetUser with invalid UUID
	fmt.Println("Test 7: GetUser (invalid UUID)")
	fmt.Println("-------------------------------")
	resp7 := &GetUserResponse{}
	err = conn.Invoke(ctx, "/auth.AuthService/GetUser", &GetUserRequest{UserId: "not-a-uuid"}, resp7)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
		fmt.Println("PASS: Invalid UUID correctly rejected!")
	} else {
		printJSON(resp7)
	}
	fmt.Println()

	fmt.Println("==========================================")
	fmt.Println("  All Tests Completed!")
	fmt.Println("==========================================")
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("%+v\n", v)
		return
	}
	fmt.Println(string(data))
}
