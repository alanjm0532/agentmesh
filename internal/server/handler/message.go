package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"unicode/utf8"

	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/alanjm0532/agentmesh/internal/server/ctxmeta"
	"github.com/alanjm0532/agentmesh/internal/server/entity"
	"github.com/alanjm0532/agentmesh/internal/server/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func (h *Handler) SendMessage(c *gin.Context, identity ctxmeta.Identity) {
	var input protocol.SendMessageRequest
	if !bindMessageJSON(c, &input) {
		return
	}
	result, err := h.messageService.Send(c, identity.UserID, input)
	if err != nil {
		messageError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) UnreadMessages(c *gin.Context, identity ctxmeta.Identity) {
	var input protocol.UnreadMessagesQuery
	if err := c.ShouldBindQuery(&input); err != nil {
		util.InvalidRequest(c)
		return
	}
	result, err := h.messageService.Unread(c, identity.UserID, input.From, input.Limit)
	if err != nil {
		messageError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) MessageHistory(c *gin.Context, identity ctxmeta.Identity) {
	var input protocol.MessageHistoryQuery
	if err := c.ShouldBindQuery(&input); err != nil {
		util.InvalidRequest(c)
		return
	}
	result, err := h.messageService.History(c, identity.UserID, input.Peer, input.Before, input.Limit)
	if err != nil {
		messageError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ReadMessages(c *gin.Context, identity ctxmeta.Identity) {
	var input protocol.ReadMessagesRequest
	if !bindMessageJSON(c, &input) {
		return
	}
	if err := h.messageService.Read(c, identity.UserID, input); err != nil {
		messageError(c, err)
		return
	}
	c.JSON(http.StatusOK, map[string]string{"message": "Read position updated"})
}

func bindMessageJSON(c *gin.Context, input any) bool {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil || !utf8.Valid(data) || json.Unmarshal(data, input) != nil || binding.Validator.ValidateStruct(input) != nil {
		util.InvalidRequest(c)
		return false
	}
	return true
}

func messageError(c *gin.Context, err error) {
	if util.HandleContextError(c, err) {
		return
	}
	status := http.StatusInternalServerError
	message := "Unable to process message request"
	code := "server_error"
	switch {
	case errors.Is(err, entity.ErrInvalidRequest):
		status, code, message = http.StatusBadRequest, "invalid_request", "Invalid message, cursor, or parameters; body must be UTF-8 without NUL, contain 1 to 131072 bytes, and use none/text/plain"
	case errors.Is(err, entity.ErrUserNotFound):
		status, code, message = http.StatusNotFound, "user_not_found", "Recipient not found"
	case errors.Is(err, entity.ErrConversationAccess):
		status, code, message = http.StatusForbidden, "conversation_access_denied", "Conversation access denied"
	case errors.Is(err, entity.ErrMessageConflict):
		status, code, message = http.StatusConflict, "message_conflict", "client_message_id was already used for a different message"
	default:
		_ = c.Error(err)
	}
	c.JSON(status, protocol.ErrorResponse{Error: code, Message: message})
}
