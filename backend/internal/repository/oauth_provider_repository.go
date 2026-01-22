package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

type OAuthProviderRepository struct {
	db *Database
}

func NewOAuthProviderRepository(db *Database) *OAuthProviderRepository {
	return &OAuthProviderRepository{db: db}
}

func (r *OAuthProviderRepository) CreateClient(ctx context.Context, client *models.OAuthClient) error {
	client.CreatedAt = time.Now()
	client.UpdatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(client).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *OAuthProviderRepository) GetClientByID(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
	client := new(models.OAuthClient)

	err := r.db.NewSelect().
		Model(client).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("oauth client not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth client by id: %w", err)
	}

	return client, nil
}

func (r *OAuthProviderRepository) GetClientByClientID(ctx context.Context, clientID string) (*models.OAuthClient, error) {
	client := new(models.OAuthClient)

	err := r.db.NewSelect().
		Model(client).
		Where("client_id = ?", clientID).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("oauth client not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth client by client_id: %w", err)
	}

	return client, nil
}

func (r *OAuthProviderRepository) UpdateClient(ctx context.Context, client *models.OAuthClient) error {
	client.UpdatedAt = time.Now()

	result, err := r.db.NewUpdate().
		Model(client).
		Column("name", "description", "logo_url", "client_type", "redirect_uris",
			"allowed_grant_types", "allowed_scopes", "default_scopes", "access_token_ttl",
			"refresh_token_ttl", "id_token_ttl", "require_pkce", "require_consent",
			"first_party", "is_active", "updated_at").
		WherePK().
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update oauth client: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("oauth client not found")
	}

	return nil
}

func (r *OAuthProviderRepository) DeleteClient(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.OAuthClient)(nil)).
		Set("is_active = ?", false).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete oauth client: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("oauth client not found")
	}

	return nil
}

func (r *OAuthProviderRepository) HardDeleteClient(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.OAuthClient)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to hard delete oauth client: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("oauth client not found")
	}

	return nil
}

func (r *OAuthProviderRepository) ListClients(ctx context.Context, ownerID *uuid.UUID, page, perPage int) ([]*models.OAuthClient, int, error) {
	clients := make([]*models.OAuthClient, 0)

	query := r.db.NewSelect().
		Model(&clients).
		Where("o_auth_client.is_active = ?", true)

	if ownerID != nil {
		query = query.Where("owner_id = ?", *ownerID)
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count oauth clients: %w", err)
	}

	offset := (page - 1) * perPage

	err = query.
		Relation("Owner").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list oauth clients: %w", err)
	}

	return clients, total, nil
}

func (r *OAuthProviderRepository) ListActiveClients(ctx context.Context) ([]*models.OAuthClient, error) {
	clients := make([]*models.OAuthClient, 0)

	err := r.db.NewSelect().
		Model(&clients).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list active oauth clients: %w", err)
	}

	return clients, nil
}

func (r *OAuthProviderRepository) CreateAuthorizationCode(ctx context.Context, code *models.AuthorizationCode) error {
	code.CreatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(code).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *OAuthProviderRepository) GetAuthorizationCode(ctx context.Context, codeHash string) (*models.AuthorizationCode, error) {
	code := new(models.AuthorizationCode)

	err := r.db.NewSelect().
		Model(code).
		Where("code_hash = ?", codeHash).
		Relation("Client").
		Relation("User").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("authorization code not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization code: %w", err)
	}

	return code, nil
}

func (r *OAuthProviderRepository) MarkAuthorizationCodeUsed(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.AuthorizationCode)(nil)).
		Set("used = ?", true).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to mark authorization code as used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("authorization code not found")
	}

	return nil
}

func (r *OAuthProviderRepository) DeleteExpiredAuthorizationCodes(ctx context.Context) (int64, error) {
	result, err := r.db.NewDelete().
		Model((*models.AuthorizationCode)(nil)).
		Where("expires_at < ?", time.Now()).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to delete expired authorization codes: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

func (r *OAuthProviderRepository) CreateAccessToken(ctx context.Context, token *models.OAuthAccessToken) error {
	token.CreatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(token).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *OAuthProviderRepository) GetAccessToken(ctx context.Context, tokenHash string) (*models.OAuthAccessToken, error) {
	token := new(models.OAuthAccessToken)

	err := r.db.NewSelect().
		Model(token).
		Where("token_hash = ?", tokenHash).
		Relation("Client").
		Relation("User").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("oauth access token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth access token: %w", err)
	}

	return token, nil
}

