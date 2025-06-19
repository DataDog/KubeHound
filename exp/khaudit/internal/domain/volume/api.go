package volume

import (
	"context"
	"errors"
)

var (
	// ErrNoResult is returned when no result is found.
	ErrNoResult = errors.New("no result found")
)

// Reader is the interface for reading volume data.
type Reader interface {
	// GetVolumes returns the volumes.
	GetVolumes(ctx context.Context, runID string, filter Filter) ([]Volume, error)
	// GetMountedHostPaths returns the mounted host paths.
	GetMountedHostPaths(ctx context.Context, runID string, filter Filter) ([]MountedHostPath, error)
}

// Filter represents the filter for volumes.
type Filter struct {
	// Type is the type of volume to filter by.
	Type *string
	// Namespace is the namespace to filter by.
	Namespace *string
	// App is the app to filter by.
	App *string
	// Team is the team to filter by.
	Team *string
	// Image is the image to filter by.
	Image *string
	// SourcePath is the source path to filter by.
	SourcePath *string
}

// Volume represents a volume.
type Volume struct {
	// Name is the name of the volume.
	Name string `json:"name"`
	// Type is the type of the volume.
	Type string `json:"type"`
	// Namespace is the namespace of the volume.
	Namespace string `json:"namespace"`
	// App is the app of the volume.
	App string `json:"app"`
	// Team is the team of the volume.
	Team string `json:"team"`
	// SourcePath is the source path of the volume.
	SourcePath string `json:"sourcePath"`
	// MountPath is the mount path of the volume.
	MountPath string `json:"mountPath"`
	// ReadOnly is true if the volume is read only.
	ReadOnly bool `json:"readOnly"`
}

// MountedHostPath represents a mounted host path.
type MountedHostPath struct {
	// SourcePath is the source path of the volume.
	SourcePath string `json:"sourcePath"`
	// Namespace is the namespace of the volume.
	Namespace string `json:"namespace"`
	// App is the app of the volume.
	App string `json:"app"`
	// Team is the team of the volume.
	Team string `json:"team"`
	// Image is the image of the volume.
	Image string `json:"image"`
}
