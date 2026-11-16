package service

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/internal/server/entity"
	"github.com/alanjm0532/agentmesh/internal/server/repository"
	"github.com/google/uuid"
)

type MessageService struct {
	messages *repository.MessageRepository
	users    *repository.UserRepository
}

func NewMessageService(messages *repository.MessageRepository, users *repository.UserRepository) *MessageService {
	return &MessageService{messages: messages, users: users}
}

func (s *MessageService) Send(ctx context.Context, sender uuid.UUID, input protocol.SendMessageRequest) (protocol.Message, error) {
	clientID, err := uuid.Parse(input.ClientMessageID)
	if err != nil || clientID == uuid.Nil || !protocol.ValidMessageBody(input.Body) || input.EncryptionAlgorithm != protocol.EncryptionAlgorithmNone || input.ContentType != "text/plain" {
		return protocol.Message{}, entity.ErrInvalidRequest
	}
	recipient, err := s.peer(ctx, input.Recipient)
	if err != nil {
		return protocol.Message{}, err
	}
	senderUser, err := s.users.GetByID(ctx, sender)
	if err != nil {
		return protocol.Message{}, err
	}
	msg, err := s.messages.Send(ctx, sender, recipient.ID, clientID, input)
	if err != nil {
		return protocol.Message{}, err
	}
	result := messageResponse(msg)
	result.SenderUsername = senderUser.Username
	return result, nil
}

func (s *MessageService) peer(ctx context.Context, name string) (*ent.User, error) {
	u, err := s.users.GetByUsername(ctx, name)
	if ent.IsNotFound(err) {
		return nil, entity.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func emptyUnread() protocol.UnreadMessagesResponse {
	return protocol.UnreadMessagesResponse{Messages: []protocol.Message{}, ReadPositions: []protocol.ReadPosition{}}
}

func (s *MessageService) Unread(ctx context.Context, self uuid.UUID, from string, limit int) (protocol.UnreadMessagesResponse, error) {
	page := emptyUnread()
	if limit < 1 || limit > protocol.MaxMessagePageSize {
		return page, entity.ErrInvalidRequest
	}
	var conversationID *uuid.UUID
	var selfConversation bool
	if from != "" {
		peer, err := s.peer(ctx, from)
		if err != nil {
			return page, err
		}
		cv, err := s.messages.Direct(ctx, self, peer.ID)
		if ent.IsNotFound(err) {
			return page, nil
		}
		if err != nil {
			return page, err
		}
		conversationID = &cv.ID
		selfConversation = peer.ID == self
	}
	rows, err := s.messages.Unread(ctx, self, conversationID, selfConversation, limit+1)
	if err != nil {
		return page, err
	}
	page.HasMore = len(rows) > limit
	if page.HasMore {
		rows = rows[:limit]
	}
	positions := make(map[string]int)
	for _, row := range rows {
		page.Messages = append(page.Messages, messageResponse(row))
		key := row.ConversationID.String()
		if i, ok := positions[key]; ok {
			page.ReadPositions[i].LastReadSeq = row.Seq
		} else {
			positions[key] = len(page.ReadPositions)
			page.ReadPositions = append(page.ReadPositions, protocol.ReadPosition{ConversationID: key, LastReadSeq: row.Seq})
		}
	}
	return page, nil
}

type historyCursor struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	Before         int64     `json:"before"`
}

func (s *MessageService) History(ctx context.Context, self uuid.UUID, name, before string, limit int) (protocol.MessageHistoryResponse, error) {
	page := protocol.MessageHistoryResponse{Messages: []protocol.Message{}}
	if limit < 1 || limit > protocol.MaxMessagePageSize {
		return page, entity.ErrInvalidRequest
	}
	peer, err := s.peer(ctx, name)
	if err != nil {
		return page, err
	}
	var cursor historyCursor
	if before != "" {
		if len(before) > 256 {
			return page, entity.ErrInvalidRequest
		}
		data, err := base64.RawURLEncoding.DecodeString(before)
		if err != nil || json.Unmarshal(data, &cursor) != nil || cursor.Before < 1 || cursor.ConversationID == uuid.Nil {
			return page, entity.ErrInvalidRequest
		}
	}
	cv, err := s.messages.Direct(ctx, self, peer.ID)
	if ent.IsNotFound(err) {
		if before != "" {
			return page, entity.ErrInvalidRequest
		}
		return page, nil
	}
	if err != nil {
		return page, err
	}
	if before != "" && cursor.ConversationID != cv.ID {
		return page, entity.ErrInvalidRequest
	}
	rows, err := s.messages.History(ctx, self, cv.ID, cursor.Before, limit+1)
	if err != nil {
		return page, err
	}
	page.HasMore = len(rows) > limit
	if page.HasMore {
		rows = rows[:limit]
	}
	for _, row := range rows {
		page.Messages = append(page.Messages, messageResponse(row))
	}
	if page.HasMore {
		data, err := json.Marshal(historyCursor{ConversationID: cv.ID, Before: rows[len(rows)-1].Seq})
		if err != nil {
			return page, err
		}
		page.NextCursor = base64.RawURLEncoding.EncodeToString(data)
	}
	return page, nil
}

func (s *MessageService) Read(ctx context.Context, self uuid.UUID, input protocol.ReadMessagesRequest) error {
	if len(input.Reads) == 0 || len(input.Reads) > protocol.MaxMessagePageSize {
		return entity.ErrInvalidRequest
	}
	positions := make([]repository.ReadCursor, 0, len(input.Reads))
	for _, position := range input.Reads {
		id, err := uuid.Parse(position.ConversationID)
		if err != nil || id == uuid.Nil || position.LastReadSeq < 0 {
			return entity.ErrInvalidRequest
		}
		positions = append(positions, repository.ReadCursor{ConversationID: id, Seq: position.LastReadSeq})
	}
	return s.messages.Read(ctx, self, positions)
}

func messageResponse(m *ent.Message) protocol.Message {
	name := ""
	if m.Edges.Sender != nil {
		name = m.Edges.Sender.Username
	}
	return protocol.Message{ID: m.ID.String(), ConversationID: m.ConversationID.String(), Seq: m.Seq, SenderID: m.SenderID.String(), SenderUsername: name, ClientMessageID: m.ClientMessageID.String(), EncryptionAlgorithm: protocol.EncryptionAlgorithm(m.EncryptionAlgorithm), ContentType: m.ContentType, Body: m.Body, CreatedAt: m.CreatedAt.UTC()}
}
