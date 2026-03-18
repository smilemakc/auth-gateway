package service

type AdminService struct {
	*AdminUserService
	*AdminAPIKeyService
	*AdminAuditService
	*AdminStatsService
	*AdminBulkService
}

func NewAdminService(
	userRepo UserStore,
	apiKeyRepo APIKeyStore,
	auditRepo AuditStore,
	oauthRepo OAuthStore,
	rbacRepo RBACStore,
	backupCodeRepo BackupCodeStore,
	appRepo ApplicationStore,
	bcryptCost int,
	db TransactionDB,
) *AdminService {
	return &AdminService{
		AdminUserService: &AdminUserService{
			userRepo:       userRepo,
			rbacRepo:       rbacRepo,
			oauthRepo:      oauthRepo,
			backupCodeRepo: backupCodeRepo,
			apiKeyRepo:     apiKeyRepo,
			auditRepo:      auditRepo,
			appRepo:        appRepo,
			bcryptCost:     bcryptCost,
			db:             db,
		},
		AdminAPIKeyService: &AdminAPIKeyService{
			apiKeyRepo: apiKeyRepo,
			userRepo:   userRepo,
		},
		AdminAuditService: &AdminAuditService{
			auditRepo: auditRepo,
		},
		AdminStatsService: &AdminStatsService{
			userRepo:   userRepo,
			auditRepo:  auditRepo,
			apiKeyRepo: apiKeyRepo,
			oauthRepo:  oauthRepo,
			rbacRepo:   rbacRepo,
		},
		AdminBulkService: &AdminBulkService{
			userRepo:   userRepo,
			appRepo:    appRepo,
			bcryptCost: bcryptCost,
		},
	}
}
