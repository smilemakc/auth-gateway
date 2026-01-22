package models

import (
	"time"
)

// SCIMUser represents a SCIM 2.0 User resource
type SCIMUser struct {
	Schemas    []string    `json:"schemas"`
	ID         string      `json:"id"`
	ExternalID string      `json:"externalId,omitempty"`
	UserName   string      `json:"userName"`
	Name       SCIMName    `json:"name"`
	Emails     []SCIMEmail `json:"emails"`
	Active     bool        `json:"active"`
	Meta       SCIMMeta    `json:"meta"`
}

// SCIMName represents SCIM name structure
type SCIMName struct {
	Formatted  string `json:"formatted,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
	GivenName  string `json:"givenName,omitempty"`
	MiddleName string `json:"middleName,omitempty"`
}

// SCIMEmail represents SCIM email structure
type SCIMEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMMeta represents SCIM meta information
type SCIMMeta struct {
	ResourceType string    `json:"resourceType"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
	Location     string    `json:"location"`
	Version      string    `json:"version,omitempty"`
}

// SCIMGroup represents a SCIM 2.0 Group resource
type SCIMGroup struct {
	Schemas     []string     `json:"schemas"`
	ID          string       `json:"id"`
	ExternalID  string       `json:"externalId,omitempty"`
	DisplayName string       `json:"displayName"`
	Members     []SCIMMember `json:"members,omitempty"`
	Meta        SCIMMeta     `json:"meta"`
}

// SCIMMember represents a SCIM group member
type SCIMMember struct {
	Value   string `json:"value"`
	Ref     string `json:"$ref,omitempty"`
	Display string `json:"display,omitempty"`
	Type    string `json:"type,omitempty"`
}

// SCIMListResponse represents a SCIM list response
type SCIMListResponse struct {
	TotalResults int           `json:"totalResults"`
	ItemsPerPage int           `json:"itemsPerPage"`
	StartIndex   int           `json:"startIndex"`
	Schemas      []string      `json:"schemas"`
	Resources    []interface{} `json:"Resources"`
}

// SCIMError represents a SCIM error response
type SCIMError struct {
	Schemas  []string `json:"schemas"`
	Detail   string   `json:"detail"`
	Status   string   `json:"status"`
	ScimType string   `json:"scimType,omitempty"`
}

// SCIMServiceProviderConfig represents SCIM service provider configuration
type SCIMServiceProviderConfig struct {
	Schemas               []string          `json:"schemas"`
	Patch                 SCIMFeature       `json:"patch"`
	Bulk                  SCIMBulkFeature   `json:"bulk"`
	Filter                SCIMFilterFeature `json:"filter"`
	ChangePassword        SCIMFeature       `json:"changePassword"`
	Sort                  SCIMFeature       `json:"sort"`
	ETag                  SCIMFeature       `json:"etag"`
	AuthenticationSchemes []SCIMAuthScheme  `json:"authenticationSchemes"`
	Meta                  SCIMMeta          `json:"meta"`
}

// SCIMFeature represents a SCIM feature
type SCIMFeature struct {
	Supported bool `json:"supported"`
}

// SCIMBulkFeature represents SCIM bulk feature
type SCIMBulkFeature struct {
	Supported      bool `json:"supported"`
	MaxOperations  int  `json:"maxOperations"`
	MaxPayloadSize int  `json:"maxPayloadSize"`
}

// SCIMFilterFeature represents SCIM filter feature
type SCIMFilterFeature struct {
	Supported  bool `json:"supported"`
	MaxResults int  `json:"maxResults"`
}

// SCIMAuthScheme represents SCIM authentication scheme
type SCIMAuthScheme struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SCIMSchema represents a SCIM schema definition
type SCIMSchema struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Attributes  []SCIMAttribute `json:"attributes"`
	Meta        SCIMMeta        `json:"meta"`
}

// SCIMAttribute represents a SCIM attribute definition
type SCIMAttribute struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	MultiValued     bool     `json:"multiValued"`
	Description     string   `json:"description,omitempty"`
	Required        bool     `json:"required"`
	CanonicalValues []string `json:"canonicalValues,omitempty"`
	CaseExact       bool     `json:"caseExact"`
	Mutability      string   `json:"mutability"`
	Returned        string   `json:"returned"`
	Uniqueness      string   `json:"uniqueness,omitempty"`
}

// SCIMPatchOperation represents a SCIM PATCH operation
type SCIMPatchOperation struct {
	Op    string      `json:"op"` // "add", "remove", "replace"
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// SCIMPatchRequest represents a SCIM PATCH request
type SCIMPatchRequest struct {
	Schemas    []string             `json:"schemas"`
	Operations []SCIMPatchOperation `json:"Operations"`
}
