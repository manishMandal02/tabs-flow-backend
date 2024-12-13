package main

import (
	"fmt"
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/auth"
	"github.com/manishMandal02/tabsflow-backend/internal/notes"
	"github.com/manishMandal02/tabsflow-backend/internal/notifications"
	"github.com/manishMandal02/tabsflow-backend/internal/spaces"
	"github.com/manishMandal02/tabsflow-backend/internal/users"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

// lambda authorizer simple moc
func authorizer(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow paddle webhook, without auth tokens
		if r.URL.Path == "/users/subscription/webhook" {
			next.ServeHTTP(w, r)
			return
		}

		c, err := r.Cookie("session")

		if err != nil {
			http_api.ErrorRes(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		sessionId, userId, err := auth.GetSessionValues(c.Value)

		if err != nil || sessionId == "" || userId == "" {
			http_api.ErrorRes(w, "Unauthorized", http.StatusUnauthorized)

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
	searchIndexTable := db.NewSearchIndexTable()

	emailQueue := events.NewEmailQueue()
	notificationQueue := events.NewNotificationQueue()
	// client := &http.Client{}

	httpClient := http.DefaultClient

	paddle, err := users.NewPaddleSubscriptionClient()

	if err != nil {
		panic(err)
	}

	mux.Handle("/auth/", auth.Router(ddb, emailQueue))
	mux.Handle("/users/", authorizer(users.Router(ddb, emailQueue, httpClient, paddle)))
	mux.Handle("/spaces/", authorizer(spaces.Router(ddb, notificationQueue)))
	mux.Handle("/notes/", authorizer(notes.Router(ddb, searchIndexTable, notificationQueue)))
	mux.Handle("/notifications/", authorizer(notifications.Router(ddb)))

	// handle unknown service routes
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http_api.ErrorRes(w, "Unknown Service", http.StatusNotFound)
	}))

	fmt.Println("Running auth service on port 8080")

	err = http.ListenAndServe(":8080", mux)

	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
