package shared

const (
	VolumeTypeHost      = "HostPath"
	VolumeTypeProjected = "Projected"
	VolumeTypeSecret    = "Secret"
	VolumeTypeConfigMap = "ConfigMap"
)

const (
	TokenTypeSA       = "ServiceAccount"
	TokenTypeBoostrap = "Bootstrap"
	TokenTypeOIDC     = "OIDC"
)

const (
	IdentityTypeSA    = "ServiceAccount"
	IdentityTypeUser  = "User"
	IdentityTypeGroup = "Group"
)

type CompromiseType int

const (
	CompromiseNone CompromiseType = iota
	CompromiseSimulated
	CompromiseKnown
)

type EndpointExposureType int

const (
	EndpointExposureNone      EndpointExposureType = iota
	EndpointExposureClusterIP                      // Container port exposed to cluster
	EndpointExposureNodeIP                         // Kubernetes endpoint exposed outside the cluster
	EndpointExposureExternal                       // Kubernetes endpoint exposed outside the cluster
	EndpointExposurePublic                         // External DNS API endpoint
)
