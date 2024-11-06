package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/suite"
)

// ResponseAssertions handles common HTTP response validations
type ResponseAssertions struct {
	suite.Suite
}

// AssertResponse is a generic response checker that returns decoded response
func (a *ResponseAssertions) AssertResponse(resp *http.Response, expectedStatus int, target interface{}) error {
	a.T().Helper()

	// Check status code
	a.Equal(expectedStatus, resp.StatusCode, "HTTP status code mismatch")

	// If no target to decode into, just return
	if target == nil {
		return nil
	}

	// Decode response
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	return nil
}

// AssertCookie checks for cookie presence and attributes
func (a *ResponseAssertions) AssertCookie(resp *http.Response, name string, options ...CookieOption) *http.Cookie {
	a.T().Helper()

	for _, cookie := range resp.Cookies() {
		if cookie.Name == name {
			// Apply all cookie validation options
			for _, opt := range options {
				opt(a, cookie)
			}
			return cookie
		}
	}
	a.Fail("Cookie not found: " + name)
	return nil
}

// CookieOption defines cookie validation options
type CookieOption func(*ResponseAssertions, *http.Cookie)

// Common cookie validation options
var (
	WithSecure = func(a *ResponseAssertions, c *http.Cookie) {
		a.True(c.Secure, "Cookie should be secure")
	}

	WithHTTPOnly = func(a *ResponseAssertions, c *http.Cookie) {
		a.True(c.HttpOnly, "Cookie should be HTTP only")
	}

	WithPath = func(path string) CookieOption {
		return func(a *ResponseAssertions, c *http.Cookie) {
			a.Equal(path, c.Path, "Cookie path mismatch")
		}
	}

	WithDomain = func(domain string) CookieOption {
		return func(a *ResponseAssertions, c *http.Cookie) {
			a.Equal(domain, c.Domain, "Cookie domain mismatch")
		}
	}
)

// HeaderAssertions checks response headers
func (a *ResponseAssertions) AssertHeader(resp *http.Response, key, value string) {
	a.T().Helper()

	actual := resp.Header.Get(key)
	a.Equal(value, actual, key)
}

// AssertContentType is a common header assertion
func (a *ResponseAssertions) AssertContentType(resp *http.Response, expected string) {
	a.T().Helper()

	a.AssertHeader(resp, "Content-Type", expected)
}

// MessageAssertions for queue/pub-sub message validation
type MessageAssertions struct {
	suite.Suite
}

// AssertMessageContent is a generic message content validator
func (m *MessageAssertions) AssertMessageContent(message []byte, target interface{}) error {
	if err := json.Unmarshal(message, target); err != nil {
		return fmt.Errorf("failed to decode message: %v", err)
	}
	return nil
}
