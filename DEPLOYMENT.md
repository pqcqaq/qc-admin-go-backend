# 🚀 Go Backend 模板部署指南

## 📋 部署前准备

### 系统要求

- **Go**: 1.21 或更高版本
- **数据库**: SQLite (默认) / MySQL / PostgreSQL
- **Redis**: 可选，用于缓存
- **S3存储**: 可选，用于文件存储

### 环境准备

```bash
# 检查Go版本
go version

# 检查Redis (可选)
redis-cli ping

# 检查数据库连接 (MySQL示例)
mysql -u username -p -h localhost
```

## 🔧 本地开发部署

### 1. 克隆并配置项目

```bash
# 克隆项目
git clone <your-project-repo>
cd go-backend

# 安装依赖
go mod download

# 生成数据库代码
go generate ./database/generate.go
```

### 2. 配置文件

```bash
# 复制配置文件
cp config.yaml config.local.yaml

# 编辑本地配置
vim config.local.yaml
```

示例配置：
```yaml
server:
  host: "localhost"
  port: 8080
  mode: "debug"

database:
  driver: "sqlite3"
  source: "./data/app.db"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

s3:
  endpoint: ""
  region: "us-east-1"
  bucket: "your-bucket"
  access_key: "your-access-key"
  secret_key: "your-secret-key"
```

### 3. 启动服务

```bash
# 开发模式启动
go run main.go -c config.local.yaml

# 或编译后启动
go build -o server.exe .
./server.exe -c config.local.yaml
```

## 🐳 Docker 部署

### 1. 创建 Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 生成数据库代码
RUN go generate ./database/generate.go

# 编译应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# 运行阶段
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata sqlite

WORKDIR /root/

# 复制编译后的程序
COPY --from=builder /app/main .
COPY --from=builder /app/config.prod.yaml ./config.yaml

# 创建数据目录
RUN mkdir -p /root/data

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./main"]
```

### 2. 构建和运行

```bash
# 构建镜像
docker build -t go-backend:latest .

# 运行容器
docker run -d \
  --name go-backend \
  -p 8080:8080 \
  -v $(pwd)/data:/root/data \
  -e DB_PATH="/root/data/app.db" \
  go-backend:latest
