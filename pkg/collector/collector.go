package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
)

// Generic interface to allow an ingestor to consume stream inputs from a collector.
type GenericIngestor interface {
	NodeIngestor
	PodIngestor
	RoleIngestor
	ClusterRoleIngestor
	RoleBindingIngestor
	ClusterRoleBindingIngestor
	EndpointIngestor
}

// NodeIngestor defines the interface to allow an ingestor to consume node inputs from a collector.
//
//go:generate mockery --name NodeIngestor --output mockingest --case underscore --filename node_ingestor.go --with-expecter
type NodeIngestor interface {
	IngestNode(context.Context, types.NodeType) error
	Complete(context.Context) error
}

// PodIngestor defines the interface to allow an ingestor to consume pod inputs from a collector.
//
//go:generate mockery --name PodIngestor --output mockingest --case underscore --filename pod_ingestor.go --with-expecter
type PodIngestor interface {
	IngestPod(context.Context, types.PodType) error
	Complete(context.Context) error
}

// RoleIngestor defines the interface to allow an ingestor to consume role inputs from a collector.
//
//go:generate mockery --name RoleIngestor --output mockingest --case underscore --filename role_ingestor.go --with-expecter
type RoleIngestor interface {
	IngestRole(context.Context, types.RoleType) error
	Complete(context.Context) error
}

// ClusterRoleIngestor defines the interface to allow an ingestor to consume cluster role inputs from a collector.
//
//go:generate mockery --name ClusterRoleIngestor --output mockingest --case underscore --filename cluster_role_ingestor.go --with-expecter
type ClusterRoleIngestor interface {
	IngestClusterRole(context.Context, types.ClusterRoleType) error
	Complete(context.Context) error
}

// RoleBindingIngestor defines the interface to allow an ingestor to consume role binding inputs from a collector.
//
//go:generate mockery --name RoleBindingIngestor --output mockingest --case underscore --filename role_binding_ingestor.go --with-expecter
type RoleBindingIngestor interface {
	IngestRoleBinding(context.Context, types.RoleBindingType) error
	Complete(context.Context) error
}

// ClusterRoleBindingIngestor defines the interface to allow an ingestor to consume cluster role binding inputs from a collector.
//
//go:generate mockery --name ClusterRoleBindingIngestor --output mockingest --case underscore --filename cluster_role_binding_ingestor.go --with-expecter
type ClusterRoleBindingIngestor interface {
	IngestClusterRoleBinding(context.Context, types.ClusterRoleBindingType) error
	Complete(context.Context) error
}

// EndpointIngestor defines the interface to allow an ingestor to consume endpoint slice inputs from a collector.
//
//go:generate mockery --name EndpointIngestor --output mockingest --case underscore --filename endpoint_ingestor.go --with-expecter
type EndpointIngestor interface {
	IngestEndpoint(context.Context, types.EndpointType) error
	Complete(context.Context) error
}

// MetadataIngestor defines the interface to allow an ingestor to computed metrics and metadata from a collector.
type MetadataIngestor interface {
	DumpMetadata(context.Context, Metadata) error
}

