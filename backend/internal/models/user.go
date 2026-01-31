package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID       uint     `gorm:"primaryKey" json:"id"`
	Email    string   `gorm:"uniqueIndex;not null" json:"email"`
	Password string   `json:"-"` // Never expose password in JSON
	Role     UserRole `gorm:"default:user" json:"role"`
	IsActive bool     `gorm:"default:true" json:"is_active"`

	// Password authentication
	PasswordEnabled bool `gorm:"default:true" json:"password_enabled"`

	// OIDC fields
	OIDCProvider string     `json:"oidc_provider,omitempty"`
	OIDCSubject  string     `gorm:"index" json:"oidc_subject,omitempty"`
	OIDCLinkedAt *time.Time `json:"oidc_linked_at,omitempty"`
	OIDCEnabled  bool       `gorm:"default:false" json:"oidc_enabled"`

	// Timestamps
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	LastLoginAt *time.Time     `json:"last_login_at,omitempty"`
}

// CanLoginWithPassword checks if user can login with password
func (u *User) CanLoginWithPassword() bool {
	return u.PasswordEnabled && u.Password != ""
}

// CanLoginWithOIDC checks if user can login with OIDC
func (u *User) CanLoginWithOIDC() bool {
	return u.OIDCEnabled && u.OIDCProvider != "" && u.OIDCSubject != ""
}

// IsLinkedToOIDC checks if account is linked to OIDC
func (u *User) IsLinkedToOIDC() bool {
	return u.OIDCProvider != "" && u.OIDCSubject != ""
}

type Session struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// IsExpired checks if session is expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

type CreateUserRequest struct {
	Email       string   `json:"email" binding:"required,email"`
	Password    string   `json:"password,omitempty"`
	Role        UserRole `json:"role" binding:"required,oneof=admin user"`
	OIDCEnabled bool     `json:"oidc_enabled"`
}

type UpdateUserRequest struct {
	Email       string   `json:"email,omitempty" binding:"omitempty,email"`
	Role        UserRole `json:"role,omitempty" binding:"omitempty,oneof=admin user"`
	IsActive    *bool    `json:"is_active,omitempty"`
	OIDCEnabled *bool    `json:"oidc_enabled,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password,omitempty"`
	NewPassword     string `json:"new_password" binding:"required,min=12"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         User   `json:"user"`
}

type OIDCCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

type LinkOIDCRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state" binding:"required"`
}

type UserResponse struct {
	ID              uint       `json:"id"`
	Email           string     `json:"email"`
	Role            UserRole   `json:"role"`
	IsActive        bool       `json:"is_active"`
	PasswordEnabled bool       `json:"password_enabled"`
	OIDCProvider    string     `json:"oidc_provider,omitempty"`
	OIDCEnabled     bool       `json:"oidc_enabled"`
	IsLinkedToOIDC  bool       `json:"is_linked_to_oidc"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		Role:            u.Role,
		IsActive:        u.IsActive,
		PasswordEnabled: u.PasswordEnabled,
		OIDCProvider:    u.OIDCProvider,
		OIDCEnabled:     u.OIDCEnabled,
		IsLinkedToOIDC:  u.IsLinkedToOIDC(),
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
		LastLoginAt:     u.LastLoginAt,
	}
}
