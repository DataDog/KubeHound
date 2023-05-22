package collections

type Role struct {
}

var _ Collection = (*Role)(nil) // Ensure interface compliance

func (c Role) Name() string {
	return RoleName
}

func (c Role) BatchSize() int {
	return DefaultBatchSize
}
