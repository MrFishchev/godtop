package domain

import "context"

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination mock_$GOFILE

// DockerService represents access to docker engine API
// Expect implementation by the infrastructure layer
type DockerService interface {
	GetContainer(ctx context.Context, idOrName string) (*Container, error)
	GetContainers(ctx context.Context, all bool) (*[]Container, error)
	GetVolumes(ctx context.Context) (*[]Volume, error)
}
