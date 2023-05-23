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
