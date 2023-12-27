package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/DataDog/KubeHound/pkg/ingestor/api"
	pb "github.com/DataDog/KubeHound/pkg/ingestor/api/grpc/pb"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

//go:generate protoc --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb --go-grpc_opt=paths=source_relative api.proto
type GRPCIngestorAPI struct {
	port   int
	addr   string
	puller puller.DataPuller
}

// server is used to implement the GRPC api
type server struct {
	pb.UnimplementedAPIServer
	healthgrpc.UnimplementedHealthServer
	api *GRPCIngestorAPI
}

var _ api.API = (*GRPCIngestorAPI)(nil)

func NewGRPCIngestorAPI(port int, addr string, puller puller.DataPuller) GRPCIngestorAPI {
	return GRPCIngestorAPI{
		port:   port,
		addr:   addr,
		puller: puller,
	}
}

func (g *GRPCIngestorAPI) Ingest(ctx context.Context, clusterName string, runID string) error {
	archivePath, err := g.puller.Pull(ctx, clusterName, runID)
	if err != nil {
		return err
	}
	g.puller.Close(ctx, archivePath)

	return nil
}

func (g *GRPCIngestorAPI) Listen(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", g.addr, g.port))
	if err != nil {
		return err
	}
	s := grpc.NewServer()

	// So we have an endpoint easily accessible for k8s health checks.
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(s, healthcheck)

	pb.RegisterAPIServer(s, &server{})
	log.I.Infof("server listening at %v", lis.Addr())
	err = s.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}

// Ingest is just a GRPC wrapper around the Ingest method from the service
func (s *server) Ingest(ctx context.Context, in *pb.IngestRequest) (*pb.IngestResponse, error) {
	err := s.api.Ingest(ctx, in.GetClustername(), in.GetRunId())
	if err != nil {
		return nil, err
	}
	return &pb.IngestResponse{}, nil
}
