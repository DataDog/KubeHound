package collections

type RoleBinding struct {
}

var _ Collection = (*RoleBinding)(nil) // Ensure interface compliance

func (c RoleBinding) Name() string {
	return RoleBindingName
}

func (c RoleBinding) BatchSize() int {
	return DefaultBatchSize
}
