package main

import (
	"godtop/infrastructure"
	"godtop/interfaces"
	"log"
)

func main() {
	handler := interfaces.Handler{
		DockerService: infrastructure.CreateDockerService(),
		HostService:   infrastructure.CreateHostService(),
	}

	if err := handler.RunServer(8080); err != nil {
		log.Fatal(err)
	}
}
