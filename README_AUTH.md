# 用户认证系统API文档

## 概述

这是一个完整的用户认证系统，支持多种认证方式包括密码、邮箱、手机号等，并实现了验证码发送和验证功能。

## 认证方式

系统支持以下认证方式：
- `password`: 密码认证
- `email`: 邮箱认证（需要验证码）
- `phone`: 手机号认证（需要验证码）
- `sms`: 短信认证（需要验证码）
- `oauth`: OAuth认证（需要验证码）
- `totp`: TOTP认证（需要验证码）

## API接口

### 1. 发送验证码

**POST** `/api/v1/auth/send-verify-code`

发送验证码到指定的标识符（邮箱/手机号等）。

**请求体:**
```json
{
    "senderType": "email",           // 发送方式: email|phone|sms
    "purpose": "register",           // 用途: register|login|reset_password
    "identifier": "test@example.com" // 标识符（邮箱/手机号等）
}
```

**响应:**
```json
{
    "success": true,
    "data": {
        "message": "验证码发送成功"
    }
}
```

**限制:**
- 同一标识符30秒内只能发送一次验证码
- 验证码有效期为15分钟

### 2. 验证验证码

**POST** `/api/v1/auth/verify-code`

验证验证码是否正确（测试接口）。

**请求体:**
```json
{
    "senderType": "email",           // 发送方式
    "purpose": "register",           // 用途
    "identifier": "test@example.com", // 标识符
    "code": "123456"                 // 验证码
}
```

**响应:**
```json
{
    "success": true,
    "data": {
        "message": "验证码验证成功"
    }
}
```

### 3. 用户注册

**POST** `/api/v1/auth/register`

用户注册，支持密码注册和验证码注册。

**请求体（密码注册）:**
```json
{
    "credentialType": "password",    // 认证类型
    "identifier": "testuser",        // 用户名
    "secret": "Test123!@#",          // 密码
    "username": "测试用户"           // 显示名称
}
```

**请求体（验证码注册）:**
```json
{
    "credentialType": "email",       // 认证类型
    "identifier": "test@example.com", // 邮箱
    "verifyCode": "123456",          // 验证码
    "username": "测试用户"           // 显示名称
}
```

**响应:**
```json
{
    "success": true,
    "data": {
        "user": {
            "id": 1,
            "name": "测试用户",
            "status": "active",
            "createTime": "2025-08-18 15:30:00",
            "updateTime": "2025-08-18 15:30:00"
        },
        "message": "注册成功"
    }
}
```

### 4. 用户登录

**POST** `/api/v1/auth/login`

用户登录，支持密码登录和验证码登录。

**请求体（密码登录）:**
```json
{
    "credentialType": "password",    // 认证类型
    "identifier": "testuser",        // 用户名
    "secret": "Test123!@#"           // 密码
}
```

**请求体（验证码登录）:**
```json
{
    "credentialType": "email",       // 认证类型
    "identifier": "test@example.com", // 邮箱
    "verifyCode": "123456"           // 验证码
}
```

**响应:**
```json
{
    "success": true,
    "data": {
        "user": {
            "id": 1,
            "name": "测试用户",
            "status": "active",
            "createTime": "2025-08-18 15:30:00",
            "updateTime": "2025-08-18 15:30:00"
        },
        "message": "登录成功"
    }
}
```

**安全特性:**
- 密码错误5次后账号锁定30分钟
- 支持账号状态检查
- 记录最后使用时间

### 5. 重置密码

**POST** `/api/v1/auth/reset-password`

重置用户密码，支持原密码重置和验证码重置。

**请求体（原密码重置）:**
```json
{
    "credentialType": "password",    // 认证类型
    "identifier": "testuser",        // 用户名
    "oldPassword": "OldPass123!",    // 原密码
    "newPassword": "NewPass123!"     // 新密码
}
```

**请求体（验证码重置）:**
```json
{
    "credentialType": "email",       // 认证类型
    "identifier": "test@example.com", // 邮箱
    "verifyCode": "123456",          // 验证码
    "newPassword": "NewPass123!"     // 新密码
}
```

**响应:**
```json
{
    "success": true,
    "data": {
        "message": "密码重置成功"
    }
}
```

## 错误响应

所有API在出错时都会返回统一的错误格式：

```json
{
    "success": false,
    "code": 400,
    "message": "请求参数格式错误",
    "timestamp": "2025-08-18T15:30:00Z",
    "path": "/api/v1/auth/login"
}
```

### 常见错误码

- `400`: 请求参数错误
- `401`: 未授权（登录失败、密码错误等）
- `404`: 用户不存在
- `409`: 资源冲突（用户已存在）
- `500`: 服务器内部错误
- `1001`: 用户不存在
- `1002`: 用户已存在
- `2001`: 数据库错误
- `3001`: 数据验证错误

## 验证码发送器

系统采用策略模式实现验证码发送，支持：

### 邮箱发送器（EmailCodeSender）
- 发送类型：`email`
- 支持HTML邮件模板
- 可配置SMTP服务器

### 手机发送器（PhoneCodeSender）
- 发送类型：`phone`
- 支持国内外手机号
- 可配置短信服务商

### SMS发送器（SMSCodeSender）
- 发送类型：`sms`
- 通用SMS发送接口
- 支持多种SMS网关

## 密码安全

- 使用Argon2算法进行密码哈希
- 每个密码使用随机盐值
- 支持密码强度验证
- 防止暴力破解攻击

## 数据库设计

### 认证表（sys_credentials）
- 支持一用户多认证方式
- 记录认证失败次数和锁定时间
- 支持OAuth提供商信息
- 软删除支持

### 验证码表（sys_verify_codes）
- 记录验证码发送历史
- 支持验证码重复使用检查
- 自动过期机制
- 发送状态跟踪

## 使用示例

### 完整的注册流程

1. **发送注册验证码**
```bash
curl -X POST http://localhost:8080/api/v1/auth/send-verify-code \
  -H "Content-Type: application/json" \
  -d '{
    "senderType": "email",
    "purpose": "register",
    "identifier": "user@example.com"
  }'
```

2. **用户注册**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "credentialType": "email",
    "identifier": "user@example.com",
    "verifyCode": "123456",
    "username": "新用户"
  }'
```

### 完整的登录流程

1. **密码登录**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "credentialType": "password",
    "identifier": "username",
    "secret": "password123"
  }'
```

2. **验证码登录**
```bash
# 先发送验证码
curl -X POST http://localhost:8080/api/v1/auth/send-verify-code \
  -H "Content-Type: application/json" \
  -d '{
    "senderType": "email",
    "purpose": "login",
    "identifier": "user@example.com"
  }'

# 然后登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "credentialType": "email",
    "identifier": "user@example.com",
    "verifyCode": "123456"
  }'
```

## 开发说明

### 扩展验证码发送器

要添加新的验证码发送器，请：

1. 实现`verifycode.Sender`接口
2. 在`DefaultSenderFactory`中注册
3. 添加对应的发送类型常量

### 自定义认证类型

要添加新的认证类型，请：

1. 在`credential.go` schema中添加新的枚举值
2. 在认证函数中添加对应的处理逻辑
3. 更新API文档和验证规则

### 配置说明

- 验证码有效期：15分钟
- 验证码发送间隔：30秒
- 登录失败锁定次数：5次
- 账号锁定时间：30分钟
- 密码哈希算法：Argon2ID

这些参数都可以在代码中进行调整。
