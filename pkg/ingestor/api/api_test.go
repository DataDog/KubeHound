package api

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	mocksNotifier "github.com/DataDog/KubeHound/pkg/ingestor/notifier/mocks"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/blob"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/mocks"
	"github.com/stretchr/testify/mock"
)

func TestIngestorAPI_Ingest(t *testing.T) {
	t.Parallel()
	type fields struct {
		cfg *config.KubehoundConfig
	}
	type args struct {
		clusterName string
		runID       string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		mock    func(puller *mocks.DataPuller, notifier *mocksNotifier.Notifier)
	}{
		{
			name: "Pulling invalid bucket name",
			fields: fields{
				cfg: config.MustLoadEmbedConfig(),
			},
			args: args{
				clusterName: "test-cluster",
				runID:       "test-run-id",
			},
			wantErr: true,
			mock: func(puller *mocks.DataPuller, notifier *mocksNotifier.Notifier) {
				puller.On("Pull", mock.Anything, "test-cluster", "test-run-id").Return("", blob.ErrInvalidBucketName)
			},
		},
		{
			name: "Pull archive successfully",
			fields: fields{
				cfg: config.MustLoadEmbedConfig(),
			},
			args: args{
				clusterName: "test-cluster",
				runID:       "test-run-id",
			},
			wantErr: false,
			mock: func(puller *mocks.DataPuller, notifier *mocksNotifier.Notifier) {
				puller.On("Pull", mock.Anything, "test-cluster", "test-run-id").Return("/tmp/kubehound/kh-1234/test-cluster/test-run-id", nil)
				puller.On("Close", mock.Anything, "/tmp/kubehound/kh-1234/test-cluster/test-run-id").Return(nil)
				puller.On("Extract", mock.Anything, "/tmp/kubehound/kh-1234/test-cluster/test-run-id").Return(nil)
				notifier.On("Notify", mock.Anything, "test-cluster", "test-run-id").Return(nil)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockedPuller := mocks.NewDataPuller(t)
			mockedNotifier := mocksNotifier.NewNotifier(t)
			g := NewIngestorAPI(tt.fields.cfg, mockedPuller, mockedNotifier)
			tt.mock(mockedPuller, mockedNotifier)
			if err := g.Ingest(context.TODO(), tt.args.clusterName, tt.args.runID); (err != nil) != tt.wantErr {
				t.Errorf("IngestorAPI.Ingest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
