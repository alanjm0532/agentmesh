package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/alanjm0532/agentmesh/internal/server/util"
	"github.com/google/uuid"
)

type UserToken struct{ ent.Schema }

func (UserToken) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "users_token"}}
}

func (UserToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }).Immutable(),
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable(),
		field.String("name").MaxLen(100).NotEmpty(),
		field.Enum("status").Values("pending", "approved", "active", "denied", "revoked").Default("pending"),
		field.String("device_code_hash").MaxLen(64).Optional().Nillable().Unique().Sensitive(),
		field.String("user_code").MaxLen(9).Optional().Nillable().Unique().Sensitive(),
		field.String("token_hash").MaxLen(64).Optional().Nillable().Unique().Sensitive(),
		field.Time("expires_at"),
		field.Time("created_at").Default(util.NowUTC).Immutable(),
	}
}

func (UserToken) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("tokens").Field("user_id").Unique()}
}

func (UserToken) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "status"), index.Fields("expires_at")}
}
