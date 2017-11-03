package auth

type Scope string

const (
	Public       Scope = "public"
	RidesRead    Scope = "rides.read"
	Offline      Scope = "offline"
	RidesRequest Scope = "rides.request"
	Profile      Scope = "rides.profile"
)

// Possible values for the Reason field in StatusError.
const (
	InvalidToken         = "invalid_token"
	TokenExpired         = "token_expired"
	InsufficientScope    = "insufficient_scope"
	UnsupportedGrantType = "unsupported_grant_type"
)
