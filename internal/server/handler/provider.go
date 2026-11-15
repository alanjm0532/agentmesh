package handler

import (
	"context"
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewHandler,
)

func (h *Handler) APIPing(ctx context.Context, c *gin.Context) {
	c.JSON(http.StatusOK, protocol.PingResponse{Message: "pong"})
}
