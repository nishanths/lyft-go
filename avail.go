package lyft

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// TODO: Richer StatusError for all methods.

// RideType is returned by the client's RideTypes method.
type RideType struct {
	DisplayName string  `json:"display_name"`
	RideType    string  `json:"ride_type"`
	ImageURL    string  `json:"image_url"`
	Pricing     Pricing `json:"pricing_details"`
	Seats       int     `json:"seats"`
}

type Pricing struct {
	Base            int    `json:"base_charge"`
	PerMile         int    `json:"cost_per_mile"`
	PerMinute       int    `json:"cost_per_minute"`
	Minimum         int    `json:"cost_minimum"`
	TrustAndService int    `json:"trust_and_service"`
	Currency        string `json:"currency"`
	CancelPenalty   int    `json:"cancel_penalty_amout"`
}

func formatFloat(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}

// RideTypes returns the ride types available at the location.
// The rideType is optional. If set, details will be returned for the specified
// ride type only. If no ride types are available, the error will
// be a StatusError.
//
// TODO: what does the error look like when no rides are available?
func (c *Client) RideTypes(lat, lng float64, rideType string) ([]RideType, error) {
	vals := make(url.Values)
	vals.Set("lat", formatFloat(lat))
	vals.Set("lng", formatFloat(lng))
	if rideType != "" {
		vals.Set("ride_type", rideType)
	}
	r, err := http.NewRequest("GET", c.BaseURL+"/v1/ridetypes?"+vals.Encode(), nil)
	if err != nil {
		return nil, err
	}

	rsp, err := c.do(r)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	// TODO: this has more details in the error response.
	if rsp.StatusCode != 200 {
		return nil, NewStatusError(rsp)
	}

	var response struct {
		RideTypes []RideType `json:"ride_types"`
	}
	if err := json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.RideTypes, nil
}

// Auxiliary type for unmarshaling.
// This type corresponds to "cost_estimates" in the Lyft API reference.
type rideEstimate struct {
	RideType       string  `json:"ride_type"`
	DisplayName    string  `json:"display_name"`
	MaximumCost    int     `json:"estimated_cost_cents_max"`
	MinimumCost    int     `json:"estimated_cost_cents_min"`
	Distance       float64 `json:"estimated_distance_miles"`
	Duration       int64   `json:"estimated_distance_seconds"`
	PrimetimeToken string  `json:"primetime_confirmation_token"`
	CostToken      string  `json:"cost_token"`
	Valid          bool    `json:"is_valid_estimate"`
}

// RideEstimate is returned by the client's RideEstimates method.
type RideEstimate struct {
	RideType       string
	DisplayName    string
	MaximumCost    int           // Estimated maximum cost of the ride.
	MinimumCost    int           // Estimated minimum cost of the ride.
	Distance       float64       // Estimated distance of the ride; in miles.
	Duration       time.Duration // Estimated duration of the ride.
	PrimetimeToken string        // DEPRECATED; see CostToken and https://developer.lyft.com/reference#availability-ride-estimates.
	CostToken      string
	Valid          bool // If false, MaximumCost and MinimumCost may be invalid.
}

func (r *RideEstimate) UnmarshalJSON(p []byte) error {
	var aux rideEstimate
	if err := json.Unmarshal(p, &aux); err != nil {
		return err
	}
	r.RideType = aux.RideType
	r.DisplayName = aux.DisplayName
	r.MaximumCost = aux.MaximumCost
	r.MinimumCost = aux.MinimumCost
	r.Distance = aux.Distance
	r.Duration = time.Second * time.Duration(aux.Duration)
	r.PrimetimeToken = aux.PrimetimeToken
	r.CostToken = aux.CostToken
	r.Valid = aux.Valid
	return nil
}

// IgnoreArg is a sentinel value that can be used when calling a function
// that has an optional float64 argument.
const IgnoreArg float64 = -181 // so that valid longitudes aren't ignored.

