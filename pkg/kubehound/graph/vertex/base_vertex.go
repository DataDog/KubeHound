package vertex

import (
	"github.com/DataDog/KubeHound/pkg/config"
)

type BaseVertex struct {
	cfg *config.VertexBuilderConfig
}

func (v *BaseVertex) Initialize(cfg *config.VertexBuilderConfig) error {
	v.cfg = cfg
	return nil
}

func (v *BaseVertex) BatchSize() int {
	return v.cfg.BatchSize
}
