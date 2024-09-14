package auth

type AuthHandler struct {
	repo authRepository
}

func newAuthHandler(repo authRepository) *AuthHandler {
	return &AuthHandler{
		repo: repo,
	}
}

func (h *AuthHandler) sendOTP() error {
	return nil
}

func (h *AuthHandler) VerifyOTP(opt int32) error {
	return nil
}

func (h *AuthHandler) generatedToken() error {
	return nil
}

func (h *AuthHandler) validateToken() error {
	return nil
}

// remove jwt token
func (h *AuthHandler) logout() error {
	return nil
}
