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

	"github.com/nishanths/lyft"
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

// Token is returned by GenerateToken.
type Token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expires      time.Duration
	Scopes       []string
}

type token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expires      int64  // seconds
	Scopes       string // space delimited
}

// RefToken is returned by RefreshToken.
type RefToken struct {
	AccessToken string
	TokenType   string
	Expires     time.Duration
	Scopes      []string
}

type refToken struct {
	AccessToken string
	TokenType   string
	Expires     int64  // seconds
	Scopes      string // space delimited
}

func GenerateTokenHeader(c *http.Client, baseURL, clientID, clientSecret, code string) (Token, http.Header, error) {
	body := fmt.Sprintf(`{"grant_type": "authorization_code", "code": "%s"}`, code)
	r, err := http.NewRequest("POST", baseURL+"/oauth/token", strings.NewReader(body))
	if err != nil {
		return Token{}, nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.SetBasicAuth(clientID, clientSecret)

	rsp, err := c.Do(r)
	if err != nil {
		return Token{}, nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return Token{}, rsp.Header, lyft.NewStatusError(rsp)
	}

	var g token
	if err := json.NewDecoder(rsp.Body).Decode(&g); err != nil {
		return Token{}, rsp.Header, err
	}
	return Token{
		AccessToken:  g.AccessToken,
		RefreshToken: g.RefreshToken,
		TokenType:    g.TokenType,
		Expires:      time.Second * time.Duration(g.Expires),
		Scopes:       strings.Fields(g.Scopes),
	}, rsp.Header, nil
}

// GenerateToken creates a new access token using the authorization code
// obtained from Lyft's authorization redirect. The access token
// returned can be used in lyft.Client. baseURL is typically lyft.BaseURL.
func GenerateToken(c *http.Client, baseURL, clientID, clientSecret, code string) (Token, error) {
	g, _, err := GenerateTokenHeader(c, baseURL, clientID, clientSecret, code)
	return g, err
}

func RefreshTokenHeader(c *http.Client, baseURL, clientID, clientSecret, refreshToken string) (RefToken, http.Header, error) {
	body := fmt.Sprintf(`{"grant_type": "refresh_token", "refresh_token": "%s"}`, refreshToken)
	r, err := http.NewRequest("POST", baseURL+"/oauth/token", strings.NewReader(body))
	if err != nil {
		return RefToken{}, nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.SetBasicAuth(clientID, clientSecret)

	rsp, err := c.Do(r)
	if err != nil {
		return RefToken{}, nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return RefToken{}, rsp.Header, lyft.NewStatusError(rsp)
	}

	var ref refToken
	if err := json.NewDecoder(rsp.Body).Decode(&ref); err != nil {
		return RefToken{}, rsp.Header, err
	}
	return RefToken{
		AccessToken: ref.AccessToken,
		TokenType:   ref.TokenType,
		Expires:     time.Second * time.Duration(ref.Expires),
		Scopes:      strings.Fields(ref.Scopes),
	}, rsp.Header, nil
}

// RefreshToken refreshes the access token associated with refreshToken.
// See Token for obtaining access/refresh token pairs.
// baseURL is typically lyft.BaseURL.
func RefreshToken(c *http.Client, baseURL, clientID, clientSecret, refreshToken string) (RefToken, error) {
	r, _, err := RefreshTokenHeader(c, baseURL, clientID, clientSecret, refreshToken)
	return r, err
}

func RevokeTokenHeader(c *http.Client, baseURL, clientID, clientSecret, accessToken string) (http.Header, error) {
	body := fmt.Sprintf(`{"token": "%s"}`, accessToken)
	r, err := http.NewRequest("POST", baseURL+"/oauth/revoke_refresh_token", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.SetBasicAuth(clientID, clientSecret)

	rsp, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return rsp.Header, lyft.NewStatusError(rsp)
	}
	return rsp.Header, nil
}

// RevokeToken revokes the supplied access token.
// baseURL is typically lyft.BaseURL.
func RevokeToken(c *http.Client, baseURL, clientID, clientSecret, accessToken string) error {
	_, err := RevokeTokenHeader(c, baseURL, clientID, clientSecret, accessToken)
	return err
}
