// Package webhook provides types and utility functions for handling
// Lyft webhooks.
package webhook // import "go.avalanche.space/lyft/webhook"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"go.avalanche.space/lyft"
)

// Event types.
const (
	EventRideStatusUpdated = "ride.status.updated"
	EventReceiptReady      = "ride.receipt.ready"
)

const SandboxEventPrefix = "sandboxevent"

type Event struct {
	EventID   string
	URL       string
	Occurred  time.Time
	EventType string
	Detail    lyft.RideDetail
}

func (e *Event) IsSandbox() bool {
	return strings.HasPrefix(e.EventID, SandboxEventPrefix)
}

func (e *Event) UnmarshalJSON(p []byte) error {
	type event struct {
		EventID   string          `json:"event_id"`
		URL       string          `json:"href"`
		Occurred  string          `json:"occurred_at"`
		EventType string          `json:"event_type"`
		Detail    lyft.RideDetail `json:"event"`
	}
	var aux event
	if err := json.Unmarshal(p, &aux); err != nil {
		return err
	}
	e.EventID = aux.EventID
	e.URL = aux.URL
	if aux.Occurred != "" {
		o, err := time.Parse(lyft.TimeLayout, aux.Occurred)
		if err != nil {
			return err
		}
		e.Occurred = o
	}
	e.EventID = aux.EventType
	e.Detail = aux.Detail
	return nil
}

func VerifySignature(body, signature, verificationToken []byte) bool {
	mac := hmac.New(sha256.New, verificationToken)
	mac.Write(body)
	bodySignature := mac.Sum(nil)
	return hmac.Equal([]byte(base64.StdEncoding.EncodeToString(bodySignature)), signature)
}

func drainAndClose(r io.ReadCloser) {
	io.Copy(ioutil.Discard, r)
	r.Close()
}
