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
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
)

// lambda authorizer simple moc
func authorizer(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow paddle webhook, without auth tokens
		if r.URL.Path == "/users/subscription/webhook" {
			next.ServeHTTP(w, r)
			return
		}

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

		next.ServeHTTP(w, r)
	})
}

func main() {

	// load config
	config.Init()

	mux := http.NewServeMux()

	ddb := db.New()
	emailQueue := events.NewEmailQueue()
	// client := &http.Client{}

	client := http.DefaultClient

	mux.Handle("/auth/", auth.Router())
	mux.Handle("/users/", authorizer(users.Router(ddb, emailQueue, client)))
	mux.Handle("/spaces/", authorizer(spaces.Router()))
	mux.Handle("/notes/", authorizer(notes.Router()))
	mux.Handle("/notifications/", authorizer(notifications.Router()))

	// handle unknown service routes
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unknown Service", http.StatusNotFound)
	}))

	fmt.Println("Running auth service on port 8080")

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
