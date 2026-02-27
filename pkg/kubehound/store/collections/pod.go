package collections

type Pod struct {
}

var _ Collection = (*Pod)(nil) // Ensure interface compliance

func (c Pod) Name() string {
	return PodName
}
