package users

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserHandler_UserById(t *testing.T) {

	tests := []struct {
		name           string
		userID         string
		mockUser       *User
		mockErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userID: "user123",
			mockUser: &User{
				Id:    "user123",
				Email: "test@example.com",
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty ID",
			userID:         "",
			mockUser:       nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errMsg.invalidUserId,
		},
		{
			name:           "User Not Found",
			userID:         "nonexistent",
			mockUser:       nil,
			mockErr:        errors.New(errMsg.userNotFound),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errMsg.userNotFound,
		},
		{
			name:           "Internal Server Error",
			userID:         "user123",
			mockUser:       nil,
			mockErr:        errors.New("internal error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   errMsg.getUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository
			mockRepo := new(mockUserRepository)
			mockRepo.On("getUserByID", tt.userID).Return(tt.mockUser, tt.mockErr)

			// Initialize handler
			handler := newUserHandler(mockRepo)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			rec := httptest.NewRecorder()

			// Call handler
			handler.userById(rec, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Assert response body if specified
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}

			// Verify mock calls
			mockRepo.AssertExpectations(t)
		})
	}
}
