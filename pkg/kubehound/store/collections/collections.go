package collections

const (
	DefaultBatchSize = 1000
)

const (
	NodeName          = "nodes"
	PodName           = "pods"
	ContainerName     = "containers"
	VolumeName        = "volumes"
	RoleName          = "roles"
	RoleBindingName   = "rolebindings"
	IdentityName      = "identities"
	PermissionSetName = "permissionsets"
	EndpointName      = "endpoints"
)

type IndexType int

const (
	SingleIndex IndexType = iota
	TextIndex
)

type IndexSpec struct {
	Field string
	Type  IndexType
}

type IndexSpecRegistry map[string][]IndexSpec

// Collection provides a common abstraction of a SQL database table or a NoSQL object
// collection to work with the storedb provider interface.
type Collection interface {
	// Name returns the name of the collection.
	Name() string

	// BatchSize returns the batch size of bulk inserts (and threshold for triggering a flush).
	BatchSize() int

	Indices() []IndexSpec
}

// TODO registry
