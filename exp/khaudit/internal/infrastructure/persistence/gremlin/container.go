package gremlin

import (
	"context"
	"errors"
	"fmt"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/container"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/kubehound"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type containerRespository struct {
	conn Connection
}

// Containers returns a new instance of the container repository.
func Containers(conn Connection) container.Reader {
	return &containerRespository{
		conn: conn,
	}
}

func (r *containerRespository) CountByNamespaces(ctx context.Context, cluster, runID string, filter container.NamespaceAggregationFilter, resultChan chan<- container.NamespaceAggregation) error {
	// Check arguments.
	if cluster == "" {
		return errors.New("cluster is required")
	}
	if runID == "" {
		return errors.New("runID is required")
	}
	if resultChan == nil {
		return errors.New("resultChan is required")
	}

	// Execute the query.
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		t := g.V().
			Has("class", "Container").
			Has("runID", runID)

		// Apply filters.
		if len(filter.Namespaces) > 0 {
			// Filter by namespaces.
			t = t.Has("namespace", gremlingo.P.Within(stringArray(filter.Namespaces)...))
		} else if len(filter.ExcludedNamespaces) > 0 {
			// Filter by excluded namespaces.
			t = t.Not(gremlingo.T__.Has("namespace", gremlingo.P.Within(stringArray(filter.ExcludedNamespaces)...)))
		}

		return t.GroupCount().By("namespace").
			Order().By(gremlingo.Column.Values, gremlingo.Order.Desc).
			Unfold().
			ToList()
	})
	if err != nil {
		return fmt.Errorf("unable to execute query: %w", err)
	}
	if len(results) == 0 {
		return container.ErrNoResult
	}

	// Iterate over the results.
	for _, row := range results {
		obj := row.GetInterface()

		// Decode as map[any]any.
		m, ok := obj.(map[any]any)
		if !ok {
			return fmt.Errorf("unexpected result type: %T", obj)
		}

		for k, v := range m {
			ns, ok := k.(string)
			if !ok {
				return fmt.Errorf("unexpected namespace type: %T", k)
			}

			count, ok := v.(int64)
			if !ok {
				return fmt.Errorf("unexpected count type: %T", v)
			}

			// Send the result to the channel.
			select {
			case <-ctx.Done():
				return nil
			case resultChan <- container.NamespaceAggregation{
				RunID:     runID,
				Cluster:   cluster,
				Namespace: ns,
				Count:     count,
			}:
			}
		}
	}

	return nil
}

func (r *containerRespository) GetAttackPathProfiles(_ context.Context, cluster, runID string, filter container.AttackPathFilter) ([]container.AttackPath, error) {
	// Check arguments.
	if runID == "" {
		return nil, errors.New("runID is required")
	}

	// Execute the query.
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		t := g.V().
			Has("class", "Container").
			Has("runID", runID)

		// Apply filters.
		if filter.Namespace != nil {
			// Filter by namespace.
			t = t.Has("namespace", *filter.Namespace)
		} else if len(filter.ExcludedNamespaces) > 0 {
			// Filter by excluded namespaces.
			t = t.Not(gremlingo.T__.Has("namespace", gremlingo.P.Within(stringArray(filter.ExcludedNamespaces)...)))
		}
		if filter.Image != nil {
			// Filter by image.
			t = t.Has("image", *filter.Image)
		}
		if filter.App != nil {
			// Filter by app.
			t = t.Has("app", *filter.App)
		}
		if filter.Team != nil {
			// Filter by team.
			t = t.Has("team", *filter.Team)
		}

		// Default target class is "Node" for container escape.
		targetClass := "Node"
		if filter.TargetClass != nil {
			targetClass = *filter.TargetClass
		}

		// Default time limit is 3000.
		timeLimit := int64(3000)
		if filter.TimeLimit != nil {
			timeLimit = *filter.TimeLimit
		}

		// Return the traversal.
		return t.
			Repeat(gremlingo.T__.OutE().InV().SimplePath().TimeLimit(timeLimit)).
			Until(gremlingo.T__.Has("class", targetClass).Or().Loops().Is(10)).
			Has("class", targetClass).
			GroupCount().By(gremlingo.T__.Path().By(gremlingo.T__.Label())).
			Unfold().
			ToList()
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}
	if len(results) == 0 {
		return nil, container.ErrNoResult
	}

	attackPaths := make([]container.AttackPath, 0)
	for _, row := range results {
		obj := row.GetInterface()

		// Decode as map[any]any.
		m, ok := obj.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("unexpected result type: %T", obj)
		}

		for k, v := range m {
			// Decode the path.
			path, ok := k.(*gremlingo.Path)
			if !ok {
				return nil, fmt.Errorf("unexpected path type: %T", k)
			}

			// Decode the count.
			count, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("unexpected count type: %T", v)
			}

			// Build the attack path.
			atp := container.AttackPath{
				ContainerCount: count,
				Cluster:        cluster,
				RunID:          runID,
				Path:           make([]string, len(path.Objects)),
			}

			// Append the steps.
			for i, step := range path.Objects {
				atp.Path[i] = fmt.Sprintf("%v", step)
			}

			// Append the attack path.
			attackPaths = append(attackPaths, atp)
		}
	}

	return attackPaths, nil
}

