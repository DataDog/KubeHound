package pipeline

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// Group encapsulates a collection of object ingest pipelines that can be run in parallel.s
type Group struct {
	Name    string         // Name of the ingest group
	Ingests []ObjectIngest // Parallelized object ingest pipelines
}

// Run executes all the object ingest pipelines in parallel and returns when all complete.
func (g *Group) Run(outer context.Context, deps *Dependencies) error {
	ctx, cancelGroup := context.WithCancelCause(outer)
	defer cancelGroup(nil)

	l := log.Trace(ctx)
	l.Infof("Starting %s ingests", g.Name)

	// Run the group ingests in parallel and cancel all on any errors. Note we deliberately avoid
	// using a worker pool here as have a small, fixed number of tasks to run in parallel.
	wg := &sync.WaitGroup{}
	for _, ingest := range g.Ingests {
		i := ingest

		// Don't spin off a goroutine until we initialize successfully
		err := i.Initialize(ctx, deps)
		if err != nil {
			l.Errorf("%s initialization: %v", i.Name(), err)
			cancelGroup(err)

			break
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer i.Close(ctx) // defers are LIFO

			l.Infof("Running ingest %s", i.Name())

			err := i.Run(ctx)
			if err != nil {
				l.Errorf("%s run: %v", i.Name(), err)
				cancelGroup(err)
			}
		}()
	}

	l.Infof("Waiting for %s ingests to complete", g.Name)
	wg.Wait()

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	l.Infof("Completed %s ingest", g.Name)

	return nil
}
