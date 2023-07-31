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

type EndpointAccessType int

const (
	EndpointAccessNone     EndpointAccessType = iota
	EndpointAccessInternal                    // Container port exposed to cluster
	EndpointAccessExternal                    // Kubernetes endpoint exposed outside the cluster
	EndpointAccessPublic                      // Public API endpoint
)
