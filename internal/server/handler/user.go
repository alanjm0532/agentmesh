package handler

import (
	"errors"
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/pkg/protocol"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) register(c *gin.Context) {
	var input protocol.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid registration request"})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), input.Username, input.Password)
	if errors.Is(err, bcrypt.ErrPasswordTooLong) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Password must not exceed 72 bytes"})
		return
	}
	if ent.IsConstraintError(err) {
		c.JSON(http.StatusConflict, gin.H{"message": "User already exists"})
		return
	}
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to register user"})
		return
	}

	// Return an explicit response so password hashes never leave the server.
	c.JSON(http.StatusCreated, protocol.RegisterResponse{
		ID:          user.ID.String(),
		Username:    user.Username,
		AccessToken: user.AccessToken,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}
