package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"fmt"
	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/alanjm0532/agentmesh/internal/server/util"
	"github.com/google/uuid"
)

type Message struct{ ent.Schema }

func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }).Immutable(),
		field.UUID("conversation_id", uuid.UUID{}).Immutable(),
		field.Int64("seq").Positive().Immutable(),
		field.UUID("sender_id", uuid.UUID{}).Immutable(),
		field.UUID("client_message_id", uuid.UUID{}).Immutable(),
		field.String("encryption_algorithm").MaxLen(100).Default("none").Immutable(),
		field.String("content_type").MaxLen(100).Default("text/plain").Immutable(),
		field.Text("body").SchemaType(map[string]string{dialect.MySQL: "longtext"}).Validate(func(s string) error {
			if !protocol.ValidMessageBody(s) {
				return fmt.Errorf("body must be valid UTF-8 without NUL and contain 1 to %d bytes", protocol.MaxMessageBytes)
			}
			return nil
		}).Sensitive().Immutable(),
		field.Time("created_at").Default(util.NowUTC).Immutable(),
	}
}

func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("conversation", Conversation.Type).Ref("messages").Field("conversation_id").Unique().Required().Immutable(),
		edge.From("sender", User.Type).Ref("messages").Field("sender_id").Unique().Required().Immutable(),
	}
}

func (Message) Indexes() []ent.Index {
	return []ent.Index{index.Fields("conversation_id", "seq").Unique(), index.Fields("sender_id", "client_message_id").Unique()}
}
