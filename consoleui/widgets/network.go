package widgets

import (
	"context"
	"fmt"
	ui "godtop/consoleui/termui"
	"godtop/domain"
	"godtop/domain/utils"
	"godtop/infrastructure"
	"log"
	"time"
)

type NetworkWidget struct {
	*ui.Table
	updateInterval time.Duration
	Containers     map[string]*domain.ContainerStats
	DockerEngine   domain.DockerService
}

func NewNetworkWidget() *NetworkWidget {
	self := &NetworkWidget{
		Table:          ui.NewTable(),
		updateInterval: time.Second * 3,
		Containers:     make(map[string]*domain.ContainerStats),
	}

	self.Title = "Network Usage"
	self.Header = []string{"Container", "Tx/s", "Rx/s"}
	self.ColGap = 2
	self.ColResizer = func() {
		self.ColWidths = []int{
			utils.MaxInt(4, (self.Inner.Dx()-10)/3),
			utils.MaxInt(4, (self.Inner.Dx()-30)/3),
			utils.MaxInt(4, (self.Inner.Dx()-30)/3),
		}
	}

	self.DockerEngine = infrastructure.CreateDockerService()
	self.update()

	go func() {
		for range time.NewTicker(self.updateInterval).C {
			self.Lock()
			self.update()
			self.Unlock()
		}
	}()

	return self
}

func (w *NetworkWidget) update() {

	go func() {

		ctx := context.Background()

		containers, err := w.DockerEngine.GetContainers(ctx, false)
		if err != nil {
			log.Printf("unable to get containers: %v", err.Error())
			return
		}

		for _, container := range *containers {
			stats, err := w.getContainerStats(ctx, container.ID)
			if err != nil {
				log.Printf("unable to get container's stats")
				continue
			}

			w.Containers[container.ID] = stats
		}

		w.updateTable(containers)
	}()
}

func (w *NetworkWidget) updateTable(containers *[]domain.Container) {
	w.Rows = make([][]string, len(w.Containers))

	i := 0
	for id, stats := range w.Containers {
		w.Rows[i] = make([]string, 3)
		w.Rows[i][0] = utils.GetContainerNameOrId(id, containers)
		w.Rows[i][1] = fmt.Sprintf("%v B", stats.TxBytes)
		w.Rows[i][2] = fmt.Sprintf("%v B", stats.RxBytes)
		i++
	}
}

func (w *NetworkWidget) getContainerStats(ctx context.Context, containerId string) (*domain.ContainerStats, error) {
	stats, err := w.DockerEngine.GetContainerStats(ctx, containerId, false)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
