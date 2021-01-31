package main

import (
	"godtop/infrastructure"
	"godtop/interfaces"
	"log"
)

func main() {
	handler := interfaces.Handler{
		Service: infrastructure.Create(),
	}

	if err := handler.RunServer(8080); err != nil {
		log.Fatal(err)
	}
}
