package service

import "github.com/alanjm0532/agentmesh/internal/model"

// HelloService 提供 HTTP 和 MCP 共用的问候逻辑。
type HelloService struct{}

func (s *HelloService) Hello(input model.HelloInput) model.HelloOutput {
	return model.HelloOutput{Message: "Hello, " + input.Name + "!"}
}
