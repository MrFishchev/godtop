package utils

import "godtop/domain"

func GetContainerNameOrId(id string, containers *[]domain.Container) string {
	var container *domain.Container = nil
	for _, c := range *containers {
		if id == c.ID {
			container = &c
			break
		}
	}

	if container == nil || len(container.Names) < 1 {
		return id
	}

	return container.Names[0]
}
