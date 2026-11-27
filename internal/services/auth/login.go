package auth

import (
	"errors"
	"strings"

	"github.com/brewbits-co/releasedesk/internal/domains/user"
	"github.com/brewbits-co/releasedesk/internal/values"
	"github.com/brewbits-co/releasedesk/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

func (s *service) Login(username string, plainTextPassword string) (user.User, error) {
	userEntity, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return user.User{}, user.ErrWrongCredentials
	}

	correctPassword := userEntity.Password.CompareWith(plainTextPassword)
	if !correctPassword {
		return user.User{}, user.ErrWrongCredentials
	}

	return userEntity, nil
}

func (s *service) GetUser(id int) (user.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *service) ListUsers() ([]user.User, error) {
	return s.userRepo.Find()
}

func (s *service) CreateUser(username, firstName, lastName, email, password string, role values.Role) (user.User, error) {
	account, err := user.NewUser(username, firstName, lastName, email, password)
	if err != nil {
		return user.User{}, err
	}
	if role != 0 {
		account.Role = role
	}
	if err = s.userRepo.Save(&account); err != nil {
		return user.User{}, err
	}
	return account, nil
}

func (s *service) ChangePassword(userID int, current, next string) error {
	account, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if !account.MustChangePassword && !account.Password.CompareWith(current) {
		return user.ErrWrongCredentials
	}
	if !validator.IsPasswordStrong(next) {
		return errors.New("password must be at least 8 characters and include upper, lower, and a digit")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(next), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(userID, string(hash), false)
}

func (s *service) UpdateProfile(userID int, firstName, lastName, email string) error {
	if !validator.IsEmail(email) {
		return user.ErrInvalidEmail
	}
	if strings.TrimSpace(firstName) == "" || strings.TrimSpace(lastName) == "" {
		return errors.New("first and last name are required")
	}
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	return s.userRepo.UpdateProfile(userID, strings.TrimSpace(firstName), strings.TrimSpace(lastName), strings.TrimSpace(email))
}
