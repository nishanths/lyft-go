// Package threeleg provides functions for working with the three-legged
// OAuth flow described at https://developer.lyft.com/v1/docs/authentication#section-3-legged-flow-for-accessing-user-specific-endpoints.
package threeleg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AuthorizationURL is the URL that a user should be directed to, in order for the user
// to grant the list of permissions your application is requesting.
func AuthorizationURL(clientID string, scopes []string, state string) string {
	v := make(url.Values)
	v.Set("client_id", clientID)
	v.Set("response_type", "code") // only possible value
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

// GenerateTokenResponse is returned by GenerateToken.
type GenerateTokenResponse struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expires      time.Duration
	Scopes       []string
}

type generateTokenResponse struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expires      int64    // seconds
	Scopes       []string // space delimited
}

// RefreshTokenResponse is returned by RefreshToken.
type RefreshTokenResponse struct {
	AccessToken string
	TokenType   string
	Expires     time.Duration
	Scopes      []string
}

// GenerateToken creates a new access token using the authorization code
// obtained from Lyft's authorization redirect. The access token
// returned can be used in lyft.Client. baseURL is typically lyft.BaseURL.
func GenerateToken(c *http.Client, baseURL, clientID, clientSecret, code string) (GenerateTokenResponse, error) {
	body := fmt.Sprintf(`{"grant_type": "authorization_code", "code": "%s"}`, code)
	r, err := http.NewRequest("POST", baseURL+"/oauth/token", strings.NewReader(body))
	if err != nil {
		return GenerateTokenResponse{}, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.SetBasicAuth(clientID, clientSecret)

	rsp, err := c.Do(r)
	if err != nil {
		return GenerateTokenResponse{}, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return GenerateTokenResponse{}, lyft.NewStatusError(rsp)
	}

	var g generateTokenResponse
	if err := json.NewDecoder(rsp.Body).Decode(&g); err != nil {
		return GenerateTokenResponse{}, err
	}
	return GenerateTokenResponse{
		AccessToken:  g.AccessToken,
		RefreshToken: g.RefreshToken,
		TokenType:    g.TokenType,
		Expires:      time.Second * time.Duration(g.Expires),
		Scopes:       strings.Fields(g.Scopes),
	}, nil
}

// RefreshToken refreshes the access token associated with refreshToken.
// See GenerateTokenResponse for obtaining access/refresh token pairs.
// baseURL is typically lyft.BaseURL.
func RefreshToken(c *http.Client, baseURL, clientID, clientSecret, refreshToken string) (RefreshTokenResponse, error) {
	panic("not impl")
}

// RevokeToken revokes the supplied access token.
// baseURL is typically lyft.BaseURL.
func RevokeToken(c *http.Client, baseURL, clientID, clientSecret, accessToken string) error {
	panic("not impl")
}
