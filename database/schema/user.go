package schema

import (
	"go-backend/database/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_users"},
	}
}

// Mixin returns User mixed-in fields.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
		// 用户对象所包含的图片
		MultiAttachmentsMixin{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(100),
		field.Int("age").
			Positive().
			Optional(),
		field.Enum("sex").
			Values("male", "female", "other").
			Default("other").
			Comment("性别"),
		field.Enum("status").
			Values("active", "inactive", "banned").
			Default("active").
			Comment("用户状态"),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user_roles", UserRole.Type),    // 一对多: 一个用户对应多个 user_roles
		edge.To("credentials", Credential.Type), // 一对多: 一个用户对应多个 credentials
	}
}
