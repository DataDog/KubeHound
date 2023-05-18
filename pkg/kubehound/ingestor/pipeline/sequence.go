package pipeline

import (
	"context"
)

type Sequence struct {
	Name   string
	Groups []Group
}

func (s *Sequence) Run(ctx context.Context, deps *Dependencies) error {
	// Stages run in sequence
	// TODO cancel everything on error
	for _, s := range s.Groups {
		err := s.Run(ctx, deps)
		if err != nil {
			return err
		}
	}

	return nil
}
