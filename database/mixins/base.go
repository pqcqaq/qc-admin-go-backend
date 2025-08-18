package mixins

import (
	"context"
	"fmt"
	"time"

	pkgent "go-backend/database/ent"
	"go-backend/database/ent/hook"
	"go-backend/database/ent/intercept"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/sony/sonyflake"
)

// BaseMixin 包含所有基础字段的mixin
type BaseMixin struct {
	mixin.Schema
}

func IDHook() ent.Hook {
	sf := sonyflake.NewSonyflake(sonyflake.Settings{})
	type IDSetter interface {
		SetID(uint64)
	}
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// 只在创建操作时设置ID
			if m.Op() != ent.OpCreate {
				return next.Mutate(ctx, m)
			}

			is, ok := m.(IDSetter)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation %T", m)
			}
			id, err := sf.NextID()
			if err != nil {
				return nil, err
			}
			is.SetID(id)
			return next.Mutate(ctx, m)
		})
	}
}

func (BaseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			GoType(uint64(0)).
			Unique().
			Immutable().
			Comment("主键ID"),
		field.Time("create_time").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Uint64("create_by").
			Optional().
			Comment("创建人ID"),
		field.Time("update_time").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
		field.Uint64("update_by").
			Optional().
			Comment("更新人ID"),
	}
}

// Hooks 返回基础字段的审计钩子
func (BaseMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		AuditHook,
		IDHook(),
	}
}

// SoftDeleteMixin 软删除mixin - 基于 ent 官方实现
type SoftDeleteMixin struct {
	mixin.Schema
}

func (SoftDeleteMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("delete_time"), // 软删除时间索引
		index.Fields("delete_by"),   // 软删除人ID索引
	}
}

// Fields 返回软删除需要的字段
func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("delete_time").
			Optional().
			Comment("删除时间"),
		field.Uint64("delete_by").
			Optional().
			Comment("删除人ID"),
	}
}

type softDeleteKey struct{}

// SkipSoftDelete 返回一个跳过软删除拦截器的新上下文
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

// Interceptors 返回软删除的拦截器 - 关键修复
func (d SoftDeleteMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		// 使用正确的 intercept.TraverseFunc 实现查询拦截
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			// 跳过软删除，意味着包含已软删除的实体
			if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
				return nil
			}
			d.P(q)
			return nil
		}),
	}
}

// Hooks 返回软删除的钩子
func (d SoftDeleteMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			func(next ent.Mutator) ent.Mutator {
				return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
					// 跳过软删除，进行物理删除
					if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
						return next.Mutate(ctx, m)
					}

					mx, ok := m.(interface {
						SetOp(ent.Op)
						Client() *pkgent.Client
						SetDeleteTime(time.Time)
						SetUpdateTime(time.Time)
						SetDeleteBy(uint64)
						SetUpdateBy(uint64)
						WhereP(...func(*sql.Selector))
					})
					if !ok {
						return nil, fmt.Errorf("unexpected mutation type %T", m)
					}

					// 添加软删除过滤条件（确保只删除未被软删除的记录）
					d.P(mx)

					// 将删除操作转换为更新操作
					mx.SetOp(ent.OpUpdate)
					now := time.Now()
					mx.SetDeleteTime(now)
					mx.SetUpdateTime(now)

					// 设置删除人信息
					userID := getUserIDFromContext(ctx)
					if userID > 0 {
						mx.SetDeleteBy(userID)
						mx.SetUpdateBy(userID)
					}

					return mx.Client().Mutate(ctx, m)
				})
			},
			ent.OpDeleteOne|ent.OpDelete,
		),
	}
}

// P 为查询添加存储级别的断言
func (d SoftDeleteMixin) P(w interface{ WhereP(...func(*sql.Selector)) }) {
	w.WhereP(
		sql.FieldIsNull(d.Fields()[0].Descriptor().Name), // delete_time
	)
}

// AuditLogger 审计日志接口
type AuditLogger interface {
	SetCreateTime(time.Time)
	CreateTime() (value time.Time, exists bool)
	SetCreateBy(uint64)
	CreateBy() (id uint64, exists bool)
	SetUpdateTime(time.Time)
	UpdateTime() (value time.Time, exists bool)
	SetUpdateBy(uint64)
	UpdateBy() (id uint64, exists bool)
}

// SoftDeleteLogger 软删除审计接口
type SoftDeleteLogger interface {
	SetDeleteTime(time.Time)
	DeleteTime() (value time.Time, exists bool)
	SetDeleteBy(uint64)
	DeleteBy() (id uint64, exists bool)
	SetUpdateTime(time.Time)
	UpdateTime() (value time.Time, exists bool)
	SetUpdateBy(uint64)
	UpdateBy() (id uint64, exists bool)
	SetOp(ent.Op)
	Where(...func(*sql.Selector))
}

// AuditHook 审计钩子
func AuditHook(next ent.Mutator) ent.Mutator {
	return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		ml, ok := m.(AuditLogger)
		if !ok {
			return next.Mutate(ctx, m)
		}

		userID := getUserIDFromContext(ctx)
		now := time.Now()

		switch op := m.Op(); {
		case op.Is(ent.OpCreate):
			ml.SetCreateTime(now)
			if _, exists := ml.CreateBy(); !exists && userID > 0 {
				ml.SetCreateBy(userID)
			}
			ml.SetUpdateTime(now)
			if _, exists := ml.UpdateBy(); !exists && userID > 0 {
				ml.SetUpdateBy(userID)
			}
		case op.Is(ent.OpUpdateOne | ent.OpUpdate):
			ml.SetUpdateTime(now)
			if _, exists := ml.UpdateBy(); !exists && userID > 0 {
				ml.SetUpdateBy(userID)
			}
		}
		return next.Mutate(ctx, m)
	})
}

// getUserIDFromContext 从上下文中获取用户ID
func getUserIDFromContext(ctx context.Context) uint64 {
	if userID, ok := ctx.Value("user_id").(uint64); ok {
		return userID
	}
	if userID, ok := ctx.Value("UserIDKey").(uint64); ok {
		return userID
	}
	return 0
}
