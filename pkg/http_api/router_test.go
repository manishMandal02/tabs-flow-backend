package http_api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Router() http.Handler {
	r := http_api.NewRouter("/test")

	r.Use(func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("userId", "123")
	})

	r.GET("/hello", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http_api.ErrorRes(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userId := r.PathValue("userId")

		if userId == "" {
			http_api.ErrorRes(w, "User ID not found", http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "Hello! from > GET /test/hello")
	})

	r.POST("/hello", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http_api.ErrorRes(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fmt.Fprintf(w, "Hello! from > POST /test/hello")
	})

	return r
}

var testCases = []struct {
	name           string
	path           string
	method         string
	expectedStatus int
	expectedBody   string
}{
	{
		name:           "GET /test/hello/extra > not found",
		path:           "/test/hello/extra",
		method:         "GET",
		expectedStatus: http.StatusNotFound,
	},
	{
		name:           "Delete /test/hello/extra > not found",
		path:           "/test/hello",
		method:         "DELETE",
		expectedStatus: http.StatusNotFound,
	},
	{
		name:           "GET /test/hello > success",
		path:           "/test/hello",
		method:         "GET",
		expectedStatus: http.StatusOK,
		expectedBody:   "Hello! from > GET /test/hello",
	},
	{
		name:           "POST /test/hello > success",
		path:           "/test/hello",
		method:         "POST",
		expectedStatus: http.StatusOK,
		expectedBody:   "Hello! from > POST /test/hello",
	},
}

func TestRouter(t *testing.T) {

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := Router()

			// create test request
			req := httptest.NewRequest(tc.method, tc.path, nil)

			// recorder
			w := httptest.NewRecorder()
			// serve request
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				// assertions
				t.Errorf("Status [Want] %d | [Actual] %d", tc.expectedStatus, w.Code)
			}

			if tc.expectedBody != "" && w.Body.String() != tc.expectedBody {
				t.Errorf("Body [Want] %s | [Actual] %s", tc.expectedBody, w.Body.String())
			}
		})
	}
}
