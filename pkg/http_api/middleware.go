package http_api

import (
	"net/http"
	"slices"
	"strings"

	"github.com/manishMandal02/tabsflow-backend/config"
)

func SetAllowOriginHeader() Handler {
	return func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")
		referrer := r.Header.Get("Referrer")

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
