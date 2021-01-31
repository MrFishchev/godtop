package infrastructure

import (
	"context"
	"godtop/domain"
)

type dockerEngine struct {
}

func Create() *dockerEngine {
	return &dockerEngine{}
}

//GetAllContainers returns list of domain.Container
func (d dockerEngine) GetAllContainers(ctx context.Context) ([]*domain.Container, error) {
	containers := []*domain.Container{
		{
			ID:   1,
			Name: "Audiomining",
		},
		{
			ID:   2,
			Name: "ProfileBuilder",
		},
	}

	return containers, nil
}

//GetContainer returns domain.Container
func (d dockerEngine) GetContainer(ctx context.Context, id int) (*domain.Container, error) {
	container := &domain.Container{
		ID:   1,
		Name: "Audiomining",
	}

	return container, nil
}
