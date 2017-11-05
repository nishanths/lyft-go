// Package auth and its subpackages define types and functions
// related to Lyft's OAuth flows.
package auth

// Scopes.
const (
	Public       = "public"
	RidesRead    = "rides.read"
	Offline      = "offline"
	RidesRequest = "rides.request"
	Profile      = "rides.profile"
)

// SandboxSecret returns the sandboxed form of an non-sandboxed client secret.
// See https://developer.lyft.com/v1/docs/sandbox.
func SandboxSecret(clientSecret string) string {
	return "SANDBOX-" + clientSecret
}
