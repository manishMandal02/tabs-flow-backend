package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/auth"
	"github.com/manishMandal02/tabsflow-backend/internal/notes"
	"github.com/manishMandal02/tabsflow-backend/internal/notifications"
	"github.com/manishMandal02/tabsflow-backend/internal/spaces"
	"github.com/manishMandal02/tabsflow-backend/internal/users"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

// lambda authorizer simple moc
func authorizer(next http_api.Handler) http_api.Handler {
	return func(w http.ResponseWriter, r *http.Request) {

		token, err := r.Cookie("access_token")

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(token.Value)

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			return
		}

		_, emailOK := claims["sub"]
		userId, userIdOK := claims["user_id"].(string)
		_, sIdOK := claims["session_id"]
		expiryTime, expiryOK := claims["exp"].(float64)

		if !emailOK || !sIdOK || !expiryOK || !userIdOK {
			logger.Dev("Error getting token claims")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if int64(expiryTime) < time.Now().Unix() {
			// token expired, redirect to login
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// token valid, allow
		r.Header.Set("UserId", userId)

		next(w, r)
	}
}

func main() {

	// load config
	config.Init()

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/", auth.Router)
	mux.HandleFunc("/users/", authorizer(users.Router))
	mux.HandleFunc("/spaces/", authorizer(spaces.Router))
	mux.HandleFunc("/notes/", authorizer(notes.Router))
	mux.HandleFunc("/notifications/", authorizer(notifications.Router))

	// handle unknown service routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unknown Service", http.StatusNotFound)
	})

	fmt.Println("Running auth service on port 8080")

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
