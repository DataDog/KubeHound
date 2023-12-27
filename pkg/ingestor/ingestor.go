package ingestor

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	grpcapi "github.com/DataDog/KubeHound/pkg/ingestor/api/grpc"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/blob"
)

func Launch(ctx context.Context) error {
	port := config.DefaultIngestorAPIPort
	addr := config.DefaultIngestorAPIAddr

	puller, err := blob.NewBlobStoragePuller()
	if err != nil {
		return err
	}
	ingestorApi := grpcapi.NewGRPCIngestorAPI(port, addr, puller)
	err = ingestorApi.Listen(ctx)
	if err != nil {
		return err
	}

	return nil
}
