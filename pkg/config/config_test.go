package config

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMustLoadConfig(t *testing.T) {
	t.Parallel()
	type args struct {
		configPath string
	}
	tests := []struct {
		name    string
		args    args
		want    *KubehoundConfig
		wantErr bool
	}{
		{
			name: "Setup correct config for file collector",
			args: args{
				configPath: "./testdata/kubehound-file-collector.yaml",
			},
			want: &KubehoundConfig{
				Storage: StorageConfig{
					RetryDelay: DefaultRetryDelay,
					Retry:      DefaultRetry,
					Wipe:       true,
				},
				Collector: CollectorConfig{
					Type: CollectorTypeFile,
					File: &FileCollectorConfig{
						Directory: "cluster-data/",
						Archive: &FileArchiveConfig{
							NoCompress: DefaultArchiveNoCompress,
						},
					},
					// This is always set as the default value
					Live: &K8SAPICollectorConfig{
						PageSize:           500,
						PageBufferSize:     10,
						RateLimitPerSecond: 100,
					},
				},
				MongoDB: MongoDBConfig{
					URL:               "mongodb://localhost:27017",
					ConnectionTimeout: DefaultConnectionTimeout,
				},
				JanusGraph: JanusGraphConfig{
					URL:               "ws://localhost:8182/gremlin",
					ConnectionTimeout: DefaultConnectionTimeout,
					WriterTimeout:     defaultJanusGraphWriterTimeout,
					WriterMaxRetry:    defaultJanusGraphWriterMaxRetry,
					WriterWorkerCount: defaultJanusGraphWriterWorkerCount,
				},
				Telemetry: TelemetryConfig{
					Statsd: StatsdConfig{
						URL: "127.0.0.1:8125",
					},
					Profiler: ProfilerConfig{
						Period:      DefaultProfilerPeriod,
						CPUDuration: DefaultProfilerCPUDuration,
					},
				},
				Builder: BuilderConfig{
					Vertex: VertexBuilderConfig{
						BatchSize:      250,
						BatchSizeSmall: 50,
					},
					Edge: EdgeBuilderConfig{
						LargeClusterOptimizations: DefaultLargeClusterOptimizations,
						WorkerPoolSize:            5,
						WorkerPoolCapacity:        100,
						BatchSize:                 250,
						BatchSizeSmall:            50,
						BatchSizeClusterImpact:    10,
					},
				},
				Ingestor: IngestorConfig{
					API: IngestorAPIConfig{
						Endpoint: "",
						Insecure: false,
					},
					Blob: &BlobConfig{
						BucketUrl: "",
						Region:    "",
					},
					TempDir:        "/tmp/kubehound",
					ArchiveName:    "archive.tar.gz",
					MaxArchiveSize: DefaultMaxArchiveSize,
				},
				Dynamic: DynamicConfig{
					ClusterName: "test-cluster",
				},
			},
			wantErr: false,
		},
		{
			name: "Setup correct config for k8s collector",
			args: args{
				configPath: "./testdata/kubehound-k8s-collector.yaml",
			},
			want: &KubehoundConfig{
				Storage: StorageConfig{
					RetryDelay: DefaultRetryDelay,
					Retry:      DefaultRetry,
					Wipe:       true,
				},
				Collector: CollectorConfig{
					Type: CollectorTypeK8sAPI,
					File: &FileCollectorConfig{
						Archive: &FileArchiveConfig{
							NoCompress: DefaultArchiveNoCompress,
						},
					},
					Live: &K8SAPICollectorConfig{
						PageSize:           500,
						PageBufferSize:     10,
						RateLimitPerSecond: 100,
					},
				},
				MongoDB: MongoDBConfig{
					URL:               "mongodb://localhost:27017",
					ConnectionTimeout: DefaultConnectionTimeout,
				},
				JanusGraph: JanusGraphConfig{
					URL:               "ws://localhost:8182/gremlin",
					ConnectionTimeout: DefaultConnectionTimeout,
					WriterTimeout:     defaultJanusGraphWriterTimeout,
					WriterMaxRetry:    defaultJanusGraphWriterMaxRetry,
					WriterWorkerCount: defaultJanusGraphWriterWorkerCount,
				},
				Telemetry: TelemetryConfig{
					Statsd: StatsdConfig{
						URL: "127.0.0.1:8125",
					},
					Profiler: ProfilerConfig{
						Period:      DefaultProfilerPeriod,
						CPUDuration: DefaultProfilerCPUDuration,
					},
				},
				Builder: BuilderConfig{
					Vertex: VertexBuilderConfig{
						BatchSize:      1000,
						BatchSizeSmall: 50,
					},
					Edge: EdgeBuilderConfig{
						LargeClusterOptimizations: true,
						WorkerPoolSize:            5,
						WorkerPoolCapacity:        50,
						BatchSize:                 1000,
						BatchSizeSmall:            100,
						BatchSizeClusterImpact:    5,
					},
				},
				Ingestor: IngestorConfig{
					API: IngestorAPIConfig{
						Endpoint: "",
						Insecure: false,
					},
					Blob: &BlobConfig{
						BucketUrl: "",
						Region:    "",
					},
					TempDir:        "/tmp/kubehound",
					ArchiveName:    "archive.tar.gz",
					MaxArchiveSize: DefaultMaxArchiveSize,
				},
			},
			wantErr: false,
		},
		{
			name: "Setup incorrect config",
			args: args{
				configPath: "./testdata/non-existing.yaml",
			},
			want:    &KubehoundConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := viper.New()
			cfg, err := NewConfig(context.TODO(), v, tt.args.configPath)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, cfg)
			}
		})
	}
}
