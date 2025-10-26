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

// OauthApplication OAuth 应用(系统作为OAuth提供方时注册的第三方客户端应用)
type OauthApplication struct {
	ent.Schema
}

func (OauthApplication) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth_applications"},
	}
}

func (OauthApplication) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
		events.EventMixin{},
	}
}

func (OauthApplication) Fields() []ent.Field {
	return []ent.Field{
		field.String("client_id").
			MaxLen(128).
			Unique().
			Immutable().
			Comment("客户端ID"),
		field.String("client_secret").
			MaxLen(512).
			Sensitive().
			Comment("客户端密钥"),
		field.String("name").
			MaxLen(64).
			NotEmpty().
			Comment("应用名称"),
		field.JSON("redirect_uris", []string{}).
			Comment("允许的重定向URI列表"),
		field.Bool("is_confidential").
			Default(true).
			Comment("是否为保密客户端"),
		field.JSON("scopes", []string{}).
			Comment("应用可请求的权限范围"),
		field.Enum("able_state").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("启用状态"),
		field.Uint64("system_id").
			Optional().
			Comment("所属系统ID"),
	}
}

func (OauthApplication) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("authorization_codes", OauthAuthorizationCode.Type).
			Comment("生成的授权码"),
		edge.To("tokens", OauthToken.Type).
			Comment("颁发的令牌"),
		edge.To("user_authorizations", OauthUserAuthorization.Type).
			Comment("用户授权记录"),
	}
}

func (OauthApplication) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("client_id"),
		index.Fields("name", "delete_time").Unique(),
	}
}

// OauthAuthorizationCode OAuth 授权码
type OauthAuthorizationCode struct {
	ent.Schema
}

func (OauthAuthorizationCode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth_authorization_codes"},
	}
}

func (OauthAuthorizationCode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (OauthAuthorizationCode) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			MaxLen(128).
			Unique().
			Comment("授权码(UUID)"),
		field.Uint64("application_id").
			Comment("关联的应用ID"),
		field.Uint64("user_id").
			Comment("关联的用户ID"),
		field.String("redirect_uri").
			MaxLen(512).
			Comment("重定向URI"),
		field.JSON("scope", []string{}).
			Comment("授权的权限范围"),
		field.Time("expires_at").
			Comment("过期时间(推荐10分钟)"),
		field.Time("used_at").
			Optional().
			Nillable().
			Comment("使用时间(一次性使用)"),
		field.String("code_challenge").
			MaxLen(128).
			Optional().
			Comment("PKCE代码挑战(预留)"),
		field.String("code_challenge_method").
			MaxLen(10).
			Optional().
			Comment("PKCE挑战方法(S256/plain)"),
	}
}

func (OauthAuthorizationCode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("application", OauthApplication.Type).
			Ref("authorization_codes").
			Field("application_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		edge.To("token", OauthToken.Type).
			Unique().
			Comment("换取的令牌"),
		edge.From("user_authorization", OauthUserAuthorization.Type).
			Ref("code").
			Unique(),
	}
}

func (OauthAuthorizationCode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code"),
		index.Fields("application_id", "user_id"),
		index.Fields("expires_at"),
	}
}

// OauthToken OAuth 令牌
type OauthToken struct {
	ent.Schema
}

func (OauthToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth_tokens"},
	}
}

func (OauthToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (OauthToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("access_token").
			MaxLen(1024).
			Unique().
			Comment("访问令牌"),
		field.String("refresh_token").
			MaxLen(1024).
			Unique().
			Comment("刷新令牌"),
		field.Uint64("application_id").
			Comment("关联的应用ID"),
		field.Uint64("user_id").
			Comment("关联的用户ID"),
		field.JSON("scope", []string{}).
			Comment("令牌的权限范围"),
		field.Time("access_expires_at").
			Comment("访问令牌过期时间(推荐1小时)"),
		field.Time("refresh_expires_at").
			Comment("刷新令牌过期时间(推荐30天)"),
		field.Time("revoked_at").
			Optional().
			Nillable().
			Comment("撤销时间"),
		field.Time("last_used_at").
			Optional().
			Nillable().
			Comment("最后使用时间(审计)"),
	}
}

func (OauthToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("application", OauthApplication.Type).
			Ref("tokens").
			Field("application_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		edge.From("authorization_code", OauthAuthorizationCode.Type).
			Ref("token").
			Unique(),
		edge.From("user_authorization", OauthUserAuthorization.Type).
			Ref("token").
			Unique(),
	}
}

func (OauthToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("access_token"),
		index.Fields("refresh_token"),
		index.Fields("application_id", "user_id"),
		index.Fields("access_expires_at"),
		index.Fields("refresh_expires_at"),
	}
}

// OauthUserAuthorization 用户授权记录
type OauthUserAuthorization struct {
	ent.Schema
}

func (OauthUserAuthorization) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth_user_authorizations"},
	}
}

func (OauthUserAuthorization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
		events.EventMixin{},
	}
}

func (OauthUserAuthorization) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("user_id").
			Comment("用户ID"),
		field.Uint64("application_id").
			Comment("应用ID"),
		field.Time("authorized_at").
			Comment("授权时间"),
		field.Enum("usage_state").
			Values("unused", "granted", "denied", "revoked").
			Default("unused").
			Comment("授权状态"),
		field.JSON("scope", []string{}).
			Comment("授权的权限范围"),
	}
}