//go:generate mockery --name CollectorClient --output mockcollector --case underscore --filename collector_client.go --with-expecter
type CollectorClient interface { //nolint: interfacebloat
	services.Dependency

	// ClusterInfo returns the target cluster information for the current run.
	ClusterInfo(ctx context.Context) (*config.ClusterInfo, error)

	// Compute the metrics and gather all the metadata and dump it through the ingestor.DumpMetadata
	ComputeMetadata(ctx context.Context, ingestor MetadataIngestor) error

	// StreamNodes will iterate through all NodeType objects collected by the collector and invoke the ingestor.IngestNode method on each.
	// Once all the NodeType objects have been exhausted the ingestor.Complete method will be invoked to signal the end of the stream.
	StreamNodes(ctx context.Context, ingestor NodeIngestor) error

	// StreamPods will iterate through all PodType objects collected by the collector and invoke the ingestor.IngestPod method on each.
	// Once all the PodType objects have been exhausted the ingestor.Complete method will be invoked to signal the end of the stream.
	StreamPods(ctx context.Context, ingestor PodIngestor) error

	// StreamRoles will iterate through all RoleType objects collected by the collector and invoke ingestor.IngestRole method on each.
	// Once all the RoleType objects have been exhausted the ingestor.Complete method will be invoked to signal the end of the stream.
	StreamRoles(ctx context.Context, ingestor RoleIngestor) error

	// StreamClusterRoles will iterate through all ClusterRoleType objects collected by the collector and invoke the ingestor.IngestRole method on each.
	// Once all the ClusterRoleType objects have been exhausted the ingestor.Complete method will be invoked to signal the end of the stream.
	StreamClusterRoles(ctx context.Context, ingestor ClusterRoleIngestor) error

	// StreamRoleBindings will iterate through all RoleBindingType objects collected by the collector and invoke the ingestor.IngestRoleBinding method on each.
	// Once all the RoleBindingType objects have been exhausted the ingestor.Complete method will be invoked to signal the end of the stream.
	StreamRoleBindings(ctx context.Context, ingestor RoleBindingIngestor) error

	// StreamClusterRoleBindings will iterate through all ClusterRoleBindingType objects collected by the collector and invoke the ingestor.ClusterRoleBinding method on each.
	// Once all the ClusterRoleBindingType objects have been exhausted the ingestor.Complete method will be invoked to signal the end of the stream.
	StreamClusterRoleBindings(ctx context.Context, ingestor ClusterRoleBindingIngestor) error

	// StreamEndpoints will iterate through all EndpointType objects collected by the collector and invoke the ingestor.IngestEndpoint method on each.
	// Once all the EndpointType objects have been exhausted the ingestor.Complete method will be invoked to signal the end of the stream.
	StreamEndpoints(ctx context.Context, ingestor EndpointIngestor) error

	// Close cleans up any resources used by the collector client implementation. Client cannot be reused after this call.
	Close(ctx context.Context) error
}

// ClientFactory creates an initialized instance of a collector client based on the provided application configuration.
func ClientFactory(ctx context.Context, cfg *config.KubehoundConfig) (CollectorClient, error) {
	switch {
	case cfg.Collector.Type == config.CollectorTypeK8sAPI:
		return NewK8sAPICollector(ctx, cfg)
	case cfg.Collector.Type == config.CollectorTypeFile:
		return NewFileCollector(ctx, cfg)
	default:
		return nil, fmt.Errorf("collector type not supported: %s", cfg.Collector.Type)
	}
}

type collectorTags struct {
	pod                []string
	role               []string
	rolebinding        []string
	endpoint           []string
	node               []string
	clusterrole        []string
	clusterrolebinding []string
	baseTags           []string
}

func newCollectorTags() *collectorTags {
	return &collectorTags{
		pod:                tag.GetBaseTagsWith(tag.Collector(FileCollectorName), tag.Entity(tag.EntityPods)),
		role:               tag.GetBaseTagsWith(tag.Collector(FileCollectorName), tag.Entity(tag.EntityRoles)),
		rolebinding:        tag.GetBaseTagsWith(tag.Collector(FileCollectorName), tag.Entity(tag.EntityRolebindings)),
		endpoint:           tag.GetBaseTagsWith(tag.Collector(FileCollectorName), tag.Entity(tag.EntityEndpoints)),
		node:               tag.GetBaseTagsWith(tag.Collector(FileCollectorName), tag.Entity(tag.EntityNodes)),
		clusterrole:        tag.GetBaseTagsWith(tag.Collector(FileCollectorName), tag.Entity(tag.EntityClusterRoles)),
		clusterrolebinding: tag.GetBaseTagsWith(tag.Collector(FileCollectorName), tag.Entity(tag.EntityClusterRolebindings)),
		baseTags:           tag.GetBaseTags(),
	}
}

type Metrics struct {
	DumpTime             time.Time     `json:"dump_time"`
	RunDuration          time.Duration `json:"run_duration"`
	TotalWaitTime        time.Duration `json:"total_wait_time"`
	ThrottlingPercentage float64       `json:"throttling_percentage"`
}

type Metadata struct {
	RunID       string  `json:"run_id"`
	ClusterName string  `json:"cluster"`
	Metrics     Metrics `json:"metrics"`
}
