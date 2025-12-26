package service

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// SAMLService handles SAML 2.0 IdP operations
type SAMLService struct {
	spRepo     SAMLSPRepository
	userRepo   UserStore
	rbacRepo   RBACStore
	logger     *logger.Logger
	issuer     string
	baseURL    string
	privateKey interface{} // Will be loaded from key manager or config
	cert       *x509.Certificate
}

// SAMLSPRepository defines interface for SAML SP operations
type SAMLSPRepository interface {
	Create(ctx context.Context, sp *models.SAMLServiceProvider) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.SAMLServiceProvider, error)
	GetByEntityID(ctx context.Context, entityID string) (*models.SAMLServiceProvider, error)
	List(ctx context.Context, page, pageSize int) ([]*models.SAMLServiceProvider, int, error)
	Update(ctx context.Context, sp *models.SAMLServiceProvider) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// NewSAMLService creates a new SAML service
func NewSAMLService(
	spRepo SAMLSPRepository,
	userRepo UserStore,
	rbacRepo RBACStore,
	logger *logger.Logger,
	issuer string,
	baseURL string,
) *SAMLService {
	return &SAMLService{
		spRepo:   spRepo,
		userRepo: userRepo,
		rbacRepo: rbacRepo,
		logger:   logger,
		issuer:   issuer,
		baseURL:  baseURL,
	}
}

// CreateSP creates a new SAML Service Provider
func (s *SAMLService) CreateSP(ctx context.Context, req *models.CreateSAMLSPRequest) (*models.SAMLServiceProvider, error) {
	sp := &models.SAMLServiceProvider{
		Name:        req.Name,
		EntityID:    req.EntityID,
		ACSURL:      req.ACSURL,
		SLOURL:      req.SLOURL,
		X509Cert:    req.X509Cert,
		MetadataURL: req.MetadataURL,
		IsActive:    true,
	}

	if err := s.spRepo.Create(ctx, sp); err != nil {
		if err == models.ErrAlreadyExists {
			return nil, models.NewAppError(409, "SAML SP with this entity ID already exists")
		}
		return nil, err
	}

	return sp, nil
}

// GetSP retrieves a SAML SP by ID
func (s *SAMLService) GetSP(ctx context.Context, id uuid.UUID) (*models.SAMLServiceProvider, error) {
	return s.spRepo.GetByID(ctx, id)
}

// ListSPs retrieves all SAML SPs with pagination
func (s *SAMLService) ListSPs(ctx context.Context, page, pageSize int) ([]*models.SAMLServiceProvider, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return s.spRepo.List(ctx, page, pageSize)
}

// UpdateSP updates a SAML SP
func (s *SAMLService) UpdateSP(ctx context.Context, id uuid.UUID, req *models.UpdateSAMLSPRequest) (*models.SAMLServiceProvider, error) {
	sp, err := s.spRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		sp.Name = *req.Name
	}
	if req.EntityID != nil {
		sp.EntityID = *req.EntityID
	}
	if req.ACSURL != nil {
		sp.ACSURL = *req.ACSURL
	}
	if req.SLOURL != nil {
		sp.SLOURL = *req.SLOURL
	}
	if req.X509Cert != nil {
		sp.X509Cert = *req.X509Cert
	}
	if req.MetadataURL != nil {
		sp.MetadataURL = *req.MetadataURL
	}
	if req.IsActive != nil {
		sp.IsActive = *req.IsActive
	}

	if err := s.spRepo.Update(ctx, sp); err != nil {
		return nil, err
	}

	return sp, nil
}

// DeleteSP deletes a SAML SP
func (s *SAMLService) DeleteSP(ctx context.Context, id uuid.UUID) error {
	return s.spRepo.Delete(ctx, id)
}

// GetSPByEntityID retrieves a SAML SP by EntityID
func (s *SAMLService) GetSPByEntityID(ctx context.Context, entityID string) (*models.SAMLServiceProvider, error) {
	return s.spRepo.GetByEntityID(ctx, entityID)
}

