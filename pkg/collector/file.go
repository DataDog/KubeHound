package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// Expect a file structure of the following
// |____<namespace>
// | |____rolebindings.rbac.authorization.k8s.io.json
// | |____pods.json
// | |____roles.rbac.authorization.k8s.io.json
// |____<namespace>
// | |____rolebindings.rbac.authorization.k8s.io.json
// | |____pods.json
// | |____roles.rbac.authorization.k8s.io.json
// |____nodes.json
// |____clusterroles.rbac.authorization.k8s.io.json
// |____clusterrolebindings.rbac.authorization.k8s.io.json
const (
	nodePath                = "nodes.json"
	clusterRolesPath        = "clusterroles.rbac.authorization.k8s.io.json"
	clusterRoleBindingsPath = "clusterrolebindings.rbac.authorization.k8s.io.json"
	podPath                 = "pods.json"
	rolesPath               = "roles.rbac.authorization.k8s.io.json"
	roleBindingsPath        = "rolebindings.rbac.authorization.k8s.io.json"
)

const (
	FileCollectorName = "local-file-collector"
)

type FileCollector struct {
	basePath string
}

func NewFile(cfg *config.KubehoundConfig) (CollectorClient, error) {
	file, err := os.Stat(basePath)
	if err != nil {
		return nil, fmt.Errorf("error querying base path: %w", err)
	}

	if !file.IsDir() {
		return nil, fmt.Errorf("base path must be a directory")
	}

	return &FileCollector{
		basePath: basePath,
	}, nil
}

func (c *FileCollector) Name() string {
	return FileCollectorName
}

func (c *FileCollector) HealthCheck(ctx context.Context) (bool, error) {
	// CHeck dir exists
	// CHeck access??
	return true, nil
}

func (c *FileCollector) Close(ctx context.Context) error {
	// NOP for this implementation
	return nil
}

func (c *FileCollector) loadPodsNamespace(ctx context.Context, fp string, callback PodProcessor) error {
	bytes, err := os.ReadFile(fp)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", fp, err)
	}

	if len(bytes) == 0 {
		return nil
	}

	var podList corev1.PodList
	err = json.Unmarshal(bytes, &podList)
	if err != nil {
		return fmt.Errorf("error unmarshalling PodList JSON: %v", err)
	}

	for _, p := range podList.Items {
		err = callback(ctx, &p)
		if err != nil {
			return fmt.Errorf("error processing pod %s::%s: %w", p.Namespace, p.Name, err)
		}
	}

	return nil
}

func (c *FileCollector) LoadPods(ctx context.Context, callback PodProcessor, complete Complete) error {
	err := filepath.WalkDir(i.basePath, func(path string, d fs.DirEntry, err error) error {
		if path == i.basePath || !d.IsDir() {
			// Skip files
			return nil
		}

		f := filepath.Join(path, podPath)
		log.I.Debugf("Ingesting pods from file %s", f)

		return i.loadPodsNamespace(ctx, f, callback)
	})

	if err != nil {
		return err
	}

	return complete(ctx)
}

func (c *FileCollector) StreamRoles(ctx context.Context, callback RoleProcessor, complete Complete) error {
	err := filepath.WalkDir(i.basePath, func(path string, d fs.DirEntry, err error) error {
		if path == i.basePath || !d.IsDir() {
			// Skip files
			return nil
		}

		f := filepath.Join(path, rolesPath)
		log.I.Debugf("Ingesting roles from file %s", f)

		return i.loadRolesNamespace(ctx, f, callback)
	})

	if err != nil {
		return err
	}

	return complete(ctx)
}

func (c *FileCollector) loadRolesNamespace(ctx context.Context, fp string, callback RoleProcessor) error {
	bytes, err := os.ReadFile(fp)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", fp, err)
	}

	if len(bytes) == 0 {
		return nil
	}

	var roleList rbacv1.RoleList
	err = json.Unmarshal(bytes, &roleList)
	if err != nil {
		return fmt.Errorf("error unmarshalling RoleList JSON: %v", err)
	}

	for _, r := range roleList.Items {
		err = callback(ctx, &r)
		if err != nil {
			return fmt.Errorf("error processing Role %s::%s: %w", r.Namespace, r.Name, err)
		}
	}

	return nil
}

