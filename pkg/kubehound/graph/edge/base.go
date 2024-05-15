package edge

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
)

type BaseEdge struct {
	cfg     *config.EdgeBuilderConfig
	runtime *config.DynamicConfig
}

func (e *BaseEdge) Initialize(cfg *config.EdgeBuilderConfig, runtime *config.DynamicConfig) error {
	e.cfg = cfg
	e.runtime = runtime

	return nil
}

func (e *BaseEdge) BatchSize() int {
	return e.cfg.BatchSize
}

func (e *BaseEdge) Traversal() types.EdgeTraversal {
	return adapter.DefaultEdgeTraversal()
}
