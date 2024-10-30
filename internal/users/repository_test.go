package users

import "github.com/stretchr/testify/mock"

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) getUserByID(id string) (*User, error) {
	args := m.Called(id)
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockUserRepository) insertUser(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepository) updateUser(id, name string) error {
	return nil
}

func (m *mockUserRepository) deleteAccount(id string) error {
	return nil
}

func (m *mockUserRepository) getAllPreferences(id string) (*preferences, error) {
	return nil, nil
}

func (m *mockUserRepository) setPreferences(userId, sk string, pData interface{}) error {
	return nil
}

func (m *mockUserRepository) updatePreferences(userId, sk string, pData interface{}) error {
	return nil
}

func (m *mockUserRepository) getSubscription(userId string) (*subscription, error) {
	return nil, nil
}

func (m *mockUserRepository) setSubscription(userId string, s *subscription) error {
	return nil
}

func (m *mockUserRepository) updateSubscription(userId string, sData *subscription) error {
	return nil
}
