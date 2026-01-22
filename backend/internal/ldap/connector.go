package ldap

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// Connector handles LDAP/AD connections and operations
type Connector struct {
	config *models.LDAPConfig
	conn   *ldap.Conn
}

// NewConnector creates a new LDAP connector
func NewConnector(config *models.LDAPConfig) *Connector {
	return &Connector{
		config: config,
	}
}

// Connect establishes a connection to the LDAP server
func (c *Connector) Connect() error {
	address := fmt.Sprintf("%s:%d", c.config.Server, c.config.Port)

	var err error
	if c.config.UseSSL {
		c.conn, err = ldap.DialTLS("tcp", address, &tls.Config{
			ServerName:         c.config.Server,
			InsecureSkipVerify: c.config.Insecure,
		})
	} else {
		c.conn, err = ldap.Dial("tcp", address)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to LDAP server: %w", err)
	}

	// Set timeout
	c.conn.SetTimeout(30 * time.Second)

	// Start TLS if needed
	if c.config.UseTLS && !c.config.UseSSL {
		err = c.conn.StartTLS(&tls.Config{
			ServerName:         c.config.Server,
			InsecureSkipVerify: c.config.Insecure,
		})
		if err != nil {
			c.conn.Close()
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Bind with credentials
	err = c.conn.Bind(c.config.BindDN, c.config.BindPassword)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to bind to LDAP: %w", err)
	}

	return nil
}

// Close closes the LDAP connection
func (c *Connector) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// TestConnection tests the LDAP connection
func (c *Connector) TestConnection() error {
	if err := c.Connect(); err != nil {
		return err
	}
	defer c.Close()

	// Try to search for users to verify connection
	searchRequest := ldap.NewSearchRequest(
		c.config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		c.config.UserSearchFilter,
		[]string{"dn"},
		nil,
	)

	_, err := c.conn.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("failed to search LDAP: %w", err)
	}

	return nil
}

// Authenticate authenticates a user against LDAP
func (c *Connector) Authenticate(username, password string) (*models.User, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}
	defer c.Close()

	// Search for user
	searchBase := c.config.UserSearchBase
	if searchBase == "" {
		searchBase = c.config.BaseDN
	}

	searchFilter := fmt.Sprintf("(&%s(%s=%s))", c.config.UserSearchFilter, c.config.UserIDAttribute, ldap.EscapeFilter(username))
	searchRequest := ldap.NewSearchRequest(
		searchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		[]string{"dn", c.config.UserEmailAttribute, c.config.UserNameAttribute},
		nil,
	)

	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search user: %w", err)
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	if len(sr.Entries) > 1 {
		return nil, fmt.Errorf("multiple users found")
	}

	entry := sr.Entries[0]
	userDN := entry.DN

	// Try to bind with user's password
	userConn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.config.Server, c.config.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to create user connection: %w", err)
	}
	defer userConn.Close()

	if c.config.UseTLS || c.config.UseSSL {
		if c.config.UseSSL {
			userConn, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", c.config.Server, c.config.Port), &tls.Config{
				ServerName:         c.config.Server,
				InsecureSkipVerify: c.config.Insecure,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to connect with TLS: %w", err)
			}
		} else {
			err = userConn.StartTLS(&tls.Config{
				ServerName:         c.config.Server,
				InsecureSkipVerify: c.config.Insecure,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to start TLS: %w", err)
			}
		}
	}

	err = userConn.Bind(userDN, password)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Extract user attributes
	user := &models.User{}
	user.Email = entry.GetAttributeValue(c.config.UserEmailAttribute)
	user.Username = username
	user.FullName = entry.GetAttributeValue(c.config.UserNameAttribute)

	return user, nil
}

// SearchUsers searches for users in LDAP
func (c *Connector) SearchUsers(filter string) ([]*LDAPUser, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}
	defer c.Close()

	searchBase := c.config.UserSearchBase
	if searchBase == "" {
		searchBase = c.config.BaseDN
	}

	searchFilter := c.config.UserSearchFilter
	if filter != "" {
		searchFilter = fmt.Sprintf("(&%s%s)", c.config.UserSearchFilter, filter)
	}

	searchRequest := ldap.NewSearchRequest(
		searchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		[]string{"dn", c.config.UserIDAttribute, c.config.UserEmailAttribute, c.config.UserNameAttribute},
		nil,
	)

	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	users := make([]*LDAPUser, 0, len(sr.Entries))
	for _, entry := range sr.Entries {
		user := &LDAPUser{
			DN:       entry.DN,
			ID:       entry.GetAttributeValue(c.config.UserIDAttribute),
			Email:    entry.GetAttributeValue(c.config.UserEmailAttribute),
			FullName: entry.GetAttributeValue(c.config.UserNameAttribute),
		}
		users = append(users, user)
	}

	return users, nil
}

// SearchGroups searches for groups in LDAP
func (c *Connector) SearchGroups(filter string) ([]*LDAPGroup, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}
	defer c.Close()

	searchBase := c.config.GroupSearchBase
	if searchBase == "" {
		searchBase = c.config.BaseDN
	}

	searchFilter := c.config.GroupSearchFilter
	if filter != "" {
		searchFilter = fmt.Sprintf("(&%s%s)", c.config.GroupSearchFilter, filter)
	}

	searchRequest := ldap.NewSearchRequest(
		searchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchFilter,
		[]string{"dn", c.config.GroupIDAttribute, c.config.GroupNameAttribute, c.config.GroupMemberAttribute},
		nil,
	)

	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search groups: %w", err)
	}

	groups := make([]*LDAPGroup, 0, len(sr.Entries))
	for _, entry := range sr.Entries {
		group := &LDAPGroup{
			DN:      entry.DN,
			ID:      entry.GetAttributeValue(c.config.GroupIDAttribute),
			Name:    entry.GetAttributeValue(c.config.GroupNameAttribute),
			Members: entry.GetAttributeValues(c.config.GroupMemberAttribute),
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// LDAPUser represents a user from LDAP
type LDAPUser struct {
	DN       string
	ID       string
	Email    string
	FullName string
}

// LDAPGroup represents a group from LDAP
type LDAPGroup struct {
	DN      string
	ID      string
	Name    string
	Members []string
}
