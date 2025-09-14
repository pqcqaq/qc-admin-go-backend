package schema

import (
	"go-backend/database/events"
	"go-backend/database/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// APIAuth API认证实体表
type APIAuth struct {
	ent.Schema
}

func (APIAuth) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_api_auth"},
	}
}

func (APIAuth) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
		events.EventMixin{},
	}
}

func (APIAuth) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Comment("API名称"),
		field.String("description").Optional().Comment("API描述"),
		field.String("method").NotEmpty().Comment("HTTP方法"),
		field.String("path").NotEmpty().Comment("API路径"),
		field.Bool("is_public").Default(false).Comment("是否为公开API，true表示允许所有请求通过"),
		field.Bool("is_active").Default(true).Comment("是否启用"),
		field.JSON("metadata", map[string]interface{}{}).Optional().Comment("额外的元数据信息"),
	}
}

func (APIAuth) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("method", "path", "delete_time").Unique(),
		index.Fields("name", "delete_time").Unique(),
		index.Fields("is_public"),
		index.Fields("is_active"),
		index.Fields("method"),
		index.Fields("path"),
	}
}

func (APIAuth) Edges() []ent.Edge {
	return []ent.Edge{
		// 一个API可以关联多个Permission
		edge.To("permissions", Permission.Type),
	}
}
