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

// Scope 权限域（菜单、页面、按钮）
type Scope struct {
	ent.Schema
}

// Annotations 返回 Scope 的注解
func (Scope) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_scopes"},
	}
}

func (Scope) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Scope) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(), // 名称
		field.Enum("type").
			Values("menu", "page", "button").
			Default("menu").
			Comment("权限域类型"),
		field.String("icon").Optional(),        // 图标
		field.String("description").Optional(), // 描述
		field.String("action").Optional(),      // 操作（如：read/write/delete）
		field.String("path").Optional(),        // 路径（菜单路由/页面URL）
		field.String("component").Optional(),   // 组件（如：Vue组件路径）
		field.String("redirect").Optional(),    // 重定向路径
		field.Int("order").Default(0).Comment("排序"),
		field.Bool("hidden").Default(false).Comment("是否隐藏"),
		field.Bool("disabled").Default(false).Comment("是否禁用"),
		field.Uint64("parent_id").Optional(), // 父级ID，用于层级菜单
	}
}

func (Scope) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id", "order"),           // 有序菜单构建
		index.Fields("type", "hidden", "disabled"),   // 菜单渲染筛选
		index.Fields("type", "parent_id"),            // 按类型分层查询
		index.Fields("name", "delete_time").Unique(), // 名称搜索
	}
}

func (Scope) Edges() []ent.Edge {
	return []ent.Edge{
		// 自引用关系：一个scope可以有多个子scope
		edge.To("children", Scope.Type).
			From("parent").
			Field("parent_id").
			Unique(), // parent关系是唯一的

		// 一个 Scope 属于一个权限
		edge.From("permission", Permission.Type).
			Ref("scope").
			Unique(),
	}
}
