// Package threeleg provides functions for working with the three-legged
// OAuth flow described at https://developer.lyft.com/v1/docs/authentication#section-3-legged-flow-for-accessing-user-specific-endpoints.
package threeleg

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AuthorizationURL returns the URL that a user should be directed to, in order
// to grant the list of permissions your application is requesting.
func AuthorizationURL(clientID string, scopes []string, state string) string {
	v := make(url.Values)
	v.Set("client_id", clientID)
	v.Set("scopes", strings.Join(scopes, " "))
	v.Set("state", state)
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
	Scopes       []string
}

type RefreshTokenResponse struct {
	AccessToken string
	TokenType   string
	Expires     time.Duration
	Scopes      []string
}

func GenerateToken(c *http.Client, clientID, clientSecret, code string) (GenerateTokenResponse, error) {
	panic("not impl")
}

func RefreshToken(c *http.Client, clientID, clientSecret, refreshToken string) (RefreshTokenResponse, error) {
	panic("not impl")
}

func RevokeToken(c *http.Client, clientID, clientSecret, accessToken string) error {
	panic("not impl")
}
