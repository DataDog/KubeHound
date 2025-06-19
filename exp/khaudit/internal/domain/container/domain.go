package container

// Container represents a container in the system.
type Container struct {
	RunID     string `json:"runID"`
	Cluster   string `json:"cluster"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	App       string `json:"app"`
	Team      string `json:"team"`
	Namespace string `json:"namespace"`
}

// NamespaceAggregation represents the aggregation of containers by namespace.
type NamespaceAggregation struct {
	RunID     string `json:"runID"`
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Count     int64  `json:"count"`
}

// AttackPath represents an attack path.
type AttackPath struct {
	RunID          string   `json:"runID"`
	Cluster        string   `json:"cluster"`
	Path           []string `json:"path"`
	ContainerCount int64    `json:"containerCount"`
}
