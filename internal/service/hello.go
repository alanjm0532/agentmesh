package service

import (
	"github.com/alanjm0532/agentmesh/internal/model"
	"go.uber.org/zap"
)

// HelloService 提供 HTTP 和 MCP 共用的问候逻辑。
type HelloService struct {
	logger *zap.Logger
}

func NewHelloService(logger *zap.Logger) *HelloService {
	return &HelloService{logger: logger.Named("hello")}
}

func (s *HelloService) Hello(input model.HelloInput) model.HelloOutput {
	s.logger.Info("HelloService.Hello", zap.String("name", input.Name))
	return model.HelloOutput{Message: "Hello, " + input.Name + "!"}
}
