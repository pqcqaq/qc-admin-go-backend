# RBAC API 测试指南

本文档提供了完整的RBAC（基于角色的访问控制）API测试命令，展示了角色继承和事件系统的工作方式。

## 前置条件

确保后端服务已启动：
```bash
go run main_rbac_test.go
```

## 1. 权限域（Scope）管理

### 创建权限域
```bash
# 创建根权限域
curl -X POST http://localhost:8080/api/v1/rbac/scopes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "系统管理",
    "type": "menu",
    "description": "系统管理模块",
    "order": 1
  }'

# 创建子权限域
curl -X POST http://localhost:8080/api/v1/rbac/scopes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "用户管理",
    "type": "page",
    "description": "用户管理页面",
    "parentId": "1",
    "order": 1
  }'
```

### 获取权限域树形结构
```bash
curl -X GET http://localhost:8080/api/v1/rbac/scopes/tree
```

### 获取权限域列表
```bash
# 获取所有权限域
curl -X GET http://localhost:8080/api/v1/rbac/scopes/all

# 分页获取权限域
curl -X GET "http://localhost:8080/api/v1/rbac/scopes?page=1&pageSize=10"
```

## 2. 权限（Permission）管理

### 创建权限
```bash
# 创建用户查看权限
curl -X POST http://localhost:8080/api/v1/rbac/permissions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "查看用户",
    "action": "user.read",
    "description": "查看用户信息的权限",
    "scopeId": "2"
  }'

# 创建用户编辑权限
curl -X POST http://localhost:8080/api/v1/rbac/permissions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "编辑用户",
    "action": "user.write",
    "description": "编辑用户信息的权限",
    "scopeId": "2"
  }'

# 创建用户删除权限
curl -X POST http://localhost:8080/api/v1/rbac/permissions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "删除用户",
    "action": "user.delete",
    "description": "删除用户的权限",
    "scopeId": "2"
  }'
```

### 获取权限列表
```bash
# 获取所有权限
curl -X GET http://localhost:8080/api/v1/rbac/permissions/all

# 分页获取权限
curl -X GET "http://localhost:8080/api/v1/rbac/permissions?page=1&pageSize=10"

# 按权限域搜索
curl -X GET "http://localhost:8080/api/v1/rbac/permissions?scopeId=2"
```

## 3. 角色（Role）管理与继承

### 创建基础角色
```bash
# 创建超级管理员角色
curl -X POST http://localhost:8080/api/v1/rbac/roles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "超级管理员",
    "description": "系统超级管理员，拥有所有权限"
  }'

# 创建管理员角色
curl -X POST http://localhost:8080/api/v1/rbac/roles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "管理员",
    "description": "普通管理员角色"
  }'

# 创建用户管理员角色（继承管理员）
curl -X POST http://localhost:8080/api/v1/rbac/roles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "用户管理员",
    "description": "负责用户管理的管理员",
    "inheritsFrom": ["2"]
  }'
```

### 测试角色继承循环检测
```bash
# 尝试创建循环继承（这应该失败）
curl -X POST http://localhost:8080/api/v1/rbac/roles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试角色",
    "description": "测试循环继承",
    "inheritsFrom": ["3", "1"]
  }'

# 尝试让角色1继承角色3（应该失败，因为角色3已经继承了角色2）
curl -X PUT http://localhost:8080/api/v1/rbac/roles/1 \
  -H "Content-Type: application/json" \
  -d '{
    "inheritsFrom": ["3"]
  }'
```

### 分配角色权限
```bash
# 为超级管理员分配所有权限
curl -X POST http://localhost:8080/api/v1/rbac/roles/1/permissions \
  -H "Content-Type: application/json" \
  -d '{
    "permissionIds": ["1", "2", "3"]
  }'

# 为管理员分配查看和编辑权限
curl -X POST http://localhost:8080/api/v1/rbac/roles/2/permissions \
  -H "Content-Type: application/json" \
  -d '{
    "permissionIds": ["1", "2"]
  }'
```

### 获取角色信息
```bash
# 获取所有角色
curl -X GET http://localhost:8080/api/v1/rbac/roles/all

# 获取单个角色（包含继承信息）
curl -X GET http://localhost:8080/api/v1/rbac/roles/3
```

## 4. 用户角色关联

### 分配用户角色
```bash
# 为用户分配角色
curl -X POST http://localhost:8080/api/v1/rbac/user-roles \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 1,
    "roleId": 3
  }'

# 分配多个用户角色
curl -X POST http://localhost:8080/api/v1/rbac/user-roles \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 2,
    "roleId": 2
  }'
```

### 查询用户权限
```bash
# 获取用户的所有角色
curl -X GET http://localhost:8080/api/v1/rbac/user-roles/users/1/roles

# 获取用户的所有权限（包括继承的）
curl -X GET http://localhost:8080/api/v1/rbac/user-roles/users/1/permissions

# 检查用户是否有特定权限
curl -X GET http://localhost:8080/api/v1/rbac/user-roles/users/1/permissions/2/check
```

### 撤销用户角色
```bash
# 撤销用户角色
curl -X DELETE http://localhost:8080/api/v1/rbac/user-roles/users/1/roles/3
```

## 5. 事件系统验证

观察服务器日志，在创建和更新角色时应该看到类似以下的事件日志：

```
收到事件: role.pre.create, 数据: map[name:用户管理员 inheritsFrom:[2]]
检查角色继承循环性...
角色继承验证通过
```

## 6. 角色继承权限测试

### 验证权限继承
```bash
# 1. 首先确认管理员角色(ID:2)有权限1和2
curl -X GET http://localhost:8080/api/v1/rbac/roles/2

# 2. 确认用户管理员角色(ID:3)继承自管理员角色
curl -X GET http://localhost:8080/api/v1/rbac/roles/3

# 3. 为用户分配用户管理员角色
curl -X POST http://localhost:8080/api/v1/rbac/user-roles \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 1,
    "roleId": 3
  }'

# 4. 检查用户是否继承了管理员角色的权限
curl -X GET http://localhost:8080/api/v1/rbac/user-roles/users/1/permissions
# 应该看到权限1和2，即使用户管理员角色本身没有直接分配这些权限
```

## 错误处理测试

### 测试各种错误情况
```bash
# 1. 创建重复角色名称
curl -X POST http://localhost:8080/api/v1/rbac/roles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "超级管理员",
    "description": "重复名称测试"
  }'

# 2. 访问不存在的角色
curl -X GET http://localhost:8080/api/v1/rbac/roles/999

# 3. 分配不存在的权限
curl -X POST http://localhost:8080/api/v1/rbac/roles/1/permissions \
  -H "Content-Type: application/json" \
  -d '{
    "permissionIds": ["999"]
  }'

# 4. 为不存在的用户分配角色
curl -X POST http://localhost:8080/api/v1/rbac/user-roles \
  -H "Content-Type: application/json" \
  -d '{
    "userId": 999,
    "roleId": 1
  }'
```

## 注意事项

1. **事件系统**：所有角色的创建和更新操作都会触发事件系统进行循环继承检测
2. **权限继承**：子角色会自动继承父角色的所有权限
3. **数据一致性**：系统会确保角色关系的一致性，防止循环继承
4. **错误处理**：所有API都有完善的错误处理和验证

## 预期结果

- 角色继承应该正常工作，子角色拥有父角色的所有权限
- 循环继承检测应该阻止无效的角色关系
- 事件系统应该在每次角色操作时正确触发
- 用户权限查询应该包含通过角色继承获得的所有权限
