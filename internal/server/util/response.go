package util

import (
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/alanjm0532/agentmesh/internal/server/entity"
	"github.com/gin-gonic/gin"
)

func InvalidRequest(c *gin.Context) {
	c.JSON(http.StatusBadRequest, protocol.ErrorResponse{Error: entity.ErrInvalidRequest.Error(), Message: "Invalid request parameters"})
}
