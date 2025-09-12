package schema

import (
	"go-backend/database/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Credential holds the schema definition for the Credential entity.
type Credential struct {
	ent.Schema
}

func (Credential) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_credentials"},
	}
}

// Mixin returns SysCredentials mixed-in fields.
func (Credential) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

// Fields of the SysCredentials.
func (Credential) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("user_id").
			Comment("用户ID"),
		field.Enum("credential_type").
			Values("password", "email", "oauth", "phone", "totp").
			Comment("认证类型"),
		field.String("identifier").
			MaxLen(255).
			Comment("认证标识符(用户名/邮箱/手机号/OAuth Provider ID等)"),
		field.String("secret").
			MaxLen(500).
			Optional().
			Sensitive().
			Comment("认证密钥(密码hash/token等)"),
		field.String("salt").
			MaxLen(100).
			Optional().
			Sensitive().
			Comment("密码盐值"),
		field.String("provider").
			MaxLen(50).
			Optional().
			Comment("OAuth提供商(google/github/wechat等)"),
		field.Bool("is_verified").
			Default(false).
			Comment("是否已验证"),
		field.Time("verified_at").
			Optional().
			Nillable().
			Comment("验证时间"),
		field.Time("last_used_at").
			Optional().
			Nillable().
			Comment("最后使用时间"),
		field.Time("expires_at").
			Optional().
			Nillable().
			Comment("过期时间"),
		field.Int("failed_attempts").
			Default(0).
			Comment("失败尝试次数"),
		field.Time("locked_until").
			Optional().
			Nillable().
			Comment("锁定到期时间"),
		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("额外信息"),
	}
}

// Edges of the SysCredentials.
func (Credential) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("credentials").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes of the SysCredentials.
func (Credential) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "credential_type", "delete_time"),
		index.Fields("credential_type", "identifier", "delete_time"),
		index.Fields("provider", "identifier", "delete_time"),
		index.Fields("is_verified", "credential_type", "delete_time"),
		index.Fields("expires_at", "delete_time"),
	}
}
