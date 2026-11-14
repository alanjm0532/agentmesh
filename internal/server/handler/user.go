package handler

import (
	"errors"
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/internal/server/entity"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"github.com/alanjm0532/agentmesh/internal/server/util"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	userService  *service.UserService
	tokenService *service.TokenService
}

func NewHandler(userService *service.UserService, tokenService *service.TokenService) *Handler {
	return &Handler{userService: userService, tokenService: tokenService}
}

func (h *Handler) Login(c *gin.Context) {
	var input entity.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		util.InvalidRequest(c)
		return
	}

	user, err := h.userService.Login(c.Request.Context(), input.Username, input.Password)
	if errors.Is(err, entity.ErrPasswordTooLong) {
		util.InvalidRequest(c)
		return
	}
	if errors.Is(err, entity.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid username or password"})
		return
	}
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to log in"})
		return
	}
	if err = saveUserSession(c, user); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to save session"})
		return
	}
	c.JSON(http.StatusOK, profileResponse(user))
}

func (h *Handler) Profile(c *gin.Context) {
	user, err := h.userService.Profile(c.Request.Context(), c.MustGet("userID").(uuid.UUID))
	if ent.IsNotFound(err) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Not logged in"})
		return
	}
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to load profile"})
		return
	}
	c.JSON(http.StatusOK, profileResponse(user))
}

func (h *Handler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to log out"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func saveUserSession(c *gin.Context, user *ent.User) error {
	session := sessions.Default(c)
	session.Clear()
	// Keep credentials and access tokens out of the signed cookie payload.
	session.Set("user", user.ID.String())
	return session.Save()
}

func profileResponse(user *ent.User) entity.ProfileResponse {
	return entity.ProfileResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: user.CreatedAt.UTC(),
		UpdatedAt: user.UpdatedAt.UTC(),
	}
}
