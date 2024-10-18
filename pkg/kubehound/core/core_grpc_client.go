package core

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	pb "github.com/DataDog/KubeHound/pkg/ingestor/api/grpc/pb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func getGrpcConn(ingestorConfig config.IngestorConfig) (*grpc.ClientConn, error) {
	var dialOpt grpc.DialOption
	if ingestorConfig.API.Insecure {
		dialOpt = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		}
		dialOpt = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}

	conn, err := grpc.NewClient(ingestorConfig.API.Endpoint, dialOpt)
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", ingestorConfig.API.Endpoint, err)
	}

	return conn, nil
}

func CoreClientGRPCIngest(ctx context.Context, ingestorConfig config.IngestorConfig, clusteName string, runID string) error {
	l := log.Logger(ctx)
	conn, err := getGrpcConn(ingestorConfig)
	if err != nil {
		return fmt.Errorf("getGrpcClient: %w", err)
	}
	defer conn.Close()
	client := pb.NewAPIClient(conn)
	l.Info("Launching ingestion", log.String("endpoint", ingestorConfig.API.Endpoint), log.String(log.FieldRunIDKey, runID))

	_, err = client.Ingest(ctx, &pb.IngestRequest{
		RunId:       runID,
		ClusterName: clusteName,
	})
	if err != nil {
		return fmt.Errorf("call Ingest (%s:%s): %w", clusteName, runID, err)
	}

	return nil
}

func CoreClientGRPCRehydrateLatest(ctx context.Context, ingestorConfig config.IngestorConfig) error {
	l := log.Logger(ctx)
	conn, err := getGrpcConn(ingestorConfig)
	if err != nil {
		return fmt.Errorf("getGrpcClient: %w", err)
	}
	defer conn.Close()
	client := pb.NewAPIClient(conn)

	l.Info("Launching rehydratation [latest]", log.String("endpoint", ingestorConfig.API.Endpoint))
	results, err := client.RehydrateLatest(ctx, &pb.RehydrateLatestRequest{})
	if err != nil {
		return fmt.Errorf("call rehydratation (latest): %w", err)
	}

	for _, res := range results.IngestedCluster {
		l.Info("Rehydrated cluster", log.String(log.FieldClusterKey, res.ClusterName), log.Time("time", res.Date.AsTime()), log.String("key", res.Key))
	}

	return nil
}
