package schema

import (
	"go-backend/database/mixins"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type Scan struct {
	ent.Schema
}

// Mixin returns User mixed-in fields.
func (Scan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
		// 对象所包含的图片
		AttachmentsMixin{},
	}
}

// Fields of the User.
func (Scan) Fields() []ent.Field {
	return []ent.Field{
		field.String("content").
			NotEmpty().
			MaxLen(255),
		field.Int("length").
			Positive().
			Comment("内容长度"),
		field.Bool("success").
			Comment("是否成功"),
	}
}
