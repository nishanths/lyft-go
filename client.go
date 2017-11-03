package lyft

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// Client is a client for the Lyft API.
//
// ClientID, ClientSecret, and AccessTokens must be specified. When making
// a request to the Lyft API, a client tries the access tokens until one
// of them succeeds. Typically, most a client will only specify
// a single access token.
//
// Client methods that make requests to the Lyft API are safe to use concurrently,
// except if the client's fields are being modified at the same time.
type Client struct {
	ClientID     string
	ClientSecret string
	AccessTokens []string

	HTTPClient *http.Client // uses http.DefaultClient if nil
	Headers    http.Header  // optional extra headers
	BaseURL    string       // the base URL of the API; uses BaseURL if empty. Useful in testing.
}

// AddToken appends the access token t to the client's list of access tokens
// if it does not already exist.
func (c *Client) AddToken(t string) {
	if indexOf(c.AccessTokens, t) == -1 {
		c.AccessTokens = append(c.AccessTokens, t)
	}
}

// RemoveToken removes the access token t from the client's list of access
// tokens.
func (c *Client) RemoveToken(t string) {
	if idx := indexOf(c.AccessTokens, t); idx != -1 {
		copy(c.AccessTokens[i:], c.AccessTokens[i+1:])
		c.AccessTokens[len(c.AccessTokens)-1] = ""
		c.AccessTokens = c.AccessTokens[:len(a)-1]
	}
}

// StatusError is returned when the HTTP roundtrip succeeded, but there
// was error was indicated via the HTTP status code.
type StatusError struct {
	StatusCode int
	Reason     string
	Body       io.ReadCloser
}

// NewStatusError constructs a StatusError from the response. It exists
// solely so that subpackages (such as package auth) can create a
// StatusError using the canonical way. Not meant for external use.
func NewStatusError(rsp *http.Response) *StatusError {
	return &StatusError{
		StatusCode: rsp.StatusCode,
		Reason:     rsp.Header.Get("error"),
		Body:       rsp.Body,
	}
}

func (s *StatusError) Error() string {
	if s.Reason != "" {
		return fmt.Sprintf("%s: status code: %d", s.Reason, s.StatusCode)
	}
	return fmt.Sprintf("status code: %d", s.StatusCode)
}

// Bytes returns the contents of Body. Body is closed automatically
// after reading. Errors during reading or closing are not reported.
// This mainly serves as a convenience method.
func (s *StatusError) Bytes() []byte {
	defer s.Body.Close()
	b, _ := ioutil.ReadAll(s.Body)
	return b
}

func indexOf(v []string, target string) int {
	for i, s := range v {
		if s == target {
			return i
		}
	}
	return -1
}
