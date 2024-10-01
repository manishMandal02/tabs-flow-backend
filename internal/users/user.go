package users

import (
	"encoding/json"
	"io"

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

func userFromJSON(body io.Reader) (*User, error) {

	var u *User

	err := json.NewDecoder(body).Decode(&u)

	if err != nil {
		return nil, err
	}

	return u, nil
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
	getUser       string
	userNotFound  string
	userExists    string
	createUser    string
	updateUser    string
	deleteUser    string
	invalidUserId string
}{
	getUser:       "Error getting user",
	userNotFound:  "User not found",
	userExists:    "User already exits",
	createUser:    "Error creating user",
	updateUser:    "Error updating user",
	deleteUser:    "Error deleting user",
	invalidUserId: "Invalid user id",
}
