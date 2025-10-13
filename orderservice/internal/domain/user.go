package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
// This is a pure domain model without infrastructure concerns
type User struct {
	ID        string
	Name      string
	Email     string
	AvatarURL string
	CreatedAt time.Time
}

// NewUser creates a new user with a generated ID
func NewUser(name, email string) *User {
	return &User{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     email,
		AvatarURL: "",
		CreatedAt: time.Now(),
	}
}

// UpdateAvatar updates the user's avatar URL
func (u *User) UpdateAvatar(avatarURL string) {
	u.AvatarURL = avatarURL
}

// Validate performs basic validation on user fields
func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("name is required")
	}

	if strings.TrimSpace(u.Email) == "" {
		return errors.New("email is required")
	}

	// Basic email validation
	if !strings.Contains(u.Email, "@") || !strings.Contains(u.Email, ".") {
		return errors.New("email format is invalid")
	}

	return nil
}
