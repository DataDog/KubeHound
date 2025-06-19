package ingestion

// Ingestion represents an ingestion in the system.
type Ingestion struct {
	// RunID is the identifier of the run.
	RunID string `json:"runID"`
	// Cluster is the identifier of the cluster.
	Cluster string `json:"cluster"`
}
