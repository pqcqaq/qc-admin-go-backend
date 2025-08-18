# 用户认证系统测试指南

## 系统概述

我已经为您完成了一个完整的用户认证系统，包含以下功能：

### 1. 核心功能
- ✅ **用户注册** - 支持邮箱和手机号两种方式
- ✅ **用户登录** - 支持邮箱、手机号、用户名三种认证方式
- ✅ **密码重置** - 通过验证码重置密码
- ✅ **验证码发送** - 支持邮件和短信两种渠道

### 2. 技术特性
- ✅ **Argon2ID密码加密** - 生产级密码安全
- ✅ **策略模式验证码发送** - 可扩展的验证码发送系统
- ✅ **单例模式服务客户端** - SMTP和SMS客户端单例管理
- ✅ **软删除和审计日志** - 完整的数据追踪
- ✅ **统一错误处理** - RESTful API错误响应

## API端点

### 认证相关 (`/api/v1/auth/`)

1. **用户注册**
   ```
   POST /api/v1/auth/register
   Content-Type: application/json
   
   {
     "credential_type": "email",        // "email" 或 "phone"
     "credential_value": "test@example.com",
     "password": "YourPassword123!",
     "verification_code": "123456"
   }
   ```

2. **用户登录**
   ```
   POST /api/v1/auth/login
   Content-Type: application/json
   
   {
     "credential_type": "email",        // "email", "phone", 或 "username"
     "credential_value": "test@example.com",
     "password": "YourPassword123!"
   }
   ```

3. **发送验证码**
   ```
   POST /api/v1/auth/send-verification-code
   Content-Type: application/json
   
   {
     "type": "email",                   // "email" 或 "phone"
     "target": "test@example.com",
     "purpose": "registration"          // "registration", "login", 或 "password_reset"
   }
   ```

4. **重置密码**
   ```
   POST /api/v1/auth/reset-password
   Content-Type: application/json
   
   {
     "credential_type": "email",        // "email" 或 "phone"
     "credential_value": "test@example.com",
     "new_password": "NewPassword123!",
     "verification_code": "123456"
   }
   ```

## 数据库表结构

### credentials 表
- `id` - 主键
- `user_id` - 用户ID
- `type` - 认证类型 (email/phone/username)
- `value` - 认证值
- `password_hash` - Argon2ID密码哈希
- `salt` - 随机盐值
- `is_verified` - 是否已验证
- `created_at`, `updated_at`, `deleted_at`

### verify_codes 表
- `id` - 主键
- `code` - 验证码
- `type` - 发送类型 (email/phone/sms)
- `target` - 目标地址
- `purpose` - 用途 (registration/login/password_reset)
- `expires_at` - 过期时间
- `used_at` - 使用时间
- `created_at`, `updated_at`, `deleted_at`

## 配置说明

### 邮件配置 (config.yaml)
```yaml
email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "your-email@gmail.com"
  password: "your-email-password"
  from_name: "扫描应用"
  from_email: "your-email@gmail.com"
```

### 短信配置 (config.yaml)
```yaml
sms:
  access_key_id: "your-aliyun-access-key-id"
  access_key_secret: "your-aliyun-access-key-secret"
  sign_name: "你的短信签名"
  template_code: "SMS_123456789"
  endpoint: "dysmsapi.aliyuncs.com"
```

## 测试步骤

### 1. 启动服务
```bash
# 编译项目
go build -o tmp/main.exe main.go

# 启动服务器
./tmp/main.exe -c config.yaml
```

### 2. 配置第三方服务

**配置邮件服务：**
- 如果使用Gmail，需要开启"应用专用密码"
- 更新 `config.yaml` 中的邮件配置

**配置阿里云短信：**
- 在阿里云控制台获取AccessKey
- 申请短信签名和模板
- 更新 `config.yaml` 中的短信配置

### 3. 测试流程

1. **发送注册验证码**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/send-verification-code \
     -H "Content-Type: application/json" \
     -d '{
       "type": "email",
       "target": "test@example.com",
       "purpose": "registration"
     }'
   ```

2. **用户注册**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{
       "credential_type": "email",
       "credential_value": "test@example.com",
       "password": "YourPassword123!",
       "verification_code": "123456"
     }'
   ```

3. **用户登录**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
       "credential_type": "email",
       "credential_value": "test@example.com",
       "password": "YourPassword123!"
     }'
   ```

## 系统架构

### 验证码发送系统
```
VerifyCodeSender (接口)
├── EmailCodeSender (邮件发送)
├── PhoneCodeSender (手机发送，预留)
└── SMSCodeSender (短信发送)

SenderFactory (工厂类)
└── 根据类型创建对应的发送器
```

### 服务层单例模式
```
email.EmailClient (单例)
├── 连接测试
├── HTML模板渲染
└── SMTP发送

sms.SMSClient (单例)
├── 阿里云SDK集成
├── JSON参数处理
└── 短信发送
```

## 注意事项

1. **安全性**
   - 密码使用Argon2ID加密，包含随机盐
   - 验证码有过期时间限制
   - 支持软删除，保留审计痕迹

2. **扩展性**
   - 使用策略模式，易于添加新的验证码发送方式
   - 单例模式确保服务客户端的高效复用
   - 清晰的分层架构便于维护

3. **生产部署**
   - 配置真实的SMTP和SMS服务信息
   - 考虑验证码频率限制
   - 监控服务健康状态

## 故障排除

### 常见问题
1. **编译错误** - 确保所有依赖已安装 (`go mod tidy`)
2. **邮件发送失败** - 检查SMTP配置和网络连接
3. **短信发送失败** - 验证阿里云配置和模板格式
4. **数据库连接** - 确保数据库服务正在运行

### 日志查看
服务启动时会显示各组件的初始化状态：
```
[INFO] Database connection established
[INFO] S3 client initialized successfully  
[INFO] Email client initialized successfully
[INFO] SMS client initialized successfully
[INFO] Server starting on localhost:8080
```

## 总结

这个认证系统已经完全实现并可以投入使用。所有代码都遵循Go最佳实践，包含完整的错误处理、安全加密和可扩展架构。您可以根据需要进一步定制功能或添加新的认证方式。