func (c *FileCollector) loadRoleBindingsNamespace(ctx context.Context, fp string, callback RoleBindingProcessor) error {
	bytes, err := os.ReadFile(fp)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", fp, err)
	}

	if len(bytes) == 0 {
		return nil
	}

	var roleBindingList rbacv1.RoleBindingList
	err = json.Unmarshal(bytes, &roleBindingList)
	if err != nil {
		return fmt.Errorf("error unmarshalling RoleBindingList JSON: %v", err)
	}

	for _, r := range roleBindingList.Items {
		err = callback(ctx, &r)
		if err != nil {
			return fmt.Errorf("error processing Rolebinding %s::%s: %w", r.Namespace, r.Name, err)
		}
	}

	return nil
}

func (c *FileCollector) StreamRoleBindings(ctx context.Context, callback RoleBindingProcessor, complete Complete) error {
	err := filepath.WalkDir(c.basePath, func(path string, d fs.DirEntry, err error) error {
		if path == c.basePath || !d.IsDir() {
			// Skip files
			return nil
		}

		f := filepath.Join(path, roleBindingsPath)
		log.I.Debugf("Collecting rolebindings from file %s", f)

		return c.loadRoleBindingsNamespace(ctx, f, callback)
	})

	if err != nil {
		return err
	}

	return complete(ctx)
}

func (c *FileCollector) StreamNodes(ctx context.Context, callback NodeProcessor, complete Complete) error {
	path := filepath.Join(c.basePath, nodePath)
	list, err := readList[corev1.NodeList](ctx, path)

	for _, item := range list.Items {
		i := types.NodeType(&item)
		err = callback(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing K8s node %s::%s: %w", i.Namespace, i.Name, err)
		}
	}

	return complete(ctx)
}

func (c *FileCollector) StreamClusterRoles(ctx context.Context, callback ClusterRoleProcessor, complete Complete) error {
	path := filepath.Join(c.basePath, clusterRolesPath)
	list, err := readList[rbacv1.ClusterRoleList](ctx, path)

	for _, item := range list.Items {
		i := types.ClusterRoleType(&item)
		err = callback(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing k8s cluster role %s: %w", i.Name, err)
		}
	}

	return complete(ctx)
}

func (c *FileCollector) StreamClusterRoleBindings(ctx context.Context, callback ClusterRoleBindingProcessor, complete Complete) error {
	path := filepath.Join(c.basePath, clusterRoleBindingsPath)
	list, err := readList[rbacv1.ClusterRoleBindingList](ctx, path)

	for _, item := range list.Items {
		i := types.ClusterRoleBindingType(&item)
		err = callback(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing K8s cluster role binding %s: %w", i.Name, err)
		}
	}

	return complete(ctx)
}

list, err := readList[rbacv1.ClusterRoleBindingList](ctx, path)

	for _, item := range list.Items {
		i := types.ClusterRoleBindingType(&item)
		err = callback(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing K8s cluster role binding %s: %w", i.Name, err)
		}
	}

func streamObjectsNamespace[] {
	list, err := readList[rbacv1.ClusterRoleBindingList](ctx, path)

	for _, item := range list.Items {
		i := types.ClusterRoleBindingType(&item)
		err = callback(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing K8s cluster role binding %s: %w", i.Name, err)
		}
	}


}

func (c *FileCollector) streamDir(ctx context.Context, callback RoleProcessor, complete Complete) error {
	err := filepath.WalkDir(c.basePath, func(path string, d fs.DirEntry, err error) error {
		if path == c.basePath || !d.IsDir() {
			// Skip files
			return nil
		}

		f := filepath.Join(path, podPath)
		log.I.Debugf("Collecting pods from file %s", f)

		return .loadPodsNamespace(ctx, f, callback)
	})

	if err != nil {
		return err
	}

	return complete(ctx)
}

// This implementation reads the entire array into memory at once.
func readList[Tl types.ListInputType](ctx context.Context, inputPath string) (Tl, error) {

	var inputList Tl
	bytes, err := os.ReadFile(inputPath)
	if err != nil {
		return inputList, fmt.Errorf("read file %s: %v", inputPath, err)
	}

	if len(bytes) == 0 {
		return inputList, nil
	}

	err = json.Unmarshal(bytes, &inputList)
	if err != nil {
		return inputList, fmt.Errorf("unmarshalling %T JSON: %w", inputList, err)
	}

	return inputList, nil
}
