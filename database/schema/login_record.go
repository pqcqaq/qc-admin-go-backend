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

// LoginRecord holds the schema definition for the LoginRecord entity.
type LoginRecord struct {
	ent.Schema
}

func (LoginRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_login_records"},
	}
}

// Mixin returns LoginRecord mixed-in fields.
func (LoginRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

// Fields of the LoginRecord.
func (LoginRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("user_id").
			Comment("用户ID"),
		field.String("identifier").
			MaxLen(255).
			Comment("登录标识符(用户名/邮箱/手机号等)"),
		field.Enum("credential_type").
			Values("password", "email", "oauth", "phone", "totp").
			Comment("登录方式"),
		field.String("ip_address").
			MaxLen(45).
			Comment("登录IP地址(支持IPv6)"),
		field.String("user_agent").
			MaxLen(512).
			Optional().
			Comment("用户代理信息"),
		field.String("device_info").
			MaxLen(255).
			Optional().
			Comment("设备信息"),
		field.String("location").
			MaxLen(255).
			Optional().
			Comment("登录地点"),
		field.Enum("status").
			Values("success", "failed", "locked").
			Comment("登录状态"),
		field.String("failure_reason").
			MaxLen(255).
			Optional().
			Comment("失败原因"),
		field.String("session_id").
			MaxLen(128).
			Optional().
			Comment("会话ID"),
		field.Time("logout_time").
			Optional().
			Nillable().
			Comment("退出时间"),
		field.Int("duration").
			Optional().
			Comment("会话持续时间(秒)"),
		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("额外元数据"),
		field.Uint64("client_id").
			Optional().
			Nillable().
			Comment("请求设备ID"),
	}
}

// Edges of the LoginRecord.
func (LoginRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("login_records").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes of the LoginRecord.
func (LoginRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "create_time"),
		index.Fields("status", "create_time"),
		index.Fields("credential_type", "create_time"),
		index.Fields("ip_address", "create_time"),
		index.Fields("session_id"),
	}
}
