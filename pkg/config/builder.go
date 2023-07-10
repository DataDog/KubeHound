package config

type VertexBuilderConfig struct {
	BatchSize int `mapstructure:"batch_size"`
}

type EdgeBuilderConfig struct {
	LargeCluster           bool `mapstructure:"large_cluster"`
	WorkerPoolSize         int  `mapstructure:"worker_pool_size"`
	BatchSizeDefault       int  `mapstructure:"batch_size_default"`
	BatchSizeSmall         int  `mapstructure:"batch_size_small"`
	BatchSizeClusterImpact int  `mapstructure:"batch_size_cluster_impact"`
}

type BuilderConfig struct {
	Vertex VertexBuilderConfig `mapstructure:"vertex"`
	Edge   EdgeBuilderConfig   `mapstructure:"edge"`
}
