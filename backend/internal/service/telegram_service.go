package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

var (
	ErrBotNotFound      = errors.New("telegram bot not found")
	ErrAccountNotFound  = errors.New("telegram account not found")
	ErrBotAccessExists  = errors.New("bot access already exists")
	ErrInvalidAuthData  = errors.New("invalid telegram auth data")
)

type TelegramService struct {
	botRepo    TelegramBotStore
	userTgRepo UserTelegramStore
	appRepo    ApplicationStore
	log        *logger.Logger
}

func NewTelegramService(botRepo TelegramBotStore, userTgRepo UserTelegramStore, appRepo ApplicationStore, log *logger.Logger) *TelegramService {
	return &TelegramService{
		botRepo:    botRepo,
		userTgRepo: userTgRepo,
		appRepo:    appRepo,
		log:        log,
	}
}

func (s *TelegramService) CreateBot(ctx context.Context, appID uuid.UUID, req *models.CreateTelegramBotRequest) (*models.TelegramBot, error) {
	if _, err := s.appRepo.GetApplicationByID(ctx, appID); err != nil {
		return nil, ErrApplicationNotFound
	}

	bot := &models.TelegramBot{
		ID:            uuid.New(),
		ApplicationID: appID,
		BotToken:      req.BotToken,
		BotUsername:   req.BotUsername,
		DisplayName:   req.DisplayName,
		IsAuthBot:     req.IsAuthBot,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.botRepo.Create(ctx, bot); err != nil {
		s.log.Error("failed to create telegram bot", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
			"bot_username":   req.BotUsername,
		})
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	s.log.Info("telegram bot created", map[string]interface{}{
		"bot_id":         bot.ID.String(),
		"application_id": appID.String(),
		"bot_username":   req.BotUsername,
		"is_auth_bot":    req.IsAuthBot,
	})

	return bot, nil
}

func (s *TelegramService) GetBot(ctx context.Context, id uuid.UUID) (*models.TelegramBot, error) {
	bot, err := s.botRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrBotNotFound
	}
	return bot, nil
}

func (s *TelegramService) ListBotsByApp(ctx context.Context, appID uuid.UUID) ([]*models.TelegramBot, error) {
	return s.botRepo.ListByApp(ctx, appID)
}

func (s *TelegramService) ListAuthBotsByApp(ctx context.Context, appID uuid.UUID) ([]*models.TelegramBot, error) {
	return s.botRepo.ListAuthBotsByApp(ctx, appID)
}

func (s *TelegramService) UpdateBot(ctx context.Context, id uuid.UUID, req *models.UpdateTelegramBotRequest) (*models.TelegramBot, error) {
	bot, err := s.botRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrBotNotFound
	}

	if req.BotToken != nil {
		bot.BotToken = *req.BotToken
	}
	if req.BotUsername != nil {
		bot.BotUsername = *req.BotUsername
	}
	if req.DisplayName != nil {
		bot.DisplayName = *req.DisplayName
	}
	if req.IsAuthBot != nil {
		bot.IsAuthBot = *req.IsAuthBot
	}
	if req.IsActive != nil {
		bot.IsActive = *req.IsActive
	}

	bot.UpdatedAt = time.Now()

	if err := s.botRepo.Update(ctx, bot); err != nil {
		s.log.Error("failed to update telegram bot", map[string]interface{}{
			"error":  err.Error(),
			"bot_id": id.String(),
		})
		return nil, fmt.Errorf("failed to update telegram bot: %w", err)
	}

	s.log.Info("telegram bot updated", map[string]interface{}{
		"bot_id":       id.String(),
		"bot_username": bot.BotUsername,
	})

	return s.botRepo.GetByID(ctx, id)
}

func (s *TelegramService) DeleteBot(ctx context.Context, id uuid.UUID) error {
	if _, err := s.botRepo.GetByID(ctx, id); err != nil {
		return ErrBotNotFound
	}

	if err := s.botRepo.Delete(ctx, id); err != nil {
		s.log.Error("failed to delete telegram bot", map[string]interface{}{
			"error":  err.Error(),
			"bot_id": id.String(),
		})
		return fmt.Errorf("failed to delete telegram bot: %w", err)
	}

	s.log.Info("telegram bot deleted", map[string]interface{}{
		"bot_id": id.String(),
	})

	return nil
}

