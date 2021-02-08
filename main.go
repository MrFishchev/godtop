package main

import (
	"fmt"
	"godtop/consoleui"
)

// @title Godtop
// @version 1.0
// @description Docker graphical activity monitor

// @contact.name Aleksey Fishchev
// @contact.email mrfishchev@seniorvlogger.com

// @license.name MIT
// @license.url https://github.com/MrFishchev/godtop/blob/main/LICENSE

// @BasePath /api
func main() {
	// handler := interfaces.Handler{
	// 	DockerService: infrastructure.CreateDockerService(),
	// 	HostService:   infrastructure.CreateHostService(),
	// }

	// if err := handler.RunServer(8080); err != nil {
	// 	log.Fatal(err)
	// }

	err := consoleui.Run()
	if err != nil {
		fmt.Printf("Cannot run consoleui: %v", err.Error())
	}
}
