package schema

import (
	"go-backend/database/mixins"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Logging holds the schema definition for the Logging entity.
type Logging struct {
	ent.Schema
}

// Mixin returns User mixed-in fields.
func (Logging) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

// Fields of the User.
func (Logging) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("level").
			Values("debug", "info", "error", "warn", "fatal").
			Default("info"),
		// type
		field.Enum("type").
			Values("Error", "Panic", "manul").
			Default("manul"),
		// message
		field.String("message").
			NotEmpty().
			MaxLen(500),
		field.String("method").
			Optional().
			MaxLen(127),
		// path
		field.String("path").
			Optional().
			MaxLen(255),
		field.String("ip").
			Optional().
			MaxLen(45), // IPv6 support
		// query
		field.String("query").
			Optional().
			MaxLen(1000),
		//code
		field.Int("code").
			Optional().
			Positive(),
		// user_agent
		field.String("user_agent").
			Optional().
			MaxLen(512),
		// data
		field.JSON("data", map[string]any{}).
			Optional(),
		// stack
		field.String("stack").
			Optional().
			MaxLen(8192), // 8KB for stack trace
	}
}
