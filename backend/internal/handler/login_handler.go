package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

const (
	// Cookie names
	sessionCookieName = "auth_session"
	otpStateCookie    = "otp_login_state"

	// Cookie settings
	sessionMaxAge  = 24 * 60 * 60 // 24 hours in seconds
	otpStateMaxAge = 10 * 60      // 10 minutes for OTP flow
)

// LoginHandler handles the login page for OAuth flows
type LoginHandler struct {
	authService  *service.AuthService
	otpService   *service.OTPService
	jwtService   *jwt.Service
	logger       *logger.Logger
	secureCookie bool
}

// NewLoginHandler creates a new login handler
func NewLoginHandler(
	authService *service.AuthService,
	otpService *service.OTPService,
	jwtService *jwt.Service,
	logger *logger.Logger,
	secureCookie bool,
) *LoginHandler {
	return &LoginHandler{
		authService:  authService,
		otpService:   otpService,
		jwtService:   jwtService,
		logger:       logger,
		secureCookie: secureCookie,
	}
}

// LoginPage renders the login page
// @Summary Login Page
// @Description HTML login page for OAuth authorization flows
// @Tags OAuth Provider - Login
// @Produce html
// @Param return_to query string false "URL to redirect after successful login"
// @Success 200 {string} string "HTML login page"
// @Router /login [get]
func (h *LoginHandler) LoginPage(c *gin.Context) {
	returnTo := c.Query("return_to")
	errorMsg := c.Query("error")

	// Check if user already has a valid session
	if h.isAuthenticated(c) {
		if returnTo != "" {
			c.Redirect(http.StatusTemporaryRedirect, returnTo)
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	html := h.renderLoginPage(returnTo, errorMsg, "", "")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// LoginSubmit handles login form submission
// @Summary Submit Login
// @Description Process login form submission with password or initiate OTP flow
// @Tags OAuth Provider - Login
// @Accept application/x-www-form-urlencoded
// @Produce html
// @Param login_type formData string true "Login type: password or otp"
// @Param identifier formData string true "Email, phone, or username"
// @Param password formData string false "Password (for password login)"
// @Param return_to formData string false "URL to redirect after login"
// @Success 302 {string} string "Redirect to return_to or OTP verification"
// @Failure 200 {string} string "HTML login page with error"
// @Router /login [post]
func (h *LoginHandler) LoginSubmit(c *gin.Context) {
	loginType := c.PostForm("login_type")
	identifier := c.PostForm("identifier")
	password := c.PostForm("password")
	returnTo := c.PostForm("return_to")

	if identifier == "" {
		h.renderLoginError(c, returnTo, "Please enter email, phone, or username", "", "")
		return
	}

	switch loginType {
	case "password":
		h.handlePasswordLogin(c, identifier, password, returnTo)
	case "otp":
		h.handleOTPRequest(c, identifier, returnTo)
	default:
		h.renderLoginError(c, returnTo, "Invalid login type", "", "")
	}
}

// OTPVerifyPage renders the OTP verification page
// @Summary OTP Verification Page
// @Description HTML page for entering OTP code
// @Tags OAuth Provider - Login
// @Produce html
// @Success 200 {string} string "HTML OTP verification page"
// @Router /login/otp [get]
func (h *LoginHandler) OTPVerifyPage(c *gin.Context) {
	returnTo := c.Query("return_to")
	identifier := c.Query("identifier")
	errorMsg := c.Query("error")

	// Verify OTP state cookie exists
	_, err := c.Cookie(otpStateCookie)
	if err != nil {
		loginURL := "/login"
		if returnTo != "" {
			loginURL = fmt.Sprintf("/login?return_to=%s&error=%s",
				url.QueryEscape(returnTo),
				url.QueryEscape("OTP session expired, please try again"))
		}
		c.Redirect(http.StatusTemporaryRedirect, loginURL)
		return
	}

	html := h.renderOTPVerifyPage(identifier, returnTo, errorMsg)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// OTPVerifySubmit handles OTP code verification
// @Summary Verify OTP Code
// @Description Verify the OTP code and complete login
// @Tags OAuth Provider - Login
// @Accept application/x-www-form-urlencoded
// @Produce html
// @Param code formData string true "6-digit OTP code"
// @Param identifier formData string true "Email or phone"
// @Param return_to formData string false "URL to redirect after login"
// @Success 302 {string} string "Redirect to return_to"
// @Failure 200 {string} string "HTML OTP page with error"
// @Router /login/otp [post]
func (h *LoginHandler) OTPVerifySubmit(c *gin.Context) {
	code := c.PostForm("code")
	identifier := c.PostForm("identifier")
	returnTo := c.PostForm("return_to")

	// Verify OTP state cookie
	_, err := c.Cookie(otpStateCookie)
	if err != nil {
		h.redirectToLogin(c, returnTo, "OTP session expired, please try again")
		return
	}

	if code == "" || len(code) != 6 {
		h.renderOTPError(c, identifier, returnTo, "Please enter a valid 6-digit code")
		return
	}

	// Determine if identifier is email or phone
	var verifyReq *models.VerifyOTPRequest
	if utils.IsValidEmail(identifier) {
		verifyReq = &models.VerifyOTPRequest{
			Email: &identifier,
			Code:  code,
			Type:  models.OTPTypeLogin,
		}
	} else if utils.IsValidPhone(identifier) {
		normalizedPhone := utils.NormalizePhone(identifier)
		verifyReq = &models.VerifyOTPRequest{
			Phone: &normalizedPhone,
			Code:  code,
			Type:  models.OTPTypeLogin,
		}
	} else {
		h.renderOTPError(c, identifier, returnTo, "Invalid identifier format")
		return
	}

	// Verify OTP
	response, err := h.otpService.VerifyOTP(c.Request.Context(), verifyReq)
	if err != nil {
		h.logger.Error("OTP verification failed", map[string]interface{}{
			"error":      err.Error(),
			"identifier": identifier,
		})
		h.renderOTPError(c, identifier, returnTo, "Verification failed, please try again")
		return
	}

	if !response.Valid || response.User == nil {
		h.renderOTPError(c, identifier, returnTo, "Invalid or expired code")
		return
	}

	// Clear OTP state cookie
	c.SetCookie(otpStateCookie, "", -1, "/", "", h.secureCookie, true)

	// Create session and redirect
	h.createSessionAndRedirect(c, response.User, returnTo)
}

// Logout handles user logout
// @Summary Logout
// @Description Clear session and redirect to login
// @Tags OAuth Provider - Login
// @Success 302 {string} string "Redirect to login page"
// @Router /logout [get]
func (h *LoginHandler) Logout(c *gin.Context) {
	c.SetCookie(sessionCookieName, "", -1, "/", "", h.secureCookie, true)

	returnTo := c.Query("return_to")
	if returnTo != "" {
		c.Redirect(http.StatusTemporaryRedirect, returnTo)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, "/login")
}

// SessionMiddleware validates session cookies and sets user context
func (h *LoginHandler) SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionToken, err := c.Cookie(sessionCookieName)
		if err != nil || sessionToken == "" {
			c.Next()
			return
		}

		claims, err := h.jwtService.ValidateAccessToken(sessionToken)
		if err != nil {
			c.SetCookie(sessionCookieName, "", -1, "/", "", h.secureCookie, true)
			c.Next()
			return
		}

		c.Set(utils.UserIDKey, claims.UserID)
		c.Set(utils.UserEmailKey, claims.Email)
		c.Set(utils.UserRolesKey, claims.Roles)

		c.Next()
	}
}

// handlePasswordLogin processes password-based login
func (h *LoginHandler) handlePasswordLogin(c *gin.Context, identifier, password, returnTo string) {
	if password == "" {
		h.renderLoginError(c, returnTo, "Password is required", identifier, "password")
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)
	deviceInfo := utils.GetDeviceInfoFromContext(c)

	var signInReq models.SignInRequest
	if utils.IsValidEmail(identifier) {
		signInReq.Email = utils.NormalizeEmail(identifier)
	} else if utils.IsValidPhone(identifier) {
		normalizedPhone := utils.NormalizePhone(identifier)
		signInReq.Phone = &normalizedPhone
	} else {
		// Assume it's a username, try as email
		signInReq.Email = identifier
	}
	signInReq.Password = password

	// SignIn internally calls generateAuthResponse which creates session via SessionService
	appID, _ := utils.GetApplicationIDFromContext(c)
	authResp, err := h.authService.SignIn(c.Request.Context(), &signInReq, ip, userAgent, deviceInfo, appID)
	if err != nil {
		h.logger.Error("Password login failed", map[string]interface{}{
			"error":      err.Error(),
			"identifier": identifier,
			"ip":         ip,
		})

		errorMsg := "Invalid credentials"
		if appErr, ok := err.(*models.AppError); ok {
			errorMsg = appErr.Message
		}
		h.renderLoginError(c, returnTo, errorMsg, identifier, "password")
		return
	}

	// Check if 2FA is required
	if authResp.Requires2FA {
		h.renderLoginError(c, returnTo, "Two-factor authentication is required. Please use the API.", identifier, "password")
		return
	}

	// Set session cookie with the access token
	h.setSessionCookie(c, authResp.AccessToken)

	if returnTo != "" {
		c.Redirect(http.StatusTemporaryRedirect, returnTo)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// handleOTPRequest initiates OTP-based login
func (h *LoginHandler) handleOTPRequest(c *gin.Context, identifier, returnTo string) {
	var otpReq *models.SendOTPRequest
	appID, _ := utils.GetApplicationIDFromContext(c)

	if utils.IsValidEmail(identifier) {
		normalizedEmail := utils.NormalizeEmail(identifier)
		otpReq = &models.SendOTPRequest{
			Email:         &normalizedEmail,
			Type:          models.OTPTypeLogin,
			ApplicationID: appID,
		}
	} else if utils.IsValidPhone(identifier) {
		normalizedPhone := utils.NormalizePhone(identifier)
		otpReq = &models.SendOTPRequest{
			Phone:         &normalizedPhone,
			Type:          models.OTPTypeLogin,
			ApplicationID: appID,
		}
	} else {
		h.renderLoginError(c, returnTo, "Please enter a valid email or phone number for OTP login", identifier, "otp")
		return
	}

	if err := h.otpService.SendOTP(c.Request.Context(), otpReq); err != nil {
		h.logger.Error("Failed to send OTP", map[string]interface{}{
			"error":      err.Error(),
			"identifier": identifier,
		})

		errorMsg := "Failed to send verification code"
		if appErr, ok := err.(*models.AppError); ok {
			errorMsg = appErr.Message
		}
		h.renderLoginError(c, returnTo, errorMsg, identifier, "otp")
		return
	}

	c.SetCookie(otpStateCookie, identifier, otpStateMaxAge, "/", "", h.secureCookie, true)

	otpURL := fmt.Sprintf("/login/otp?identifier=%s", url.QueryEscape(identifier))
	if returnTo != "" {
		otpURL += "&return_to=" + url.QueryEscape(returnTo)
	}
	c.Redirect(http.StatusTemporaryRedirect, otpURL)
}

// createSessionAndRedirect creates a session for the user and redirects
func (h *LoginHandler) createSessionAndRedirect(c *gin.Context, user *models.User, returnTo string) {
	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	// GenerateTokensForUser internally creates session via SessionService
	authResp, err := h.authService.GenerateTokensForUser(c.Request.Context(), user, ip, userAgent)
	if err != nil {
		h.logger.Error("Failed to generate tokens", map[string]interface{}{
			"error":   err.Error(),
			"user_id": user.ID,
		})
		h.redirectToLogin(c, returnTo, "Login failed, please try again")
		return
	}

	h.setSessionCookie(c, authResp.AccessToken)

	if returnTo != "" {
		c.Redirect(http.StatusTemporaryRedirect, returnTo)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// setSessionCookie sets the session cookie with the access token
func (h *LoginHandler) setSessionCookie(c *gin.Context, accessToken string) {
	c.SetCookie(
		sessionCookieName,
		accessToken,
		sessionMaxAge,
		"/",
		"",
		h.secureCookie,
		true,
	)
}

// isAuthenticated checks if the user has a valid session
func (h *LoginHandler) isAuthenticated(c *gin.Context) bool {
	sessionToken, err := c.Cookie(sessionCookieName)
	if err != nil || sessionToken == "" {
		return false
	}

	_, err = h.jwtService.ValidateAccessToken(sessionToken)
	return err == nil
}

// renderLoginError renders the login page with an error
func (h *LoginHandler) renderLoginError(c *gin.Context, returnTo, errorMsg, identifier, activeTab string) {
	html := h.renderLoginPage(returnTo, errorMsg, identifier, activeTab)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// renderOTPError renders the OTP page with an error
func (h *LoginHandler) renderOTPError(c *gin.Context, identifier, returnTo, errorMsg string) {
	html := h.renderOTPVerifyPage(identifier, returnTo, errorMsg)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// redirectToLogin redirects to login page with error
func (h *LoginHandler) redirectToLogin(c *gin.Context, returnTo, errorMsg string) {
	loginURL := "/login"
	if returnTo != "" || errorMsg != "" {
		loginURL += "?"
		params := url.Values{}
		if returnTo != "" {
			params.Set("return_to", returnTo)
		}
		if errorMsg != "" {
			params.Set("error", errorMsg)
		}
		loginURL += params.Encode()
	}
	c.Redirect(http.StatusTemporaryRedirect, loginURL)
}

// renderLoginPage generates the HTML for the login page
func (h *LoginHandler) renderLoginPage(returnTo, errorMsg, identifier, activeTab string) string {
	errorHTML := ""
	if errorMsg != "" {
		errorHTML = fmt.Sprintf(`
            <div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
                %s
            </div>`, escapeHTML(errorMsg))
	}

	passwordTabClass := "bg-blue-600 text-white"
	otpTabClass := "bg-gray-200 text-gray-700 hover:bg-gray-300"
	passwordFormDisplay := "block"
	otpFormDisplay := "none"

	if activeTab == "otp" {
		passwordTabClass = "bg-gray-200 text-gray-700 hover:bg-gray-300"
		otpTabClass = "bg-blue-600 text-white"
		passwordFormDisplay = "none"
		otpFormDisplay = "block"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Sign In - Auth Gateway</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        .tab-btn { transition: all 0.2s ease; }
        .input-field { transition: all 0.2s ease; }
        .input-field:focus { transform: translateY(-1px); }
        .submit-btn { transition: all 0.2s ease; }
        .submit-btn:hover:not(:disabled) { transform: translateY(-1px); }
        .submit-btn:disabled { opacity: 0.7; cursor: not-allowed; }
        .spinner { border: 2px solid transparent; border-top-color: white; border-radius: 50%%; animation: spin 0.8s linear infinite; }
        @keyframes spin { to { transform: rotate(360deg); } }
    </style>
</head>
<body class="min-h-screen bg-gray-100 flex items-center justify-center p-4">
    <div class="max-w-md w-full bg-white rounded-2xl shadow-xl overflow-hidden">
        <div class="bg-slate-900 p-8 text-center">
            <div class="mx-auto bg-blue-600 w-12 h-12 rounded-lg flex items-center justify-center mb-4">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path>
                </svg>
            </div>
            <h2 class="text-2xl font-bold text-white">Sign In</h2>
            <p class="text-slate-400 mt-2 text-sm">Enter your credentials to continue</p>
        </div>

        <div class="p-8">
            %s

            <!-- Login Type Tabs -->
            <div class="flex mb-6 bg-gray-100 rounded-lg p-1">
                <button type="button" id="password-tab" class="tab-btn flex-1 py-2 px-4 rounded-md text-sm font-medium %s" onclick="switchTab('password')">
                    Password
                </button>
                <button type="button" id="otp-tab" class="tab-btn flex-1 py-2 px-4 rounded-md text-sm font-medium %s" onclick="switchTab('otp')">
                    Email/SMS Code
                </button>
            </div>

            <!-- Password Login Form -->
            <form id="password-form" method="POST" action="/login" class="space-y-4" style="display: %s;">
                <input type="hidden" name="login_type" value="password">
                <input type="hidden" name="return_to" value="%s">

                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Email, Phone, or Username</label>
                    <input
                        type="text"
                        name="identifier"
                        required
                        class="input-field w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
                        placeholder="user@example.com"
                        value="%s"
                        autocomplete="username"
                    >
                </div>

                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Password</label>
                    <input
                        type="password"
                        name="password"
                        required
                        class="input-field w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
                        placeholder="Enter your password"
                        autocomplete="current-password"
                    >
                </div>

                <button
                    type="submit"
                    class="submit-btn w-full py-3 px-4 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg shadow-sm flex justify-center items-center"
                >
                    <span class="btn-text">Sign In</span>
                    <div class="spinner w-5 h-5 hidden"></div>
                </button>
            </form>

            <!-- OTP Login Form -->
            <form id="otp-form" method="POST" action="/login" class="space-y-4" style="display: %s;">
                <input type="hidden" name="login_type" value="otp">
                <input type="hidden" name="return_to" value="%s">

                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Email or Phone Number</label>
                    <input
                        type="text"
                        name="identifier"
                        required
                        class="input-field w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
                        placeholder="user@example.com or +1234567890"
                        value="%s"
                        autocomplete="username"
                    >
                </div>

                <p class="text-sm text-gray-500">
                    We'll send you a one-time code to verify your identity.
                </p>

                <button
                    type="submit"
                    class="submit-btn w-full py-3 px-4 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg shadow-sm flex justify-center items-center"
                >
                    <span class="btn-text">Send Code</span>
                    <div class="spinner w-5 h-5 hidden"></div>
                </button>
            </form>

            <div class="mt-6 text-center text-sm text-gray-500">
                <a href="/api/auth/password/reset/request" class="text-blue-600 hover:underline">Forgot password?</a>
            </div>
        </div>
    </div>

    <script>
        function switchTab(tab) {
            const passwordTab = document.getElementById('password-tab');
            const otpTab = document.getElementById('otp-tab');
            const passwordForm = document.getElementById('password-form');
            const otpForm = document.getElementById('otp-form');

            if (tab === 'password') {
                passwordTab.className = 'tab-btn flex-1 py-2 px-4 rounded-md text-sm font-medium bg-blue-600 text-white';
                otpTab.className = 'tab-btn flex-1 py-2 px-4 rounded-md text-sm font-medium bg-gray-200 text-gray-700 hover:bg-gray-300';
                passwordForm.style.display = 'block';
                otpForm.style.display = 'none';
            } else {
                passwordTab.className = 'tab-btn flex-1 py-2 px-4 rounded-md text-sm font-medium bg-gray-200 text-gray-700 hover:bg-gray-300';
                otpTab.className = 'tab-btn flex-1 py-2 px-4 rounded-md text-sm font-medium bg-blue-600 text-white';
                passwordForm.style.display = 'none';
                otpForm.style.display = 'block';
            }
        }

        document.querySelectorAll('form').forEach(form => {
            form.addEventListener('submit', function(e) {
                const btn = this.querySelector('button[type="submit"]');
                const btnText = btn.querySelector('.btn-text');
                const spinner = btn.querySelector('.spinner');

                btn.disabled = true;
                btnText.classList.add('hidden');
                spinner.classList.remove('hidden');
            });
        });
    </script>
</body>
</html>`,
		errorHTML,
		passwordTabClass, otpTabClass,
		passwordFormDisplay, escapeHTML(returnTo), escapeHTML(identifier),
		otpFormDisplay, escapeHTML(returnTo), escapeHTML(identifier),
	)
}

// renderOTPVerifyPage generates the HTML for the OTP verification page
func (h *LoginHandler) renderOTPVerifyPage(identifier, returnTo, errorMsg string) string {
	errorHTML := ""
	if errorMsg != "" {
		errorHTML = fmt.Sprintf(`
            <div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
                %s
            </div>`, escapeHTML(errorMsg))
	}

	maskedIdentifier := maskIdentifier(identifier)

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Verify Code - Auth Gateway</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        .otp-input {
            width: 3rem;
            height: 3.5rem;
            text-align: center;
            font-size: 1.5rem;
            font-weight: bold;
            transition: all 0.2s ease;
        }
        .otp-input:focus { transform: translateY(-2px); box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); }
        .submit-btn { transition: all 0.2s ease; }
        .submit-btn:hover:not(:disabled) { transform: translateY(-1px); }
        .submit-btn:disabled { opacity: 0.7; cursor: not-allowed; }
        .spinner { border: 2px solid transparent; border-top-color: white; border-radius: 50%%; animation: spin 0.8s linear infinite; }
        @keyframes spin { to { transform: rotate(360deg); } }
    </style>
</head>
<body class="min-h-screen bg-gray-100 flex items-center justify-center p-4">
    <div class="max-w-md w-full bg-white rounded-2xl shadow-xl overflow-hidden">
        <div class="bg-slate-900 p-8 text-center">
            <div class="mx-auto bg-green-600 w-12 h-12 rounded-lg flex items-center justify-center mb-4">
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path>
                </svg>
            </div>
            <h2 class="text-2xl font-bold text-white">Enter Verification Code</h2>
            <p class="text-slate-400 mt-2 text-sm">We sent a code to %s</p>
        </div>

        <div class="p-8">
            %s

            <form method="POST" action="/login/otp" class="space-y-6">
                <input type="hidden" name="identifier" value="%s">
                <input type="hidden" name="return_to" value="%s">
                <input type="hidden" name="code" id="code-input">

                <div class="flex justify-center gap-2">
                    <input type="text" maxlength="1" class="otp-input border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none" data-index="0" inputmode="numeric" pattern="[0-9]" autocomplete="one-time-code">
                    <input type="text" maxlength="1" class="otp-input border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none" data-index="1" inputmode="numeric" pattern="[0-9]">
                    <input type="text" maxlength="1" class="otp-input border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none" data-index="2" inputmode="numeric" pattern="[0-9]">
                    <span class="flex items-center text-gray-400 text-2xl">-</span>
                    <input type="text" maxlength="1" class="otp-input border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none" data-index="3" inputmode="numeric" pattern="[0-9]">
                    <input type="text" maxlength="1" class="otp-input border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none" data-index="4" inputmode="numeric" pattern="[0-9]">
                    <input type="text" maxlength="1" class="otp-input border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none" data-index="5" inputmode="numeric" pattern="[0-9]">
                </div>

                <button
                    type="submit"
                    id="verify-btn"
                    disabled
                    class="submit-btn w-full py-3 px-4 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg shadow-sm flex justify-center items-center"
                >
                    <span class="btn-text">Verify</span>
                    <div class="spinner w-5 h-5 hidden"></div>
                </button>
            </form>

            <div class="mt-6 text-center">
                <p class="text-sm text-gray-500 mb-2">Didn't receive a code?</p>
                <form method="POST" action="/login" class="inline">
                    <input type="hidden" name="login_type" value="otp">
                    <input type="hidden" name="identifier" value="%s">
                    <input type="hidden" name="return_to" value="%s">
                    <button type="submit" class="text-blue-600 hover:underline text-sm font-medium">
                        Resend Code
                    </button>
                </form>
            </div>

            <div class="mt-4 text-center">
                <a href="/login%s" class="text-gray-500 hover:text-gray-700 text-sm">
                    Back to login
                </a>
            </div>
        </div>
    </div>

    <script>
        const inputs = document.querySelectorAll('.otp-input');
        const codeInput = document.getElementById('code-input');
        const verifyBtn = document.getElementById('verify-btn');

        function updateCodeInput() {
            let code = '';
            inputs.forEach(input => {
                code += input.value;
            });
            codeInput.value = code;
            verifyBtn.disabled = code.length !== 6;
        }

        inputs.forEach((input, index) => {
            input.addEventListener('input', (e) => {
                const value = e.target.value.replace(/[^0-9]/g, '');
                e.target.value = value;

                if (value && index < inputs.length - 1) {
                    inputs[index + 1].focus();
                }
                updateCodeInput();
            });

            input.addEventListener('keydown', (e) => {
                if (e.key === 'Backspace' && !e.target.value && index > 0) {
                    inputs[index - 1].focus();
                }
            });

            input.addEventListener('paste', (e) => {
                e.preventDefault();
                const paste = (e.clipboardData || window.clipboardData).getData('text');
                const digits = paste.replace(/[^0-9]/g, '').slice(0, 6);

                digits.split('').forEach((digit, i) => {
                    if (inputs[i]) {
                        inputs[i].value = digit;
                    }
                });

                const nextEmptyIndex = Math.min(digits.length, inputs.length - 1);
                inputs[nextEmptyIndex].focus();
                updateCodeInput();
            });
        });

        inputs[0].focus();

        document.querySelector('form').addEventListener('submit', function(e) {
            if (codeInput.value.length !== 6) {
                e.preventDefault();
                return;
            }

            const btn = verifyBtn;
            const btnText = btn.querySelector('.btn-text');
            const spinner = btn.querySelector('.spinner');

            btn.disabled = true;
            btnText.classList.add('hidden');
            spinner.classList.remove('hidden');
        });
    </script>
</body>
</html>`,
		escapeHTML(maskedIdentifier),
		errorHTML,
		escapeHTML(identifier),
		escapeHTML(returnTo),
		escapeHTML(identifier),
		escapeHTML(returnTo),
		buildReturnToParam(returnTo),
	)
}

// Helper functions

func escapeHTML(s string) string {
	replacer := map[rune]string{
		'&':  "&amp;",
		'<':  "&lt;",
		'>':  "&gt;",
		'"':  "&quot;",
		'\'': "&#39;",
	}
	result := ""
	for _, r := range s {
		if repl, ok := replacer[r]; ok {
			result += repl
		} else {
			result += string(r)
		}
	}
	return result
}

func maskIdentifier(identifier string) string {
	if utils.IsValidEmail(identifier) {
		parts := splitEmail(identifier)
		if len(parts) == 2 {
			local := parts[0]
			domain := parts[1]
			if len(local) > 2 {
				return local[:2] + "***@" + domain
			}
			return local[:1] + "***@" + domain
		}
	}

	if utils.IsValidPhone(identifier) {
		if len(identifier) > 4 {
			return "***" + identifier[len(identifier)-4:]
		}
	}

	return "***"
}

func splitEmail(email string) []string {
	for i, r := range email {
		if r == '@' {
			return []string{email[:i], email[i+1:]}
		}
	}
	return []string{email}
}

func buildReturnToParam(returnTo string) string {
	if returnTo == "" {
		return ""
	}
	return "?return_to=" + url.QueryEscape(returnTo)
}
