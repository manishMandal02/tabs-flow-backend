package test_utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// NewHTTPClient creates a configured HTTP client with cookie support
func NewHTTPClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Jar:     jar,
		Timeout: time.Second * 30,
	}, nil
}

// DoRequest is a helper for making HTTP requests
func (s *E2ETestSuite) DoRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader *bytes.Buffer

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, s.Config.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return s.HTTPClient.Do(req)
}
