package config

import (
	"testing"

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
			name: "Setup correct config",
			args: args{
				configPath: "./testdata/kubehound.yaml",
			},
			want: &KubehoundConfig{
				Collector: CollectorConfig{
					Type: CollectorTypeFile,
					File: &FileCollectorConfig{
						Directory: "cluster-data/",
					},
				},
				MongoDB: MongoDBConfig{
					URL: "mongodb://localhost:27017",
				},
				Telemetry: TelemetryConfig{
					Statsd: StatsdConfig{
						URL: "127.0.0.1:8125",
					},
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg, err := NewConfig(tt.args.configPath)
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
