package main

import (
	"log"

	"github.com/alanjm0532/agentmesh/configs"
	"github.com/alanjm0532/agentmesh/internal/handler"
	"github.com/alanjm0532/agentmesh/internal/service"
)

func main() {
	cfg := configs.Load()
	helloService := &service.HelloService{}
	router := handler.NewRouter(helloService)

	if err := router.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
