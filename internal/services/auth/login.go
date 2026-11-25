package auth

import (
	"github.com/brewbits-co/releasedesk/internal/domains/user"
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
