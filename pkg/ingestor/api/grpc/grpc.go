package grpc

import (
	"context"
	"net"

	"github.com/DataDog/KubeHound/pkg/ingestor/api"
	pb "github.com/DataDog/KubeHound/pkg/ingestor/api/grpc/pb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// server is used to implement the GRPC api
type server struct {
	// grpc related embeds
	pb.UnimplementedAPIServer
	healthgrpc.UnimplementedHealthServer

	// api is the actual implementation of the api
	// this is the only gRPC wrapper around it
	// Most function should be basics type conversion and call to s.api.<function>
	api *api.IngestorAPI
}

// Ingest is just a GRPC wrapper around the Ingest method from the API package
func (s *server) Ingest(ctx context.Context, in *pb.IngestRequest) (*pb.IngestResponse, error) {
	err := s.api.Ingest(ctx, in.GetClusterName(), in.GetRunId())
	if err != nil {
		log.I.Errorf("Ingest failed: %v", err)

		return nil, err
	}

	return &pb.IngestResponse{}, nil
}

// RehydrateLatest is just a GRPC wrapper around the RehydrateLatest method from the API package
func (s *server) RehydrateLatest(ctx context.Context, in *pb.RehydrateLatestRequest) (*pb.RehydrateLatestResponse, error) {
	res, err := s.api.RehydrateLatest(ctx)
	if err != nil {
		log.I.Errorf("Ingest failed: %v", err)

		return nil, err
	}

	return &pb.RehydrateLatestResponse{
		IngestedCluster: res,
	}, nil
}

// Listen starts the GRPC server with the generic api implementation
// It uses the config from the passed API for address and ports
func Listen(ctx context.Context, api *api.IngestorAPI) error {
	lis, err := net.Listen("tcp", api.Cfg.Ingestor.API.Endpoint)
	if err != nil {
		return err
	}
	s := grpc.NewServer()

	// So we have an endpoint easily accessible for k8s health checks.
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(s, healthcheck)

	// Register reflection service on gRPC server.
	reflection.Register(s)

	pb.RegisterAPIServer(s, &server{
		api: api,
	})
	log.I.Infof("server listening at %v", lis.Addr())
	err = s.Serve(lis)
	if err != nil {
		return err
	}

	return nil
}
