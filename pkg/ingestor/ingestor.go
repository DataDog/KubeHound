package ingestor

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor/api"
	"github.com/DataDog/KubeHound/pkg/ingestor/api/grpc"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/blob"
)

func Launch(ctx context.Context, cfg *config.KubehoundConfig) error {
	puller, err := blob.NewBlobStoragePuller(cfg)
	if err != nil {
		return err
	}
	ingestorApi := api.NewIngestorAPI(cfg, puller, nil)

	err = grpc.Listen(ctx, ingestorApi)
	if err != nil {
		return err
	}

	return nil
}
