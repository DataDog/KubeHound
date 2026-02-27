package collections

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

// Collection provides a common abstraction of a database table.
type Collection interface {
	// Name returns the name of the collection (table).
	Name() string
}

func GetCollections() []string {
	return []string{
		NodeName,
		PodName,
		ContainerName,
		VolumeName,
		RoleName,
		RoleBindingName,
		IdentityName,
		PermissionSetName,
		EndpointName,
	}
}
