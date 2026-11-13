package handler

import (
	"errors"
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) login(c *gin.Context) {
	var input LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid login request"})
		return
	}

	user, err := h.userService.Login(c.Request.Context(), input.Username, input.Password)
	if errors.Is(err, bcrypt.ErrPasswordTooLong) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Password must not exceed 72 bytes"})
		return
	}
	if errors.Is(err, service.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid username or password"})
		return
	}
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to log in"})
		return
	}
	if err := saveUserSession(c, user); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to save session"})
		return
	}
	c.JSON(http.StatusOK, profileResponse(user))
}

func (h *Handler) profile(c *gin.Context) {
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

func (h *Handler) logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to log out"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func (h *Handler) token(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	user, err := h.userService.Profile(c.Request.Context(), c.MustGet("userID").(uuid.UUID))
	if ent.IsNotFound(err) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Not logged in"})
		return
	}
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to load token"})
		return
	}
	c.JSON(http.StatusOK, TokenResponse{AccessToken: user.AccessToken})
}

func (h *Handler) resetToken(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	user, err := h.userService.ResetAccessToken(c.Request.Context(), c.MustGet("userID").(uuid.UUID))
	if ent.IsNotFound(err) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Not logged in"})
		return
	}
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to reset token"})
		return
	}
	c.JSON(http.StatusOK, TokenResponse{AccessToken: user.AccessToken})
}

func saveUserSession(c *gin.Context, user *ent.User) error {
	session := sessions.Default(c)
	session.Clear()
	// Keep credentials and access tokens out of the signed cookie payload.
	session.Set("user", user.ID.String())
	return session.Save()
}

func profileResponse(user *ent.User) ProfileResponse {
	return ProfileResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: user.CreatedAt.UTC(),
		UpdatedAt: user.UpdatedAt.UTC(),
	}
}
