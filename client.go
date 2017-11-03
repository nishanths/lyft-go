package lyft

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Client struct {
	ClientID     string
	ClientSecret string
	AccessTokens map[string]struct{}

	HTTPClient *http.Client // uses http.DefaultClient if nil
	Headers    http.Header  // optional extra headers
	BaseURL    string       // the base URL of the API; uses BaseURL if empty. Useful in testing.
}

const (
	InvalidToken      = "invalid_token"
	TokenExpired      = "token_expired"
	InsufficientScope = "insufficient_scope"
)

type StatusError struct {
	StatusCode int
	Reason     string
	Body       io.ReadCloser
}

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

func (s *StatusError) Unauthorized() bool {
	return s.StatusCode == http.StatusUnauthorized
}

// Bytes returns the contents of Body. Body is closed automatically
// after reading. Errors during reading or closing are not reported.
// This mainly serves as a convenience method.
func (s *StatusError) Bytes() []byte {
	defer s.Body.Close()
	b, _ := ioutil.ReadAll(s.Body)
	return b
}
