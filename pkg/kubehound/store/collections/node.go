package collections

type Node struct {
}

var _ Collection = (*Node)(nil) // Ensure interface compliance

func (c Node) Name() string {
	return NodeName
}

func (c Node) BatchSize() int {
	return DefaultBatchSize
}
