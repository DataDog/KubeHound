package collector

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/globals/types"
)

type Complete func(context.Context) error
type NodeProcessor func(context.Context, *types.NodeType) error
type PodProcessor func(context.Context, *types.PodType) error
type RoleProcessor func(context.Context, *types.RoleType) error
type ClusterRoleProcessor func(context.Context, *types.ClusterRoleType) error
type RoleBindingProcessor func(context.Context, *types.RoleBindingType) error
type ClusterRoleBindingProcessor func(context.Context, *types.ClusterRoleBindingType) error

type CollectorClient interface {
	HealthCheck(ctx context.Context) (bool, error)
	StreamNodes(ctx context.Context, callback NodeProcessor, complete Complete) error
	StreamPods(ctx context.Context, callback PodProcessor, complete Complete) error
	StreamRoles(ctx context.Context, callback RoleProcessor, complete Complete) error
	StreamClusterRoles(ctx context.Context, callback ClusterRoleProcessor, complete Complete) error
	StreamRoleBindings(ctx context.Context, callback RoleBindingProcessor, complete Complete) error
	StreamClusterRoleBindings(ctx context.Context, callback ClusterRoleBindingProcessor, complete Complete) error
	Close(ctx context.Context) error
}

func ClientFactory(ctx context.Context, cfg *config.KubehoundConfig) (CollectorClient, error) {
	return nil, globals.ErrNotImplemented
}
