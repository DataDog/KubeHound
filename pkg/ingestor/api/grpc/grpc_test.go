package grpc

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/blob"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/mocks"
	"github.com/stretchr/testify/mock"
)

func TestGRPCIngestorAPI_Ingest(t *testing.T) {
	type fields struct {
		puller puller.DataPuller
		cfg    *config.KubehoundConfig
	}
	type args struct {
		ctx         context.Context
		clusterName string
		runID       string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		mock    func(puller *mocks.DataPuller)
	}{
		{
			name: "Pulling invalid bucket name",
			fields: fields{
				cfg: config.MustLoadEmbedConfig(),
			},
			args: args{
				ctx:         context.Background(),
				clusterName: "test-cluster",
				runID:       "test-run-id",
			},
			wantErr: true,
			mock: func(puller *mocks.DataPuller) {
				puller.On("Pull", mock.Anything, "test-cluster", "test-run-id").Return("", blob.ErrInvalidBucketName)
			},
		},
		{
			name: "Pull archive successfully",
			fields: fields{
				cfg: config.MustLoadEmbedConfig(),
			},
			args: args{
				ctx:         context.Background(),
				clusterName: "test-cluster",
				runID:       "test-run-id",
			},
			wantErr: false,
			mock: func(puller *mocks.DataPuller) {
				puller.On("Pull", mock.Anything, "test-cluster", "test-run-id").Return("/tmp/kubehound/kh-1234/test-cluster/test-run-id", nil)
				puller.On("Close", mock.Anything, "/tmp/kubehound/kh-1234/test-cluster/test-run-id").Return(nil)
				puller.On("Extract", mock.Anything, "/tmp/kubehound/kh-1234/test-cluster/test-run-id").Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedPuller := mocks.NewDataPuller(t)
			g := NewGRPCIngestorAPI(tt.fields.cfg, mockedPuller)
			tt.mock(mockedPuller)
			if err := g.Ingest(tt.args.ctx, tt.args.clusterName, tt.args.runID); (err != nil) != tt.wantErr {
				t.Errorf("GRPCIngestorAPI.Ingest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
