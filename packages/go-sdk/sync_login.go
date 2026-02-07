package authgateway

import (
	"context"
	"database/sql"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/proto"
)

type LoginSyncer struct {
	db     *sql.DB
	logger Logger
}

func (ls *LoginSyncer) SyncOnLogin(ctx context.Context, validation *proto.ValidateTokenResponse) error {
	if validation == nil {
		return nil
	}

	userID := validation.GetUserId()
	if userID == "" {
		return nil
	}

	_, err := ls.db.ExecContext(ctx, upsertUserSQL,
		userID,
		validation.GetEmail(),
		validation.GetUsername(),
		"",
		validation.GetIsActive(),
	)
	if err != nil {
		ls.logger.Error("login sync: upsert user %s failed: %v", userID, err)
		return err
	}

	ls.logger.Info("login sync: synced user %s", userID)
	return nil
}
