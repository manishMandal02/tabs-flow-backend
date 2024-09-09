package user

type User struct {
	ID         string `json:"id"`
	FullName   string `json:"full_name"`
	Email      string `json:"email"`
	ProfilePic string `json:"profile_pic"`
}

func newUser(id, fullName, email, profilePic string) *User {
	return &User{
		ID:         id,
		FullName:   fullName,
		Email:      email,
		ProfilePic: profilePic,
	}
}

// func (u *User) ToMap() map[string]interface{} {
// 	return map[string]interface{}{
// 		"id":          u.ID,
// 		"full_name":   u.FullName,
// 		"email":       u.Email,
// 		"profile_pic": u.ProfilePic,
// 	}
// }
