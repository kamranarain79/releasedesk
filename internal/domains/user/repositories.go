package user

type UserRepository interface {
	GetByID(id int) (User, error)
	GetByUsername(username string) (User, error)
	Find() ([]User, error)
	Save(user *User) error
	UpdatePassword(id int, hash string, mustChange bool) error
	UpdateProfile(id int, firstName, lastName, email string) error
}
