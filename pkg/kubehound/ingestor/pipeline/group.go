package pipeline

import (
	"context"
	"sync"
)

type Group struct {
	Name    string
	Ingests []ObjectIngest
}

func (s *Group) Run(outer context.Context, deps *Dependencies) error {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(outer)

	// TODO cancel everything on error
	for _, ingest := range s.Ingests {
		i := ingest
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := i.Initialize(ctx, deps)
			if err != nil {
				// TODO
			}

			err = i.Run(ctx)
			if err != nil {
				// TOOD
			}
		}()
	}

	wg.Wait()

	return nil
}
