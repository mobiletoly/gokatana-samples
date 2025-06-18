package model

//go:generate go tool gobetter -input $GOFILE

// AuthUser represents a user in the authentication system
type AuthUser struct { //+gob:Constructor
	ID            string
	Email         string
	PasswordHash  string
	FirstName     string
	LastName      string
	IsActive      bool
	EmailVerified bool
	CreatedAt     string
	UpdatedAt     string
}
