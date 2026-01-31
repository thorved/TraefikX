package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/auth"
	"github.com/traefikx/backend/internal/config"
	"github.com/traefikx/backend/internal/models"
)

// OIDCLogin initiates OIDC login flow
func (h *AuthHandler) OIDCLogin(c *gin.Context) {
	if !auth.IsOIDCEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OIDC is not configured"})
		return
	}

	state := auth.GenerateOIDCState(0) // 0 means new user
	authURL := auth.GetOIDCAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// OIDCCallback handles the OIDC callback
func (h *AuthHandler) OIDCCallback(c *gin.Context) {
	if !auth.IsOIDCEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OIDC is not configured"})
		return
	}

	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code or state parameter"})
		return
	}

	// Validate state
	oidcState, valid := auth.ValidateOIDCState(state)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired state"})
		return
	}

	// Exchange code for token
	token, err := auth.ExchangeOIDCCode(code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Get user info from OIDC provider
	userInfo, err := auth.GetOIDCUserInfo(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}

	if userInfo.Email == "" {
		log.Printf("Error: OIDC provider returned empty email. Subject: %s", userInfo.Subject)
		c.JSON(http.StatusBadRequest, gin.H{"error": "OIDC provider did not return an email address"})
		return
	}

	// Check if this is a linking flow
	if oidcState.LinkToUser > 0 {
		h.handleOIDCLink(c, oidcState.LinkToUser, userInfo)
		return
	}

	log.Printf("OIDC Callback: Subject=%s, Email=%s", userInfo.Subject, userInfo.Email)

	// Find or create user
	var user models.User
	result := h.db.Where("oidc_subject = ? AND oidc_provider = ?", userInfo.Subject, config.AppConfig.OIDCProviderName).First(&user)

	if result.Error != nil {
		log.Printf("OIDC user not found by subject. Searching by email: %s", userInfo.Email)

		// User doesn't exist, check if there's a user with same email
		// Use lowercase comparison to ensure we match regardless of casing
		emailResult := h.db.Where("lower(email) = ?", strings.ToLower(userInfo.Email)).First(&user)
		log.Printf("Email search result: ID=%d, Email=%s, Error=%v", user.ID, user.Email, emailResult.Error)

		if user.ID == 0 {
			// User does not exist and auto-creation is disabled
			log.Printf("OIDC login failed: No user found with email %s", userInfo.Email)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Account not found",
				"details": "No account exists with this email address. Please contact an administrator.",
			})
			return
		} else {
			// Link existing user with OIDC (email matches)
			log.Printf("Linking existing user ID=%d with OIDC", user.ID)

			user.OIDCProvider = config.AppConfig.OIDCProviderName
			user.OIDCSubject = userInfo.Subject
			user.OIDCEnabled = true
			now := time.Now()
			user.OIDCLinkedAt = &now

			// Ensure email is set (in case DB record was weird)
			if user.Email == "" {
				user.Email = userInfo.Email
			}

			if err := h.db.Save(&user).Error; err != nil {
				log.Printf("Failed to link user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link user", "details": err.Error()})
				return
			}
		}
	} else {
		// Update existing OIDC user
		now := time.Now()
		user.LastLoginAt = &now
		h.db.Save(&user)
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is disabled"})
		return
	}

	// Generate tokens
	tokenPair, err := auth.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Create session
	session := models.Session{
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	h.db.Create(&session)

	// Return tokens
	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	})
}

// OIDCLinkInit initiates account linking flow
func (h *AuthHandler) OIDCLinkInit(c *gin.Context) {
	if !auth.IsOIDCEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "OIDC is not configured"})
		return
	}

	userID, _ := c.Get("userID")
	state := auth.GenerateOIDCState(userID.(uint))
	authURL := auth.GetOIDCAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// OIDCUnlink removes OIDC link from account
func (h *AuthHandler) OIDCUnlink(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user has password enabled
	if !user.CanLoginWithPassword() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot unlink OIDC without password login enabled"})
		return
	}

	// Remove OIDC link
	user.OIDCProvider = ""
	user.OIDCSubject = ""
	user.OIDCEnabled = false
	user.OIDCLinkedAt = nil

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OIDC account unlinked successfully"})
}

// GetOIDCStatus returns OIDC configuration status
func (h *AuthHandler) GetOIDCStatus(c *gin.Context) {
	cfg := config.AppConfig

	c.JSON(http.StatusOK, gin.H{
		"enabled":       auth.IsOIDCEnabled(),
		"provider_name": cfg.OIDCProviderName,
	})
}

func (h *AuthHandler) handleOIDCLink(c *gin.Context, userID uint, userInfo *auth.OIDCUserInfo) {
	// Get the user we want to link to
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if another user is already linked to this OIDC account
	var existingUser models.User
	if err := h.db.Where("oidc_subject = ? AND id != ?", userInfo.Subject, userID).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Another account is already linked to this OIDC identity"})
		return
	}

	// Check if email matches (if user has email)
	if user.Email != userInfo.Email {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email mismatch",
			"details": gin.H{
				"account_email": user.Email,
				"oidc_email":    userInfo.Email,
			},
		})
		return
	}

	// Link OIDC to user
	user.OIDCProvider = config.AppConfig.OIDCProviderName
	user.OIDCSubject = userInfo.Subject
	user.OIDCEnabled = true
	now := time.Now()
	user.OIDCLinkedAt = &now

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OIDC account linked successfully",
		"user":    user.ToResponse(),
	})
}
