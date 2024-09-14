package user

import (
	"encoding/json"
	"strings"

	"github.com/go-playground/validator/v10"
)

type User struct {
	Id         string `json:"id" dynamodbav:"PK" validate:"required"`
	FullName   string `json:"fullName" dynamodbav:"FullName" validate:"required"`
	Email      string `json:"email" dynamodbav:"Email" validate:"required,email"`
	ProfilePic string `json:"profilePic" dynamodbav:"ProfilePic" validate:"url"`
}

type userWithSK struct {
	*User
	SK string `json:"sk" dynamodbav:"SK"`
}

// func newUser(id, fullName, email, profilePic string) *User {
// 	return &User{
// 		ID:         id,
// 		FullName:   fullName,
// 		Email:      email,
// 		ProfilePic: profilePic,
// 	}
// }

func (u *User) fromJSON(body string) error {

	decoder := json.NewDecoder(strings.NewReader(body))

	err := decoder.Decode(&u)

	if err != nil {
		return err
	}
	return nil
}

func (u *User) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(u)

	if err != nil {
		return err
	}

	return nil

}

var errMsg = struct {
	GetUser      string
	UserNotFound string
	CreateUser   string
	UpdateUser   string
	DeleteUser   string
}{
	GetUser:      "Error getting user",
	UserNotFound: "User not found",
	CreateUser:   "Error creating user",
	UpdateUser:   "Error updating user",
	DeleteUser:   "Error deleting user",
}