```

### 3. Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    volumes:
      - ./data:/root/data

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: appdb
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

启动所有服务：
```bash
docker-compose up -d
```

## ☁️ 云服务部署

### AWS ECS 部署

1. **创建任务定义**

```json
{
  "family": "go-backend",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::123456789012:role/ecsTaskExecutionRole",
  "containerDefinitions": [
    {
      "name": "go-backend",
      "image": "your-account.dkr.ecr.region.amazonaws.com/go-backend:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "DB_HOST",
          "value": "your-rds-endpoint"
        },
        {
          "name": "REDIS_HOST",
          "value": "your-elasticache-endpoint"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/go-backend",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

2. **创建服务**

```bash
aws ecs create-service \
  --cluster your-cluster \
  --service-name go-backend-service \
  --task-definition go-backend:1 \
  --desired-count 2 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-12345,subnet-67890],securityGroups=[sg-abcdef],assignPublicIp=ENABLED}"
```

### Kubernetes 部署

1. **创建 Deployment**

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-backend
  labels:
    app: go-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-backend
  template:
    metadata:
      labels:
        app: go-backend
    spec:
      containers:
      - name: go-backend
        image: your-registry/go-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: REDIS_HOST
          value: "redis-service"
        - name: S3_BUCKET
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: s3-bucket
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

2. **创建 Service**

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: go-backend-service
spec:
  selector:
    app: go-backend
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer
```

3. **部署到集群**

```bash
# 应用配置
kubectl apply -f k8s/

# 检查部署状态
kubectl get pods -l app=go-backend
kubectl get services
```

## 🔧 生产环境优化

### 1. 性能优化

```yaml
# config.prod.yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"  # 生产模式

database:
  driver: "postgres"
  source: "host=db-host user=dbuser password=dbpass dbname=proddb sslmode=require"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: "5m"

redis:
  addr: "redis-cluster:6379"
  password: "redis-password"
  db: 0
  pool_size: 10
  min_idle_conns: 5

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

### 2. 安全配置

```bash
# 环境变量方式配置敏感信息
export DB_PASSWORD="your-secure-password"
export REDIS_PASSWORD="redis-secure-password"
export S3_SECRET_KEY="your-s3-secret"
export JWT_SECRET="your-jwt-secret"

# 启动应用
./server -c config.prod.yaml
```

### 3. 反向代理配置 (Nginx)

```nginx
# /etc/nginx/sites-available/go-backend
upstream go_backend {
    server 127.0.0.1:8080;
    # 如果有多个实例
    # server 127.0.0.1:8081;
    # server 127.0.0.1:8082;
}

server {
    listen 80;
    server_name your-domain.com;

    # HTTPS重定向
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/certificate.crt;
    ssl_certificate_key /path/to/private.key;

    location / {
        proxy_pass http://go_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    location /health {
        proxy_pass http://go_backend/health;
        access_log off;
    }

    # 文件上传大小限制
    client_max_body_size 50M;
}
```

## 📊 监控和日志

### 1. 日志收集

```yaml
# docker-compose.logging.yml
version: '3.8'

services:
  app:
    build: .
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  # 使用ELK Stack
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.14.0
    environment:
      - discovery.type=single-node
    volumes:
      - es_data:/usr/share/elasticsearch/data

  logstash:
    image: docker.elastic.co/logstash/logstash:7.14.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf

  kibana:
    image: docker.elastic.co/kibana/kibana:7.14.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200

volumes:
  es_data:
```

### 2. 健康检查脚本

```bash
#!/bin/bash
# scripts/health_check.sh

ENDPOINT="http://localhost:8080/health"
TIMEOUT=10

response=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout $TIMEOUT $ENDPOINT)

if [ $response -eq 200 ]; then
    echo "Service is healthy"
    exit 0
else
    echo "Service is unhealthy (HTTP $response)"
    exit 1
fi
```

### 3. 系统服务配置 (systemd)

```ini
# /etc/systemd/system/go-backend.service
[Unit]
Description=Go Backend Service
After=network.target

[Service]
Type=simple
User=appuser
WorkingDirectory=/opt/go-backend
ExecStart=/opt/go-backend/server -c /opt/go-backend/config.prod.yaml
Restart=always
RestartSec=5
Environment=PATH=/usr/bin:/usr/local/bin
Environment=DB_PASSWORD=your-db-password
Environment=REDIS_PASSWORD=your-redis-password

# 日志配置
StandardOutput=journal
StandardError=journal
SyslogIdentifier=go-backend

# 安全配置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/go-backend/data

[Install]
WantedBy=multi-user.target
```

启用服务：
```bash
sudo systemctl enable go-backend
sudo systemctl start go-backend
sudo systemctl status go-backend
```

## 🚨 故障排除

### 常见问题

1. **数据库连接失败**
   ```bash
   # 检查数据库连接
   telnet db-host 5432
   
   # 检查用户权限
   psql -h db-host -U username -d dbname
   ```

2. **Redis连接失败**
   ```bash
   # 检查Redis连接
   redis-cli -h redis-host ping
   
   # 检查Redis配置
   redis-cli -h redis-host info
   ```

3. **文件上传失败**
   ```bash
   # 检查S3权限
   aws s3 ls s3://your-bucket
   
   # 检查磁盘空间
   df -h
   ```

### 日志分析

```bash
# 查看应用日志
tail -f /var/log/go-backend/app.log

# 使用jq解析JSON日志
tail -f app.log | jq '.level, .message, .timestamp'

# 查看错误日志
grep "ERROR" app.log | tail -20
```

---

✅ **部署完成！** 您的Go Backend服务现在应该正在运行。访问 `http://your-domain.com/health` 检查服务状态。
