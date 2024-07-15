package backend

import (
	"bytes"
	"context"
	"fmt"
	"os"
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
	DefaultDatadogComposePath  = "docker-compose.datadog.yaml"
	DefaultUIProfile           = []string{DevUIProfile}

	DevUIProfile = "jupyter"
)

func loadProject(ctx context.Context, composeFilePaths []string, profiles []string) (*types.Project, error) {
	var project *types.Project
	var err error

	switch {
	case len(composeFilePaths) != 0 && len(composeFilePaths[0]) != 0:
		log.I.Infof("Loading backend from file %s", composeFilePaths)
		project, err = loadComposeConfig(ctx, composeFilePaths, profiles)
	default:
		log.I.Infof("Loading backend from default embedded")
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
	var hostname string

	// Adding datadog setup
	ddAPIKey, ddAPIKeyOk := os.LookupEnv("DD_API_KEY")
	ddAPPKey, ddAPPKeyOk := os.LookupEnv("DD_API_KEY")

	if ddAPIKeyOk && ddAPPKeyOk {
		DefaultReleaseComposePaths = append(DefaultReleaseComposePaths, DefaultDatadogComposePath)
		hostname, err = os.Hostname()
		if err != nil {
			hostname = "kubehound"
		}

	}

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
		Environment: map[string]string{
			"DD_API_KEY":      ddAPIKey,
			"DD_APP_KEY":      ddAPPKey,
			"DOCKER_HOSTNAME": hostname,
		},
	}

	return loader.LoadWithContext(ctx, opts, loader.WithProfiles(profiles))
}

func loadEmbeddedDockerCompose(_ context.Context, filepath string, dockerComposeFileData map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	var localYaml map[interface{}]interface{}
	var localData []byte
	var err error

	if strings.HasSuffix(filepath, ".tpl") {
		// Setting the version tag for the release dynamically
		version := map[string]string{"VersionTag": config.BuildVersion}

		// For local version (when the version is "dirty", using latest to have a working binary)
		// For any branch outside of main, using latest image as the current tag will cover (including the commit sha in the tag)
		if strings.HasSuffix(config.BuildBranch, "dirty") || config.BuildBranch != "main" {
			log.I.Warnf("Loading the kubehound images with tag latest - dev branch detected")
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
	} else {
		localData, err = embedconfigdocker.F.ReadFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("reading embed config: %w", err)
		}
	}

	err = yaml.Unmarshal(localData, &localYaml)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return mergeMaps(dockerComposeFileData, localYaml), nil
}
