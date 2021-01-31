package domain

import "context"

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination mock_$GOFILE

// DockerService represents access to docker engine API
// Expect implementation by the infrastructure layer
type DockerService interface {
	GetContainer(ctx context.Context, id int) (*Container, error)
	GetAllContainers(ctx context.Context) ([]*Container, error)
}
