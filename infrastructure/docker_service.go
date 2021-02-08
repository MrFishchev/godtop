package infrastructure

import (
	"bytes"
	"context"
	"errors"
	"godtop/domain"
	"godtop/domain/utils"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/tidwall/gjson"
)

type dockerEngine struct {
}

func CreateDockerService() *dockerEngine {
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

	return nil, errors.New("cannot find a container")
}

func (d dockerEngine) GetContainerStats(ctx context.Context, containerId string, stream bool) (*domain.ContainerStats, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	response, err := cli.ContainerStatsOneShot(ctx, containerId)
	if err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	_, err = io.Copy(&buff, response.Body)
	if err != nil {
		return nil, err
	}
	jsonBytes := buff.Bytes()

	result := domain.ContainerStats{}
	result.RxBytes, result.TxBytes = getNetworkStats(&jsonBytes)
	result.UsedMemory, result.MemoryUsage = getMemoryStats(&jsonBytes)
	result.CpuUsage = getCpuStats(&jsonBytes)

	return &result, nil
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
					Size:        utils.GetDirectorySize(mount.Source),
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

func getNetworkStats(jsonBytes *[]byte) (rx int64, tx int64) {
	eth0 := gjson.GetBytes(*jsonBytes, "networks.eth0")
	if eth0.Type.String() != "Null" {
		rx = eth0.Get("rx_bytes").Int()
		tx = eth0.Get("tx_bytes").Int()
	}

	return rx, tx
}

func getMemoryStats(jsonBytes *[]byte) (usedMemory int64, memoryUsage float32) {
	//used_memory = memory_stats.usage - memory_stats.stats.cache
	//available_memory = memory_stats.limit
	//memory_usage% = (used_memory / available_memory) * 100.0
	memory := gjson.GetBytes(*jsonBytes, "memory_stats")
	if memory.Type.String() != "Null" {
		usage := memory.Get("usage")
		cache := memory.Get("stats.cache")
		if usage.Type.String() != "Null" && cache.Type.String() != "Null" {
			usedMemory = usage.Int() - cache.Int()
		}

		available := memory.Get("limit")
		if available.Type.String() != "Null" {
			memoryUsage = (float32(usedMemory) / float32(available.Int())) * 100.0
		}
	}

	return usedMemory, memoryUsage
}

func getCpuStats(jsonBytes *[]byte) float32 {
	//cpu_delta = cpu_stats.cpu_usage.total_usage - precpu_stats.cpu_usage.total_usage
	//system_cpu_delta = cpu_stats.system_cpu_usage - precpu_stats.system_cpu_usage
	//number_cpus = cpu_stats.online_cpus (if older lenght(cpu_stats.cpu_usage.percpu_usage))
	//cpu_usage% = (cpu_delta / system_cpu_delta) * number_cpus * 100.0
	cpu := gjson.GetBytes(*jsonBytes, "cpu_stats")
	precpu := gjson.GetBytes(*jsonBytes, "precpu_stats")
	if cpu.Type.String() == "Null" || precpu.Type.String() == "Null" {
		return 0
	}

	totalUsage := cpu.Get("cpu_usage.total_usage")
	preTotalUsage := precpu.Get("cpu_usage.total_usage")
	if totalUsage.Type.String() == "Null" {
		return 0
	}

	cpuDelta := totalUsage.Int() - preTotalUsage.Int()
	systemUsage := cpu.Get("system_cpu_usage")
	preSystemUsage := precpu.Get("system_cpu_usage")
	if systemUsage.Type.String() == "Null" {
		return 0
	}

	systemCpuDelta := systemUsage.Int() - preSystemUsage.Int()
	number_cpus := cpu.Get("online_cpus")

	return float32(cpuDelta/systemCpuDelta) * float32(number_cpus.Int()) * 100.0
}

//endregion
