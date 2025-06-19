package renderer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/container"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/ingestion"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/permission"
	"github.com/DataDog/KubeHound/exp/khaudit/internal/domain/volume"
)

// Markdown creates a new markdown renderer.
func Markdown(ingestions ingestion.Reader, permissions permission.Reader, containers container.Reader, volumes volume.Reader, namespaces []string, assumeGroup string) Renderer {
	return &renderer{
		ingestions:  ingestions,
		permissions: permissions,
		containers:  containers,
		namespaces:  namespaces,
		volumes:     volumes,
		assumeGroup: assumeGroup,
	}
}

// -----------------------------------------------------------------------------

type renderer struct {
	ingestions  ingestion.Reader
	permissions permission.Reader
	containers  container.Reader
	volumes     volume.Reader
	namespaces  []string
	assumeGroup string
}

func (r *renderer) Render(ctx context.Context, writer io.Writer, cluster string, runID string) error {
	// Check arguments.
	if writer == nil {
		return errors.New("writer is nil")
	}

	// Render the cluster metrics.
	if err := r.renderClusterMetrics(ctx, writer); err != nil {
		return fmt.Errorf("unable to render cluster metrics: %w", err)
	}

	// Render cluster roles overview.
	if err := r.renderClusterRolesOverview(ctx, writer, runID); err != nil {
		return fmt.Errorf("unable to render cluster roles overview: %w", err)
	}

	// Render container escape.
	if err := r.renderContainerEscape(ctx, writer, cluster, runID); err != nil {
		return fmt.Errorf("unable to render container escape: %w", err)
	}

	return nil
}

// -----------------------------------------------------------------------------

var clusterMetricsMarkdownTemplate = `# Cluster Metrics

## Edge Metrics

| Edge Label | Count |
| ---------- | ----- |
{{ range $key, $value := .EdgeMetrics -}}
| {{ $key }} | {{ $value }} |
{{ end }}

## Vertex Metrics

| Vertex Label | Count |
| ----------- | ----- |
{{ range $key, $value := .VertexMetrics -}}
| {{ $key }} | {{ $value }} |
{{ end }}

`

func (r *renderer) renderClusterMetrics(ctx context.Context, writer io.Writer) error {
	slog.Info("getting edge metrics")
	// Get Edge metrics.
	edgeMetrics, err := r.ingestions.GetEdgeCountPerClasses(ctx)
	if err != nil {
		if !errors.Is(err, ingestion.ErrNoResult) {
			return fmt.Errorf("unable to get edge metrics: %w", err)
		}
	}

	slog.Info("getting vertex metrics")
	// Get Vertex metrics.
	vertexMetrics, err := r.ingestions.GetVertexCountPerClasses(ctx)
	if err != nil {
		if !errors.Is(err, ingestion.ErrNoResult) {
			return fmt.Errorf("unable to get vertex metrics: %w", err)
		}
	}

	slog.Info("rendering cluster metrics")
	// Render the cluster metrics.
	return template.Must(template.New("cluster-metrics").Parse(clusterMetricsMarkdownTemplate)).Execute(writer, struct {
		EdgeMetrics   map[string]int64
		VertexMetrics map[string]int64
	}{
		EdgeMetrics:   edgeMetrics,
		VertexMetrics: vertexMetrics,
	})
}

// -----------------------------------------------------------------------------

var clusterRolesOverviewMarkdownTemplate = `# Cluster Roles Overview

The initial step of the attack is to find a reachable pod, aka a pod that can be
exploited to gain access to the cluster.

We are currently looking at the ` + "`{{ .DefaultGroup }}`" + ` group to assume that 
any employee member of this group is allowed to ` + "`kubectl exec`" + ` into any 
pod in the cluster.

## Mutable Pods

We need to start analysing what we can do with our identity to save time during 
further investigations. Most DD clusters have protection, preventing you from 
testing the attack path from the initial step. 

The following list indicates the number of pods that can be exploited directly 
via ` + "`kubectl exec`" + ` or indirectly via a pod manifest alteration.

| Namespace | Pod Count |
| --------- | --------- |
{{ range $key, $value := .ReachablePods -}}
| {{ $key }} | {{ $value }} |
{{ end }}

### Direct access to pods via ` + "`kubectl exec`" + ` for ` + "`{{ .DefaultGroup }}`" + ` group

Immediately exploitable pods where anyone at DD member of ` + "`{{ .DefaultGroup }}`" + ` group can jump into.

` + "```" + `
{{ .DirectAccessPodCount }}
` + "```" + `

Reachable pods via ` + "`kubectl exec`" + ` for ` + "`{{ .DefaultGroup }}`" + ` grouped by namespace.

| Group Name | Namespace | Pod Count |
| ---------- | --------- | --------- |
{{ range .DirectAccessReachablePods -}}
| {{ .GroupName }} | {{ .Namespace }} | {{ .PodCount }} |
{{ end }}

> TODO: Exploit chain validation from pods using this image.

`

