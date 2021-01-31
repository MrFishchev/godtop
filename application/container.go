package application

import (
	"context"
	"godtop/domain"
)

type ContainerInteractor struct {
	Service domain.DockerService
}

//Get returns container by id or name
func (i *ContainerInteractor) Get(ctx context.Context, nameOrId string) (*domain.Container, error) {
	return i.Service.GetContainer(ctx, nameOrId)
}

//GetAll returns all running containers
func (i *ContainerInteractor) GetRunning(ctx context.Context) (*[]domain.Container, error) {
	return i.Service.GetContainers(ctx, false)
}

//GetAll returns all existing containers
func (i *ContainerInteractor) GetAll(ctx context.Context) (*[]domain.Container, error) {
	return i.Service.GetContainers(ctx, true)
}
