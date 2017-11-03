package auth

type Scope string

const (
	Public       Scope = "public"
	RidesRead    Scope = "rides.read"
	Offline      Scope = "offline"
	RidesRequest Scope = "rides.request"
	Profile      Scope = "rides.profile"
)
