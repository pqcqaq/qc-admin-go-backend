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

// Role 角色表
type Role struct {
	ent.Schema
}

func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_roles"},
	}
}

func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
		events.EventMixin{}, // 添加事件驱动mixin
	}
}

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
	}
}

func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		// 角色 <-> 用户 (多对多)
		edge.To("user_roles", UserRole.Type),

		// 角色 <-> 权限 (多对多)
		edge.To("role_permissions", RolePermission.Type),

		// 角色继承 (多继承) self-referencing
		edge.To("inherits_from", Role.Type).
			From("inherited_by"),

		// 可以被多个client_devices引用
		edge.From("client_device", ClientDevice.Type).Ref("roles"),
	}
}

func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "delete_time").Unique(), // 角色名称搜索
	}
}

// Permission 权限表
type Permission struct {
	ent.Schema
}

func (Permission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_permissions"},
	}
}

func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("action").NotEmpty(), // 比如 aaa.read/bbb.write/ccc.delete
		field.String("description").Optional(),
		field.Bool("is_public").Default(false), // 是否公开权限，任何用户都将拥有这些权限
	}
}

func (Permission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("action"),                       // 按操作类型查询
		index.Fields("name", "delete_time").Unique(), // 权限名称搜索
	}
}

func (Permission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("role_permissions", RolePermission.Type),
		// 权限包含一个 Scope (菜单/页面/按钮)
		edge.To("scope", Scope.Type).
			Unique(),
		// 一个权限可以被多个APIAuth引用
		edge.From("api_auths", APIAuth.Type).
			Ref("permissions"),
	}
}

// UserRole 用户-角色关联表
type UserRole struct {
	ent.Schema
}

func (UserRole) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_user_role"},
	}
}

func (UserRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (UserRole) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("user_id"),
		field.Uint64("role_id"),
	}
}

func (UserRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "role_id", "delete_time").Unique(), // 防重复用户角色（核心索引）
		index.Fields("user_id"), // 查询用户的所有角色
		index.Fields("role_id"), // 查询角色下的所有用户
	}
}

func (UserRole) Edges() []ent.Edge {
	return []ent.Edge{
		// 每个 UserRole 指向一个 User
		edge.From("user", User.Type).
			Ref("user_roles").
			Field("user_id").
			Unique().
			Required(),

		// 每个 UserRole 指向一个 Role
		edge.From("role", Role.Type).
			Ref("user_roles").
			Field("role_id").
			Unique().
			Required(),
	}
}

// RolePermission 角色-权限关联表
type RolePermission struct {
	ent.Schema
}

func (RolePermission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_role_permission"},
	}
}

func (RolePermission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (RolePermission) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("role_id"),
		field.Uint64("permission_id"),
	}
}

func (RolePermission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id", "permission_id", "delete_time").Unique(), // 防重复角色权限（核心索引）
		index.Fields("role_id"),       // 查询角色的所有权限
		index.Fields("permission_id"), // 查询权限被哪些角色使用
	}
}

func (RolePermission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("role", Role.Type).
			Ref("role_permissions").
			Field("role_id").
			Unique().
			Required(),

		edge.From("permission", Permission.Type).
			Ref("role_permissions").
			Field("permission_id").
			Unique().
			Required(),
	}
}
