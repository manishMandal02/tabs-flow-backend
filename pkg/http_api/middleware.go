package http_api

import (
	"net/http"
	"slices"

	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

func SetAllowOriginHeader() Handler {
	return func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")

		logger.Dev("origin: %v", origin)

		if origin == "" {
			ErrorRes(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		if !slices.Contains(config.AllowedOrigins, origin) {
			ErrorRes(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		w.Header().Add("Access-Control-Allow-Origin", origin)
	}
}
