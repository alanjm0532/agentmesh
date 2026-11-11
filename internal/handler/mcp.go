package handler

import (
	"context"
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (h *Handler) mcpHandler() http.Handler {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "gin-mcp-demo",
		Version: "1.0.0",
	}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hello",
		Description: "传入名字，返回问候语",
	}, h.callHello)

	// 将无状态 Streamable HTTP handler 交给 Gin 托管。
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
	})
}

func (h *Handler) callHello(_ context.Context, _ *mcp.CallToolRequest, input model.HelloInput) (*mcp.CallToolResult, model.HelloOutput, error) {
	return nil, h.helloService.Hello(input), nil
}
