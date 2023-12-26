package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/DataDog/KubeHound/pkg/ingestor/api"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	grpc "google.golang.org/grpc"
)

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api.proto
type GRPCIngestorApi struct {
	port   int
	addr   string
	puller puller.DataPuller
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	UnimplementedAPIServer
	api *GRPCIngestorApi
}

var _ api.API = (*GRPCIngestorApi)(nil)

func NewGRPCIngestorAPI(port int, addr string, puller puller.DataPuller) GRPCIngestorApi {
	return GRPCIngestorApi{
		port:   port,
		addr:   addr,
		puller: puller,
	}
}

func (g *GRPCIngestorApi) Ingest(ctx context.Context, clusterName string, runID string) error {
	archivePath, err := g.puller.Pull(ctx, clusterName, runID)
	if err != nil {
		return err
	}
	g.puller.Close(ctx, archivePath)

	return nil
}

func (g *GRPCIngestorApi) Listen(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	RegisterAPIServer(s, &server{})
	log.I.Infof("server listening at %v", lis.Addr())
	err = s.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}

// Ingest is just a GRPC wrapper around the Ingest method from the service
func (s *server) Ingest(ctx context.Context, in *IngestRequest) (*IngestResponse, error) {
	err := s.api.Ingest(ctx, in.GetClustername(), in.GetRunId())
	if err != nil {
		return nil, err
	}
	return &IngestResponse{}, nil
}
