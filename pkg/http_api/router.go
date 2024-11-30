package http_api

import (
	"net/http"
	"strings"

	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type Handler func(w http.ResponseWriter, r *http.Request)

type Route struct {
	Method       string
	PathSegments []string
	Handlers     []Handler
}

type Router struct {
	base       string
	routes     []*Route
	middleware []Handler
}

type IRouter interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	Use(handlers Handler)
	GET(path string, handlers ...Handler)
	POST(path string, handlers ...Handler)
	PATCH(path string, handlers ...Handler)
	DELETE(path string, handlers ...Handler)
}

func (r *Route) Match(method, path string) (bool, map[string]string) {
	if r.Method != method {
		return false, nil
	}

	// split path
	segments := strings.Split(strings.Trim(path, "/"), "/")

	if len(segments) != len(r.PathSegments) {
		return false, nil
	}

	params := make(map[string]string)

	for i, s := range r.PathSegments {
		if strings.HasPrefix(s, ":") {
			params[s[1:]] = segments[i]
		} else if s != segments[i] {
			return false, nil
		}
	}

	return true, params
}

func NewRouter(base string) IRouter {
	return &Router{
		base:       base,
		routes:     []*Route{},
		middleware: []Handler{},
	}
}

func (r *Router) AddRoute(method, path string, handlers []Handler) {

	segments := strings.Split(strings.Trim(path, "/"), "/")
	r.routes = append(r.routes, &Route{
		Method:       method,
		PathSegments: segments,
		Handlers:     handlers,
	})
}

func (r *Router) Use(middleware Handler) {
	r.middleware = append(r.middleware, middleware)
}

func (r *Router) GET(path string, handlers ...Handler) {
	r.AddRoute(http.MethodGet, path, handlers)
}

func (r *Router) POST(path string, handlers ...Handler) {
	r.AddRoute(http.MethodPost, path, handlers)

}

func (r *Router) PATCH(path string, handlers ...Handler) {
	r.AddRoute(http.MethodPatch, path, handlers)
}

func (r *Router) DELETE(path string, handlers ...Handler) {
	r.AddRoute(http.MethodDelete, path, handlers)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger.Info("Router req: [Method]: %v \n[Path]: %v", req.Method, req.URL.Path)
	for _, route := range r.routes {

		match, params := route.Match(req.Method, strings.TrimPrefix(req.URL.Path, r.base))
		if match {

			// wrap the original ResponseWriter
			rw := &responseWriterWritten{ResponseWriter: w}

			logger.Info("params: %v", params)

			// set path values
			for key, value := range params {
				req.SetPathValue(key, value)
			}
			// run middleware
			if len(r.middleware) > 0 {
				for _, m := range r.middleware {
					m(rw, req)
					if rw.HasWritten() {
						// stop the req, if header was written by a middleware
						return
					}
				}
			}

			// run handlers
			if len(route.Handlers) == 1 {
				route.Handlers[0](w, req)
			} else {
				for _, h := range route.Handlers {
					h(rw, req)
				}
			}
			return
		}

	}

	ErrorRes(w, ErrorRouteNotFound, http.StatusNotFound)
}
