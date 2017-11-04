package lyft

import (
	"bytes"
	"fmt"
	"net/http"
)

const BaseURL = "https://api.lyft.com/v1"

// Client is a client for the Lyft API. AccessToken must be specified.
// Methods that make requests to the Lyft API are safe to use concurrently,
// except when the client's fields are being modified at the same time.
type Client struct {
	AccessToken string

	// Optional.
	HTTPClient *http.Client // Uses http.DefaultClient if nil.
	Header     http.Header  // Extra header to add.
	BaseURL    string       // The base URL of the API; uses the package-level BaseURL if empty. Useful in testing.
}

func (c *Client) do(r *http.Request) (*http.Response, error) {
	c.addHeader(r.Header)
	c.authorize(r.Header)

	client := http.DefaultClient
	if c.HTTPClient != nil {
		client = c.HTTPClient
	}
	return client.Do(r)
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
// was error was indicated via the HTTP status code.
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
