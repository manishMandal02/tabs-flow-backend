package user

import "github.com/manishMandal02/tabsflow-backend/pkg/database"

type userRepository interface {
	getUserByID(id string) (*User, error)
	getUserByEmail(email string) (*User, error)
	createUser(user *User) error
	updateUser(user *User) error
	deleteUser(id string) error
}

type userRepo struct {
	db database.DDB
}

func newUserRepository(db database.DDB) userRepository {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) getUserByID(id string) (*User, error) {
	// TODO - implement
	return nil, nil
}
func (r *userRepo) getUserByEmail(email string) (*User, error) {
	// TODO - implement
	return nil, nil
}

func (r *userRepo) createUser(user *User) error {
	// TODO - implement
	return nil
}

func (r *userRepo) updateUser(user *User) error {
	// TODO - implement
	return nil
}

func (r *userRepo) deleteUser(id string) error {
	// TODO - implement
	return nil
}
