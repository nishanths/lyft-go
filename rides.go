package lyft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
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

type CostTokenInfo struct {
	PrimetimePercentage string
	PrimetimeMultiplier float64
	PrimetimeToken      string
	CostToken           string
	TokenDuration       time.Duration
	ErrorURI            string
}

func newCostTokenInfo(body io.Reader) (CostTokenInfo, error) {
	var c CostTokenInfo
	err := json.NewDecoder(body).Decode(&c)
	return c, err
}

func (c *CostTokenInfo) UnmarshalJSON(p []byte) error {
	type costTokenInfo struct {
		PrimetimePercentage string  `json:"primetime_percentage"`
		PrimetimeMultiplier float64 `json:"primetime_multiplier"`
		PrimetimeToken      string  `json:"primetime_confirmation_token"`
		CostToken           string  `json:"cost_token"`
		// Is this seriously a string? Even Swagger says so.
		// http://petstore.swagger.io/?url=https://api.lyft.com/v1/spec
		TokenDuration string `json:"token_duration"` // in seconds
		ErrorURI      string `json:"error_uri"`
	}
	var aux costTokenInfo
	if err := json.Unmarshal(p, &aux); err != nil {
		return err
	}
	c.PrimetimePercentage = aux.PrimetimePercentage
	c.PrimetimeMultiplier = aux.PrimetimeMultiplier
	c.PrimetimeToken = aux.PrimetimeToken
	c.CostToken = aux.CostToken
	i, err := strconv.ParseInt(aux.TokenDuration, 10, 64)
	if err != nil {
		return err
	}
	c.TokenDuration = time.Second * time.Duration(i)
	c.ErrorURI = aux.ErrorURI
	return nil
}

var _ error = (*RideRequestError)(nil)

type RideRequestError struct {
	ErrorInfo                // Fields may be empty
	Cost      *CostTokenInfo // May be nil
}

func newRideRequestError(rsp *http.Response) *RideRequestError {
	var eiBuf bytes.Buffer
	eiBuf.ReadFrom(rsp.Body)
	ciBuf := bytes.NewBuffer(eiBuf.Bytes())

	ei := newErrorInfo(&eiBuf, rsp.Header)
	ci, err := newCostTokenInfo(ciBuf)

	ret := &RideRequestError{
		ErrorInfo: ei,
	}
	if err == nil {
		ret.Cost = &ci
	}
	return ret
}

func (c *RideRequestError) Error() string {
	if c.Reason != "" && c.Description != "" {
		return fmt.Sprintf("%s: %s", c.Reason, c.Description)
	} else if c.Reason != "" {
		return c.Reason
	}
	return "<ride request error>"
}

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
//
// If further action (such as confirming the cost) is required before the
// ride can be successfully created, the error will be of type *RideRequestError.
// This corresponds to the 400 status code documented in Lyft's API reference.
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
		return CreatedRide{}, rsp.Header, newRideRequestError(rsp)
	default:
		return CreatedRide{}, rsp.Header, NewStatusError(rsp)
	}
}

// func (c *Client) CancelRide()

// SetDestination updates the ride's destination to the supplied location.
// The location's Address field is optional.
func (c *Client) SetDestination(rideID string, loc Location) (Location, http.Header, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(loc); err != nil {
		return Location{}, nil, err
	}
	r, err := http.NewRequest("PUT", fmt.Sprintf("%s/v1/rides/%s/destination", c.base(), rideID), &buf)
	if err != nil {
		return Location{}, nil, err
	}
	r.Header.Set("Content-Type", "application/json")

	rsp, err := c.do(r)
	if err != nil {
		return Location{}, nil, err
	}
	defer rsp.Body.Close()

	switch rsp.StatusCode {
	case 200:
		var ret Location
		if err := json.NewDecoder(rsp.Body).Decode(&ret); err != nil {
			return Location{}, rsp.Header, err
		}
		return ret, rsp.Header, nil
	default:
		return Location{}, rsp.Header, NewStatusError(rsp)
	}
}

// func (c *Client) RideReceipt()

// TODO: Implement these.
// func (c *Client) RideDetails()
// func (c *Client) RateRide()
