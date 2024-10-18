package grpc

import (
	"context"
	"net"

	"github.com/DataDog/KubeHound/pkg/dump"
	"github.com/DataDog/KubeHound/pkg/ingestor/api"
	pb "github.com/DataDog/KubeHound/pkg/ingestor/api/grpc/pb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// On macOS you need to install protobuf (`brew install protobuf`)
// Need to install: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
//go:generate protoc --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb --go-grpc_opt=paths=source_relative ./api.proto

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
	l := log.Logger(ctx)
	// Rebuilding the path for the dump archive file
	dumpResult, err := dump.NewDumpResult(in.GetClusterName(), in.GetRunId(), true)
	if err != nil {
		return nil, err
	}
	key := dumpResult.GetFullPath()

	err = s.api.Ingest(ctx, key)
	if err != nil {
		l.Error("Ingest failed", log.ErrorField(err))

		return nil, err
	}

	return &pb.IngestResponse{}, nil
}

// RehydrateLatest is just a GRPC wrapper around the RehydrateLatest method from the API package
func (s *server) RehydrateLatest(ctx context.Context, in *pb.RehydrateLatestRequest) (*pb.RehydrateLatestResponse, error) {
	l := log.Logger(ctx)
	res, err := s.api.RehydrateLatest(ctx)
	if err != nil {
		l.Error("Ingest failed", log.ErrorField(err))

		return nil, err
	}

	return &pb.RehydrateLatestResponse{
		IngestedCluster: res,
	}, nil
}

// Listen starts the GRPC server with the generic api implementation
// It uses the config from the passed API for address and ports
func Listen(ctx context.Context, api *api.IngestorAPI) error {
	l := log.Logger(ctx)
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
	l.Infof("server listening at %v", lis.Addr())
	err = s.Serve(lis)
	if err != nil {
		return err
	}

	return nil
}