func (r *containerRespository) GetVulnerables(ctx context.Context, cluster, runID string, filter container.AttackPathFilter, resultChan chan<- container.Container) error {
	// Check arguments.
	if runID == "" {
		return errors.New("runID is required")
	}
	if cluster == "" {
		return errors.New("cluster is required")
	}

	// Execute the query.
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		t := g.V().
			Has("class", "Container").
			Has("runID", runID)

		// Apply filters.
		if filter.Namespace != nil {
			// Filter by namespace.
			t = t.Has("namespace", *filter.Namespace)
		} else if len(filter.ExcludedNamespaces) > 0 {
			// Filter by excluded namespaces.
			t = t.Not(gremlingo.T__.Has("namespace", gremlingo.P.Within(stringArray(filter.ExcludedNamespaces)...)))
		}
		if filter.Image != nil {
			// Filter by image.
			t = t.Has("image", *filter.Image)
		}
		if filter.App != nil {
			// Filter by app.
			t = t.Has("app", *filter.App)
		}
		if filter.Team != nil {
			// Filter by team.
			t = t.Has("team", *filter.Team)
		}

		// Default target class is "Node" for container escape.
		targetClass := "Node"
		if filter.TargetClass != nil {
			targetClass = *filter.TargetClass
		}

		// Default time limit is 3000.
		timeLimit := int64(3000)
		if filter.TimeLimit != nil {
			timeLimit = *filter.TimeLimit
		}

		// Return the traversal.
		return t.
			Where(
				gremlingo.T__.Repeat(
					gremlingo.T__.OutE().InV().SimplePath().TimeLimit(timeLimit),
				).
					Until(gremlingo.T__.Has("class", targetClass).Or().Loops().Is(10)).
					Has("class", targetClass).
					Limit(1),
			).
			Dedup().By("image").
			ValueMap("namespace", "app", "team", "image").
			ToList()
	})
	if err != nil {
		return fmt.Errorf("unable to execute query: %w", err)
	}
	if len(results) == 0 {
		return container.ErrNoResult
	}

	for _, row := range results {
		obj := row.GetInterface()

		// Decode as map[any]any.
		m, ok := obj.(map[any]any)
		if !ok {
			return fmt.Errorf("unexpected result type: %T", obj)
		}

		c := container.Container{
			Cluster: cluster,
			RunID:   runID,
		}
		for k, v := range m {
			switch k {
			case "namespace":
				values, ok := v.([]any)
				if !ok {
					return fmt.Errorf("unexpected namespace type: %T", v)
				}
				if len(values) == 0 {
					continue
				}
				c.Namespace, ok = values[0].(string)
				if !ok {
					return fmt.Errorf("unexpected namespace type: %T", values[0])
				}
			case "app":
				values, ok := v.([]any)
				if !ok {
					return fmt.Errorf("unexpected app type: %T", v)
				}
				if len(values) == 0 {
					continue
				}
				c.App, ok = values[0].(string)
				if !ok {
					return fmt.Errorf("unexpected app type: %T", values[0])
				}
			case "team":
				values, ok := v.([]any)
				if !ok {
					return fmt.Errorf("unexpected team type: %T", v)
				}
				if len(values) == 0 {
					continue
				}
				c.Team, ok = values[0].(string)
				if !ok {
					return fmt.Errorf("unexpected team type: %T", values[0])
				}
			case "image":
				values, ok := v.([]any)
				if !ok {
					return fmt.Errorf("unexpected image type: %T", v)
				}
				if len(values) == 0 {
					continue
				}
				c.Image, ok = values[0].(string)
				if !ok {
					return fmt.Errorf("unexpected image type: %T", values[0])
				}
			default:
				return fmt.Errorf("unexpected key: %v", k)
			}
		}

		// Send the result to the channel.
		select {
		case <-ctx.Done():
			return nil
		case resultChan <- c:
		}
	}

	return nil
}

