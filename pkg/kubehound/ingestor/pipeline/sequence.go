package pipeline

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

type Sequence struct {
	Name   string
	Groups []Group
}

func (s *Sequence) Run(ctx context.Context, deps *Dependencies) error {
	l := log.Trace(ctx, log.WithComponent(globals.IngestorComponent))

	l.Infof("Starting ingest sequence %s", s.Name)
	for _, g := range s.Groups {
		l.Infof("Running ingest group %s", g.Name)
		err := g.Run(ctx, deps)
		if err != nil {
			return fmt.Errorf("group %s ingest: %w", g.Name, err)
		}
	}

	l.Infof("Completed ingest sequence %s", s.Name)
	return nil
}
