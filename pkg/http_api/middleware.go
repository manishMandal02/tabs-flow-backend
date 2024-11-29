package http_api

import (
	"net/http"
	"slices"

	"github.com/manishMandal02/tabsflow-backend/config"
)

func SetAllowOriginHeader() Handler {
	return func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")

		if origin == "" {
			http.Error(w, "Origin not allowed", http.StatusForbidden)
		}

		if !slices.Contains(config.AllowedOrigins, origin) {
			http.Error(w, "Origin not allowed", http.StatusForbidden)
		}

		w.Header().Add("Access-Control-Allow-Origin", origin)
	}
}
