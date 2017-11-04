// Package lyft provides a client for Lyft's v1 HTTP API.
// Lyft's API reference is available at https://developer.lyft.com/v1/docs/overview.
//
// It does not yet support webhooks, rich error details, extracting the Request-ID header,
// rate limiting, and the sandbox routes.
package lyft

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

// BaseURL is the base URL for Lyft's v1 HTTP API.
const BaseURL = "https://api.lyft.com/v1"

// Client is a client for the Lyft API.
// AccessToken must be set for a client to be ready to use. The rest of the
// fields are optional. Methods are goroutine safe, unless the
// client's fields are being modified at the same time.
type Client struct {
	AccessToken string

	// Optional.
	HTTPClient *http.Client // Uses http.DefaultClient if nil.
	Header     http.Header  // Extra header to add.
	BaseURL    string       // The base URL of the API; uses the package-level BaseURL if empty. Useful in testing.

	debug bool // Dump requests/responses using package log's default logger.
}

func (c *Client) base() string {
	if c.BaseURL == "" {
		return BaseURL
	}
	return c.BaseURL
}

func (c *Client) do(r *http.Request) (*http.Response, error) {
	// Set up headers and add credentials.
	c.addHeader(r.Header)
	c.authorize(r.Header)

	// Determine the HTTP client to use.
	client := http.DefaultClient
	if c.HTTPClient != nil {
		client = c.HTTPClient
	}

	if c.debug {
		dump, err := httputil.DumpRequestOut(r, true)
		if err != nil {
			log.Printf("dumping request: %s", err)
		} else {
			log.Printf("%s", dump)
		}
	}

	// Do the request.
	rsp, err := client.Do(r)

	if c.debug {
		dump, err := httputil.DumpResponse(rsp, true)
		if err != nil {
			log.Printf("dumping response: %s", err)
		} else {
			log.Printf("%s", dump)
		}
	}

	return rsp, err
}

// addHeader adds the key/values in c.Header to h.
func (c *Client) addHeader(h http.Header) {
	for key, values := range c.Header {
		for _, v := range values {
			h.Add(key, v)
		}
	}
}

// authorize modifies the header to include the access token
// in the Authorization field, as expected by the Lyft API. Useful when
// constructing a request manually.
func (c *Client) authorize(h http.Header) {
	h.Add("Authorization", "Bearer "+c.AccessToken)
}

// StatusError is returned when the HTTP roundtrip succeeded, but there
// was error was indicated via the HTTP status code, typically due to an
// application level error.
type StatusError struct {
	StatusCode   int
	Reason       string
	ResponseBody bytes.Buffer
}

// NewStatusError constructs a StatusError from the response. It exists
// solely so that subpackages (such as package auth) can create a
// StatusError using the canonical way. Not meant for external use.
// Does not close rsp.Body.
func NewStatusError(rsp *http.Response) *StatusError {
	var buf bytes.Buffer
	buf.ReadFrom(rsp.Body) // ignore errors
	return &StatusError{
		StatusCode:   rsp.StatusCode,
		Reason:       rsp.Header.Get("error"),
		ResponseBody: buf,
	}
}

func (s *StatusError) Error() string {
	if s.Reason != "" {
		return fmt.Sprintf("%s: status code: %d", s.Reason, s.StatusCode)
	}
	return fmt.Sprintf("status code: %d", s.StatusCode)
}

func index(v []string, target string) int {
	for i, s := range v {
		if s == target {
			return i
		}
	}
	return -1
}
