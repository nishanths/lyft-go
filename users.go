package lyft

import (
	"encoding/json"
	"net/http"
)

func (c *Client) RideHistory() {
	panic("not implemented")
}

// UserProfile is returned by the client's UserProfile method.
type UserProfile struct {
	ID        string `json:"id"` // Authenticated user's ID.
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Ridden    bool   `json:"has_taken_a_ride"` // Whether the user has taken at least one ride.
}

func (c *Client) UserProfile(id string) (UserProfile, error) {
	r, err := http.NewRequest("GET", c.BaseURL+"/v1/profile", nil)
	if err != nil {
		return UserProfile{}, err
	}

	rsp, err := c.do(r)
	if err != nil {
		return UserProfile{}, err
	}
	defer rsp.Body.Close()

	// TODO: this has more details in the error response.
	if rsp.StatusCode != 200 {
		return UserProfile{}, NewStatusError(rsp)
	}

	var p UserProfile
	if err := json.NewDecoder(rsp.Body).Decode(&p); err != nil {
		return UserProfile{}, err
	}
	return p, nil
}
