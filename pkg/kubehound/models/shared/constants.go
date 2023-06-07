package shared

const (
	VolumeTypeHost      = "HostPath"
	VolumeTypeProjected = "Projected"
)

const (
	TokenTypeSA       = "ServiceAccount"
	TokenTypeBoostrap = "Bootstrap"
	TokenTypeOIDC     = "OIDC"
)

type CompromiseType int

const (
	CompromiseNone CompromiseType = iota
	CompromiseSimulated
	CompromiseKnown
)
