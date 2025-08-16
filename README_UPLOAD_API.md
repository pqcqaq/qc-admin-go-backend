# 文件上传 API 文档

## 概述

本项目实现了两类文件上传功能：

### 第一类：分离式上传（预签名URL方式）
适用于大文件上传，前端直接上传到S3，后端不经手文件流

### 第二类：直接上传（表单上传）
适用于小文件上传，前端通过表单上传到后端，后端转发到S3

## API 接口

### 1. 准备上传 - POST /api/attachments/prepare-upload

获取上传凭证，返回预签名URL用于直接上传到S3。

**请求体：**
```json
{
  "filename": "example.pdf",
  "content_type": "application/pdf",
  "size": 1024000,
  "bucket": "my-bucket",  // 可选，不提供则使用配置中的默认bucket
  "tag1": "document",     // 可选
  "tag2": "important",    // 可选
  "tag3": "2024"         // 可选
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "upload_url": "https://s3.amazonaws.com/bucket/path?presigned=xxx",
    "upload_session_id": "abc123def456",
    "expires_at": 1672531200,
    "attachment_id": 123
  },
  "message": "上传凭证生成成功"
}
```

### 2. 确认上传 - POST /api/attachments/confirm-upload

上传完成后通知后端，更新文件状态。

**请求体：**
```json
{
  "upload_session_id": "abc123def456",
  "etag": "d41d8cd98f00b204e9800998ecf8427e",  // 可选
  "actual_size": 1024000                        // 可选，实际文件大小
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": 123,
    "create_time": "2024-01-01T00:00:00Z",
    "update_time": "2024-01-01T00:01:00Z",
    "filename": "example.pdf",
    "path": "uploads/2024/01/01/abc123def456_example.pdf",
    "url": "https://s3.amazonaws.com/bucket/uploads/2024/01/01/abc123def456_example.pdf",
    "content_type": "application/pdf",
    "size": 1024000,
    "etag": "d41d8cd98f00b204e9800998ecf8427e",
    "bucket": "my-bucket",
    "storage_provider": "s3",
    "status": "uploaded",
    "upload_session_id": "abc123def456",
    "tag1": "document",
    "tag2": "important",
    "tag3": "2024"
  },
  "message": "文件上传确认成功"
}
```

### 3. 直接上传 - POST /api/attachments/direct-upload

直接通过表单上传文件到后端。

**请求（multipart/form-data）：**
- `file`: 文件（必需）
- `bucket`: 存储桶（可选）
- `tag1`: 标签1（可选）
- `tag2`: 标签2（可选）
- `tag3`: 标签3（可选）

**响应：**
```json
{
  "success": true,
  "message": "文件上传成功",
  "attachment": {
    "id": 124,
    "create_time": "2024-01-01T00:00:00Z",
    "update_time": "2024-01-01T00:00:00Z",
    "filename": "example.jpg",
    "path": "uploads/2024/01/01/1704067200000000000_example.jpg",
    "url": "https://s3.amazonaws.com/bucket/uploads/2024/01/01/1704067200000000000_example.jpg",
    "content_type": "image/jpeg",
    "size": 512000,
    "etag": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "bucket": "my-bucket",
    "storage_provider": "s3",
    "status": "uploaded",
    "tag1": "image",
    "tag2": "",
    "tag3": ""
  }
}
```

## 使用流程

### 分离式上传流程：

1. 前端调用 `/api/attachments/prepare-upload` 获取上传凭证
2. 前端使用返回的 `upload_url` 直接上传文件到S3
3. 上传成功后，前端调用 `/api/attachments/confirm-upload` 通知后端
4. 后端验证文件存在性并更新状态为 `uploaded`

### 直接上传流程：

1. 前端通过表单提交文件到 `/api/attachments/direct-upload`
2. 后端接收文件并上传到S3
3. 后端创建附件记录并返回完整信息

## 配置要求

确保在配置文件中正确设置S3相关配置：

```yaml
s3:
  endpoint: ""                    # S3端点，AWS留空，MinIO等填写完整URL
  region: "us-east-1"            # AWS区域
  access_key_id: "your-key"      # 访问密钥ID
  secret_access_key: "your-secret" # 访问密钥
  bucket: "default-bucket"       # 默认存储桶
  force_path_style: false        # 是否使用路径样式URL
```

## 错误处理

所有接口都使用统一的错误格式：

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求数据格式错误",
    "details": "具体错误信息"
  }
}
```

常见错误码：
- `VALIDATION_ERROR`: 请求参数验证失败
- `NOT_FOUND`: 资源不存在
- `INTERNAL_SERVER_ERROR`: 服务器内部错误
- `DATABASE_ERROR`: 数据库操作错误