func (r *OAuthProviderRepository) GetAccessTokenByID(ctx context.Context, id uuid.UUID) (*models.OAuthAccessToken, error) {
	token := new(models.OAuthAccessToken)

	err := r.db.NewSelect().
		Model(token).
		Where("id = ?", id).
		Relation("Client").
		Relation("User").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("oauth access token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth access token by id: %w", err)
	}

	return token, nil
}

func (r *OAuthProviderRepository) RevokeAccessToken(ctx context.Context, tokenHash string) error {
	now := time.Now()

	result, err := r.db.NewUpdate().
		Model((*models.OAuthAccessToken)(nil)).
		Set("is_active = ?", false).
		Set("revoked_at = ?", now).
		Where("token_hash = ?", tokenHash).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke oauth access token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("oauth access token not found")
	}

	return nil
}

func (r *OAuthProviderRepository) RevokeAllUserAccessTokens(ctx context.Context, userID, clientID uuid.UUID) error {
	now := time.Now()

	_, err := r.db.NewUpdate().
		Model((*models.OAuthAccessToken)(nil)).
		Set("is_active = ?", false).
		Set("revoked_at = ?", now).
		Where("user_id = ?", userID).
		Where("client_id = ?", clientID).
		Where("is_active = ?", true).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke all user access tokens: %w", err)
	}

	return nil
}

func (r *OAuthProviderRepository) RevokeAllClientAccessTokens(ctx context.Context, clientID uuid.UUID) error {
	now := time.Now()

	_, err := r.db.NewUpdate().
		Model((*models.OAuthAccessToken)(nil)).
		Set("is_active = ?", false).
		Set("revoked_at = ?", now).
		Where("client_id = ?", clientID).
		Where("is_active = ?", true).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke all client access tokens: %w", err)
	}

	return nil
}

func (r *OAuthProviderRepository) DeleteExpiredAccessTokens(ctx context.Context) (int64, error) {
	result, err := r.db.NewDelete().
		Model((*models.OAuthAccessToken)(nil)).
		Where("expires_at < ?", time.Now()).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to delete expired oauth access tokens: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

func (r *OAuthProviderRepository) CreateRefreshToken(ctx context.Context, token *models.OAuthRefreshToken) error {
	token.CreatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(token).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *OAuthProviderRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.OAuthRefreshToken, error) {
	token := new(models.OAuthRefreshToken)

	err := r.db.NewSelect().
		Model(token).
		Where("token_hash = ?", tokenHash).
		Relation("Client").
		Relation("User").
		Relation("AccessToken").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("oauth refresh token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth refresh token: %w", err)
	}

	return token, nil
}

func (r *OAuthProviderRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	now := time.Now()

	result, err := r.db.NewUpdate().
		Model((*models.OAuthRefreshToken)(nil)).
		Set("is_active = ?", false).
		Set("revoked_at = ?", now).
		Where("token_hash = ?", tokenHash).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke oauth refresh token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("oauth refresh token not found")
	}

	return nil
}

func (r *OAuthProviderRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID, clientID uuid.UUID) error {
	now := time.Now()

	_, err := r.db.NewUpdate().
		Model((*models.OAuthRefreshToken)(nil)).
		Set("is_active = ?", false).
		Set("revoked_at = ?", now).
		Where("user_id = ?", userID).
		Where("client_id = ?", clientID).
		Where("is_active = ?", true).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke all user refresh tokens: %w", err)
	}

	return nil
}

func (r *OAuthProviderRepository) RevokeAllClientRefreshTokens(ctx context.Context, clientID uuid.UUID) error {
	now := time.Now()

	_, err := r.db.NewUpdate().
		Model((*models.OAuthRefreshToken)(nil)).
		Set("is_active = ?", false).
		Set("revoked_at = ?", now).
		Where("client_id = ?", clientID).
		Where("is_active = ?", true).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke all client refresh tokens: %w", err)
	}

	return nil
}

func (r *OAuthProviderRepository) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	result, err := r.db.NewDelete().
		Model((*models.OAuthRefreshToken)(nil)).
		Where("expires_at < ?", time.Now()).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to delete expired oauth refresh tokens: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

func (r *OAuthProviderRepository) CreateOrUpdateConsent(ctx context.Context, consent *models.UserConsent) error {
	consent.GrantedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(consent).
		On("CONFLICT (user_id, client_id) DO UPDATE").
		Set("scopes = EXCLUDED.scopes").
		Set("granted_at = EXCLUDED.granted_at").
		Set("revoked_at = NULL").
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *OAuthProviderRepository) GetUserConsent(ctx context.Context, userID, clientID uuid.UUID) (*models.UserConsent, error) {
	consent := new(models.UserConsent)

	err := r.db.NewSelect().
		Model(consent).
		Where("user_id = ?", userID).
		Where("client_id = ?", clientID).
		Relation("Client").
		Relation("User").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user consent: %w", err)
	}

	return consent, nil
}

