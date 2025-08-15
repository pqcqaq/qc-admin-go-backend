# MinIO + Redis 依赖服务启动脚本

## 启动服务

```bash
# 启动所有依赖服务
docker-compose -f docker-compose/dependency.yaml up -d

# 查看服务状态
docker-compose -f docker-compose/dependency.yaml ps

# 查看日志
docker-compose -f docker-compose/dependency.yaml logs -f
```

## 停止服务

```bash
# 停止所有服务
docker-compose -f docker-compose/dependency.yaml down

# 停止服务并删除数据卷（谨慎使用）
docker-compose -f docker-compose/dependency.yaml down -v
```

## MinIO 访问信息

- **MinIO API地址**: <http://localhost:9000>
- **MinIO 控制台**: <http://localhost:9001>
- **用户名**: minioadmin
- **密码**: minioadmin123
- **默认存储桶**: default-bucket

## Redis 访问信息

- **Redis地址**: localhost:6379
- **无密码**

## 注意事项

1. 首次启动时，minio-init容器会自动创建 `default-bucket` 存储桶
2. 存储桶策略设置为公共读取，适合开发环境使用
3. 数据持久化到Docker卷中，重启容器数据不会丢失
4. 生产环境请修改默认密码和安全设置

## 验证服务

启动服务后，可以通过以下方式验证：

1. 访问 <http://localhost:9001> 打开MinIO控制台
2. 使用 minioadmin/minioadmin123 登录
3. 检查是否存在 default-bucket 存储桶

## 配置说明

您的 Go 应用程序的 `config.yaml` 已经更新为使用MinIO服务：

- endpoint: <http://localhost:9000>
- access_key_id: minioadmin
- secret_access_key: minioadmin123
- bucket: default-bucket
