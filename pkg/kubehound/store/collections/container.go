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

func (c Container) Indices() []IndexSpec {
	return []IndexSpec{
		{
			Field: "pod_id",
			Type:  SingleIndex,
		},
		{
			Field: "node_id",
			Type:  SingleIndex,
		},
		{
			Field: "inherited.namespace",
			Type:  SingleIndex,
		},
		{
			Field: "inherited.serviceaccount",
			Type:  SingleIndex,
		},
	}
}
