package http_api

import (
	"net/http"
)

func SetAllowOriginHeader() Handler {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Access-Control-Allow-Origin", r.URL.Hostname())
	}
}
