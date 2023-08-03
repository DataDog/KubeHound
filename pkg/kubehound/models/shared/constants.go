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

type EndpointExposureType int

const (
	EndpointExposureNone     EndpointExposureType = iota
	EndpointExposureInternal                      // Container port exposed to cluster
	EndpointExposureExternal                      // Kubernetes endpoint exposed outside the cluster
	EndpointExposurePublic                        // External DNS API endpoint
)
