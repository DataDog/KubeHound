package collections

type FakeCollection struct {
}

const (
	TestBatchSize = 5
)

var _ Collection = (*FakeCollection)(nil) // Ensure interface compliance

func (c FakeCollection) Name() string {
	return "FakeCollectionName"
}

func (c FakeCollection) BatchSize() int {
	return TestBatchSize
}
