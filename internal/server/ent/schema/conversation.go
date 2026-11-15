package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/alanjm0532/agentmesh/internal/server/util"
	"github.com/google/uuid"
)

type Conversation struct{ ent.Schema }

func (Conversation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }).Immutable(),
		field.String("type").Default("direct").Immutable(),
		field.String("direct_key").MaxLen(73).Optional().Nillable().Unique().Immutable(),
		field.Int64("last_seq").Default(0).NonNegative(),
		field.Time("created_at").Default(util.NowUTC).Immutable(),
	}
}

func (Conversation) Edges() []ent.Edge {
	return []ent.Edge{edge.To("members", ConversationMember.Type), edge.To("messages", Message.Type)}
}
