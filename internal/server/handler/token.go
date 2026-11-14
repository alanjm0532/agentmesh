package handler

import (
	"errors"
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/alanjm0532/agentmesh/internal/server/entity"
	"github.com/alanjm0532/agentmesh/internal/server/middleware"
	"github.com/alanjm0532/agentmesh/internal/server/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) RequireToken() gin.HandlerFunc { return middleware.TokenAuth(h.tokenService) }

func (h *Handler) DeviceCode(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	var input protocol.DeviceCodeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		util.InvalidRequest(c)
		return
	}
	response, err := h.tokenService.StartLogin(c.Request.Context(), input.Name)
	if err != nil {
		tokenError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) DeviceToken(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	var input protocol.DeviceTokenRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		util.InvalidRequest(c)
		return
	}
	response, err := h.tokenService.Exchange(c.Request.Context(), input.DeviceCode)
	if err != nil {
		tokenError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) CLIRequest(c *gin.Context) {
	grant, err := h.tokenService.LoginRequest(c.Request.Context(), c.Query("user_code"))
	if err != nil {
		tokenError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": grant.Name, "user_code": grant.UserCode, "expires_at": grant.ExpiresAt.UTC()})
}

func (h *Handler) AuthorizeCLI(c *gin.Context) {
	var input entity.AuthorizeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		util.InvalidRequest(c)
		return
	}
	if err := h.tokenService.Authorize(c.Request.Context(), input.UserCode, c.MustGet("userID").(uuid.UUID), *input.Approve); err != nil {
		tokenError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Authorization updated"})
}

func (h *Handler) ListTokens(c *gin.Context) {
	tokens, err := h.tokenService.List(c.Request.Context(), c.MustGet("userID").(uuid.UUID))
	if err != nil {
		tokenError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

func (h *Handler) RenameToken(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.InvalidRequest(c)
		return
	}
	var input entity.RenameTokenRequest
	if err = c.ShouldBindJSON(&input); err != nil {
		util.InvalidRequest(c)
		return
	}
	ok, err := h.tokenService.Rename(c.Request.Context(), c.MustGet("userID").(uuid.UUID), id, input.Name)
	tokenUpdateResult(c, ok, err)
}

func (h *Handler) RevokeToken(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.InvalidRequest(c)
		return
	}
	ok, err := h.tokenService.Revoke(c.Request.Context(), c.MustGet("userID").(uuid.UUID), id)
	tokenUpdateResult(c, ok, err)
}

func (h *Handler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, protocol.IdentityResponse{ID: c.MustGet("userID").(uuid.UUID).String(), Username: c.GetString("username")})
}

func (h *Handler) RevokeCurrentToken(c *gin.Context) {
	ok, err := h.tokenService.Revoke(c.Request.Context(), c.MustGet("userID").(uuid.UUID), c.MustGet("tokenID").(uuid.UUID))
	tokenUpdateResult(c, ok, err)
}

func tokenUpdateResult(c *gin.Context, ok bool, err error) {
	if err != nil {
		tokenError(c, err)
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"message": "Token not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Token updated"})
}

func tokenError(c *gin.Context, err error) {
	for _, expected := range []error{entity.ErrAuthorizationPending, entity.ErrAccessDenied, entity.ErrExpiredToken, entity.ErrInvalidGrant} {
		if errors.Is(err, expected) {
			c.JSON(http.StatusBadRequest, protocol.ErrorResponse{Error: expected.Error(), Message: expected.Error()})
			return
		}
	}
	_ = c.Error(err)
	c.JSON(http.StatusInternalServerError, protocol.ErrorResponse{Error: "server_error", Message: "Unable to process token request"})
}
