package user

type User struct {
	Id         string `json:"id" dynamodbav:"pk"`
	FullName   string `json:"fullName" dynamodbav:"FullName"`
	Email      string `json:"email" dynamodbav:"Email"`
	ProfilePic string `json:"profilePic" dynamodbav:"ProfilePic"`
}

// func newUser(id, fullName, email, profilePic string) *User {
// 	return &User{
// 		ID:         id,
// 		FullName:   fullName,
// 		Email:      email,
// 		ProfilePic: profilePic,
// 	}
// }

var errMsg = struct {
	GetUser      string
	UserNotFound string
	CreateUser   string
	UpdateUser   string
	DeleteUser   string
}{
	GetUser:      "error getting user",
	UserNotFound: "user not found",
	CreateUser:   "error creating user",
	UpdateUser:   "error updating user",
	DeleteUser:   "error deleting user",
}
