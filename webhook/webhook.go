// Package webhook provides types and utility functions for handling
// Lyft webhooks.
package webhook // import "go.avalanche.space/lyft/webhook"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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
	Event     lyft.RideDetail
}

func (e Event) IsSandbox() bool {
	return strings.HasPrefix(e.EventID, SandboxEventPrefix)
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