// RideEstimates returns the estimated cost, distance, and duration of a ride.
// The end locations are optional and are ignored if the value equals
// the package-level const IgnoreArg. rideType is also optional; if it is set, estimates
// will be returned for the specified type only.
func (c *Client) RideEstimates(startLat, startLng, endLat, endLng float64, rideType string) ([]RideEstimate, error) {
	vals := make(url.Values)
	vals.Set("start_lat", formatFloat(startLat))
	vals.Set("start_lng", formatFloat(startLng))
	if endLat != IgnoreArg {
		vals.Set("end_lat", formatFloat(endLat))
	}
	if endLng != IgnoreArg {
		vals.Set("end_lng", formatFloat(endLng))
	}
	if rideType != "" {
		vals.Set("ride_type", formatFloat(endLng))
	}
	r, err := http.NewRequest("GET", c.BaseURL+"/v1/cost?"+vals.Encode(), nil)
	if err != nil {
		return nil, err
	}

	rsp, err := c.do(r)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return nil, NewStatusError(rsp)
	}

	var response struct {
		C []RideEstimate `json:"cost_estimates"`
	}
	if err := json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.C, nil
}

// Auxiliary type for unmarshaling.
type etaEstimate struct {
	RideType    string `json:"ride_type"`
	DisplayName string `json:"display_name"`
	ETA         int64  `json:"eta_seconds"`
	Valid       bool   `json:"is_valid_estimate"`
}

// ETAEstimate is returned by the client's DriverETA method.
type ETAEstimate struct {
	RideType    string
	DisplayName string
	ETA         time.Duration
	Valid       bool // If false, ETA may be invalid.
}

func (e *ETAEstimate) UnmarshalJSON(p []byte) error {
	var aux etaEstimate
	if err := json.Unmarshal(p, &aux); err != nil {
		return err
	}
	e.RideType = aux.RideType
	e.DisplayName = aux.DisplayName
	e.ETA = time.Second * time.Duration(aux.ETA)
	e.Valid = aux.Valid
	return nil
}

// DriverETA estimates the time for the nearest driver to reach the specifed location.
// The end locations are optional and are ignored if the value equals the
// package-level const IgnoreArg. The rideType argument is also optional. If set,
// estimates will be returned for the specified type only.
func (c *Client) DriverETA(startLat, startLng, endLat, endLng float64, rideType string) ([]ETAEstimate, error) {
	vals := make(url.Values)
	vals.Set("lat", formatFloat(startLat))
	vals.Set("lng", formatFloat(startLng))
	if endLat != IgnoreArg {
		vals.Set("destination_lat", formatFloat(endLat))
	}
	if endLng != IgnoreArg {
		vals.Set("destination_lng", formatFloat(endLng))
	}
	if rideType != "" {
		vals.Set("ride_type", formatFloat(endLng))
	}
	r, err := http.NewRequest("GET", c.BaseURL+"/v1/eta?"+vals.Encode(), nil)
	if err != nil {
		return nil, err
	}

	rsp, err := c.do(r)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return nil, NewStatusError(rsp)
	}

	var response struct {
		E []ETAEstimate `json:"eta_estimates"`
	}
	if err := json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.E, nil
}

// NearbyDriver is returned by the client's DriversNearby method.
type NearbyDriver struct {
	Drivers  []DriverLocation `json:"drivers"`
	RideType string           `json:"ride_type"`
}

type DriverLocation struct {
	Locations []Location `json:"locations"` // Most recent coordinates (TODO: but in which order? WTF, Lyft API docs)
}

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

// DriversNearby returns the location of drivers near a location.
func (c *Client) DriversNearby(lat, lng float64) ([]NearbyDriver, error) {
	vals := make(url.Values)
	vals.Set("lat", formatFloat(lat))
	vals.Set("lng", formatFloat(lng))
	r, err := http.NewRequest("GET", c.BaseURL+"/v1/drivers?"+vals.Encode(), nil)
	if err != nil {
		return nil, err
	}

	rsp, err := c.do(r)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return nil, NewStatusError(rsp)
	}

	var response struct {
		N []NearbyDriver `json:"nearby_drivers"`
	}
	if err := json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.N, nil
}
