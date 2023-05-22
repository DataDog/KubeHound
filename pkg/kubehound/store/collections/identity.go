package collections

type Identity struct {
}

var _ Collection = (*Identity)(nil) // Ensure interface compliance

func (c Identity) Name() string {
	return IdentityName
}

func (c Identity) BatchSize() int {
	return DefaultBatchSize
}