func (r *renderer) renderClusterRolesOverview(ctx context.Context, writer io.Writer, runID string) error {
	slog.Info("getting reachable pods per namespace")
	// Get the number of reachable pods per namespace.
	reachablePods, err := r.permissions.GetReachablePodCountPerNamespace(ctx, runID)
	if err != nil {
		return fmt.Errorf("unable to get reachable pods: %w", err)
	}

	slog.Info("getting direct access pod count")
	// Get the number of pods that can be exploited directly.
	directAccessPodCount, err := r.permissions.GetKubectlExecutablePodCount(ctx, runID, r.assumeGroup)
	if err != nil {
		if !errors.Is(err, permission.ErrNoResult) {
			return fmt.Errorf("unable to get direct access pod count: %w", err)
		}
	}

	slog.Info("getting direct access reachable pods")
	// Get the reachable pods that can be exploited directly.
	directAccessReachablePods, err := r.permissions.GetExposedPodCountPerNamespace(ctx, runID, r.assumeGroup)
	if err != nil {
		if !errors.Is(err, permission.ErrNoResult) {
			return fmt.Errorf("unable to get direct access reachable pods: %w", err)
		}
	}

	// Sort the reachable pods by count.
	sort.Slice(directAccessReachablePods, func(i, j int) bool {
		return directAccessReachablePods[i].PodCount > directAccessReachablePods[j].PodCount
	})

	slog.Info("rendering cluster roles overview")

	// Render the cluster roles overview.
	return template.Must(template.New("cluster-roles-overview").Parse(clusterRolesOverviewMarkdownTemplate)).Execute(writer, struct {
		ReachablePods             map[string]int64
		DirectAccessReachablePods []permission.ExposedPodCount
		DefaultGroup              string
		DirectAccessPodCount      int64
	}{
		ReachablePods:             reachablePods,
		DirectAccessReachablePods: directAccessReachablePods,
		DefaultGroup:              r.assumeGroup,
		DirectAccessPodCount:      directAccessPodCount,
	})
}

// -----------------------------------------------------------------------------

var containerEscapeNamespaceMarkdownTemplate = `## {{ if .DefaultGroupHasAccess }}🟢{{ else }}❌{{ end }} - Namespace {{ .Namespace }}
{{ $namespace := .Namespace }}
### Allowed Groups

List of groups that must be exploited to gain access to the pods in the namespace.

| Namespace | Group Name |
| --------- | ---------- |
{{ range .Groups -}}
| {{ $namespace }} | {{ . }} |
{{ end }}

### Attack Path Profiles

List of attack path profiles that can be exploited to gain access to the node.

| Namespace | Attack Path Profile |
| --------- | ------------------- |
{{ range .ContainerEscapeProfiles -}}
| {{ $namespace }} | {{ join "-->" .Path }} |
{{ end }}

{{ if .HasMountedHostpathVolumes -}}### Hostpath Volumes

List of hostpath volumes that can be exploited to gain access to the node.

| Namespace | Host Path |
| --------- | ----------- |
{{ range .HostpathVolumesBySourcePath -}}
| {{ $namespace }} | {{ . }} |
{{ end }}

Host path volumes grouped by pod image.

| Namespace | Volume Name | App | Team | Image |
| --------- | ----------- | --- | ---- | ----- |
{{ range .HostpathVolumes -}}
| {{ $namespace }} | {{ .SourcePath }} | {{ .App }} | {{ .Team }} | {{ .Image }} |
{{ end }}

{{ end }}
### Initial access

We are currently looking at the ` + "`{{ .DefaultGroup }}`" + ` group to assume that 
any employee member of this group is allowed to ` + "`kubectl exec`" + ` into any 
pod in the cluster.

{{ if .DefaultGroupHasAccess }}
> The ` + "`{{ .DefaultGroup }}`" + ` group has access to the pods in the 
` + "`{{ .Namespace }}`" + ` namespace. The initial access is possible.
{{ else }}
> The ` + "`{{ .DefaultGroup }}`" + ` group does not have access to the pods in the 
` + "`{{ .Namespace }}`" + ` namespace. The initial access is blocked.
{{ end }}

`