func (r *OAuthProviderRepository) RevokeConsent(ctx context.Context, userID, clientID uuid.UUID) error {
	now := time.Now()

	result, err := r.db.NewUpdate().
		Model((*models.UserConsent)(nil)).
		Set("revoked_at = ?", now).
		Where("user_id = ?", userID).
		Where("client_id = ?", clientID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke user consent: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user consent not found")
	}

	return nil
}

func (r *OAuthProviderRepository) ListUserConsents(ctx context.Context, userID uuid.UUID) ([]*models.UserConsent, error) {
	consents := make([]*models.UserConsent, 0)

	err := r.db.NewSelect().
		Model(&consents).
		Where("user_id = ?", userID).
		Relation("Client").
		Order("granted_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list user consents: %w", err)
	}

	return consents, nil
}

func (r *OAuthProviderRepository) ListClientConsents(ctx context.Context, clientID uuid.UUID) ([]*models.UserConsent, error) {
	consents := make([]*models.UserConsent, 0)

	err := r.db.NewSelect().
		Model(&consents).
		Where("client_id = ?", clientID).
		Relation("User").
		Order("granted_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list client consents: %w", err)
	}

	return consents, nil
}

func (r *OAuthProviderRepository) CreateDeviceCode(ctx context.Context, code *models.DeviceCode) error {
	code.CreatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(code).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *OAuthProviderRepository) GetDeviceCode(ctx context.Context, deviceCodeHash string) (*models.DeviceCode, error) {
	code := new(models.DeviceCode)

	err := r.db.NewSelect().
		Model(code).
		Where("device_code_hash = ?", deviceCodeHash).
		Relation("Client").
		Relation("User").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("device code not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get device code: %w", err)
	}

	return code, nil
}

func (r *OAuthProviderRepository) GetDeviceCodeByUserCode(ctx context.Context, userCode string) (*models.DeviceCode, error) {
	code := new(models.DeviceCode)

	err := r.db.NewSelect().
		Model(code).
		Where("user_code = ?", userCode).
		Relation("Client").
		Relation("User").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("device code not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get device code by user code: %w", err)
	}

	return code, nil
}

func (r *OAuthProviderRepository) UpdateDeviceCodeStatus(ctx context.Context, id uuid.UUID, status models.DeviceCodeStatus, userID *uuid.UUID) error {
	query := r.db.NewUpdate().
		Model((*models.DeviceCode)(nil)).
		Set("status = ?", status)

	if userID != nil {
		query = query.Set("user_id = ?", *userID)
	}

	result, err := query.
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update device code status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("device code not found")
	}

	return nil
}

func (r *OAuthProviderRepository) DeleteExpiredDeviceCodes(ctx context.Context) (int64, error) {
	result, err := r.db.NewDelete().
		Model((*models.DeviceCode)(nil)).
		Where("expires_at < ?", time.Now()).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to delete expired device codes: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

func (r *OAuthProviderRepository) CreateScope(ctx context.Context, scope *models.OAuthScope) error {
	scope.CreatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(scope).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *OAuthProviderRepository) GetScopeByName(ctx context.Context, name string) (*models.OAuthScope, error) {
	scope := new(models.OAuthScope)

	err := r.db.NewSelect().
		Model(scope).
		Where("name = ?", name).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("oauth scope not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth scope by name: %w", err)
	}

	return scope, nil
}

func (r *OAuthProviderRepository) ListScopes(ctx context.Context) ([]*models.OAuthScope, error) {
	scopes := make([]*models.OAuthScope, 0)

	err := r.db.NewSelect().
		Model(&scopes).
		Order("name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list oauth scopes: %w", err)
	}

	return scopes, nil
}

func (r *OAuthProviderRepository) ListSystemScopes(ctx context.Context) ([]*models.OAuthScope, error) {
	scopes := make([]*models.OAuthScope, 0)

	err := r.db.NewSelect().
		Model(&scopes).
		Where("is_system = ?", true).
		Order("name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list system oauth scopes: %w", err)
	}

	return scopes, nil
}

func (r *OAuthProviderRepository) DeleteScope(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.OAuthScope)(nil)).
		Where("id = ?", id).
		Where("is_system = ?", false).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete oauth scope: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("oauth scope not found or is a system scope")
	}

	return nil
}
