package sql

import (
	"errors"
	"github.com/brewbits-co/releasedesk/internal/domains/user"
	"xorm.io/xorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// NewUserRepository is the constructor for userRepository
func NewUserRepository(engine *xorm.Engine) user.UserRepository {
	return &userRepository{engine: engine}
}

// userRepository is the implementation of user.UserRepository
type userRepository struct {
	engine *xorm.Engine
}

func (r *userRepository) GetByID(id int) (user.User, error) {
	var userEntity user.User
	found, err := r.engine.ID(id).Get(&userEntity)
	if err != nil || !found {
		return user.User{}, ErrUserNotFound
	}
	userEntity.FormatAuditable()

	return userEntity, nil
}

func (r *userRepository) GetByUsername(username string) (user.User, error) {
	var userEntity user.User
	found, err := r.engine.Where("username = ?", username).Get(&userEntity)
	if err != nil || !found {
		return user.User{}, ErrUserNotFound
	}
	userEntity.FormatAuditable()

	return userEntity, nil
}

func (r *userRepository) Save(userEntity *user.User) error {
	_, err := r.engine.Insert(userEntity)
	return err
}

func (r *userRepository) UpdatePassword(id int, hash string, mustChange bool) error {
	table := r.engine.Quote("user")
	_, err := r.engine.Exec("UPDATE "+table+" SET password = ?, must_change_password = ? WHERE id = ?", hash, mustChange, id)
	return err
}

func (r *userRepository) UpdateProfile(id int, firstName, lastName, email string) error {
	table := r.engine.Quote("user")
	_, err := r.engine.Exec("UPDATE "+table+" SET first_name = ?, last_name = ?, email = ? WHERE id = ?", firstName, lastName, email, id)
	return err
}

func (r *userRepository) Find() ([]user.User, error) {
	var users []user.User
	err := r.engine.Find(&users)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].FormatAuditable()
	}
	return users, nil
}
