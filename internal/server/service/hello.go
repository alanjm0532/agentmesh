package service

import (
	"github.com/alanjm0532/agentmesh/pkg/protocol"
	"go.uber.org/zap"
)

type HelloService struct {
	logger *zap.Logger
}

func NewHelloService(logger *zap.Logger) *HelloService {
	return &HelloService{logger: logger.Named("hello")}
}

func (s *HelloService) Hello(input protocol.HelloRequest) protocol.HelloResponse {
	s.logger.Info("HelloService.Hello", zap.String("name", input.Name))
	return protocol.HelloResponse{Message: "Hello, " + input.Name + "!"}
}
