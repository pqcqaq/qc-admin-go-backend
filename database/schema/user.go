package schema

import (
	"go-backend/database/mixins"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Mixin returns User mixed-in fields.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
		// 用户对象所包含的图片
		AttachmentsMixin{},
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
