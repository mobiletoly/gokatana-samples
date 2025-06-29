package model

import "time"

//go:generate go tool gobetter -input $GOFILE

// UserProfile represents a user profile in the domain model
type UserProfile struct { //+gob:Constructor
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	Height    *int      `json:"height"`
	Weight    *int      `json:"weight"`
	Gender    *string   `json:"gender"`
	BirthDate *string   `json:"birth_date"` // Using string for date to match swagger format
	IsMetric  bool      `json:"is_metric"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
