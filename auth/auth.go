package auth

// Scopes.
const (
	Public       = "public"
	RidesRead    = "rides.read"
	Offline      = "offline"
	RidesRequest = "rides.request"
	Profile      = "rides.profile"
)

// Possible values for the Reason field in StatusError.
const (
	InvalidToken         = "invalid_token"
	TokenExpired         = "token_expired"
	InsufficientScope    = "insufficient_scope"
	UnsupportedGrantType = "unsupported_grant_type"
)
