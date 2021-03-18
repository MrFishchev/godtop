package widgets

import (
	"context"
	"fmt"
	"github.com/gizak/termui/v3"
	ui "godtop/consoleui/termui"
	"godtop/domain"
	"godtop/infrastructure"
	"log"
	"sort"
	"time"
)

const MaxContainerLabelLength = 30

type CpuWidget struct {
	*ui.LineGraph
	DockerEngine        domain.DockerService
	MaxContainersCount int
	ContainersLoads     map[string]float64
	updateInterval      time.Duration
	average             float64
}

func NewCpuWidget() *CpuWidget {
	w := &CpuWidget{
		LineGraph:           ui.NewLineGraph(),
		MaxContainersCount: 5,
		updateInterval:      time.Second,
		ContainersLoads:     make(map[string]float64),
		average:             0,
	}

	w.Title = "CPU load"
	w.HorizontalScale = 5
	w.DockerEngine = infrastructure.CreateDockerService()
	w.update()

	go func() {
		for range time.NewTicker(w.updateInterval).C {
			w.Lock()
			w.update()
			w.Unlock()
		}
	}()

	return w
}

func (w *CpuWidget) Scale(i int) {
	w.LineGraph.HorizontalScale = i
}

func (w *CpuWidget) update() {
	go func() {

		ctx := context.Background()
		containers, err := w.DockerEngine.GetContainers(ctx, false)
		if err != nil {
			log.Printf("unable to get containers: %v", err.Error())
			return
		}

		containersInfo := make(map[string]*domain.ContainerStats, len(*containers))
		for _, c := range *containers {
			stats, err := w.DockerEngine.GetContainerStats(ctx, c.ID, false)
			if err != nil {
				log.Printf("unable to get stats of the container: %v (%v)", c.ID, err.Error())
				continue
			}
			stats.DisplayName = getContainerName(&c)

			containersInfo[c.ID] = stats
		}

		type pair struct {
			Key   string
			Value *domain.ContainerStats
		}

		var sortedContainersInfo []pair
		for k, v := range containersInfo {
			sortedContainersInfo = append(sortedContainersInfo, pair{k, v})
		}
		sort.Slice(sortedContainersInfo, func(i, j int) bool {
			return sortedContainersInfo[i].Value.CpuUsage > sortedContainersInfo[j].Value.CpuUsage
		})

		var count int
		for i, c := range sortedContainersInfo {
			if count >= w.MaxContainersCount {
				return
			}
			displayName := c.Value.DisplayName
			w.Data[displayName] = append(w.Data[displayName], float64(c.Value.CpuUsage))
			w.Labels[displayName] = fmt.Sprintf("%3f%%", c.Value.CpuUsage)
			w.ContainersLoads[displayName] = float64(c.Value.CpuUsage)
			w.LineColors[displayName] = termui.StandardColors[i]
			w.LabelStyles[displayName] = termui.ModifierBold
		}
	}()
}

func getContainerName(container *domain.Container) string{
	if len(container.Names)  > 0{
		name := container.Names[0]
		if len(name) > MaxContainerLabelLength {
			return name[:MaxContainerLabelLength] + "..."
		}
		return name
	}

	return container.ID[:MaxContainerLabelLength] + "..."
}
