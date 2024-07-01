package backend

import (
	"context"
	"errors"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/api/types/container"
)

var (
	currentBackend *Backend
)

type Backend struct {
	project        *types.Project
	composeService api.Service
	dockerCli      *command.DockerCli
}

func NewBackend(ctx context.Context, composeFilePaths []string) error {
	var err error
	currentBackend, err = newBackend(ctx, composeFilePaths)

	return err
}

func newBackend(ctx context.Context, composeFilePaths []string) (*Backend, error) {
	project, err := loadProject(ctx, composeFilePaths)
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

func BuildUp(ctx context.Context) error {
	return currentBackend.buildUp(ctx)
}

func (b *Backend) buildUp(ctx context.Context) error {
	log.I.Infof("Building the kubehound stack")
	err := b.composeService.Build(ctx, b.project, api.BuildOptions{
		NoCache: true,
		Pull:    true,
	})
	if err != nil {
		return err
	}

	return b.up(ctx)
}

func Up(ctx context.Context) error {
	return currentBackend.up(ctx)
}

func (b *Backend) up(ctx context.Context) error {
	log.I.Infof("Spawning the kubehound stack")

	err := b.composeService.Up(ctx, b.project, api.UpOptions{
		Create: api.CreateOptions{
			Build: &api.BuildOptions{
				NoCache: true,
				Pull:    true,
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
	if err != nil {
		return fmt.Errorf("error starting the kubehound stack: %w. Please make sure your Docker is correctly installed and that you have enough disk space available. ", err)
	}

	return nil
}

func Down(ctx context.Context) error {
	return currentBackend.down(ctx)
}

func (b *Backend) down(ctx context.Context) error {
	log.I.Info("Tearing down the kubehound stack")

	err := b.composeService.Remove(ctx, b.project.Name, api.RemoveOptions{
		Stop:    true,
		Volumes: true,
		Force:   true,
	})
	if err != nil {
		return fmt.Errorf("error shuting down kubehound stack: %w", err)
	}

	return nil
}

func Reset(ctx context.Context) error {
	return currentBackend.Reset(ctx)
}

func (b *Backend) Reset(ctx context.Context) error {
	err := b.down(ctx)
	if err != nil {
		return err
	}

	return b.up(ctx)
}

func IsStackRunning(ctx context.Context) (bool, error) {
	return currentBackend.isStackRunning(ctx)
}

func (b *Backend) isStackRunning(ctx context.Context) (bool, error) {
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
		if _, ok := containerList[container.Names[0]]; ok && container.State == "running" {
			return true, nil
		}
	}

	return false, nil
}

func Wipe(ctx context.Context) error {
	return currentBackend.wipe(ctx)
}

func (b *Backend) wipe(ctx context.Context) error {
	var err error
	log.I.Infof("Wiping the persisted backend data")

	for _, volumeID := range b.project.VolumeNames() {
		log.I.Infof("Deleting volume %s", volumeID)
		err = errors.Join(err, b.dockerCli.Client().VolumeRemove(ctx, volumeID, true))
	}

	return err
}
