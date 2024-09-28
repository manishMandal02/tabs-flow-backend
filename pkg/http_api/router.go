package http_api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type Handler func(w http.ResponseWriter, r *http.Request)

type Route struct {
	Method  string
	Path    string
	Handler Handler
}

func (r *Route) Match(method, path string) bool {
	return r.Method == method && strings.HasPrefix(path, r.Path)
}

type Router struct {
	base   string
	routes []*Route
}

func NewRouter(base string) *Router {
	return &Router{
		base:   base,
		routes: []*Route{},
	}
}

func (r *Router) AddRoute(method, path string, handler Handler) {
	r.routes = append(r.routes, &Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger.Dev(fmt.Sprintf("Router req: method: %v, path: %v", req.Method, req.URL.Path))
	for _, route := range r.routes {
		if route.Match(req.Method, strings.TrimPrefix(req.URL.Path, r.base)) {
			route.Handler(w, req)
			return
		}
	}

	http.Error(w, ErrorRouteNotFound, http.StatusNotFound)

}
