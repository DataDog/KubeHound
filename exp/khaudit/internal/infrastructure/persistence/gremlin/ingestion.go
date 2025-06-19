package gremlin

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/ingestion"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type ingestionRepository struct {
	conn Connection
}

// Ingestions creates a new Ingestion repository.
func Ingestions(conn Connection) ingestion.Reader {
	return &ingestionRepository{
		conn: conn,
	}
}

// -----------------------------------------------------------------------------

// List returns the list of ingestions.
func (r *ingestionRepository) List(ctx context.Context, filter ingestion.ListFilter) ([]ingestion.Ingestion, error) {
	// Execute the query.
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		query := g.V().
			Has("class", "Node")

		if filter.RunID != nil {
			if *filter.RunID != "" {
				query = query.Has("runID", *filter.RunID)
			}
		}
		if filter.Cluster != nil {
			if *filter.Cluster != "" {
				query = query.Has("cluster", *filter.Cluster)
			}
		}

		return query.GroupCount().By(
			gremlingo.T__.Project("cluster", "runID").
				By("cluster").By("runID")).
			Unfold().
			ToList()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}
	if len(results) == 0 {
		return nil, ingestion.ErrNoResult
	}

	// Iterate over the results.
	var ingestions []ingestion.Ingestion
	for _, row := range results {
		obj := row.GetInterface()

		// Decode as map[any]any.
		m, ok := obj.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", obj)
		}

		for k := range m {
			nodeProjectionPtr, ok := k.(*any)
			if !ok {
				return nil, fmt.Errorf("unexpected result type: %T", k)
			}
			if nodeProjectionPtr == nil {
				return nil, fmt.Errorf("unexpected nil result")
			}
			nodeProjection, ok := (*nodeProjectionPtr).(map[any]any)
			if !ok {
				return nil, fmt.Errorf("unexpected projection type: %T", *nodeProjectionPtr)
			}

			ig := ingestion.Ingestion{}
			for pk, pv := range nodeProjection {
				switch pk {
				case "cluster":
					ig.Cluster, ok = pv.(string)
					if !ok {
						return nil, fmt.Errorf("unexpected cluster type: %T", pv)
					}
				case "runID":
					ig.RunID, ok = pv.(string)
					if !ok {
						return nil, fmt.Errorf("unexpected runID type: %T", pv)
					}
				default:
					return nil, fmt.Errorf("unexpected projection key: %s", pk)
				}
			}

			// Handle context cancellation.
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				ingestions = append(ingestions, ig)
			}
		}
	}

	return ingestions, nil
}

func (r *ingestionRepository) GetEdgeCountPerClasses(ctx context.Context) (map[string]int64, error) {
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		query := g.E().GroupCount().By(gremlingo.T.Label).Unfold()
		return query.ToList()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}
	if len(results) == 0 {
		return nil, ingestion.ErrNoResult
	}

	// Iterate over the results.
	edgeCounts := make(map[string]int64)
	for _, row := range results {
		obj := row.GetInterface()

		// Decode as map[any]any.
		m, ok := obj.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", obj)
		}

		for k, v := range m {
			label, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected label type: %T", k)
			}
			count, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("unexpected count type: %T", v)
			}

			edgeCounts[label] = count
		}

		// Handle context cancellation.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	return edgeCounts, nil
}

func (r *ingestionRepository) GetVertexCountPerClasses(ctx context.Context) (map[string]int64, error) {
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		query := g.V().GroupCount().By(gremlingo.T.Label).Unfold()
		return query.ToList()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}
	if len(results) == 0 {
		return nil, ingestion.ErrNoResult
	}

	// Iterate over the results.
	vertexCounts := make(map[string]int64)
	for _, row := range results {
		obj := row.GetInterface()

		// Decode as map[any]any.
		m, ok := obj.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", obj)
		}

		for k, v := range m {
			label, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected label type: %T", k)
			}
			count, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("unexpected count type: %T", v)
			}

			vertexCounts[label] = count
		}

		// Handle context cancellation.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	return vertexCounts, nil
}
