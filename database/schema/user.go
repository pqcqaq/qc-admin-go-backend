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
		field.String("email").
			NotEmpty().
			MaxLen(255).
			Unique(),
		field.Int("age").
			Positive().
			Optional(),
		field.String("phone").
			MaxLen(20).
			Optional(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user_roles", UserRole.Type), // 一对多: 一个用户对应多个 user_roles
	}
}
