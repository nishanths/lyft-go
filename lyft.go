// Package lyft provides a client for Lyft's v1 HTTP API.
// Lyft's API reference is available at https://developer.lyft.com/v1/docs/overview.
//
// Response Header and Request-ID
//
// Methods on the client typically have a signature like:
//
//  func (c *Client) Foo (T, http.Header, error)
//
// The returned header is the the HTTP response header. It is safe to access
// when the error is nil or the error is of type *StatusError.
//
// The returned header is useful for obtaining the rate limit header and unique
// Request-ID header set by Lyft. For details, see
// https://developer.lyft.com/v1/docs/errors#section-detailed-information-on-error-codes.
//
// Errors
//
// When the HTTP roundtrip succeeds but there was an application-level error,
// the error will be of type *StatusError, which can be inspected for more
// details.
//
// Formats
//
// According to http://petstore.swagger.io/?url=https://api.lyft.com/v1/spec#/,
// currency strings returned are in ISO 4217.
//
// Usage
//
// This example shows how to obtain an access token and find the
// ride types available at a location.
//
//   // Obtain an access token using the two-legged or three-legged flows.
//   t, err := twoleg.GenerateToken(http.DefaultClient, lyft.BaseURL, os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"))
//   if err != nil {
//       log.Fatalf("error generating token: %s", err)
//   }
//
//   // Create a client.
//   c := &lyft.Client{AccessToken: t.AccessToken}
//
//   // Make requests.
//   r, header, err := c.RideTypes(37.7, -122.2)
//   if err != nil {
//       log.Fatalf("error getting ride types: %s", err)
//   }
//   log.Printf("ride types: %+v", r)
//   log.Printf("Request-ID: %s", lyft.RequestID(header))
//
// Missing Features
//
// The package does not yet support webhooks and the sandbox routes.
package lyft // import "go.avalanche.space/lyft"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
)

// BaseURL is the base URL for Lyft's v1 HTTP API.
const BaseURL = "https://api.lyft.com/v1"

// Client is a client for the Lyft API.
// AccessToken must be set for a client to be ready to use. The rest of the
// fields are optional. Methods are goroutine safe, unless the
// client's fields are being modified at the same time.
type Client struct {
	AccessToken string
	// The following fields are optional.
	HTTPClient *http.Client // Uses http.DefaultClient if nil.
	Header     http.Header  // Extra request header to add.
	BaseURL    string       // The base URL of the API; uses the package-level BaseURL if empty. Useful in tests.
	debug      bool         // Dump requests/responses using package log's default logger.
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
			log.Printf("error dumping request: %s", err)
		} else {
			log.Printf("%s", dump)
		}
	}

	// Do the request.
	rsp, err := client.Do(r)

	if c.debug {
		dump, err := httputil.DumpResponse(rsp, true)
		if err != nil {
			log.Printf("error dumping response: %s", err)
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

// Possible values for the Reason field in StatusError.
const (
	InvalidToken         = "invalid_token"
	TokenExpired         = "token_expired"
	InsufficientScope    = "insufficient_scope"
	UnsupportedGrantType = "unsupported_grant_type"
)

// StatusError is returned when the HTTP roundtrip succeeded, but there
// was error was indicated via the HTTP status code, typically due to an
// application-level error.
type StatusError struct {
	StatusCode   int
	ResponseBody bytes.Buffer
	// The following fields may be empty.
	Reason      string
	Details     []map[string]string
	Description string
}

// See https://developer.lyft.com/v1/docs/errors.
type errType struct {
	Slug        string              `json:"error"`
	Details     []map[string]string `json:"error_detail"`
	Description string              `json:"error_description"`
}

// NewStatusError constructs a StatusError from the response. It exists
// solely so that subpackages (such as package auth) can create a
// StatusError using the canonical way. Not meant for external use.
// Does not close rsp.Body.
func NewStatusError(rsp *http.Response) *StatusError {
	var buf bytes.Buffer   // for the ResponseBody
	buf.ReadFrom(rsp.Body) // ignore errors

	decodeBuf := bytes.NewBuffer(buf.Bytes()) // to parse the response body
	var errTyp errType
	decodeErr := json.NewDecoder(decodeBuf).Decode(&errTyp)

	// Determine the value for the Reason field; from the header
	// otherwise from the body.
	var e string
	v := rsp.Header["error"]
	if len(v) != 0 {
		e = v[0]
	} else if decodeErr == nil {
		e = errTyp.Slug
	}

	return &StatusError{
		StatusCode:   rsp.StatusCode,
		ResponseBody: buf,
		Reason:       e,
		Details:      errTyp.Details, // safe to access even if decodeErr != nil
		Description:  errTyp.Description,
	}
}

func (s *StatusError) Error() string {
	if s.Reason != "" {
		return fmt.Sprintf("%s: status code: %d", s.Reason, s.StatusCode)
	}
	return fmt.Sprintf("status code: %d", s.StatusCode)
}

// IsRateLimit returns whether the error arose because of running into a
// rate limit.
func IsRateLimit(err error) bool {
	if se, ok := err.(*StatusError); ok {
		return se.StatusCode == 429
	}
	return false
}

// IsTokenExpired returns true if the error arose because the access token
// expired.
func IsTokenExpired(err error) bool {
	if se, ok := err.(*StatusError); ok {
		// https://developer.lyft.com/v1/docs/authentication#section-http-status-codes
		// There doesn't seem to be a canonical way?
		return (se.StatusCode == 401 && len(se.ResponseBody.Bytes()) == 0) || se.Reason == TokenExpired
	}
	return false
}

// RequestID gets the value of the Request-ID key from a response header.
func RequestID(h http.Header) string {
	return h.Get("Request-ID")
}

// RateRemaining returns the value of X-Ratelimit-Remaining.
func RateRemaining(h http.Header) (n int, ok bool) {
	return intHeaderValue(h, "X-Ratelimit-Remaining")
}

// RateRemaining returns the value of X-Ratelimit-Limit.
func RateLimit(h http.Header) (n int, ok bool) {
	return intHeaderValue(h, "X-Ratelimit-Limit")
}

func intHeaderValue(h http.Header, k string) (int, bool) {
	vals, ok := h[k]
	if !ok || len(vals) == 0 {
		return 0, false
	}
	i, err := strconv.Atoi(vals[0])
	if err != nil {
		return 0, false
	}
	return i, true
}
