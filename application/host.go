package application

import (
	"context"
	"godtop/domain"
)

type HostInteractor struct {
	Service domain.HostService
}

//GetInfo returns info about the host system
func (i *HostInteractor) GetInfo(ctx context.Context) *domain.HostInfo {
	return i.Service.GetInfo(ctx)
}