func (OauthUserAuthorization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		edge.To("application", OauthApplication.Type).
			Field("application_id").
			Unique().
			Required(),
		edge.To("code", OauthAuthorizationCode.Type).
			Unique().
			Comment("关联的授权码"),
		edge.To("token", OauthToken.Type).
			Unique().
			Comment("关联的令牌"),
	}
}

func (OauthUserAuthorization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "application_id", "delete_time").Unique(),
		index.Fields("usage_state"),
	}
}

// OauthProvider OAuth 提供商配置(系统作为OAuth应用方时配置的第三方提供商)
type OauthProvider struct {
	ent.Schema
}

func (OauthProvider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth_providers"},
	}
}

func (OauthProvider) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
		events.EventMixin{},
	}
}

func (OauthProvider) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("type").
			Values("oak", "gitea", "github", "google", "gitlab", "microsoft", "weixin", "custom").
			Comment("提供商类型"),
		field.String("name").
			MaxLen(64).
			NotEmpty().
			Comment("提供商名称"),
		field.String("authorization_endpoint").
			MaxLen(512).
			Comment("授权端点URL"),
		field.String("token_endpoint").
			MaxLen(512).
			Comment("令牌端点URL"),
		field.String("user_info_endpoint").
			MaxLen(512).
			Comment("用户信息端点URL"),
		field.String("revoke_endpoint").
			MaxLen(512).
			Optional().
			Comment("撤销端点URL"),
		field.String("refresh_endpoint").
			MaxLen(512).
			Optional().
			Comment("刷新令牌端点URL"),
		field.String("client_id").
			MaxLen(512).
			Comment("在提供商处注册的客户端ID"),
		field.String("client_secret").
			MaxLen(512).
			Sensitive().
			Comment("客户端密钥"),
		field.String("redirect_uri").
			MaxLen(512).
			Comment("回调URI"),
		field.JSON("scopes", []string{}).
			Optional().
			Comment("请求的权限范围"),
		field.Bool("auto_register").
			Default(false).
			Comment("是否自动注册新用户"),
		field.Enum("able_state").
			Values("enabled", "disabled").
			Default("enabled").
			Comment("启用状态"),
		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("额外的元数据信息"),
	}
}

func (OauthProvider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("states", OauthState.Type).
			Comment("创建的状态码"),
		edge.To("oauth_users", OauthUser.Type).
			Comment("关联的OAuth用户"),
	}
}

func (OauthProvider) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type", "name", "delete_time").Unique(),
		index.Fields("able_state"),
	}
}

// OauthState OAuth 状态(防CSRF攻击)
type OauthState struct {
	ent.Schema
}

func (OauthState) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth_states"},
	}
}

func (OauthState) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (OauthState) Fields() []ent.Field {
	return []ent.Field{
		field.String("state").
			MaxLen(32).
			Unique().
			Comment("随机状态码"),
		field.Enum("type").
			Values("login", "bind").
			Comment("操作类型(login或bind)"),
		field.Uint64("provider_id").
			Comment("关联的提供商ID"),
		field.Uint64("user_id").
			Optional().
			Comment("发起操作的用户ID(已登录时)"),
		field.Time("expires_at").
			Comment("过期时间"),
		field.Time("used_at").
			Optional().
			Nillable().
			Comment("使用时间(一次性)"),
	}
}

func (OauthState) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("provider", OauthProvider.Type).
			Ref("states").
			Field("provider_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique(),
		edge.To("oauth_users", OauthUser.Type).
			Comment("生成的OAuth用户连接"),
	}
}

func (OauthState) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("state"),
		index.Fields("expires_at"),
	}
}

// OauthUser OAuth 用户连接(第三方OAuth登录的用户连接信息)
type OauthUser struct {
	ent.Schema
}

func (OauthUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth_users"},
	}
}

func (OauthUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
		events.EventMixin{},
	}
}

func (OauthUser) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("provider_id").
			Comment("关联的提供商ID"),
		field.Uint64("user_id").
			Comment("关联的本地用户ID"),
		field.Uint64("state_id").
			Optional().
			Comment("关联的状态ID"),
		field.String("provider_user_id").
			MaxLen(256).
			Comment("第三方提供商的用户ID"),
		field.JSON("raw_user_info", map[string]interface{}{}).
			Optional().
			Comment("第三方返回的原始用户信息"),
		field.String("access_token").
			MaxLen(1024).
			Optional().
			Sensitive().
			Comment("第三方颁发的访问令牌"),
		field.String("refresh_token").
			MaxLen(1024).
			Optional().
			Sensitive().
			Comment("第三方刷新令牌"),
		field.Time("access_expires_at").
			Optional().
			Nillable().
			Comment("访问令牌过期时间"),
		field.Time("refresh_expires_at").
			Optional().
			Nillable().
			Comment("刷新令牌过期时间"),
		field.Enum("load_state").
			Values("unload", "loaded").
			Default("unload").
			Comment("用户信息加载状态"),
	}
}

func (OauthUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("provider", OauthProvider.Type).
			Ref("oauth_users").
			Field("provider_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		edge.From("state", OauthState.Type).
			Ref("oauth_users").
			Field("state_id").
			Unique(),
	}
}

func (OauthUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider_id", "provider_user_id", "delete_time").Unique(),
		index.Fields("user_id"),
		index.Fields("load_state"),
	}
}