var containerEscapeContainerMarkdownTemplate = `#### Pods - {{ .Image.Image }}

Reachable pods via ` + "`kubectl exec`" + ` for ` + "`{{ .DefaultGroup }}`" + `.

{{ if .IsTruncated }}
> The list of reachable pods is truncated to 100.
{{ end }}

{{ if .HasReachablePods -}}
| Namespace | Pod Name | Image Name | App | Team |
| --------- | -------- | ---------- | --- | ---- |
{{ range .ReachablePods -}}
| {{ .Namespace }} | {{ .PodName }} | {{ .Image }} | {{ .App }} | {{ .Team }} |
{{ end }}
{{ else -}}> No reachable pods found for the image for the ` + "`{{ .DefaultGroup }}`" + ` group.{{ end }}

`

func (r *renderer) renderContainerEscape(ctx context.Context, writer io.Writer, cluster string, runID string) error {
	// Parse and compile the container escape template.
	containerEscapeNamespaceTemplate, err := template.New("container-escape-namespace").Funcs(template.FuncMap{
		"join": join,
	}).Parse(containerEscapeNamespaceMarkdownTemplate)
	if err != nil {
		return fmt.Errorf("unable to parse container escape template: %w", err)
	}

	// Parse and compile the container escape template.
	containerEscapeContainerTemplate, err := template.New("container-escape-container").Funcs(template.FuncMap{
		"join": join,
	}).Parse(containerEscapeContainerMarkdownTemplate)
	if err != nil {
		return fmt.Errorf("unable to parse container escape template: %w", err)
	}

	// Deduplicate the namespaces and sort them.
	r.namespaces = slices.Compact(r.namespaces)
	sort.Strings(r.namespaces)

	// If no namespaces are provided, skip the container escape audit.
	if len(r.namespaces) == 0 {
		slog.Info("no namespaces provided, skipping container escape audit")
		return nil
	}

	// Add the container escape header.
	fmt.Fprintln(writer, "# Container Escape")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "The following section focuses on each namespace with identified attack paths.")
	fmt.Fprintln(writer, "> Namespace with 🟢 are exploitable, while ❌ are not.")
	fmt.Fprintln(writer, "> Pods with 💥 have verified exploit chain.")
	fmt.Fprintln(writer, "")

	// Render the container escape for each provided namespace.
	for _, namespace := range r.namespaces {
		slog.Info("enumerating kubectl executable groups for namespace", "namespace", namespace)
		// List groups that can be exploited to gain access to the container.
		groups, err := r.permissions.GetKubectlExecutableGroupsForNamespace(ctx, runID, namespace)
		if err != nil {
			if !errors.Is(err, permission.ErrNoResult) {
				return fmt.Errorf("unable to get groups: %w", err)
			}
		}
		if len(groups) == 0 {
			// No groups found, skip the namespace.
			slog.Info("no groups found for namespace", "namespace", namespace)
		}

		// Sort the groups by name.
		sort.Strings(groups)

		// Check if the groups contains the default group.
		defaultGroupHasAccess := false
		for _, group := range groups {
			if group == r.assumeGroup {
				defaultGroupHasAccess = true
			}
		}

		slog.Info("getting container escape profiles for namespace", "namespace", namespace)

		// Get the container escape attack path profiles.
		containerEscapeProfiles, err := r.containers.GetAttackPathProfiles(ctx, cluster, runID, container.AttackPathFilter{
			Namespace: &namespace,
		})
		if err != nil {
			if !errors.Is(err, container.ErrNoResult) {
				return fmt.Errorf("unable to get container escape profiles for namespace %s: %w", namespace, err)
			}
		}

		// Sort the container escape profiles by path length.
		sort.Slice(containerEscapeProfiles, func(i, j int) bool {
			return len(containerEscapeProfiles[i].Path) < len(containerEscapeProfiles[j].Path)
		})

		slog.Info("auditing hostpath volumes for namespace", "namespace", namespace)

		// Get the hostpath volumes for the namespace.
		hostpathVolumes, err := r.volumes.GetMountedHostPaths(ctx, runID, volume.Filter{
			Namespace: &namespace,
		})
		if err != nil {
			if !errors.Is(err, volume.ErrNoResult) {
				return fmt.Errorf("unable to get hostpath volumes for namespace %s: %w", namespace, err)
			}
		}

		// Collect the hostpath volumes by source path.
		hostpathVolumesBySourcePath := []string{}
		for _, hostpath := range hostpathVolumes {
			hostpathVolumesBySourcePath = append(hostpathVolumesBySourcePath, hostpath.SourcePath)
		}

		// Sort the hostpath volumes by source path.
		sort.Strings(hostpathVolumesBySourcePath)

		// Deduplicate the hostpath volumes by source path.
		hostpathVolumesBySourcePath = slices.Compact(hostpathVolumesBySourcePath)

		// Render the container escape for the namespace.
		if err := containerEscapeNamespaceTemplate.Execute(writer, struct {
			Namespace                   string
			Groups                      []string
			DefaultGroup                string
			DefaultGroupHasAccess       bool
			ContainerEscapeProfiles     []container.AttackPath
			HostpathVolumes             []volume.MountedHostPath
			HasMountedHostpathVolumes   bool
			HostpathVolumesBySourcePath []string
		}{
			Namespace:                   namespace,
			Groups:                      groups,
			DefaultGroup:                r.assumeGroup,
			DefaultGroupHasAccess:       defaultGroupHasAccess,
			ContainerEscapeProfiles:     containerEscapeProfiles,
			HostpathVolumes:             hostpathVolumes,
			HasMountedHostpathVolumes:   len(hostpathVolumes) > 0,
			HostpathVolumesBySourcePath: hostpathVolumesBySourcePath,
		}); err != nil {
			return fmt.Errorf("unable to render container escape for namespace %s: %w", namespace, err)
		}

		// If the default group does not have access, skip the image processing.
		if !defaultGroupHasAccess {
			slog.Info("skipping container escape for namespace, default group does not have access", "namespace", namespace)
			continue
		}

		slog.Info("processing container escape for namespace", "namespace", namespace)

		// Get the container initial image.
		imageChan := make(chan container.Container, 1000)
		if err := r.containers.GetVulnerables(ctx, cluster, runID, container.AttackPathFilter{
			Namespace: &namespace,
		}, imageChan); err != nil {
			if !errors.Is(err, container.ErrNoResult) {
				return fmt.Errorf("unable to get container initial image for namespace %s: %w", namespace, err)
			}
		}

		// Close the image channel.
		close(imageChan)

		// For each image, render the container escape case.
		for image := range imageChan {
			slog.Info("processing container escape for image", "image", image.Image)

			// Get the reachable pods for the image.
			reachablePods, err := r.permissions.GetExposedNamespacePods(ctx, runID, namespace, r.assumeGroup, permission.ExposedPodFilter{
				Image: &image.Image,
			})
			if err != nil {
				if !errors.Is(err, permission.ErrNoResult) {
					return fmt.Errorf("unable to get reachable pods: %w", err)
				}
			}

			// Truncate pods to 100.
			isTruncated := false
			if len(reachablePods) > 100 {
				reachablePods = reachablePods[:100]
				isTruncated = true
			}

			// Merge with template.
			if err := containerEscapeContainerTemplate.Execute(writer, struct {
				Namespace               string
				Groups                  []string
				DefaultGroup            string
				Image                   container.Container
				ContainerEscapeProfiles []container.AttackPath
				DefaultGroupHasAccess   bool
				ReachablePods           []permission.ExposedPodNamespace
				IsTruncated             bool
				HasReachablePods        bool
			}{
				Namespace:               namespace,
				Groups:                  groups,
				DefaultGroup:            r.assumeGroup,
				DefaultGroupHasAccess:   defaultGroupHasAccess,
				Image:                   image,
				ContainerEscapeProfiles: containerEscapeProfiles,
				ReachablePods:           reachablePods,
				HasReachablePods:        len(reachablePods) > 0,
				IsTruncated:             isTruncated,
			}); err != nil {
				return fmt.Errorf("unable to render container escape for image %s: %w", image.Name, err)
			}
		}
	}

	return nil
}

func join(sep string, s []string) string {
	return strings.Join(s, sep)
}