// GetMetadata generates SAML metadata for the IdP
func (s *SAMLService) GetMetadata() (*SAMLMetadata, error) {
	// Generate metadata XML
	metadata := &SAMLMetadata{
		EntityDescriptor: EntityDescriptor{
			EntityID: s.issuer,
			IDPSSODescriptor: IDPSSODescriptor{
				ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
				SingleSignOnService: []SingleSignOnService{
					{
						Binding:  "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect",
						Location: fmt.Sprintf("%s/saml/sso", s.baseURL),
					},
					{
						Binding:  "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST",
						Location: fmt.Sprintf("%s/saml/sso", s.baseURL),
					},
				},
				SingleLogoutService: []SingleLogoutService{
					{
						Binding:  "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect",
						Location: fmt.Sprintf("%s/saml/slo", s.baseURL),
					},
				},
				KeyDescriptor: []KeyDescriptor{
					{
						Use: "signing",
						KeyInfo: KeyInfo{
							X509Data: X509Data{
								X509Certificate: s.getCertificateBase64(),
							},
						},
					},
				},
			},
		},
	}

	return metadata, nil
}

// CreateAssertion creates a SAML assertion for a user
func (s *SAMLService) CreateAssertion(ctx context.Context, userID uuid.UUID, sp *models.SAMLServiceProvider) (*SAMLResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return nil, err
	}

	// Get user roles
	roles, err := s.rbacRepo.GetUserRoles(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get user roles for SAML assertion", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		roles = []models.Role{} // Continue with empty roles
	}

	// Build role names
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// Generate assertion ID
	assertionID := generateSAMLID()

	// Create assertion
	now := time.Now()
	assertion := &SAMLAssertion{
		ID:           assertionID,
		IssueInstant: now.Format(time.RFC3339),
		Version:      "2.0",
		Issuer: Issuer{
			Value: s.issuer,
		},
		Subject: Subject{
			NameID: NameID{
				Format: "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
				Value:  user.Email,
			},
			SubjectConfirmation: SubjectConfirmation{
				Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
				SubjectConfirmationData: SubjectConfirmationData{
					NotOnOrAfter: now.Add(5 * time.Minute).Format(time.RFC3339),
					Recipient:    sp.ACSURL,
				},
			},
		},
		Conditions: Conditions{
			NotBefore:    now.Format(time.RFC3339),
			NotOnOrAfter: now.Add(5 * time.Minute).Format(time.RFC3339),
			AudienceRestriction: AudienceRestriction{
				Audience: sp.EntityID,
			},
		},
		AttributeStatement: AttributeStatement{
			Attributes: []Attribute{
				{
					Name:       "email",
					NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
					AttributeValues: []AttributeValue{
						{Value: user.Email},
					},
				},
				{
					Name:       "name",
					NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
					AttributeValues: []AttributeValue{
						{Value: user.FullName},
					},
				},
				{
					Name:       "username",
					NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
					AttributeValues: []AttributeValue{
						{Value: user.Username},
					},
				},
			},
		},
	}

	// Add roles if available
	if len(roleNames) > 0 {
		roleValues := make([]AttributeValue, len(roleNames))
		for i, role := range roleNames {
			roleValues[i] = AttributeValue{Value: role}
		}
		assertion.AttributeStatement.Attributes = append(assertion.AttributeStatement.Attributes, Attribute{
			Name:            "roles",
			NameFormat:      "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
			AttributeValues: roleValues,
		})
	}

	// Create response
	responseID := generateSAMLID()
	response := &SAMLResponse{
		ID:           responseID,
		IssueInstant: now.Format(time.RFC3339),
		Version:      "2.0",
		Destination:  sp.ACSURL,
		Issuer: Issuer{
			Value: s.issuer,
		},
		Status: Status{
			StatusCode: StatusCode{
				Value: "urn:oasis:names:tc:SAML:2.0:status:Success",
			},
		},
		Assertion: *assertion,
	}

	return response, nil
}

