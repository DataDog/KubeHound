package collections

type Endpoint struct {
}

var _ Collection = (*Endpoint)(nil) // Ensure interface compliance

func (c Endpoint) Name() string {
	return EndpointName
}