func (s *TelegramService) GetOrCreateAccount(ctx context.Context, userID uuid.UUID, telegramUserID int64, username, firstName, lastName, photoURL string, authDate time.Time) (*models.UserTelegramAccount, error) {
	account, err := s.userTgRepo.GetAccountByUserAndTgID(ctx, userID, telegramUserID)
	if err == nil {
		usernamePtr := &username
		if username == "" {
			usernamePtr = nil
		}
		lastNamePtr := &lastName
		if lastName == "" {
			lastNamePtr = nil
		}
		photoURLPtr := &photoURL
		if photoURL == "" {
			photoURLPtr = nil
		}

		account.Username = usernamePtr
		account.FirstName = firstName
		account.LastName = lastNamePtr
		account.PhotoURL = photoURLPtr
		account.AuthDate = authDate
		account.UpdatedAt = time.Now()

		if err := s.userTgRepo.UpdateAccount(ctx, account); err != nil {
			s.log.Warn("failed to update telegram account", map[string]interface{}{
				"error":            err.Error(),
				"user_id":          userID.String(),
				"telegram_user_id": telegramUserID,
			})
		}

		return account, nil
	}

	usernamePtr := &username
	if username == "" {
		usernamePtr = nil
	}
	lastNamePtr := &lastName
	if lastName == "" {
		lastNamePtr = nil
	}
	photoURLPtr := &photoURL
	if photoURL == "" {
		photoURLPtr = nil
	}

	account = &models.UserTelegramAccount{
		ID:             uuid.New(),
		UserID:         userID,
		TelegramUserID: telegramUserID,
		Username:       usernamePtr,
		FirstName:      firstName,
		LastName:       lastNamePtr,
		PhotoURL:       photoURLPtr,
		AuthDate:       authDate,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.userTgRepo.CreateAccount(ctx, account); err != nil {
		s.log.Error("failed to create telegram account", map[string]interface{}{
			"error":            err.Error(),
			"user_id":          userID.String(),
			"telegram_user_id": telegramUserID,
		})
		return nil, fmt.Errorf("failed to create telegram account: %w", err)
	}

	s.log.Info("telegram account created", map[string]interface{}{
		"account_id":       account.ID.String(),
		"user_id":          userID.String(),
		"telegram_user_id": telegramUserID,
	})

	return account, nil
}

func (s *TelegramService) ListAccountsByUser(ctx context.Context, userID uuid.UUID) ([]*models.UserTelegramAccount, error) {
	return s.userTgRepo.ListAccountsByUser(ctx, userID)
}

func (s *TelegramService) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	if err := s.userTgRepo.DeleteAccount(ctx, id); err != nil {
		s.log.Error("failed to delete telegram account", map[string]interface{}{
			"error":      err.Error(),
			"account_id": id.String(),
		})
		return fmt.Errorf("failed to delete telegram account: %w", err)
	}

	s.log.Info("telegram account deleted", map[string]interface{}{
		"account_id": id.String(),
	})

	return nil
}

func (s *TelegramService) GrantBotAccess(ctx context.Context, userID, botID, accountID uuid.UUID) (*models.UserTelegramBotAccess, error) {
	existing, err := s.userTgRepo.GetBotAccess(ctx, userID, botID)
	if err == nil && existing != nil {
		return existing, nil
	}

	access := &models.UserTelegramBotAccess{
		ID:                uuid.New(),
		UserID:            userID,
		TelegramBotID:     botID,
		TelegramAccountID: accountID,
		CanSendMessages:   true,
		AuthorizedVia:     true,
		CreatedAt:         time.Now(),
	}

	if err := s.userTgRepo.CreateBotAccess(ctx, access); err != nil {
		s.log.Error("failed to grant bot access", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
			"bot_id":  botID.String(),
		})
		return nil, fmt.Errorf("failed to grant bot access: %w", err)
	}

	s.log.Info("bot access granted", map[string]interface{}{
		"access_id": access.ID.String(),
		"user_id":   userID.String(),
		"bot_id":    botID.String(),
	})

	return access, nil
}

func (s *TelegramService) ListBotAccessByUser(ctx context.Context, userID uuid.UUID) ([]*models.UserTelegramBotAccess, error) {
	return s.userTgRepo.ListBotAccessByUser(ctx, userID)
}

func (s *TelegramService) ListBotAccessByUserAndApp(ctx context.Context, userID, appID uuid.UUID) ([]*models.UserTelegramBotAccess, error) {
	return s.userTgRepo.ListBotAccessByUserAndApp(ctx, userID, appID)
}

func (s *TelegramService) UpdateBotAccess(ctx context.Context, id uuid.UUID, canSendMessages bool) error {
	access, err := s.userTgRepo.GetBotAccess(ctx, uuid.Nil, uuid.Nil)
	if err != nil {
		return fmt.Errorf("failed to get bot access: %w", err)
	}

	access.CanSendMessages = canSendMessages

	if err := s.userTgRepo.UpdateBotAccess(ctx, access); err != nil {
		s.log.Error("failed to update bot access", map[string]interface{}{
			"error":     err.Error(),
			"access_id": id.String(),
		})
		return fmt.Errorf("failed to update bot access: %w", err)
	}

	s.log.Info("bot access updated", map[string]interface{}{
		"access_id":         id.String(),
		"can_send_messages": canSendMessages,
	})

	return nil
}

func (s *TelegramService) RevokeBotAccess(ctx context.Context, id uuid.UUID) error {
	if err := s.userTgRepo.DeleteBotAccess(ctx, id); err != nil {
		s.log.Error("failed to revoke bot access", map[string]interface{}{
			"error":     err.Error(),
			"access_id": id.String(),
		})
		return fmt.Errorf("failed to revoke bot access: %w", err)
	}

	s.log.Info("bot access revoked", map[string]interface{}{
		"access_id": id.String(),
	})

	return nil
}

func (s *TelegramService) VerifyTelegramAuth(botToken string, data map[string]string) bool {
	hash := data["hash"]
	if hash == "" {
		return false
	}

	var pairs []string
	for k, v := range data {
		if k == "hash" {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	checkString := strings.Join(pairs, "\n")

	secretKey := sha256.Sum256([]byte(botToken))

	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(checkString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedHash), []byte(hash))
}
