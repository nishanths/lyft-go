package lyft

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Ride types. May not be an exhaustive list.
const (
	RideTypeLyft    = "lyft"
	RideTypePlus    = "lyft_plus"
	RideTypeLine    = "lyft_line"
	RideTypePremier = "lyft_premier"
	RideTypeLux     = "lyft_lux"
	RideTypeLuxSUV  = "lyft_luxsuv"
)

// RideRequest is the paramters for the client's RequestRide method.
type RideRequest struct {
	Origin      Location `json:"origin"`      // Latitude and Longitude fields are required
	Destination Location `json:"destination"` // Latitude and Longitude fields are required
	RideType    string   `json:"ride_type"`   // Required
	CostToken   string   `json:"cost_token"`  // Optional
}

// CreatedRide is returned by the client's RequestRide method.
type CreatedRide struct {
	RideID      string   `json:"ride_id"`
	RideStatus  string   `json:"status"` // StatusPending for newly requested rides
	RideType    string   `json:"ride_type"`
	Origin      Location `json:"origin"`
	Destination Location `json:"destination"`
	Passenger   Person   `json:"passenger"` // The Phone field will not be set
}

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	Address   string  `json:"address"`
}

// RequestRide requests a ride for a user.
// As of 2017-11-05, Lyft Line is not fully supported. See
// https://developer.lyft.com/reference#ride-request for details.
func (c *Client) RequestRide(req RideRequest) (CreatedRide, http.Header, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return CreatedRide{}, nil, err
	}
	r, err := http.NewRequest("POST", c.base()+"/v1/rides", &buf)
	if err != nil {
		return CreatedRide{}, nil, err
	}
	r.Header.Set("Content-Type", "application/json")

	rsp, err := c.do(r)
	if err != nil {
		return CreatedRide{}, nil, err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case 201:
		var cr CreatedRide
		if err := json.NewDecoder(rsp.Body).Decode(&cr); err != nil {
			return CreatedRide{}, rsp.Header, err
		}
		return cr, rsp.Header, nil
	case 400:
		// TODO: This should use a typed error for 400 responses.
		panic("unhandled case")
	default:
		return CreatedRide{}, rsp.Header, NewStatusError(rsp)
	}
}

// func (c *Client) CancelRide()
// func (c *Client) RateRide()
// func (c *Client) SetDestination()
// func (c *Client) RideReceipt()
// func (c *Client) RideDetails()
