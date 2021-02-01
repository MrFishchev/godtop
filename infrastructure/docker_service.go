package infrastructure

import (
	"context"
	"errors"
	"godtop/domain"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type dockerEngine struct {
}

func Create() *dockerEngine {
	return &dockerEngine{}
}

//GetAllContainers returns list of docker containers
func (d dockerEngine) GetContainers(ctx context.Context, all bool) (*[]domain.Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	options := types.ContainerListOptions{
		All: all,
	}

	containers, err := cli.ContainerList(ctx, options)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Container, len(containers))
	for i, container := range containers {
		result[i] = domain.Container{
			ID:          container.ID,
			Names:       getTrimmedNames(container.Names),
			State:       container.State,
			Status:      container.Status,
			PublicPorts: getPublicPorts(container.Ports),
		}
	}

	return &result, nil
}

//GetContainer returns container by id or name even even not running
func (d dockerEngine) GetContainer(ctx context.Context, idOrName string) (*domain.Container, error) {
	containers, err := d.GetContainers(ctx, true)
	if err != nil {
		return nil, err
	}

	for _, container := range *containers {
		if container.ID == idOrName {
			return &container, nil
		}

		for _, name := range container.Names {
			if strings.TrimPrefix(name, "/") == idOrName {
				return &container, nil
			}
		}
	}

	return nil, errors.New("Cannot find a container")
}

func (d dockerEngine) GetVolumes(ctx context.Context) (*[]domain.Volume, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	var volumes []domain.Volume
	for _, container := range containers {
		for _, mount := range container.Mounts {
			if mount.Type == "bind" || mount.Type == "volume" {
				volume := domain.Volume{
					Source:      mount.Source,
					Destination: mount.Destination,
					Size:        GetDirectorySize(mount.Source),
				}
				volumes = append(volumes, volume)
			}
		}
	}

	return &volumes, nil
}

//region Private Methods

func getTrimmedNames(names []string) []string {
	result := make([]string, len(names))
	for i, name := range names {
		result[i] = strings.TrimPrefix(name, "/")
	}

	return result
}

func getPublicPorts(ports []types.Port) []uint16 {
	if len(ports) == 0 {
		return nil
	}

	var result []uint16
	for _, port := range ports {
		if port.PublicPort != 0 {
			result = append(result, port.PublicPort)
		}
	}

	return result
}

//endregion
