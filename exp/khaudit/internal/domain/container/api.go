package container

import (
	"context"
	"errors"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/kubehound"
)

var (
	// ErrNoResult is returned when no result is found.
	ErrNoResult = errors.New("no result found")
)

// Reader is the interface for reading container data.
type Reader interface {
	// CountByNamespaces returns the count of containers by namespace.
	CountByNamespaces(ctx context.Context, cluster, runID string, filter NamespaceAggregationFilter, resultChan chan<- NamespaceAggregation) error
	// GetAttackPathProfiles returns the attack path profiles.
	GetAttackPathProfiles(ctx context.Context, cluster, runID string, filter AttackPathFilter) ([]AttackPath, error)
	// GetVulnerables returns the vulnerable containers.
	GetVulnerables(ctx context.Context, cluster, runID string, filter AttackPathFilter, resultChan chan<- Container) error
	// GetAttackPaths returns the containers with attack paths.
	GetAttackPaths(ctx context.Context, cluster, runID string, filter AttackPathFilter, resultChan chan<- kubehound.AttackPath) error
}

// NamespaceAggregationFilter represents the filter for namespace aggregation.
type NamespaceAggregationFilter struct {
	// Namespace is the namespace to filter by.
	Namespaces []string
	// ExcludedNamespaces is the excluded namespaces to filter by.
	ExcludedNamespaces []string
}

// AttackPathFilter represents the filter for attack paths from containers.
type AttackPathFilter struct {
	// Namespace is the namespace to filter by.
	Namespace *string
	// Image is the image to filter by.
	Image *string
	// App is the app to filter by.
	App *string
	// Team is the team to filter by.
	Team *string
	// ExcludedNamespaces is the excluded namespaces to filter by.
	ExcludedNamespaces []string
	// TargetClass is the target class to filter by.
	TargetClass *string
	// TimeLimit is the time limit to filter by.
	TimeLimit *int64
}
