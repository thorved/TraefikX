package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/traefikx/backend/internal/auth"
	"github.com/traefikx/backend/internal/database"
	"github.com/traefikx/backend/internal/models"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Login handles password-based authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is disabled"})
		return
	}

	// Check if password login is enabled
	if !user.CanLoginWithPassword() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password login is disabled for this account"})
		return
	}

	// Verify password
	if !database.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	h.db.Save(&user)

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

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header required"})
		return
	}

	// Extract token
	parts := make([]string, 0)
	for _, p := range authHeader {
		if p == ' ' {
			break
		}
		parts = append(parts, string(p))
	}

	// In a real implementation, we would delete the session
	// For now, we just return success
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find session by refresh token
	var session models.Session
	if err := h.db.Where("token = ?", req.RefreshToken).First(&session).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Check if session is expired
	if session.IsExpired() {
		h.db.Delete(&session)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has expired"})
		return
	}

	// Get user
	var user models.User
	if err := h.db.First(&user, session.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is disabled"})
		return
	}

	// Generate new token pair
	tokenPair, err := auth.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Update session with new refresh token
	session.Token = tokenPair.RefreshToken
	session.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	h.db.Save(&session)

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	})
}

// GetMe returns current user info
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// ChangePassword allows user to change their password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// If user has a password, verify current password
	if user.Password != "" {
		if req.CurrentPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is required"})
			return
		}

		if !database.CheckPassword(req.CurrentPassword, user.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}
	}

	// Hash new password
	hashedPassword, err := database.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update user
	user.Password = hashedPassword
	user.PasswordEnabled = true
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// TogglePasswordLogin enables/disables password login for OIDC users
func (h *AuthHandler) TogglePasswordLogin(c *gin.Context) {
	userID, _ := c.Get("userID")

	type ToggleRequest struct {
		Enabled bool `json:"enabled"`
	}

	var req ToggleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// If disabling password login, user must have OIDC enabled
	if !req.Enabled && !user.CanLoginWithOIDC() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot disable password login without OIDC enabled"})
		return
	}

	user.PasswordEnabled = req.Enabled
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password login updated", "enabled": req.Enabled})
}

// RemovePassword removes password from account (OIDC-only login)
func (h *AuthHandler) RemovePassword(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user has OIDC enabled
	if !user.CanLoginWithOIDC() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove password without OIDC enabled"})
		return
	}

	user.Password = ""
	user.PasswordEnabled = false
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password removed successfully"})
}
