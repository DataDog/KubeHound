//go:build no_backend

package backend

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	embedconfigdocker "github.com/DataDog/KubeHound/deployments/kubehound"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
	"gopkg.in/yaml.v2"
)

var (
	DefaultReleaseComposePaths = []string{"docker-compose.yaml", "docker-compose.release.yaml.tpl"}
	DefaultUIProfile           = []string{DevUIProfile}

	DevUIProfile = "jupyter"
)

func loadProject(ctx context.Context, composeFilePaths []string, profiles []string) (*types.Project, error) {
	var project *types.Project
	var err error
	l := log.Logger(ctx)
	switch {
	case len(composeFilePaths) != 0 && len(composeFilePaths[0]) != 0:
		l.Info("Loading backend from file", log.Strings("path", composeFilePaths))
		project, err = loadComposeConfig(ctx, composeFilePaths, profiles)
	default:
		l.Info("Loading backend from default embedded")
		project, err = loadEmbeddedConfig(ctx, profiles)
	}

	if err != nil {
		return nil, err
	}

	// Adding labels to make the project compatible with the Compose API
	// ref: https://github.com/docker/compose/issues/11210#issuecomment-1820553483
	for i, s := range project.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     project.Name,
			api.ServiceLabel:     s.Name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  project.WorkingDir,
			api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
			api.OneoffLabel:      "False", // default, will be overridden by `run` command
		}

		project.Services[i] = s
	}
	for key, n := range project.Networks {
		n.Labels = map[string]string{
			api.ProjectLabel: project.Name,
			api.NetworkLabel: n.Name,
			api.VersionLabel: api.ComposeVersion,
		}

		project.Networks[key] = n
	}

	return project, nil
}
func loadComposeConfig(ctx context.Context, composeFilePaths []string, profiles []string) (*types.Project, error) {
	options, err := cli.NewProjectOptions(
		composeFilePaths,
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithProfiles(profiles),
	)
	if err != nil {
		return nil, err
	}

	return cli.ProjectFromOptions(ctx, options)
}

func loadEmbeddedConfig(ctx context.Context, profiles []string) (*types.Project, error) {
	var dockerComposeFileData map[interface{}]interface{}
	var err error

	for i, filePath := range DefaultReleaseComposePaths {
		dockerComposeFileData, err = loadEmbeddedDockerCompose(ctx, filePath, dockerComposeFileData)
		if err != nil {
			return nil, fmt.Errorf("loading embedded compose file %d: %w", i, err)
		}
	}

	data, err := yaml.Marshal(dockerComposeFileData)
	if err != nil {
		return nil, fmt.Errorf("reading embed config: %w", err)
	}

	opts := types.ConfigDetails{
		ConfigFiles: []types.ConfigFile{
			{
				Content: data,
			},
		},
	}

	return loader.LoadWithContext(ctx, opts, loader.WithProfiles(profiles))
}

func loadEmbeddedDockerCompose(ctx context.Context, filepath string, dockerComposeFileData map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	l := log.Logger(ctx)
	var localYaml map[interface{}]interface{}
	localData, err := embedconfigdocker.F.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("reading embed config: %w", err)
	}

	// Dynamically setting the version tag for the release using a template file
	if strings.HasSuffix(filepath, ".tpl") {
		// Setting the version tag for the release dynamically
		version := map[string]string{"VersionTag": config.BuildVersion}

		// For local version (when the version is "dirty", using latest to have a working binary)
		// For any branch outside of main, using latest image as the current tag will cover (including the commit sha in the tag)
		if strings.HasSuffix(config.BuildBranch, "dirty") || config.BuildBranch != "main" {
			l.Warn("Loading the kubehound images with tag latest - dev branch detected")
			version["VersionTag"] = "latest"
		}

		tmpl, err := template.New(filepath).ParseFS(embedconfigdocker.F, filepath)
		if err != nil {
			return nil, fmt.Errorf("new template: %w", err)
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, version)
		if err != nil {
			return nil, fmt.Errorf("executing template: %w", err)
		}
		localData = buf.Bytes()
	}

	err = yaml.Unmarshal(localData, &localYaml)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return mergeMaps(dockerComposeFileData, localYaml), nil
}
