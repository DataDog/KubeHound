package collections

type FakeCollection struct {
}

var _ Collection = (*FakeCollection)(nil) // Ensure interface compliance

func (c FakeCollection) Name() string {
	return "FakeCollectionName"
}
