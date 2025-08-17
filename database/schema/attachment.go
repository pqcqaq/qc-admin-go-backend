package schema

import (
	"go-backend/database/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// Attachment holds the schema definition for the Attachment entity.
type Attachment struct {
	ent.Schema
}

func (Attachment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_attachments"},
	}
}

// Mixin returns Attachment mixed-in fields.
func (Attachment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

// Fields of the Attachment.
func (Attachment) Fields() []ent.Field {
	return []ent.Field{
		field.String("filename").
			MaxLen(255).
			Comment("原始文件名"),
		field.String("path").
			MaxLen(500).
			Comment("文件存储路径"),
		field.String("url").
			MaxLen(500).
			Optional().
			Comment("文件访问URL"),
		field.String("content_type").
			MaxLen(100).
			Comment("文件MIME类型"),
		field.Int64("size").
			Comment("文件大小(字节)"),
		field.String("etag").
			MaxLen(100).
			Optional().
			Comment("文件ETag"),
		field.String("bucket").
			MaxLen(100).
			Comment("存储桶名称"),
		field.String("storage_provider").
			MaxLen(50).
			Default("s3").
			Comment("存储提供商"),
		field.JSON("metadata", map[string]any{}).
			Optional().
			Comment("附加元数据"),
		field.Enum("status").
			Values("uploading", "uploaded", "failed", "deleted").
			Default("uploading").
			Comment("文件状态"),
		field.String("upload_session_id").
			MaxLen(100).
			Optional().
			Comment("上传会话ID"),
		field.String("tag1").
			MaxLen(100).
			Optional().
			Comment("标签1"),
		field.String("tag2").
			MaxLen(100).
			Optional().
			Comment("标签2"),
		field.String("tag3").
			MaxLen(100).
			Optional().
			Comment("标签3"),
	}
}

// Edges of the Attachment.
func (Attachment) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Attachment.
func (Attachment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("path", "delete_time").Unique().StorageKey("idx_attachment_path"),
		index.Fields("filename").StorageKey("idx_attachment_filename"),
		index.Fields("content_type").StorageKey("idx_attachment_content_type"),
		index.Fields("status").StorageKey("idx_attachment_status"),
		index.Fields("bucket").StorageKey("idx_attachment_bucket"),
		index.Fields("delete_time", "create_time").StorageKey("idx_attachment_deleted_created"),
		index.Fields("upload_session_id").StorageKey("idx_attachment_session"),
		index.Fields("tag1").StorageKey("idx_attachment_tag1"),
		index.Fields("tag2").StorageKey("idx_attachment_tag2"),
		index.Fields("tag3").StorageKey("idx_attachment_tag3"),
	}
}

// BaseMixin 包含所有基础字段的mixin
type MultiAttachmentsMixin struct {
	mixin.Schema
}

func (MultiAttachmentsMixin) Fields() []ent.Field {
	return []ent.Field{}
}

func (MultiAttachmentsMixin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("attachments", Attachment.Type),
	}
}

type SingleAttachmentMixin struct {
	mixin.Schema
}

func (SingleAttachmentMixin) Fields() []ent.Field {
	return []ent.Field{}
}

func (SingleAttachmentMixin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("attachment", Attachment.Type).
			Unique(),
	}
}
