package gremlin

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/volume"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type volumeRepository struct {
	conn Connection
}

// Volumes creates a new Volume repository.
func Volumes(conn Connection) volume.Reader {
	return &volumeRepository{
		conn: conn,
	}
}

// -----------------------------------------------------------------------------

func (r *volumeRepository) GetVolumes(ctx context.Context, runID string, filter volume.Filter) ([]volume.Volume, error) {
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		query := g.V().Has("class", "Volume").
			Has("runID", runID)

		// Apply query filters.
		if filter.Namespace != nil {
			if *filter.Namespace != "" {
				query = query.Has("namespace", *filter.Namespace)
			}
		}
		if filter.Type != nil {
			if *filter.Type != "" {
				query = query.Has("type", *filter.Type)
			}
		}
		if filter.App != nil {
			if *filter.App != "" {
				query = query.Has("app", *filter.App)
			}
		}
		if filter.Team != nil {
			if *filter.Team != "" {
				query = query.Has("team", *filter.Team)
			}
		}

		return query.ElementMap().ToList()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	// Iterate over the results.
	volumes := make([]volume.Volume, 0)
	for _, result := range results {
		obj := result.GetInterface()

		// Decode as map[any]any.
		m, ok := obj.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", obj)
		}

		entity := volume.Volume{}
		for k, v := range m {
			switch k {
			case "name":
				entity.Name = v.(string)
			case "type":
				entity.Type = v.(string)
			case "namespace":
				entity.Namespace = v.(string)
			case "app":
				entity.App = v.(string)
			case "team":
				entity.Team = v.(string)
			case "sourcePath":
				entity.SourcePath = v.(string)
			case "mountPath":
				entity.MountPath = v.(string)
			case "readOnly":
				entity.ReadOnly = v.(bool)
			}
		}

		// Handle context cancellation.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			volumes = append(volumes, entity)
		}
	}

	return volumes, nil
}

func (r *volumeRepository) GetMountedHostPaths(ctx context.Context, runID string, filter volume.Filter) ([]volume.MountedHostPath, error) {
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		query := g.V().Has("class", "Volume").
			Has("runID", runID).
			Has("type", "HostPath")

		// Apply query filters.
		if filter.SourcePath != nil {
			if *filter.SourcePath != "" {
				query = query.Has("sourcePath", *filter.SourcePath)
			}
		}
		if filter.Namespace != nil {
			if *filter.Namespace != "" {
				query = query.Has("namespace", *filter.Namespace)
			}
		}
		if filter.App != nil {
			if *filter.App != "" {
				query = query.Has("app", *filter.App)
			}
		}
		if filter.Team != nil {
			if *filter.Team != "" {
				query = query.Has("team", *filter.Team)
			}
		}

		// Set a volume alias.
		query = query.As("v")

		// Get associated container.
		query = query.InE("VOLUME_DISCOVER").OutV().Has("class", "Container")

		// Apply image filter.
		if filter.Image != nil {
			if *filter.Image != "" {
				query = query.Has("image", *filter.Image)
			}
		}

		// Alias the container.
		query = query.As("c")

		return query.Project("sourcePath", "namespace", "app", "team", "image").
			By(gremlingo.T__.Select("v").Values("sourcePath")).
			By(gremlingo.T__.Select("v").Values("namespace")).
			By(gremlingo.T__.Select("c").Values("app")).
			By(gremlingo.T__.Select("c").Values("team")).
			By(gremlingo.T__.Select("c").Values("image")).
			Dedup().
			Order().By("sourcePath").
			ToList()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}
	if len(results) == 0 {
		return nil, volume.ErrNoResult
	}

	// Iterate over the results.
	mountedHostPaths := make([]volume.MountedHostPath, 0)
	for _, result := range results {
		obj := result.GetInterface()

		// Decode as map[any]any.
		m, ok := obj.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", obj)
		}

		entity := volume.MountedHostPath{}

		// Get the source path
		for k, v := range m {
			switch k {
			case "sourcePath":
				entity.SourcePath = v.(string)
			case "namespace":
				entity.Namespace = v.(string)
			case "app":
				entity.App = v.(string)
			case "team":
				entity.Team = v.(string)
			case "image":
				entity.Image = v.(string)
			}
		}

		// Handle context cancellation.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			mountedHostPaths = append(mountedHostPaths, entity)
		}
	}

	return mountedHostPaths, nil
}
