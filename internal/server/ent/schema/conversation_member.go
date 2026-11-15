package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/alanjm0532/agentmesh/internal/server/util"
	"github.com/google/uuid"
)

type ConversationMember struct{ ent.Schema }

func (ConversationMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }).Immutable(),
		field.UUID("conversation_id", uuid.UUID{}).Immutable(),
		field.UUID("user_id", uuid.UUID{}).Immutable(),
		field.Int64("last_read_seq").Default(0).NonNegative(),
		field.Time("joined_at").Default(util.NowUTC).Immutable(),
	}
}

func (ConversationMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("conversation", Conversation.Type).Ref("members").Field("conversation_id").Unique().Required().Immutable(),
		edge.From("user", User.Type).Ref("conversation_memberships").Field("user_id").Unique().Required().Immutable(),
	}
}

func (ConversationMember) Indexes() []ent.Index {
	return []ent.Index{index.Fields("conversation_id", "user_id").Unique(), index.Fields("user_id", "conversation_id")}
}
