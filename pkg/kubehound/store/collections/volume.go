package collections

type Volume struct {
}

var _ Collection = (*Volume)(nil) // Ensure interface compliance

func (c Volume) Name() string {
	return VolumeName
}

func (c Volume) BatchSize() int {
	return DefaultBatchSize
}

func (c Volume) Indices() []IndexSpec {
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
			Field: "container_id",
			Type:  SingleIndex,
		},
		{
			Field: "projected_id",
			Type:  SingleIndex,
		},
		{
			Field: "name",
			Type:  SingleIndex,
		},
		{
			Field: "type",
			Type:  SingleIndex,
		},
		{
			Field: "source_path",
			Type:  TextIndex,
		},
		{
			Field: "readonly",
			Type:  SingleIndex,
		},
	}
}
