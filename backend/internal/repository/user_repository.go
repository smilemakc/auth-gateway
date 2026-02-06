package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/queryopt"
	"github.com/uptrace/bun"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	db *Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.NewInsert().
		Model(user).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

// CreateWithTx creates a new user within a transaction
func (r *UserRepository) CreateWithTx(ctx context.Context, tx bun.Tx, user *models.User) error {
	_, err := tx.NewInsert().
		Model(user).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID, isActive *bool, opts ...queryopt.UserGetOption) (*models.User, error) {
	o := queryopt.BuildUserGetOptions(opts)
	user := new(models.User)

	query := r.db.NewSelect().
		Model(user).
		Where("id = ?", id)

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	if o.WithRoles {
		query = query.Relation("Roles")
	}

	err := query.Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string, isActive *bool, opts ...queryopt.UserGetOption) (*models.User, error) {
	o := queryopt.BuildUserGetOptions(opts)
	user := new(models.User)

	query := r.db.NewSelect().
		Model(user).
		Where("email = ?", email)

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	if o.WithRoles {
		query = query.Relation("Roles")
	}

	err := query.Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string, isActive *bool, opts ...queryopt.UserGetOption) (*models.User, error) {
	o := queryopt.BuildUserGetOptions(opts)
	user := new(models.User)

	query := r.db.NewSelect().
		Model(user).
		Where("username = ?", username)

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	if o.WithRoles {
		query = query.Relation("Roles")
	}

	err := query.Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	result, err := r.db.NewUpdate().
		Model(user).
		Column("full_name", "profile_picture_url", "updated_at").
		WherePK().
		Returning("updated_at").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	result, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("password_hash = ?", passwordHash).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// Delete soft deletes a user (sets is_active to false)
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("is_active = ?", false).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Where("email = ?", email).
		Exists(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// UsernameExists checks if a username already exists
func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Where("username = ?", username).
		Exists(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return exists, nil
}

// MarkEmailVerified marks a user's email as verified
func (r *UserRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("email_verified = ?", true).
		Set("email_verified_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// List retrieves a list of users with pagination
func (r *UserRepository) List(ctx context.Context, opts ...queryopt.UserListOption) ([]*models.User, error) {
	o := queryopt.BuildUserListOptions(opts)
	users := make([]*models.User, 0)

	query := r.db.NewSelect().
		Model(&users)

	if o.IsActive != nil {
		query = query.Where("is_active = ?", *o.IsActive)
	}
	if o.WithRoles {
		query = query.Relation("Roles")
	}

	err := query.
		Order("created_at DESC").
		Limit(o.Limit).
		Offset(o.Offset).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context, isActive *bool) (int, error) {
	query := r.db.NewSelect().
		Model((*models.User)(nil))

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	count, err := query.Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// UpdateTOTPSecret updates the TOTP secret for a user
func (r *UserRepository) UpdateTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error {
	result, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("totp_secret = ?", secret).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update TOTP secret: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// EnableTOTP enables TOTP 2FA for a user
func (r *UserRepository) EnableTOTP(ctx context.Context, userID uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("totp_enabled = ?", true).
		Set("totp_enabled_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to enable TOTP: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// DisableTOTP disables TOTP 2FA for a user
func (r *UserRepository) DisableTOTP(ctx context.Context, userID uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("totp_enabled = ?", false).
		Set("totp_secret = ?", nil).
		Set("totp_enabled_at = ?", nil).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to disable TOTP: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// GetByPhone retrieves a user by phone number
func (r *UserRepository) GetByPhone(ctx context.Context, phone string, isActive *bool, opts ...queryopt.UserGetOption) (*models.User, error) {
	o := queryopt.BuildUserGetOptions(opts)
	user := new(models.User)

	query := r.db.NewSelect().
		Model(user).
		Where("phone = ?", phone)

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	if o.WithRoles {
		query = query.Relation("Roles")
	}

	err := query.Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}

	return user, nil
}

// PhoneExists checks if a phone number already exists
func (r *UserRepository) PhoneExists(ctx context.Context, phone string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Where("phone = ?", phone).
		Exists(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check phone existence: %w", err)
	}

	return exists, nil
}

// MarkPhoneVerified marks a user's phone as verified
func (r *UserRepository) MarkPhoneVerified(ctx context.Context, userID uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("phone_verified = ?", true).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to mark phone as verified: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrUserNotFound
	}

	return nil
}
