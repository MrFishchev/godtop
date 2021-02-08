package widgets

import (
	"context"
	"errors"
	"fmt"
	ui "godtop/consoleui/termui"
	"godtop/domain"
	"godtop/domain/utils"
	"godtop/infrastructure"
	"log"
	"sort"
	"time"
)

type VolumesWidget struct {
	*ui.Table
	updateInterval time.Duration
	Volumes        map[string]*domain.Volume
	DockerEngine   domain.DockerService
}

func NewVolumesWidget() *VolumesWidget {
	self := &VolumesWidget{
		Table:          ui.NewTable(),
		updateInterval: time.Second,
		Volumes:        make(map[string]*domain.Volume),
	}

	self.Title = "Storage Usage"
	self.Header = []string{"Volume", "Size"}
	self.ColGap = 2
	self.ColResizer = func() {
		self.ColWidths = []int{
			utils.MaxInt(4, (self.Inner.Dx()-29)/2),
			utils.MaxInt(4, (self.Inner.Dx()-10)/2),
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

func (w *VolumesWidget) update() {
	ctx := context.Background()
	volumes, err := w.DockerEngine.GetVolumes(ctx)
	if err != nil {
		log.Printf("unable to get volumes: %v", err.Error())
		return
	}

	w.addAppearedVolumes(volumes)
	w.deleteDisappearedVolumes(volumes)
	w.updateVolumes(volumes)
	w.updateTable()
}

// converts self.Volumes into self.Rows which is a [][]string
func (w *VolumesWidget) updateTable() {
	w.sortVolumesBySize()
	w.Rows = make([][]string, len(w.Volumes))

	i := 0
	for _, volume := range w.Volumes {
		w.Rows[i] = make([]string, 2)
		w.Rows[i][0] = volume.Source
		w.Rows[i][1] = fmt.Sprint(volume.Size / 1024 / 1024)
		i++
	}
}

// updates info of volumes in self.Volumes
func (w *VolumesWidget) updateVolumes(volumes *[]domain.Volume) {
	for _, volumeInTable := range w.Volumes {
		updatedVolume, err := getUpdatedVolume(volumeInTable, volumes)
		if err != nil {
			log.Printf("cannot update %v: %v", getVolumeKey(volumeInTable), err.Error())
			continue
		}
		updateVolumeInfo(volumeInTable, updatedVolume)
	}
}

func (w *VolumesWidget) sortVolumesBySize() {
	type pair struct {
		Key   string
		Value *domain.Volume
	}

	var sortedVolumes []pair
	for k, v := range w.Volumes {
		sortedVolumes = append(sortedVolumes, pair{k, v})
	}

	sort.Slice(sortedVolumes, func(i, j int) bool {
		return sortedVolumes[i].Value.Size > sortedVolumes[j].Value.Size
	})

	w.Volumes = make(map[string]*domain.Volume, len(sortedVolumes))
	for _, volume := range sortedVolumes {
		w.Volumes[volume.Key] = volume.Value
	}
}

func (w *VolumesWidget) addAppearedVolumes(volumes *[]domain.Volume) {
	for _, volume := range *volumes {

		key := getVolumeKey(&volume)
		if _, ok := w.Volumes[key]; !ok {
			w.Volumes[key] = &volume
		}
	}
}

func (w *VolumesWidget) deleteDisappearedVolumes(volumes *[]domain.Volume) {
	toDelete := []string{}
	for volumeInTable := range w.Volumes {
		exists := false
		for _, volume := range *volumes {
			if volumeInTable == getVolumeKey(&volume) {
				exists = true
				break
			}
		}

		if !exists {
			toDelete = append(toDelete, volumeInTable)
		}
	}

	for _, volume := range toDelete {
		delete(w.Volumes, volume)
	}
}

func getVolumeKey(volume *domain.Volume) string {
	return volume.Source + ":" + volume.Destination
}

func getUpdatedVolume(searchingVolume *domain.Volume, list *[]domain.Volume) (*domain.Volume, error) {
	for _, v := range *list {
		if *searchingVolume == v {
			return &v, nil
		}
	}

	return nil, errors.New("unable to find volume")
}

func updateVolumeInfo(volume, updatedVolume *domain.Volume) {
	volume.Name = updatedVolume.Name
	volume.Size = updatedVolume.Size
}
