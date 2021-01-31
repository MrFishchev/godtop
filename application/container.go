package application

import (
	"context"
	"godtop/domain"
)

type ContainerInteractor struct {
	Service domain.DockerService
}

//Get returns container or error
func (i *ContainerInteractor) Get(ctx context.Context, id int) (*domain.Container, error) {
	return i.Service.GetContainer(ctx, id)
}

//GetAll returns all containers or error
func (i *ContainerInteractor) GetAll(ctx context.Context) ([]*domain.Container, error) {
	return i.Service.GetAllContainers(ctx)
}
