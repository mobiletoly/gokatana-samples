package model

import "time"

//go:generate go tool gobetter -input $GOFILE

// AuthUser represents a user in the authentication system
type AuthUser struct { //+gob:Constructor
	ID            string
	Email         string
	PasswordHash  string
	FirstName     string
	LastName      string
	TenantID      string
	IsActive      bool
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Tenant represents a tenant in the multi-tenant system
type Tenant struct { //+gob:Constructor
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EmailConfirmationToken represents an email confirmation token
type EmailConfirmationToken struct { //+gob:Constructor
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Email     string     `json:"email"`
	TokenHash string     `json:"token_hash"` // Hashed token/code for database storage
	Source    string     `json:"source"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// IsExpired checks if the token has expired
func (t *EmailConfirmationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used
func (t *EmailConfirmationToken) IsUsed() bool {
	return t.UsedAt != nil
}

// IsValid checks if the token is valid (not expired and not used)
func (t *EmailConfirmationToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}
