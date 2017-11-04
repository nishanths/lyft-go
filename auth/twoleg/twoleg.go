package twoleg

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/nishanths/lyft"
)

type GenerateTokenResponse struct {
	AccessToken string
	TokenType   string
	Expires     time.Duration
	Scopes      []string
}

type generateTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Expires     int64  `json:"expires_in"` // seconds
	Scopes      string `json:"scope"`      // space delimited
}

func GenerateToken(c *http.Client, clientID, clientSecret, code string) (GenerateTokenResponse, error) {
	body := strings.NewReader(`{"grant_type": "client_credentials", "scope": "public"}`)
	r, err := http.NewRequest("POST", lyft.BaseURL+"/oauth/token", body)
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
		AccessToken: g.AccessToken,
		TokenType:   g.TokenType,
		Expires:     time.Second * time.Duration(g.Expires),
		Scopes:      strings.Fields(g.Scopes),
	}, nil
}
