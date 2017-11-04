package lyft

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Ride statuses.
const (
	StatusPending    = "pending"
	StatusAccepted   = "accepted"
	StatusArrived    = "arrived"
	StatusPickedUp   = "pickedUp"
	StatusDroppedOff = "droppedOff"
	StatusCanceled   = "canceled"
	StatusUnknown    = "unknown"
)

// Ride profiles.
const (
	ProfileBusiness = "business"
	ProfilePersonal = "personal"
)

const dateTimeLayout = "2006-01-02T15:04:05-07:00" // used in unmarshaling

// Auxiliary type for unmarshaling.
type rideHistory struct {
	RideID              string            `json:"ride_id"`
	RideStatus          string            `json:"status"`
	RideType            string            `json:"ride_type"`
	Origin              rideLocation      `json:"origin"`
	Pickup              rideLocation      `json:"pickup"`
	Destination         rideLocation      `json:"destination"`
	Dropoff             rideLocation      `json:"dropoff"`
	Location            VehicleLocation   `json:"location"`
	Passenger           Person            `json:"passenger"`
	Driver              Person            `json:"driver"`
	Vehicle             Vehicle           `json:"vehicle"`
	PrimetimePercentage string            `json:"primetime_percentage"`
	Distance            float64           `json:"distance_miles"`
	Duration            float64           `json:"duration_seconds"` // Documented as float64
	Price               Price             `json:"price"`
	LineItems           []LineItem        `json:"line_items"`
	Requested           string            `json:"requested_at"`
	RideProfile         string            `json:"ride_profile"`
	BeaconColor         string            `json:"beacon_string"`
	PricingDetailsURL   string            `json:"pricing_details_url"`
	RouteURL            string            `json:"route_url"`
	CanCancel           []string          `json:"can_cancel"`
	CanceledBy          string            `json:"canceled_by"`
	CancellationPrice   cancellationPrice `json:"cancellation_price"`
	Rating              int               `json:"rating"`
	Feedback            string            `json:"feedback"`
}

func (h rideHistory) convert(res *RideHistory) error {
	var err error
	res.RideID = h.RideID
	res.RideStatus = h.RideStatus
	res.RideType = h.RideType

	err = h.Origin.convert(&res.Origin)
	if err != nil {
		return err
	}
	err = h.Pickup.convert(&res.Pickup)
	if err != nil {
		return err
	}
	err = h.Destination.convert(&res.Destination)
	if err != nil {
		return err
	}
	err = h.Dropoff.convert(&res.Dropoff)
	if err != nil {
		return err
	}

	res.Location = h.Location
	res.Passenger = h.Passenger
	res.Driver = h.Driver
	res.Vehicle = h.Vehicle
	res.PrimetimePercentage = h.PrimetimePercentage
	res.Distance = h.Distance
	res.Duration = time.Second * time.Duration(h.Duration)
	res.Price = h.Price
	res.LineItems = h.LineItems

	res.Requested, err = time.Parse(dateTimeLayout, h.Requested)
	if err != nil {
		return err
	}

	res.RideProfile = h.RideProfile
	res.BeaconColor = h.BeaconColor
	res.PricingDetailsURL = h.PricingDetailsURL
	res.RouteURL = h.RouteURL
	res.CanCancel = h.CanCancel
	res.CanceledBy = h.CanceledBy
	h.CancellationPrice.convert(&res.CancellationPrice)
	res.Rating = h.Rating
	res.Feedback = h.Feedback
	return nil
}

type rideLocation struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	Address   string  `json:"address"`
	ETA       float64 `json:"eta_seconds"` // Documented differently for origin v. destination v. dropoff, so float64 is safest
	Time      string  `json:"time"`
}

func (l rideLocation) convert(res *RideLocation) error {
	var err error
	res.Latitude = l.Latitude
	res.Longitude = l.Longitude
	res.Address = l.Address
	res.ETA = time.Second * time.Duration(l.ETA) // TODO: consider not truncating
	res.Time, err = time.Parse(dateTimeLayout, l.Time)
	return err
}

type cancellationPrice struct {
	Amount        int    `json:"amonut"`
	Currency      string `json:"currency"`
	Token         string `json:"token"`
	TokenDuration int64  `json:"token_duration"` // seconds; documented as int
}

