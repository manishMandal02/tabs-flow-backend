package http_api

import "net/http"

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPClient wrapper to ensure *http.Client implements Client interface
type HTTPClient struct {
	*http.Client
}

// Do implements Client interface
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}
