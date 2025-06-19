package gremlin

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/permission"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type permissionRepository struct {
	conn Connection
}

// Permissions creates a new Permission repository.
func Permissions(conn Connection) permission.Reader {
	return &permissionRepository{
		conn: conn,
	}
}

// -----------------------------------------------------------------------------

func (r *permissionRepository) GetReachablePodCountPerNamespace(ctx context.Context, runID string) (map[string]int64, error) {
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		query := g.V().Has("class", "PermissionSet").Has("runID", runID).
			Project("namespace", "podPatchExecCount").
			By("namespace").
			By(gremlingo.T__.OutE("POD_EXEC", "POD_PATCH").Count()).
			Group().By("namespace").By(gremlingo.T__.Select("podPatchExecCount").Sum()).
			Order(gremlingo.Scope.Local).By(gremlingo.Column.Values, gremlingo.Order.Desc).
			Unfold()

		return query.ToList()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	// Iterate over the results.
	namespaceCounts := make(map[string]int64)
	for _, row := range results {
		obj := row.GetInterface()

		// Decode as map[any]any.
		m, ok := obj.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", obj)
		}

		for k, v := range m {
			namespace, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected result type: %T", obj)
			}

			count, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("unexpected result type: %T", obj)
			}

			// Handle context cancellation.
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				namespaceCounts[namespace] = count
			}
		}
	}

	return namespaceCounts, nil
}

func (r *permissionRepository) GetKubectlExecutablePodCount(_ context.Context, runID string, groupName string) (int64, error) {
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		query := g.V().Has("class", "Identity").Has("runID", runID).
			Has("type", "Group").Has("name", groupName).
			OutE("PERMISSION_DISCOVER").InV().
			OutE("POD_EXEC").InV().
			Dedup().Count()

		return query.ToList()
	})
	if err != nil {
		return 0, fmt.Errorf("unable to execute query: %w", err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("no results found")
	}

	return results[0].GetInt64()
}

func (r *permissionRepository) GetExposedPodCountPerNamespace(ctx context.Context, runID string, groupName string) ([]permission.ExposedPodCount, error) {
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		query := g.V().Has("class", "Identity").Has("runID", runID).
			Has("type", "Group").Has("name", groupName).As("group").
			OutE("PERMISSION_DISCOVER").InV().
			Where(gremlingo.T__.OutE("POD_EXEC").InV().Dedup().Count().Is(gremlingo.P.Gt(0))).
			Dedup().
			Project("groupName", "namespace", "podCount").
			By(gremlingo.T__.Select("group").Values("name")).
			By("namespace").
			By(gremlingo.T__.OutE("POD_EXEC").InV().Dedup().Count()).
			Order().By("groupName").By("namespace").By("podCount").
			Dedup()

		return query.ToList()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	exposedPodCounts := make([]permission.ExposedPodCount, 0)
	for _, result := range results {
		obj := result.GetInterface()
		m, ok := obj.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", obj)
		}

		var model permission.ExposedPodCount
		for k, v := range m {
			switch k {
			case "groupName":
				model.GroupName = v.(string)
			case "namespace":
				model.Namespace = v.(string)
			case "podCount":
				model.PodCount = v.(int64)
			}
		}

		// Handle context cancellation.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			exposedPodCounts = append(exposedPodCounts, model)
		}
	}

	return exposedPodCounts, nil
}

func (r *permissionRepository) GetExposedNamespacePods(ctx context.Context, runID string, namespace string, groupName string, filter permission.ExposedPodFilter) ([]permission.ExposedPodNamespace, error) {
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		query := g.V().Has("class", "Identity").Has("runID", runID).
			Has("type", "Group").Has("name", groupName).As("group").
			OutE("PERMISSION_DISCOVER").InV().
			OutE("POD_EXEC").InV().
			OutE("CONTAINER_ATTACH").InV().
			Has("class", "Container").
			Has("namespace", namespace).
			// Limit to 101 to check if there are more than 100.
			Limit(101)

		// Apply filters.
		if filter.Image != nil {
			if *filter.Image != "" {
				query = query.Has("image", *filter.Image)
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

		query = query.Dedup().
			Project("namespace", "podName", "image", "app", "team").
			By("namespace").
			By("pod").
			By("image").
			By("app").
			By("team").
			Order().By("namespace").
			Dedup()

		return query.ToList()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	exposedPodNamespaces := make([]permission.ExposedPodNamespace, 0)
	for _, result := range results {
		obj := result.GetInterface()
		m, ok := obj.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", obj)
		}

		var model permission.ExposedPodNamespace
		for k, v := range m {
			switch k {
			case "namespace":
				model.Namespace = v.(string)
			case "podName":
				model.PodName = v.(string)
			case "image":
				model.Image = v.(string)
			case "app":
				model.App = v.(string)
			case "team":
				model.Team = v.(string)
			}
		}

		// Handle context cancellation.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			exposedPodNamespaces = append(exposedPodNamespaces, model)
		}
	}

	return exposedPodNamespaces, nil
}

func (r *permissionRepository) GetKubectlExecutableGroupsForNamespace(ctx context.Context, runID string, namespace string) ([]string, error) {
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		query := g.V().Has("class", "Pod").Has("runID", runID).
			Has("namespace", namespace).
			InE("POD_EXEC").OutV().
			Has("class", "PermissionSet").
			InE("PERMISSION_DISCOVER").OutV().
			Has("class", "Identity").
			Has("type", "Group").
			Dedup().
			Project("groupName").
			By("name")

		return query.ToList()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}

	if len(results) == 0 {
		return nil, permission.ErrNoResult
	}

	groups := make([]string, 0)
	for _, result := range results {
		obj := result.GetInterface()
		m, ok := obj.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", obj)
		}

		// Handle context cancellation.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			groups = append(groups, m["groupName"].(string))
		}
	}

	return groups, nil
}