func (c cancellationPrice) convert(res *CancellationPrice) {
	res.Amount = c.Amount
	res.Currency = c.Currency
	res.Token = c.Token
	res.TokenDuration = time.Second * time.Duration(c.TokenDuration) // TODO: consider not truncating
}

// RideHistory is returned by the client's RideHistory method.
// Some fields are available only if certain conditions are true
// at the time of making the request. See the API reference for details.
type RideHistory struct {
	RideID              string
	RideStatus          string
	RideType            string
	Origin              RideLocation
	Pickup              RideLocation
	Destination         RideLocation
	Dropoff             RideLocation
	Location            VehicleLocation
	Passenger           Person
	Driver              Person
	Vehicle             Vehicle
	PrimetimePercentage string
	Distance            float64
	Duration            time.Duration
	Price               Price
	LineItems           []LineItem
	Requested           time.Time
	RideProfile         string
	BeaconColor         string
	PricingDetailsURL   string
	RouteURL            string
	CanCancel           []string
	CanceledBy          string
	CancellationPrice   CancellationPrice
	Rating              int
	Feedback            string
}

type RideLocation struct {
	Latitude  float64
	Longitude float64
	Address   string
	ETA       time.Duration // Only set for RideHistory.Origin, .Destination
	Time      time.Time     // Only set for RideHistory.Pickup, .Dropoff
}

type VehicleLocation struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	Bearing   float64 `json:"bearing"`
}

type Person struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ImageURL  string `json:"image_url"`
	Rating    string `json:"rating"`
	Phone     string `json:"phone_number"` // Only set for RideHistory.Driver
}

type Vehicle struct {
	Make              string `json:"make"`
	Model             string `json:"model"`
	Year              int    `json:"year"`
	LicensePlate      string `json:"license_plate"`
	LicensePlateState string `json:"license_plate_state"`
	Color             string `json:"color"`
	ImageURL          string `json:"image_url"`
}

type Price struct {
	Amount      int    `json:"amonut"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
}

type LineItem struct {
	Amount      int    `json:"amonut"`
	Currency    string `json:"currency"`
	Description string `json:"type"`
}

type CancellationPrice struct {
	Amount        int
	Currency      string
	Token         string
	TokenDuration time.Duration
}

func (h *RideHistory) UnmarshalJSON(p []byte) error {
	var aux rideHistory
	if err := json.Unmarshal(p, &aux); err != nil {
		return err
	}
	return aux.convert(h)
}

// RideHistory returns the authenticated user's current and past rides.
// See the Lyft API reference for details on how far back the
// start and end times can go. If end is the zero time it is ignored.
// Limit specifies the maximum number of rides to return. If limit is -1,
// RideHistory requests the maximum limit documented in the API reference (50).
//
// Implementation detail: The times, in UTC, are formatted using "2006-01-02T15:04:05Z".
// For example: start.UTC().Format("2006-01-02T15:04:05Z").
func (c *Client) RideHistory(start, end time.Time, limit int32) ([]RideHistory, error) {
	const layout = "2006-01-02T15:04:05Z"

	vals := make(url.Values)
	vals.Set("start_time", start.UTC().Format(layout))
	if !end.UTC().IsZero() {
		vals.Set("end_time", end.UTC().Format(layout))
	}
	if limit == -1 {
		limit = 50 // max limit documented in the Lyft API reference
	}
	vals.Set("limit", strconv.FormatInt(int64(limit), 10))
	r, err := http.NewRequest("GET", c.base()+"/v1/rides?"+vals.Encode(), nil)
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
		R []RideHistory `json:"ride_history"`
	}
	if err := json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.R, nil
}

// UserProfile is returned by the client's UserProfile method.
type UserProfile struct {
	ID        string `json:"id"` // Authenticated user's ID.
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Ridden    bool   `json:"has_taken_a_ride"` // Whether the user has taken at least one ride.
}

// UserProfile returns the authenticated user's profile info.
func (c *Client) UserProfile(id string) (UserProfile, error) {
	r, err := http.NewRequest("GET", c.base()+"/v1/profile", nil)
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
