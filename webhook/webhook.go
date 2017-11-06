// Package webhook provides types and utility functions for handling
// Lyft webhooks.
package webhook // import "go.avalanche.space/lyft/webhook"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.avalanche.space/lyft"
)

// Event types.
const (
	RideStatusUpdated = "ride.status.updated"
	RideReceiptReady  = "ride.receipt.ready"
)

const SandboxEventPrefix = "sandboxevent"

// Event represents an event from a Lyft webhook.
// It implements json.Unmarshaler in a manner suitable for decoding
// incoming webhook request bodies.
type Event struct {
	EventID   string
	URL       string
	Occurred  time.Time
	EventType string
	Detail    lyft.RideDetail // Some fields may not be set
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

// Signature returns the value of "X-Lyft-Signature" in the header.
// The "sha256=" will have been trimmed.
func Signature(h http.Header) string {
	return strings.TrimPrefix(h.Get("X-Lyft-Signature"), "sha256=")
}

func VerifySignature(responseBody, signature, verificationToken []byte) bool {
	mac := hmac.New(sha256.New, verificationToken)
	mac.Write(responseBody)
	bodySignature := mac.Sum(nil)
	return hmac.Equal([]byte(base64.StdEncoding.EncodeToString(bodySignature)), signature)
}
