package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		// Use UUIDv7 for time-ordered primary keys across all supported databases.
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }).
			Immutable(),
		// Let the database enforce uniqueness, including concurrent registrations.
		field.String("username").Unique(),
		field.String("password").Sensitive(),
		field.String("access_token").Unique().Sensitive(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}
