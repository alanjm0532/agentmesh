package main

import (
	"github.com/alanjm0532/agentmesh/configs"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config configs.Config
	router *gin.Engine
}

func newServer(config configs.Config, router *gin.Engine) *Server {
	return &Server{config: config, router: router}
}

func (s *Server) Run() error {
	return s.router.Run(s.config.Server.Addr)
}
