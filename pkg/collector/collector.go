package collector

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
)

type Complete func(context.Context) error
type NodeProcessor func(context.Context, *types.NodeType) error
type PodProcessor func(context.Context, *types.PodType) error
type RoleProcessor func(context.Context, *types.RoleType) error
type ClusterRoleProcessor func(context.Context, *types.ClusterRoleType) error
type RoleBindingProcessor func(context.Context, *types.RoleBindingType) error
type ClusterRoleBindingProcessor func(context.Context, *types.ClusterRoleBindingType) error

//go:generate mockery --name CollectorClient --output mocks --case underscore --filename collector_client.go --with-expecter
type CollectorClient interface {
	services.Dependency

	// StreamNodes will iterate through all NodeType objects collected by the collector and invoke the NodeProcessor callback on each.
	// Once all the NodeType objects have been exhausted the Complete callback will be invoked to signal the end of the stream.
	StreamNodes(ctx context.Context, callback NodeProcessor, complete Complete) error

	// StreamPods will iterate through all PodType objects collected by the collector and invoke the PodProcessor callback on each.
	// Once all the PodType objects have been exhausted the Complete callback will be invoked to signal the end of the stream.
	StreamPods(ctx context.Context, callback PodProcessor, complete Complete) error

	// StreamRoles will iterate through all RoleType objects collected by the collector and invoke the RoleProcessor callback on each.
	// Once all the RoleType objects have been exhausted the Complete callback will be invoked to signal the end of the stream.
	StreamRoles(ctx context.Context, callback RoleProcessor, complete Complete) error

	// StreamClusterRoles will iterate through all ClusterRoleType objects collected by the collector and invoke the ClusterRoleProcessor callback on each.
	// Once all the ClusterRoleType objects have been exhausted the Complete callback will be invoked to signal the end of the stream.
	StreamClusterRoles(ctx context.Context, callback ClusterRoleProcessor, complete Complete) error

	// StreamRoleBindings will iterate through all RoleBindingType objects collected by the collector and invoke the RoleBindingProcessor callback on each.
	// Once all the RoleBindingType objects have been exhausted the Complete callback will be invoked to signal the end of the stream.
	StreamRoleBindings(ctx context.Context, callback RoleBindingProcessor, complete Complete) error

	// StreamClusterRoleBindings will iterate through all ClusterRoleBindingType objects collected by the collector and invoke the ClusterRoleBindingProcessor callback on each.
	// Once all the ClusterRoleBindingType objects have been exhausted the Complete callback will be invoked to signal the end of the stream.
	StreamClusterRoleBindings(ctx context.Context, callback ClusterRoleBindingProcessor, complete Complete) error

	// Close cleans up any resources used by the collector client implementation. Client cannot be reused after this call.
	Close(ctx context.Context) error
}

// ClientFactory creates an initialized instance of a collector client based on the provided application configuration.
func ClientFactory(ctx context.Context, cfg *config.KubehoundConfig) (CollectorClient, error) {
	return nil, globals.ErrNotImplemented
}
