package schema

import (
	"go-backend/database/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type VerifyCode struct {
	ent.Schema
}

func (VerifyCode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_verify_codes"},
	}
}

func (VerifyCode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

// Fields of the SysCredentials.
func (VerifyCode) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			MaxLen(10).
			Comment("验证码"),
		field.String("identifier").
			MaxLen(255).
			Comment("标识符（手机号/邮箱等）"),
		field.Enum("sender_type").
			Values("email", "phone", "sms").
			Comment("验证码类型"),
		field.String("send_for").
			MaxLen(50).
			Comment("发送目的"),
		field.Time("expires_at").
			Comment("过期时间"),
		field.Time("used_at").
			Optional().
			Nillable().
			Comment("使用时间"),
		field.Bool("send_success").
			Default(false).
			Comment("发送是否成功"),
		field.Time("send_at").
			Optional().
			Nillable().
			Comment("发送时间"),
	}
}

func (VerifyCode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("identifier", "sender_type", "send_for", "delete_time").Unique(), // 唯一索引，防止重复发送相同标识符的验证码
	}
}
