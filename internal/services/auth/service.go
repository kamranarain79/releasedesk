package auth

import (
	"github.com/brewbits-co/releasedesk/internal/domains/user"
	"github.com/brewbits-co/releasedesk/internal/values"
)

// Service defines the interface for handling authentication-related use cases.
type Service interface {
	// Login authenticates a user based on the provided username and password.
	// It returns the authenticated user's account information or any potential error.
	Login(username string, password string) (user.User, error)
	// GetUser returns a user by ID.
	GetUser(id int) (user.User, error)
	// ListUsers returns every user in this instance.
	// Later versions can scope this to an organization and hide inactive members.
	ListUsers() ([]user.User, error)
	CreateUser(username, firstName, lastName, email, password string, role values.Role) (user.User, error)
	ChangePassword(userID int, current, next string) error
	UpdateProfile(userID int, firstName, lastName, email string) error
}

// NewAuthService initializes a new instance of the authentication Service using the provided dependencies.
func NewAuthService(userRepo user.UserRepository) Service {
	srv := &service{
		userRepo: userRepo,
	}

	return srv
}

// service implements the auth.Service.
type service struct {
	userRepo user.UserRepository
}
