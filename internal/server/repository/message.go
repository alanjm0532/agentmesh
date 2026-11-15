package repository

import (
	"context"
	"math"
	"sort"

	"entgo.io/ent/dialect/sql"
	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/internal/server/ent/conversation"
	"github.com/alanjm0532/agentmesh/internal/server/ent/conversationmember"
	"github.com/alanjm0532/agentmesh/internal/server/ent/message"
	"github.com/alanjm0532/agentmesh/internal/server/ent/user"
	"github.com/alanjm0532/agentmesh/internal/server/entity"
	"github.com/alanjm0532/agentmesh/internal/server/util"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MessageRepository struct {
	client *ent.Client
	logger *zap.Logger
}

func NewMessageRepository(client *ent.Client, logger *zap.Logger) *MessageRepository {
	return &MessageRepository{client: client, logger: logger.Named("message_repository")}
}

func (r *MessageRepository) Direct(ctx context.Context, a, b uuid.UUID) (*ent.Conversation, error) {
	return r.client.Conversation.Query().Where(conversation.DirectKeyEQ(util.DirectKey(a, b)), conversation.TypeEQ("direct")).Only(ctx)
}

func (r *MessageRepository) Send(ctx context.Context, sender, recipient, clientID uuid.UUID, input protocol.SendMessageRequest) (*ent.Message, error) {
	var result *ent.Message
	err := util.Transaction(ctx, r.client, r.logger, func(tx *ent.Tx) error {
		existing, err := tx.Message.Query().Where(message.SenderIDEQ(sender), message.ClientMessageIDEQ(clientID)).Only(ctx)
		if err == nil {
			cv, err := tx.Conversation.Get(ctx, existing.ConversationID)
			if err != nil {
				return err
			}
			if cv.DirectKey == nil || *cv.DirectKey != util.DirectKey(sender, recipient) || existing.Body != input.Body || existing.EncryptionAlgorithm != input.EncryptionAlgorithm.String() || existing.ContentType != input.ContentType {
				return entity.ErrMessageConflict
			}
			result = existing
			return nil
		}
		if !ent.IsNotFound(err) {
			return err
		}
		cv, err := tx.Conversation.Query().Where(conversation.DirectKeyEQ(util.DirectKey(sender, recipient)), conversation.TypeEQ("direct")).Only(ctx)
		if ent.IsNotFound(err) {
			cv, err = tx.Conversation.Create().SetDirectKey(util.DirectKey(sender, recipient)).Save(ctx)
			if err != nil {
				return err
			}
			_, err = tx.ConversationMember.CreateBulk(
				tx.ConversationMember.Create().SetConversationID(cv.ID).SetUserID(sender),
				tx.ConversationMember.Create().SetConversationID(cv.ID).SetUserID(recipient),
			).Save(ctx)
		}
		if err != nil {
			return err
		}
		// UPDATE takes the write lock before reading the allocated sequence. Keeping
		// it until commit prevents a later sequence from becoming visible first.
		n, err := tx.Conversation.Update().Where(conversation.IDEQ(cv.ID), conversation.LastSeqLT(math.MaxInt64)).AddLastSeq(1).Save(ctx)
		if err != nil {
			return err
		}
		if n != 1 {
			return entity.ErrInvalidRequest
		}
		cv, err = tx.Conversation.Get(ctx, cv.ID)
		if err != nil {
			return err
		}
		result, err = tx.Message.Create().SetConversationID(cv.ID).SetSeq(cv.LastSeq).SetSenderID(sender).SetClientMessageID(clientID).
			SetEncryptionAlgorithm(input.EncryptionAlgorithm.String()).SetContentType(input.ContentType).SetBody(input.Body).Save(ctx)
		return err
	})
	return result, err
}

func (r *MessageRepository) Unread(ctx context.Context, userID uuid.UUID, conversationID *uuid.UUID, limit int) ([]*ent.Message, error) {
	q := r.client.Message.Query().Where(message.SenderIDNEQ(userID), func(s *sql.Selector) {
		members := sql.Table(conversationmember.Table)
		s.Join(members).On(s.C(message.FieldConversationID), members.C(conversationmember.FieldConversationID))
		s.Where(sql.And(sql.EQ(members.C(conversationmember.FieldUserID), userID), sql.ColumnsGT(s.C(message.FieldSeq), members.C(conversationmember.FieldLastReadSeq))))
	})
	if conversationID != nil {
		q.Where(message.ConversationIDEQ(*conversationID))
	}
	// Every conversation's rows must appear in sequence order, regardless of
	// clock skew between server instances. No wall-clock cursor is used.
	return q.WithSender(func(q *ent.UserQuery) { q.Select(user.FieldID, user.FieldUsername) }).Order(ent.Asc(message.FieldSeq), ent.Asc(message.FieldID)).Limit(limit).All(ctx)
}

func (r *MessageRepository) History(ctx context.Context, userID, conversationID uuid.UUID, before int64, limit int) ([]*ent.Message, error) {
	q := r.client.Message.Query().Where(message.ConversationIDEQ(conversationID), message.HasConversationWith(conversation.HasMembersWith(conversationmember.UserIDEQ(userID))))
	if before > 0 {
		q.Where(message.SeqLT(before))
	}
	return q.WithSender(func(q *ent.UserQuery) { q.Select(user.FieldID, user.FieldUsername) }).Order(ent.Desc(message.FieldSeq)).Limit(limit).All(ctx)
}

type ReadCursor struct {
	ConversationID uuid.UUID
	Seq            int64
}

func (r *MessageRepository) Read(ctx context.Context, userID uuid.UUID, positions []ReadCursor) error {
	// Use a stable lock order for concurrent batches.
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].ConversationID.String() < positions[j].ConversationID.String()
	})
	return util.Transaction(ctx, r.client, r.logger, func(tx *ent.Tx) error {
		for _, position := range positions {
			member, err := tx.ConversationMember.Query().Where(conversationmember.ConversationIDEQ(position.ConversationID), conversationmember.UserIDEQ(userID)).Only(ctx)
			if ent.IsNotFound(err) {
				return entity.ErrConversationAccess
			}
			if err != nil {
				return err
			}
			cv, err := tx.Conversation.Get(ctx, position.ConversationID)
			if err != nil {
				return err
			}
			if position.Seq > cv.LastSeq {
				return entity.ErrInvalidRequest
			}
			_, err = tx.ConversationMember.Update().Where(conversationmember.IDEQ(member.ID), conversationmember.LastReadSeqLT(position.Seq)).SetLastReadSeq(position.Seq).Save(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
