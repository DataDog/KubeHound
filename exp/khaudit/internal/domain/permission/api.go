package permission

import (
	"context"
	"errors"
)

// ErrNoResult is the error returned when no result is found.
var ErrNoResult = errors.New("no result found")

// Reader is the interface for the permission domain.
type Reader interface {
	// GetReachablePodCountPerNamespace returns the count of reachable pods per namespace.
	GetReachablePodCountPerNamespace(ctx context.Context, runID string) (map[string]int64, error)
	// GetKubectlExecutablePodCount returns the count of kubectl executable pods per namespace.
	GetKubectlExecutablePodCount(ctx context.Context, runID string, groupName string) (int64, error)
	// GetExposedPodCountPerNamespace returns the count of exposed pods per namespace.
	GetExposedPodCountPerNamespace(ctx context.Context, runID string, groupName string) ([]ExposedPodCount, error)
	// GetKubectlExecutableGroupsForNamespace returns the groups that have kubectl executable pods in a namespace.
	GetKubectlExecutableGroupsForNamespace(ctx context.Context, runID string, namespace string) ([]string, error)
	// GetExposedNamespacePods returns the pods that are exposed to a group in a namespace.
	GetExposedNamespacePods(ctx context.Context, runID string, namespace string, groupName string, filter ExposedPodFilter) ([]ExposedPodNamespace, error)
}

// ExposedPodCount is the count of exposed pods per namespace.
type ExposedPodCount struct {
	GroupName string `json:"group_name"`
	Namespace string `json:"namespace"`
	PodCount  int64  `json:"pod_count"`
}

// ExposedPodNamespace is the namespace of an exposed pod.
type ExposedPodNamespace struct {
	Namespace string `json:"namespace"`
	PodName   string `json:"pod_name"`
	Image     string `json:"image"`
	App       string `json:"app"`
	Team      string `json:"team"`
}

// ExposedPodFilter is the filter for exposed pods.
type ExposedPodFilter struct {
	Image *string `json:"image"`
	App   *string `json:"app"`
	Team  *string `json:"team"`
}
