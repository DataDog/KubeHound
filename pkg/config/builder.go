package config

const (
	DefaultEdgeWorkerPoolSize         = 5
	DefaultEdgeWorkerPoolCapacity     = 100
	DefaultEdgeBatchSize              = 500
	DefaultEdgeBatchSizeSmall         = DefaultEdgeBatchSize / 5
	DefaultEdgeBatchSizeClusterImpact = 10

	DefaultVertexBatchSize      = 500
	DefaultVertexBatchSizeSmall = DefaultVertexBatchSize / 5

	DefaultStopOnError = false

	DefaultLargeClusterOptimizations = true
)

// VertexBuilderConfig configures vertex builder parameters.
type VertexBuilderConfig struct {
	BatchSize      int `mapstructure:"batch_size"`       // Batch size for inserts
	BatchSizeSmall int `mapstructure:"batch_size_small"` // Batch size for expensive inserts

}

// EdgeBuilderConfig configures edge builder parameters.
type EdgeBuilderConfig struct {
	LargeClusterOptimizations bool `mapstructure:"large_cluster_optimizations"`
	WorkerPoolSize            int  `mapstructure:"worker_pool_size"`          // Number of workers for the edge builder worker pool
	WorkerPoolCapacity        int  `mapstructure:"worker_pool_capacity"`      // Work item capacity for the edge builder worker pool
	BatchSize                 int  `mapstructure:"batch_size"`                // Batch size for inserts
	BatchSizeSmall            int  `mapstructure:"batch_size_small"`          // Batch size for expensive inserts
	BatchSizeClusterImpact    int  `mapstructure:"batch_size_cluster_impact"` // Batch size for inserts impacting entire cluster e.g POD_PATCH
}

type BuilderConfig struct {
	Vertex      VertexBuilderConfig `mapstructure:"vertex"`        // Vertex builder config
	Edge        EdgeBuilderConfig   `mapstructure:"edge"`          // Edge builder config
	StopOnError bool                `mapstructure:"stop_on_error"` // Stop the building of the graph on error
}
