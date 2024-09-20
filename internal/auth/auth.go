package auth

type emailOTP struct {
	Email string `json:"email" dynamodbav:"PK"`
	OTP   string `json:"otp"`
	TTL   int64  `json:"ttl"`
}

type emailWithUserId struct {
	Email  string `json:"email" dynamodbav:"PK"`
	UserId string `json:"userId" dynamodbav:"SK"`
}

type deviceInfo struct {
	Browser  string `json:"browser" dynamodbav:"browser"`
	OS       string `json:"os" dynamodbav:"os"`
	Platform string `json:"platform" dynamodbav:"platform"`
	IsMobile bool   `json:"isMobile" dynamodbav:"isMobile"`
}

type session struct {
	Email      string      `json:"email" dynamodbav:"PK"`
	Id         string      `json:"id" dynamodbav:"SK"`
	TTL        int64       `json:"ttl" dynamodbav:"TTL"`
	DeviceInfo *deviceInfo `json:"deviceInfo" dynamodbav:"DeviceInfo"`
}

var errMsg = struct {
	sendOTP         string
	validateOTP     string
	inValidOTP      string
	createToken     string
	createSession   string
	deleteSession   string
	validateSession string
	googleAuth      string
	tokenExpired    string
	invalidToken    string
	invalidSession  string
	getUserId       string
	logout          string
}{
	sendOTP:         "Error sending OTP",
	validateOTP:     "Error validating OTP",
	inValidOTP:      "Invalid OTP",
	googleAuth:      "Error authenticating with google",
	createSession:   "Error creating session",
	deleteSession:   "Error deleting session",
	createToken:     "Error creating token",
	validateSession: "Error validating session",
	tokenExpired:    "Token expired",
	invalidToken:    "Invalid token",
	invalidSession:  "Invalid session",
	getUserId:       "Error getting user id",
	logout:          "Error logging out",
}
