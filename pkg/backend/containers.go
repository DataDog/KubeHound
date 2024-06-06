package backend

import (
	"context"
	"errors"
	"fmt"

	"strings"

	embedconfigdocker "github.com/DataDog/KubeHound/deployments/kubehound"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/api/types/container"
)

type Backend struct {
	project        *types.Project
	composeService api.Service
	dockerCli      *command.DockerCli
}

func NewBackend(ctx context.Context) (*Backend, error) {
	project, err := loadProject(ctx)
	if err != nil {
		return nil, err
	}

	dockerCli, err := newDockerCli()
	if err != nil {
		return nil, err
	}

	composeService := compose.NewComposeService(dockerCli)

	return &Backend{
		project:        project,
		dockerCli:      dockerCli,
		composeService: composeService,
	}, nil
}

func loadProject(ctx context.Context) (*types.Project, error) {
	data, err := embedconfigdocker.F.ReadFile(embedconfigdocker.DefaultComposePath)
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

	project, err := loader.LoadWithContext(ctx, opts)
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

func newDockerCli() (*command.DockerCli, error) {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, err
	}

	// Initialize the Docker client with the default options
	err = dockerCli.Initialize(flags.NewClientOptions())
	if err != nil {
		return nil, err
	}

	return dockerCli, nil
}

func (b *Backend) Up(ctx context.Context) error {
	log.I.Infof("Spawning the kubehound stack")

	return b.composeService.Up(ctx, b.project, api.UpOptions{
		Create: api.CreateOptions{
			Build: &api.BuildOptions{
				Pull: true,
			},
			Services:      b.project.ServiceNames(),
			Recreate:      api.RecreateForce,
			RemoveOrphans: true,
		},
		Start: api.StartOptions{
			Wait:  true,
			Watch: true,
		},
	})
}

func (b *Backend) Down(ctx context.Context) error {
	log.I.Infof("Tearing down the kubehound stack")

	return b.composeService.Remove(ctx, b.project.Name, api.RemoveOptions{
		Stop:    true,
		Volumes: true,
		Force:   true,
	})
}

func (b *Backend) Reset(ctx context.Context) error {
	err := b.Down(ctx)
	if err != nil {
		return err
	}

	return b.Up(ctx)
}

func (b *Backend) IsStackRunning(ctx context.Context) (bool, error) {
	containers, err := b.dockerCli.Client().ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return false, err
	}

	containerList := map[string]bool{}
	for _, service := range b.project.ServiceNames() {
		containerName := fmt.Sprintf("%s-%s", b.project.Name, service)
		containerList[containerName] = true
	}

	for _, container := range containers {
		if _, ok := containerList[container.Names[0]]; !ok && container.State == "running" {
			return true, nil
		}
	}

	return false, nil
}

func (b *Backend) Wipe(ctx context.Context) error {
	var err error
	log.I.Infof("Wipping the persisted backend data")

	for _, volumeID := range b.project.VolumeNames() {
		log.I.Infof("Deleting volume %s", volumeID)
		err = errors.Join(err, b.dockerCli.Client().VolumeRemove(ctx, volumeID, true))
	}

	return err
}
