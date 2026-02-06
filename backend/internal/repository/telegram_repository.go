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

type TelegramBotRepository struct {
	db *Database
}

func NewTelegramBotRepository(db *Database) *TelegramBotRepository {
	return &TelegramBotRepository{db: db}
}

func (r *TelegramBotRepository) Create(ctx context.Context, bot *models.TelegramBot) error {
	bot.CreatedAt = time.Now()
	bot.UpdatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(bot).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *TelegramBotRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.TelegramBot, error) {
	bot := new(models.TelegramBot)

	err := r.db.NewSelect().
		Model(bot).
		Where("tb.id = ?", id).
		Relation("Application").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("telegram bot not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get telegram bot by id: %w", err)
	}

	return bot, nil
}

func (r *TelegramBotRepository) ListByApp(ctx context.Context, appID uuid.UUID) ([]*models.TelegramBot, error) {
	bots := make([]*models.TelegramBot, 0)

	err := r.db.NewSelect().
		Model(&bots).
		Where("application_id = ?", appID).
		Order("display_name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list telegram bots by app: %w", err)
	}

	return bots, nil
}

func (r *TelegramBotRepository) ListAuthBotsByApp(ctx context.Context, appID uuid.UUID) ([]*models.TelegramBot, error) {
	bots := make([]*models.TelegramBot, 0)

	err := r.db.NewSelect().
		Model(&bots).
		Where("application_id = ?", appID).
		Where("is_auth_bot = ?", true).
		Where("is_active = ?", true).
		Order("display_name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list auth bots by app: %w", err)
	}

	return bots, nil
}

func (r *TelegramBotRepository) Update(ctx context.Context, bot *models.TelegramBot) error {
	bot.UpdatedAt = time.Now()

	result, err := r.db.NewUpdate().
		Model(bot).
		Column("bot_token", "bot_username", "display_name", "is_auth_bot", "is_active", "updated_at").
		WherePK().
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update telegram bot: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("telegram bot not found")
	}

	return nil
}

func (r *TelegramBotRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.TelegramBot)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete telegram bot: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("telegram bot not found")
	}

	return nil
}

type UserTelegramRepository struct {
	db *Database
}

func NewUserTelegramRepository(db *Database) *UserTelegramRepository {
	return &UserTelegramRepository{db: db}
}

func (r *UserTelegramRepository) CreateAccount(ctx context.Context, account *models.UserTelegramAccount) error {
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(account).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *UserTelegramRepository) GetAccountByUserAndTgID(ctx context.Context, userID uuid.UUID, telegramUserID int64) (*models.UserTelegramAccount, error) {
	account := new(models.UserTelegramAccount)

	err := r.db.NewSelect().
		Model(account).
		Where("user_id = ?", userID).
		Where("telegram_user_id = ?", telegramUserID).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("telegram account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get telegram account by user and tg id: %w", err)
	}

	return account, nil
}

func (r *UserTelegramRepository) GetAccountByTgID(ctx context.Context, telegramUserID int64) (*models.UserTelegramAccount, error) {
	account := new(models.UserTelegramAccount)

	err := r.db.NewSelect().
		Model(account).
		Where("telegram_user_id = ?", telegramUserID).
		Relation("User").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("telegram account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get telegram account by tg id: %w", err)
	}

	return account, nil
}

func (r *UserTelegramRepository) ListAccountsByUser(ctx context.Context, userID uuid.UUID) ([]*models.UserTelegramAccount, error) {
	accounts := make([]*models.UserTelegramAccount, 0)

	err := r.db.NewSelect().
		Model(&accounts).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list telegram accounts by user: %w", err)
	}

	return accounts, nil
}

func (r *UserTelegramRepository) UpdateAccount(ctx context.Context, account *models.UserTelegramAccount) error {
	account.UpdatedAt = time.Now()

	result, err := r.db.NewUpdate().
		Model(account).
		Column("username", "first_name", "last_name", "photo_url", "auth_date", "updated_at").
		WherePK().
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update telegram account: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("telegram account not found")
	}

	return nil
}

func (r *UserTelegramRepository) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.UserTelegramAccount)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete telegram account: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("telegram account not found")
	}

	return nil
}

func (r *UserTelegramRepository) CreateBotAccess(ctx context.Context, access *models.UserTelegramBotAccess) error {
	access.CreatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(access).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *UserTelegramRepository) GetBotAccess(ctx context.Context, userID, botID uuid.UUID) (*models.UserTelegramBotAccess, error) {
	access := new(models.UserTelegramBotAccess)

	err := r.db.NewSelect().
		Model(access).
		Where("user_id = ?", userID).
		Where("telegram_bot_id = ?", botID).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("bot access not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bot access: %w", err)
	}

	return access, nil
}

func (r *UserTelegramRepository) ListBotAccessByUser(ctx context.Context, userID uuid.UUID) ([]*models.UserTelegramBotAccess, error) {
	accesses := make([]*models.UserTelegramBotAccess, 0)

	err := r.db.NewSelect().
		Model(&accesses).
		Where("user_id = ?", userID).
		Relation("TelegramBot").
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list bot accesses by user: %w", err)
	}

	return accesses, nil
}

func (r *UserTelegramRepository) ListBotAccessByUserAndApp(ctx context.Context, userID, appID uuid.UUID) ([]*models.UserTelegramBotAccess, error) {
	accesses := make([]*models.UserTelegramBotAccess, 0)

	err := r.db.NewSelect().
		Model(&accesses).
		Join("JOIN telegram_bots AS tb ON tb.id = utba.telegram_bot_id").
		Where("utba.user_id = ?", userID).
		Where("tb.application_id = ?", appID).
		Relation("TelegramBot").
		Order("utba.created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list bot accesses by user and app: %w", err)
	}

	return accesses, nil
}

func (r *UserTelegramRepository) UpdateBotAccess(ctx context.Context, access *models.UserTelegramBotAccess) error {
	result, err := r.db.NewUpdate().
		Model(access).
		Column("can_send_messages", "authorized_via").
		WherePK().
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update bot access: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("bot access not found")
	}

	return nil
}

func (r *UserTelegramRepository) DeleteBotAccess(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.UserTelegramBotAccess)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete bot access: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("bot access not found")
	}

	return nil
}
