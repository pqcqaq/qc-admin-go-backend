package funcs

import (
	"context"
	"fmt"
	"go-backend/database/ent"
	"go-backend/database/ent/role"
)

// HasCircularInheritance 检查角色继承是否存在循环引用
// roleID: 当前角色ID
// parentID: 要设置的父角色ID
func HasCircularInheritance(ctx context.Context, client *ent.Client, roleID, parentID uint64) error {
	// 如果要设置自己为父角色，直接返回错误
	if roleID == parentID {
		return fmt.Errorf("角色不能继承自己")
	}

	// 使用深度优先搜索检测循环
	visited := make(map[uint64]bool)
	return dfsCheckCircular(ctx, client, parentID, roleID, visited, 0)
}

// dfsCheckCircular 使用深度优先搜索检测循环继承
// currentID: 当前遍历的角色ID
// targetID: 目标角色ID（我们要检查是否能到达的角色）
// visited: 已访问的角色ID集合
// depth: 当前继承深度
func dfsCheckCircular(ctx context.Context, client *ent.Client, currentID, targetID uint64, visited map[uint64]bool, depth int) error {
	// 检查继承深度是否超过限制（防止过深的继承链）
	const maxInheritanceDepth = 10
	if depth > maxInheritanceDepth {
		return fmt.Errorf("角色继承深度超过限制(%d层)", maxInheritanceDepth)
	}

	// 如果当前角色就是目标角色，说明存在循环
	if currentID == targetID {
		return fmt.Errorf("检测到角色继承循环")
	}

	// 如果已经访问过这个角色，说明存在循环
	if visited[currentID] {
		return fmt.Errorf("检测到角色继承循环")
	}

	// 标记当前角色为已访问
	visited[currentID] = true

	// 查询当前角色的所有父角色（inherits_from关系）
	parentRoles, err := client.Role.Query().
		Where(role.HasInheritedByWith(role.ID(currentID))).
		All(ctx)

	if err != nil {
		return fmt.Errorf("查询父角色失败: %v", err)
	}

	// 递归检查每个父角色
	for _, parentRole := range parentRoles {
		if err := dfsCheckCircular(ctx, client, parentRole.ID, targetID, visited, depth+1); err != nil {
			return err
		}
	}

	// 移除访问标记（回溯）
	delete(visited, currentID)
	return nil
}

// getAllAncestorRoles 获取角色的所有祖先角色（用于权限计算）
func GetAllAncestorRoles(ctx context.Context, client *ent.Client, roleID uint64) ([]*ent.Role, error) {
	var ancestors []*ent.Role
	visited := make(map[uint64]bool)

	err := collectAncestors(ctx, client, roleID, visited, &ancestors, 0)
	if err != nil {
		return nil, err
	}

	return ancestors, nil
}

// collectAncestors 递归收集祖先角色
func collectAncestors(ctx context.Context, client *ent.Client, roleID uint64, visited map[uint64]bool, ancestors *[]*ent.Role, depth int) error {
	// 防止无限递归
	const maxDepth = 10
	if depth > maxDepth {
		return fmt.Errorf("角色继承深度超过限制")
	}

	if visited[roleID] {
		return nil // 已经访问过，避免重复
	}
	visited[roleID] = true

	// 查询当前角色的父角色（inherits_from关系）
	parentRoles, err := client.Role.Query().
		Where(role.HasInheritedByWith(role.ID(roleID))).
		All(ctx)

	if err != nil {
		return err
	}

	for _, parentRole := range parentRoles {
		*ancestors = append(*ancestors, parentRole)
		// 递归获取父角色的祖先
		if err := collectAncestors(ctx, client, parentRole.ID, visited, ancestors, depth+1); err != nil {
			return err
		}
	}

	return nil
}
