package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/globals/types"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func addK8sObjToTypeList(_ context.Context, object any, objectList any) error {
	switch o := objectList.(type) {
	case *rbacv1.ClusterRoleList:
		val, ok := object.(types.ClusterRoleType)
		if !ok {
			return fmt.Errorf("failed to cast object to ClusterRoleType: %s", reflect.TypeOf(object).String())
		}
		o.Items = append(o.Items, *val)
	case *rbacv1.ClusterRoleBindingList:
		val, ok := object.(types.ClusterRoleBindingType)
		if !ok {
			return fmt.Errorf("failed to cast object to ClusterRoleBindingType: %s", reflect.TypeOf(object).String())
		}
		o.Items = append(o.Items, *val)
	case *rbacv1.RoleList:
		val, ok := object.(types.RoleType)
		if !ok {
			return fmt.Errorf("failed to cast object to RoleType: %s", reflect.TypeOf(object).String())
		}
		o.Items = append(o.Items, *val)
	case *rbacv1.RoleBindingList:
		val, ok := object.(types.RoleBindingType)
		if !ok {
			return fmt.Errorf("failed to cast object to RoleBindingType: %s", reflect.TypeOf(object).String())

		}
		o.Items = append(o.Items, *val)
	case *discoveryv1.EndpointSliceList:
		val, ok := object.(types.EndpointType)
		if !ok {
			return fmt.Errorf("failed to cast object to EndpointType: %s", reflect.TypeOf(object).String())
		}
		o.Items = append(o.Items, *val)
	case *corev1.PodList:
		val, ok := object.(types.PodType)
		if !ok {
			return fmt.Errorf("failed to cast object to PodType: %s", reflect.TypeOf(object).String())
		}
		o.Items = append(o.Items, *val)
	case *corev1.NodeList:
		val, ok := object.(types.NodeType)
		if !ok {
			return fmt.Errorf("failed to cast object to NodeType: %s", reflect.TypeOf(object).String())
		}
		o.Items = append(o.Items, *val)
	default:
		return fmt.Errorf("unknown object type to cast: %s", reflect.TypeOf(object).String())
	}

	return nil

}

func bufferObject[T any, V any](ctx context.Context, filePath string, buffer map[string]*T, node V) error {
	_, ok := buffer[filePath]
	if !ok {
		var newBuffer T
		buffer[filePath] = &newBuffer
	}

	return addK8sObjToTypeList(ctx, node, buffer[filePath])
}

func dumpObj[T any](ctx context.Context, buffer map[string]T, writer writer.DumperWriter) error {
	for path, buf := range buffer {
		jsonData, err := json.Marshal(buf)
		if err != nil {
			return fmt.Errorf("failed to marshal Kubernetes object: %w", err)
		}

		err = writer.Write(ctx, jsonData, path)
		if err != nil {
			return fmt.Errorf("write %s: %w", reflect.TypeOf(buf), err)
		}
		delete(buffer, path)
	}

	return nil
}