// Helper functions

func generateSAMLID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("_%x", b)
}

func (s *SAMLService) getCertificateBase64() string {
	if s.cert == nil {
		// Return empty or placeholder - should be loaded from key manager
		return ""
	}
	return base64.StdEncoding.EncodeToString(s.cert.Raw)
}

// SAML XML structures (simplified)

type SAMLMetadata struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntityDescriptor"`
	EntityDescriptor
}

type EntityDescriptor struct {
	EntityID         string           `xml:"entityID,attr"`
	IDPSSODescriptor IDPSSODescriptor `xml:"IDPSSODescriptor"`
}

type IDPSSODescriptor struct {
	ProtocolSupportEnumeration string                `xml:"protocolSupportEnumeration,attr"`
	SingleSignOnService        []SingleSignOnService `xml:"SingleSignOnService"`
	SingleLogoutService        []SingleLogoutService `xml:"SingleLogoutService"`
	KeyDescriptor              []KeyDescriptor       `xml:"KeyDescriptor"`
}

type SingleSignOnService struct {
	Binding  string `xml:"Binding,attr"`
	Location string `xml:"Location,attr"`
}

type SingleLogoutService struct {
	Binding  string `xml:"Binding,attr"`
	Location string `xml:"Location,attr"`
}

type KeyDescriptor struct {
	Use     string  `xml:"use,attr"`
	KeyInfo KeyInfo `xml:"KeyInfo"`
}

type KeyInfo struct {
	X509Data X509Data `xml:"X509Data"`
}

type X509Data struct {
	X509Certificate string `xml:"X509Certificate"`
}

type SAMLResponse struct {
	XMLName      xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol Response"`
	ID           string   `xml:"ID,attr"`
	IssueInstant string   `xml:"IssueInstant,attr"`
	Version      string   `xml:"Version,attr"`
	Destination  string   `xml:"Destination,attr"`
	Issuer       Issuer
	Status       Status
	Assertion    SAMLAssertion
}

type SAMLAssertion struct {
	XMLName            xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
	ID                 string   `xml:"ID,attr"`
	IssueInstant       string   `xml:"IssueInstant,attr"`
	Version            string   `xml:"Version,attr"`
	Issuer             Issuer
	Subject            Subject
	Conditions         Conditions
	AttributeStatement AttributeStatement
}

type Issuer struct {
	Value string `xml:",chardata"`
}

type Subject struct {
	NameID              NameID
	SubjectConfirmation SubjectConfirmation
}

type NameID struct {
	Format string `xml:"Format,attr"`
	Value  string `xml:",chardata"`
}

type SubjectConfirmation struct {
	Method                  string `xml:"Method,attr"`
	SubjectConfirmationData SubjectConfirmationData
}

type SubjectConfirmationData struct {
	NotOnOrAfter string `xml:"NotOnOrAfter,attr"`
	Recipient    string `xml:"Recipient,attr"`
}

type Conditions struct {
	NotBefore           string `xml:"NotBefore,attr"`
	NotOnOrAfter        string `xml:"NotOnOrAfter,attr"`
	AudienceRestriction AudienceRestriction
}

type AudienceRestriction struct {
	Audience string `xml:"Audience"`
}

type AttributeStatement struct {
	Attributes []Attribute `xml:"Attribute"`
}

type Attribute struct {
	Name            string           `xml:"Name,attr"`
	NameFormat      string           `xml:"NameFormat,attr"`
	AttributeValues []AttributeValue `xml:"AttributeValue"`
}

type AttributeValue struct {
	Value string `xml:",chardata"`
}

type Status struct {
	StatusCode StatusCode `xml:"StatusCode"`
}

type StatusCode struct {
	Value string `xml:"Value,attr"`
}
