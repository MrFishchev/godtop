package application

import (
	"context"
	"godtop/domain"
)

type VolumeInteractor struct {
	Service domain.DockerService
}

func (i *VolumeInteractor) GetAll(ctx context.Context) (*[]domain.Volume, error) {
	return i.Service.GetVolumes(ctx)
}
