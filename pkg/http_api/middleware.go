package http_api

import (
	"net/http"
	"slices"
	"strings"

	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

func SetAllowOriginHeader() Handler {
	return func(w http.ResponseWriter, r *http.Request) {

		logger.Dev("header: %v", r.Header)

		origin := r.Header.Get("Origin")
		referrer := r.Header.Get("Referrer")

		logger.Dev("origin: %v", origin)

		if origin == "" {
			if referrer == "" {
				ErrorRes(w, "Origin not allowed", http.StatusForbidden)
				return
			}
			origin = referrer
		}

		origin = strings.TrimSuffix(origin, "/")

		if !slices.Contains(config.AllowedOrigins, origin) {
			ErrorRes(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		w.Header().Add("Access-Control-Allow-Origin", origin)
	}
}
