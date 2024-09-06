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

func CoreClientGRPCIngest(ingestorConfig config.IngestorConfig, clusteName string, runID string) error {
	conn, err := getGrpcConn(ingestorConfig)
	if err != nil {
		return fmt.Errorf("getGrpcClient: %w", err)
	}
	defer conn.Close()
	client := pb.NewAPIClient(conn)

	log.I.Infof("Launching ingestion on %s [rundID: %s]", ingestorConfig.API.Endpoint, runID)

	_, err = client.Ingest(context.Background(), &pb.IngestRequest{
		RunId:       runID,
		ClusterName: clusteName,
	})
	if err != nil {
		return fmt.Errorf("call Ingest (%s:%s): %w", clusteName, runID, err)
	}

	return nil
}

func CoreClientGRPCRehydrateLatest(ingestorConfig config.IngestorConfig) error {
	conn, err := getGrpcConn(ingestorConfig)
	if err != nil {
		return fmt.Errorf("getGrpcClient: %w", err)
	}
	defer conn.Close()
	client := pb.NewAPIClient(conn)

	log.I.Infof("Launching rehydratation on %s [latest]", ingestorConfig.API.Endpoint)
	results, err := client.RehydrateLatest(context.Background(), &pb.RehydrateLatestRequest{})
	if err != nil {
		return fmt.Errorf("call rehydratation (latest): %w", err)
	}

	for _, res := range results.IngestedCluster {
		log.I.Infof("Rehydrated cluster: %s, date: %s, run_id: %s", res.ClusterName, res.Date.AsTime().Format("01-02-2006 15:04:05"), res.Key)
	}

	return nil
}
