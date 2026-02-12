package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

const (
	loginAlertDevicesKeyPrefix  = "login_alert:devices:"
	loginAlertCooldownKeyPrefix = "login_alert:cooldown:"
	loginAlertDevicesTTL        = 90 * 24 * time.Hour // 90 days
	loginAlertCooldownTTL       = 1 * time.Hour
)

// LoginAlertService detects logins from new devices and sends email alerts.
type LoginAlertService struct {
	redis           *RedisService
	sessionRepo     SessionStore
	emailProfileSvc *EmailProfileService
	geoService      *GeoService
	logger          *logger.Logger
}

// LoginAlertParams contains all data needed to check and send a login alert.
type LoginAlertParams struct {
	UserID    uuid.UUID
	Username  string
	Email     string
	IP        string
	UserAgent string
	Device    models.DeviceInfo
	AppID     *uuid.UUID
	IsNewUser bool // true for signup / OAuth JIT — skip alert
}

// NewLoginAlertService creates a new LoginAlertService.
func NewLoginAlertService(
	redis *RedisService,
	sessionRepo SessionStore,
	emailProfileSvc *EmailProfileService,
	geoService *GeoService,
	log *logger.Logger,
) *LoginAlertService {
	return &LoginAlertService{
		redis:           redis,
		sessionRepo:     sessionRepo,
		emailProfileSvc: emailProfileSvc,
		geoService:      geoService,
		logger:          log,
	}
}

// CheckAndAlert checks whether the device is new for the user and sends an email alert if so.
// This method is designed to be called in a goroutine — it never returns an error,
// only logs warnings on failure.
func (s *LoginAlertService) CheckAndAlert(ctx context.Context, params LoginAlertParams) {
	// Skip conditions: no email, bot, or brand-new user (signup)
	if params.Email == "" || params.Device.IsBot || params.IsNewUser {
		return
	}

	fingerprint := computeFingerprint(params.Device)
	if fingerprint == "::" || fingerprint == "" {
		return
	}

	devicesKey := loginAlertDevicesKeyPrefix + params.UserID.String()

	// Check if the device SET exists (has any members)
	members, err := s.redis.SMembers(ctx, devicesKey)
	if err != nil {
		s.logger.Warn("login_alert: failed to read device set", map[string]interface{}{
			"user_id": params.UserID.String(),
			"error":   err.Error(),
		})
		return
	}

	if len(members) == 0 {
		// SET is empty — first time for this user. Seed from existing sessions.
		s.seedFromExistingSessions(ctx, params.UserID, devicesKey)
		// Add current fingerprint
		if err := s.redis.SAdd(ctx, devicesKey, fingerprint); err != nil {
			s.logger.Warn("login_alert: failed to add fingerprint after seed", map[string]interface{}{
				"error": err.Error(),
			})
		}
		_ = s.redis.Expire(ctx, devicesKey, loginAlertDevicesTTL)
		// No alert on first initialization
		return
	}

	// Check if fingerprint is already known
	known, err := s.redis.SIsMember(ctx, devicesKey, fingerprint)
	if err != nil {
		s.logger.Warn("login_alert: failed to check fingerprint membership", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if known {
		// Known device — refresh TTL and return
		_ = s.redis.Expire(ctx, devicesKey, loginAlertDevicesTTL)
		return
	}

	// NEW device detected — apply cooldown to avoid spam
	cooldownKey := fmt.Sprintf("%s%s:%s", loginAlertCooldownKeyPrefix, params.UserID.String(), fingerprint)
	set, err := s.redis.SetNX(ctx, cooldownKey, "1", loginAlertCooldownTTL)
	if err != nil {
		s.logger.Warn("login_alert: cooldown SetNX failed", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	if !set {
		// Cooldown active — already sent alert recently
		return
	}

	// Resolve geo location
	location := "Unknown"
	if s.geoService != nil {
		if geo := s.geoService.GetLocation(ctx, params.IP); geo != nil {
			if geo.City != "" && geo.CountryName != "" {
				location = geo.City + ", " + geo.CountryName
			} else if geo.CountryName != "" {
				location = geo.CountryName
			}
		}
	}

	// Send email alert
	variables := map[string]interface{}{
		"username":    params.Username,
		"email":       params.Email,
		"ip_address":  params.IP,
		"user_agent":  params.UserAgent,
		"device_type": params.Device.DeviceType,
		"location":    location,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05 MST"),
	}

	if s.emailProfileSvc != nil {
		if err := s.emailProfileSvc.SendEmail(ctx, nil, params.AppID, params.Email, "login_alert", variables); err != nil {
			s.logger.Warn("login_alert: failed to send email", map[string]interface{}{
				"user_id": params.UserID.String(),
				"error":   err.Error(),
			})
		}
	}

	// Register the new fingerprint
	if err := s.redis.SAdd(ctx, devicesKey, fingerprint); err != nil {
		s.logger.Warn("login_alert: failed to register new fingerprint", map[string]interface{}{
			"error": err.Error(),
		})
	}
	_ = s.redis.Expire(ctx, devicesKey, loginAlertDevicesTTL)
}

// seedFromExistingSessions populates the Redis device SET from the user's active DB sessions.
func (s *LoginAlertService) seedFromExistingSessions(ctx context.Context, userID uuid.UUID, devicesKey string) {
	sessions, err := s.sessionRepo.GetUserSessions(ctx, userID)
	if err != nil {
		s.logger.Warn("login_alert: failed to get sessions for seed", map[string]interface{}{
			"user_id": userID.String(),
			"error":   err.Error(),
		})
		return
	}

	for _, sess := range sessions {
		fp := computeFingerprint(models.DeviceInfo{
			DeviceType: sess.DeviceType,
			OS:         sess.OS,
			Browser:    sess.Browser,
		})
		if fp != "" && fp != "::" {
			_ = s.redis.SAdd(ctx, devicesKey, fp)
		}
	}

	if len(sessions) > 0 {
		_ = s.redis.Expire(ctx, devicesKey, loginAlertDevicesTTL)
	}
}

// computeFingerprint builds a device fingerprint string: "devicetype:os:browser"
// Versions are stripped so that browser/OS updates don't trigger false positives.
func computeFingerprint(device models.DeviceInfo) string {
	return stripVersion(device.DeviceType) + ":" + stripVersion(device.OS) + ":" + stripVersion(device.Browser)
}

// versionRegex matches the first digit (possibly preceded by a space) and everything after it.
var versionRegex = regexp.MustCompile(`[\s.]*\d.*$`)

// stripVersion removes version numbers from a string and lowercases the result.
// "Chrome 120.0.6099" → "chrome", "macOS 14.2" → "macos", "iOS 17.2" → "ios"
func stripVersion(s string) string {
	s = versionRegex.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	// Remove remaining spaces (e.g. "Mac OS" → "macos")
	s = strings.ReplaceAll(s, " ", "")
	return s
}
