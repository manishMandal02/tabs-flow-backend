package main

import (
	"fmt"
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/auth"
	"github.com/manishMandal02/tabsflow-backend/internal/notes"
	"github.com/manishMandal02/tabsflow-backend/internal/spaces"
	"github.com/manishMandal02/tabsflow-backend/internal/users"
)

func main() {

	// load config
	config.Init()

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/", auth.Router)
	mux.HandleFunc("/users/", users.Router)
	mux.HandleFunc("/spaces/", spaces.Router)
	mux.HandleFunc("/notes/", notes.Router)

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
