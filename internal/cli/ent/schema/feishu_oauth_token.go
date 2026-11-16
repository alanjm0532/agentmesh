package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type FeishuOAuthToken struct{ ent.Schema }

func (FeishuOAuthToken) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "feishu_oauth_tokens"}}
}

func (FeishuOAuthToken) Fields() []ent.Field {
	return []ent.Field{
		// 配置文件作为授权身份的命名空间，多个配置可以共用 CLI 数据库。
		field.String("profile"),
		field.String("app_id"),
		field.Text("access_token").Sensitive(),
		field.Text("refresh_token").Sensitive(),
		field.Int64("expires_at"),
		field.Int64("refresh_expires_at"),
		field.Text("scope"),
	}
}

func (FeishuOAuthToken) Indexes() []ent.Index {
	return []ent.Index{index.Fields("profile", "app_id").Unique()}
}
