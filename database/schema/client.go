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

type ClientDevice struct {
	ent.Schema
}

func (ClientDevice) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_clients"},
	}
}

func (ClientDevice) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (ClientDevice) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(100).
			Comment("设备名称"),
		field.String("code").
			NotEmpty().
			MinLen(64).
			MaxLen(64).
			Comment("设备标识字符串(生成)"),
		field.String("description").
			Optional().
			MaxLen(255).
			Comment("备注"),
		field.Bool("enabled").
			Default(true).
			Comment("是否启用"),
		field.Uint64("access_token_expiry").
			Positive().
			Min(1000).
			Comment("accessToken超时时间(ms)"),
		field.Uint64("refresh_token_expiry").
			Positive().
			Min(1000).
			Comment("refreshToken超时时间(ms)"),
		field.Bool("anonymous").
			Default(true).
			Comment("允许所有角色登录"),
	}
}

func (ClientDevice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code", "delete_time").Unique(),
	}
}

func (ClientDevice) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联多个角色，被关联的角色允许登录该终端设备
		edge.To("roles", Role.Type),
	}
}
