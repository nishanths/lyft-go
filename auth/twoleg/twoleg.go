package twoleg

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/nishanths/lyft"
	"github.com/nishanths/lyft/auth"
)

type GenerateTokenResponse struct {
	AccessToken string
	TokenType   string
	Expires     time.Duration
	Scopes      []auth.Scope
}

type generateTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"` // seconds
	Scope       string `json:"scope"`      // space delimited
}

func GenerateToken(c *http.Client, clientID, clientSecret, code string) (GenerateTokenResponse, error) {
	body := strings.NewReader(`{"grant_type": "client_credentials", "scope": "public"}`)
	r, err := http.NewRequest("POST", lyft.BaseURL+"/oauth/token", body)
	if err != nil {
		return GenerateTokenResponse{}, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.SetBasicAuth(clientID, clientSecret)

	rsp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		return nil, lyft.NewStatusError(rsp)
	}

	var g generateTokenResponse
	if err := json.NewDecoder(rsp.Body).Decode(&g); err != nil {
		return nil, err
	}

	fields := strings.Fields(g.Scope)
	scopes := make([]auth.Scope, len(fields))
	for i, f := range fields {
		scopes[i] = auth.Scope(f)
	}

	return GenerateTokenResponse{
		AccessToken: g.AccessToken,
		TokenType:   g.TokenType,
		Expires:     time.Second * g.ExpiresIn,
		Scopes:      scopes,
	}, nil
}