func (r *containerRespository) GetAttackPaths(ctx context.Context, cluster, runID string, filter container.AttackPathFilter, resultChan chan<- kubehound.AttackPath) error {
	// Check arguments.
	if runID == "" {
		return errors.New("runID is required")
	}
	if cluster == "" {
		return errors.New("cluster is required")
	}
	if resultChan == nil {
		return errors.New("resultChan is required")
	}

	// Execute the query.
	results, err := r.conn.Query(func(g *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
		t := g.V().
			Has("class", "Container").
			Has("runID", runID)

		// Apply filters.
		if filter.Namespace != nil {
			// Filter by namespace.
			t = t.Has("namespace", *filter.Namespace)
		} else if len(filter.ExcludedNamespaces) > 0 {
			// Filter by excluded namespaces.
			t = t.Not(gremlingo.T__.Has("namespace", gremlingo.P.Within(stringArray(filter.ExcludedNamespaces)...)))
		}
		if filter.Image != nil {
			// Filter by image.
			t = t.Has("image", *filter.Image)
		}
		if filter.App != nil {
			// Filter by app.
			t = t.Has("app", *filter.App)
		}
		if filter.Team != nil {
			// Filter by team.
			t = t.Has("team", *filter.Team)
		}

		// Default target class is "Node" for container escape.
		targetClass := "Node"
		if filter.TargetClass != nil {
			targetClass = *filter.TargetClass
		}

		// Default time limit is 5000.
		timeLimit := int64(5000)
		if filter.TimeLimit != nil {
			timeLimit = *filter.TimeLimit
		}

		// Return the traversal.
		return t.
			Repeat(gremlingo.T__.OutE().InV().SimplePath().TimeLimit(timeLimit)).
			Until(gremlingo.T__.Has("class", targetClass).Or().Loops().Is(10)).
			Path().By(gremlingo.T__.ElementMap()).
			ToList()
	})
	if err != nil {
		return fmt.Errorf("unable to execute query: %w", err)
	}
	if len(results) == 0 {
		return container.ErrNoResult
	}

	for _, row := range results {
		obj := row.GetInterface()

		// Decode as *gremlingo.Path.
		path, ok := obj.(*gremlingo.Path)
		if !ok {
			return fmt.Errorf("unexpected result type: %T", obj)
		}

		var tuples []kubehound.HexTuple
		for _, step := range path.Objects {
			stepMap, ok := step.(map[any]any)
			if !ok {
				return fmt.Errorf("unexpected step type: %T", step)
			}

			label, ok := stepMap["label"].(string)
			if !ok {
				return fmt.Errorf("unexpected label type: %T", stepMap["label"])
			}

			switch label {
			case "Container", "Node", "Endpoint", "Group", "Identity", "PermissionSet", "Pod", "Volume":
				// Extract the subject ID.
				subjectID, ok := stepMap["id"].(int64)
				if !ok {
					return fmt.Errorf("unexpected subject ID type: %T", stepMap["id"])
				}

				for k, v := range stepMap {
					switch k {
					case "id":
						// Skip the subject ID.
						continue
					default:
						tuples = append(tuples, kubehound.HexTuple{
							Subject:   "urn:vertex:" + fmt.Sprintf("%d", subjectID),
							Predicate: fmt.Sprintf("urn:property:%v", k),
							Value:     fmt.Sprintf("%v", v),
							DataType:  fmt.Sprintf("%T", v),
							Language:  "",
							Graph:     "",
						})
					}
				}

			case "CONTAINER_ATTACH", "ENDPOINT_EXPLOIT", "CE_MODULE_LOAD", "CE_NSENTER",
				"CE_PRIV_MOUNT", "CE_SYS_PTRACE", "CE_UMH_CORE_PATTERN", "CE_VAR_LOG_SYMLINK",
				"EXPLOIT_HOST_READ", "EXPLOIT_HOST_TRAVERSE", "EXPLOIT_HOST_WRITE",
				"IDENTITY_ASSUME", "PERMISSION_DISCOVER", "POD_ATTACH", "POD_CREATE",
				"POD_EXEC", "POD_PATCH", "ROLE_BIND", "SHARE_PS_NAMESPACE", "TOKEN_BRUTEFORCE",
				"TOKEN_LIST", "TOKEN_STEAL", "VOLUME_ACCESS", "VOLUME_DISCOVER":
				// Extract the subject ID.
				relationKey, ok := stepMap["id"].(map[string]any)
				if !ok {
					return fmt.Errorf("unexpected subject ID type: %T", stepMap["id"])
				}

				subjectID, ok := relationKey["relationId"]
				if !ok {
					return fmt.Errorf("relationId is not found")
				}

				for k, v := range stepMap {
					switch k {
					case "id":
						// Skip the subject ID.
						continue
					case "IN":
						// Manually add the IN relation.
						refMap, ok := v.(map[any]any)
						if !ok {
							return fmt.Errorf("unexpected IN type: %T", v)
						}

						// Add the IN relation.
						tuples = append(tuples, kubehound.HexTuple{
							Subject:   "urn:edge:" + fmt.Sprintf("%v", subjectID),
							Predicate: "urn:property:in",
							Value:     "urn:vertex:" + fmt.Sprintf("%v", refMap["id"]),
							DataType:  "",
							Language:  "",
							Graph:     "",
						})
					case "OUT":
						// Manually add the OUT relation.
						refMap, ok := v.(map[any]any)
						if !ok {
							return fmt.Errorf("unexpected OUT type: %T", v)
						}

						// Add the OUT relation.
						tuples = append(tuples, kubehound.HexTuple{
							Subject:   "urn:edge:" + fmt.Sprintf("%v", subjectID),
							Predicate: "urn:property:out",
							Value:     "urn:vertex:" + fmt.Sprintf("%v", refMap["id"]),
							DataType:  "",
							Language:  "",
							Graph:     "",
						})
					default:
						tuples = append(tuples, kubehound.HexTuple{
							Subject:   "urn:edge:" + fmt.Sprintf("%v", subjectID),
							Predicate: fmt.Sprintf("urn:property:%v", k),
							Value:     fmt.Sprintf("%v", v),
							DataType:  fmt.Sprintf("%T", v),
							Language:  "",
							Graph:     "",
						})
					}
				}
			default:
				return fmt.Errorf("unexpected label: %v", label)
			}
		}

		// Send the result to the channel.
		select {
		case <-ctx.Done():
			return nil
		case resultChan <- kubehound.AttackPath(tuples):
		}
	}

	return nil
}
