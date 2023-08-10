package collections

type PermissionSet struct {
}

var _ Collection = (*PermissionSet)(nil) // Ensure interface compliance

func (c PermissionSet) Name() string {
	return PermissionSetName
}

func (c PermissionSet) BatchSize() int {
	return DefaultBatchSize
}
