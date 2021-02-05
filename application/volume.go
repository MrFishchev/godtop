package application

import (
	"context"
	"godtop/domain"
)

type VolumeInteractor struct {
	Service domain.DockerService
}

//GetAll returns all mounted volumes in a container
func (i *VolumeInteractor) GetAll(ctx context.Context) (*[]domain.Volume, error) {
	return i.Service.GetVolumes(ctx)
}
