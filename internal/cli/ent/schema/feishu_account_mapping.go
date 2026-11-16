package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type FeishuAccountMapping struct{ ent.Schema }

func (FeishuAccountMapping) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "feishu_account_mappings"}}
}

func (FeishuAccountMapping) Fields() []ent.Field {
	return []ent.Field{
		field.String("server"),
		field.String("app_id"),
		field.String("username"),
		field.String("email"),
		// 绑定更新时保留创建时间，仅更新修改时间。
		field.Time("created_at").Default(func() time.Time { return time.Now().UTC() }).Immutable(),
		field.Time("updated_at").Default(func() time.Time { return time.Now().UTC() }).
			UpdateDefault(func() time.Time { return time.Now().UTC() }),
	}
}

func (FeishuAccountMapping) Indexes() []ent.Index {
	return []ent.Index{index.Fields("server", "app_id", "username").Unique()}
}
