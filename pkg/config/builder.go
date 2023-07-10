package config

const (
	DefaultEdgeWorkerPoolSize         = 10
	DefaultEdgeWorkerPoolCapacity     = 100
	DefaultEdgeBatchSize              = 500
	DefaultEdgeBatchSizeSmall         = DefaultEdgeBatchSize / 8
	DefaultEdgeBatchSizeClusterImpact = 1

	DefaultVertexBatchSize = 500
)

type VertexBuilderConfig struct {
	BatchSize int `mapstructure:"batch_size"`
}

type EdgeBuilderConfig struct {
	LargeCluster           bool `mapstructure:"large_cluster"`
	WorkerPoolSize         int  `mapstructure:"worker_pool_size"`
	WorkerPoolCapacity     int  `mapstructure:"worker_pool_capacity"`
	BatchSize              int  `mapstructure:"batch_size"`
	BatchSizeSmall         int  `mapstructure:"batch_size_small"`
	BatchSizeClusterImpact int  `mapstructure:"batch_size_cluster_impact"`
}

type BuilderConfig struct {
	Vertex VertexBuilderConfig `mapstructure:"vertex"`
	Edge   EdgeBuilderConfig   `mapstructure:"edge"`
}
