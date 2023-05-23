package collections

type Container struct {
}

var _ Collection = (*Container)(nil) // Ensure interface compliance

func (c Container) Name() string {
	return ContainerName
}

func (c Container) BatchSize() int {
	return DefaultBatchSize
}
