package handlers

import (
	"context"
	"fmt"

	"go-backend/database/ent"
	"go-backend/database/events"
	"go-backend/internal/funcs"
	"go-backend/pkg/database"

	entpkg "entgo.io/ent"
)

// RoleInheritanceHandler 角色继承检查处理器
type RoleInheritanceHandler struct{}

// Handle 处理角色继承事件
func (h *RoleInheritanceHandler) Handle(ctx context.Context, event *events.Event) error {
	// 只处理角色相关的 pre-create 和 pre-update 事件
	if event.EntityType != "Role" ||
		(event.Type != events.EventTypePreUpdate && event.Type != events.EventTypePreCreate) {
		return nil
	}

	// 类型断言获取角色变更信息
	roleMutation, ok := event.Mutation.(*ent.RoleMutation)
	if !ok {
		return nil // 不是角色变更，忽略
	}

	// 获取数据库客户端
	client := database.Client
	if client == nil {
		return fmt.Errorf("无法获取数据库客户端")
	}

	// 对于新建角色，roleID 需要从变更中获取或者使用0（表示新建）
	var roleID uint64
	if id, exists := roleMutation.ID(); exists {
		roleID = id
	} else {
		// 新建角色的情况，此时还没有ID，我们可以用0表示
		// 但需要确保继承关系的检查逻辑能够处理这种情况
		roleID = 0
	}

	// 检查要添加的父角色（inherits_from）
	if inheritsFromIDs := roleMutation.InheritsFromIDs(); len(inheritsFromIDs) > 0 {
		for _, parentID := range inheritsFromIDs {
			// 对于新建角色，我们只需要检查父角色之间是否存在循环
			// 对于更新角色，需要检查是否会形成循环
			if roleID == 0 {
				// 新建角色：检查要设置的父角色之间是否存在循环
				if err := h.checkParentCircularInheritance(ctx, client, inheritsFromIDs); err != nil {
					return fmt.Errorf("设置角色继承失败: %v", err)
				}
				break // 只需要检查一次
			} else {
				// 更新角色：检查是否会与现有角色形成循环
				if err := funcs.HasCircularInheritance(ctx, client, roleID, parentID); err != nil {
					return fmt.Errorf("设置角色继承失败: %v", err)
				}
			}
		}
	}

	// 检查要添加的子角色（inherited_by）
	if inheritedByIDs := roleMutation.InheritedByIDs(); len(inheritedByIDs) > 0 {
		for _, childID := range inheritedByIDs {
			if roleID == 0 {
				// 新建角色：检查要设置的子角色是否会形成循环
				if err := h.checkChildCircularInheritance(ctx, client, inheritedByIDs); err != nil {
					return fmt.Errorf("设置角色继承失败: %v", err)
				}
				break // 只需要检查一次
			} else {
				// 更新角色：检查是否会形成循环
				if err := funcs.HasCircularInheritance(ctx, client, childID, roleID); err != nil {
					return fmt.Errorf("设置角色继承失败: %v", err)
				}
			}
		}
	}

	return nil
}

// SupportsEvent 检查是否支持指定的事件类型和实体类型
func (h *RoleInheritanceHandler) SupportsEvent(eventType events.EventType, entityType string, operation entpkg.Op) bool {
	return entityType == "Role" &&
		(eventType == events.EventTypePreUpdate || eventType == events.EventTypePreCreate) &&
		(operation == entpkg.OpUpdate || operation == entpkg.OpUpdateOne || operation == entpkg.OpCreate)
}

// checkParentCircularInheritance 检查新建角色的父角色列表是否存在循环
func (h *RoleInheritanceHandler) checkParentCircularInheritance(ctx context.Context, client *ent.Client, parentIDs []uint64) error {
	// 检查父角色列表中是否有重复
	seen := make(map[uint64]bool)
	for _, parentID := range parentIDs {
		if seen[parentID] {
			return fmt.Errorf("父角色列表中存在重复的角色ID: %d", parentID)
		}
		seen[parentID] = true

		// 检查每个父角色与其他父角色之间是否存在继承关系
		for _, otherParentID := range parentIDs {
			if parentID != otherParentID {
				if err := funcs.HasCircularInheritance(ctx, client, parentID, otherParentID); err != nil {
					return fmt.Errorf("父角色 %d 和 %d 之间存在继承关系，会形成循环: %v", parentID, otherParentID, err)
				}
			}
		}
	}
	return nil
}

// checkChildCircularInheritance 检查新建角色的子角色列表是否存在循环
func (h *RoleInheritanceHandler) checkChildCircularInheritance(ctx context.Context, client *ent.Client, childIDs []uint64) error {
	// 检查子角色列表中是否有重复
	seen := make(map[uint64]bool)
	for _, childID := range childIDs {
		if seen[childID] {
			return fmt.Errorf("子角色列表中存在重复的角色ID: %d", childID)
		}
		seen[childID] = true

		// 检查每个子角色与其他子角色之间是否存在继承关系
		for _, otherChildID := range childIDs {
			if childID != otherChildID {
				if err := funcs.HasCircularInheritance(ctx, client, otherChildID, childID); err != nil {
					return fmt.Errorf("子角色 %d 和 %d 之间存在继承关系，会形成循环: %v", childID, otherChildID, err)
				}
			}
		}
	}
	return nil
}

// RegisterRoleInheritanceHandler 注册角色继承检查处理器
func RegisterRoleInheritanceHandler() {
	handler := &RoleInheritanceHandler{}
	// 注册处理创建和更新事件
	events.Subscribe(events.EventTypePreCreate, handler)
	events.Subscribe(events.EventTypePreUpdate, handler)
}
