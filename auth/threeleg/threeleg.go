package threeleg

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/nishanths/lyft/auth"
)

// AuthorizationURL returns the URL that a user should be directed to, in order
// to grant the list of permissions your application is requesting.
func AuthorizationURL(clientID string, scopes []auth.Scope, state string) string {
	s := make([]string, len(scopes))
	for i, scope := range scopes {
		s[i] = string(scope)
	}
	v := url.Values{
		"client_id": []string{clientID},
		"scopes":    s,
		"state":     []string{state},
	}
	return fmt.Sprintf("https://api.lyft.com/oauth/authorize?%s", v.Encode())
}

// AuthorizationCode retrieves the authorization code sent in the
// authorization redirect request from Lyft.
// If ReadForm hasn't been called on the request already, it will be
// called during the process.
func AuthorizationCode(r *http.Request) string {
	return r.FormValue("code")
}

type GenerateTokenResponse struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expires      time.Duration
	Scopes       []auth.Scope
}

type RefreshTokenResponse struct {
	AccessToken string
	TokenType   string
	Expires     time.Duration
	Scopes      []auth.Scope
}

func GenerateToken(c *http.Client, clientID, clientSecret, code string) (GenerateTokenResponse, error) {

}

func RefreshToken(c *http.Client, clientID, clientSecret, refreshToken string) (RefreshTokenResponse, error) {

}

func RevokeToken(c *http.Client, clientID, clientSecret, accessToken string) error {

}
