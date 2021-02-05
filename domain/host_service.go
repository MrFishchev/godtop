package domain

import "context"

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination mock_$GOFILE

// HostService represents access to the host system
// Expect implementation by the infrastructure layer
type HostService interface {
	GetInfo(ctx context.Context) *HostInfo
}
