package ingestor

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	grpcapi "github.com/DataDog/KubeHound/pkg/ingestor/api/grpc"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/blob"
)

func Launch(ctx context.Context, cfg *config.KubehoundConfig) error {
	puller, err := blob.NewBlobStoragePuller(cfg)
	if err != nil {
		return err
	}
	ingestorApi := grpcapi.NewGRPCIngestorAPI(cfg, puller)
	err = ingestorApi.Listen(ctx)
	if err != nil {
		return err
	}

	return nil
}
