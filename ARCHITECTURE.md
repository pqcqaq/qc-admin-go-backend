# Go Backend 模板架构文档

## 📖 架构概述

这是一个企业级Go后端项目模板的技术架构文档。模板采用分层架构设计，集成了现代Web开发的最佳实践，提供了完整的用户管理、文件处理、缓存等基础设施，并包含扫描管理作为业务逻辑示例。

## 🏗️ 设计理念

### 核心原则

1. **分层架构** - 清晰的职责分离
2. **依赖注入** - 低耦合，高内聚
3. **配置外部化** - 多环境支持
4. **错误集中处理** - 统一错误处理机制
5. **代码生成** - 减少重复劳动
6. **云原生** - 容器化和微服务友好

### 技术选择

- **Web框架**: Gin - 高性能，中间件丰富
- **ORM**: Ent - 类型安全，代码生成
- **缓存**: Redis - 高性能分布式缓存
- **存储**: S3 兼容 - 云存储标准
- **配置**: Viper - 多格式配置支持

## 技术架构图

```mermaid
graph TB
    %% 客户端层
    subgraph "客户端层"
        Web[Web前端]
        Mobile[移动端App]
        API_Client[API客户端]
    end
    
    %% 负载均衡
    LB[负载均衡器<br/>Nginx/HAProxy]
    
    %% 应用服务层
    subgraph "应用服务层"
        App1[Go Backend服务1]
        App2[Go Backend服务2]
        App3[Go Backend服务N]
    end
    
    %% 中间件层
    subgraph "中间件组件"
        Redis[(Redis缓存)]
        MQ[消息队列]
    end
    
    %% 数据存储层
    subgraph "数据存储层"
        Database[(主数据库<br/>SQLite/MySQL/PostgreSQL)]
        ReadDB[(只读副本)]
        S3[(AWS S3<br/>文件存储)]
    end
    
    %% 监控日志
    subgraph "监控与日志"
        Log[日志系统]
        Monitor[监控系统]
        Metrics[指标收集]
    end
    
    %% 连接关系
    Web --> LB
    Mobile --> LB
    API_Client --> LB
    
    LB --> App1
    LB --> App2
    LB --> App3
    
    App1 --> Redis
    App2 --> Redis
    App3 --> Redis
    
    App1 --> Database
    App2 --> Database
    App3 --> Database
    
    App1 --> ReadDB
    App2 --> ReadDB
    App3 --> ReadDB
    
    App1 --> S3
    App2 --> S3
    App3 --> S3
    
    App1 --> Log
    App2 --> Monitor
    App3 --> Metrics
    
    style Web fill:#e1f5fe
    style Mobile fill:#e1f5fe
    style Database fill:#fff3e0
    style Redis fill:#ffebee
    style S3 fill:#f3e5f5
```

## 应用内部架构

```mermaid
graph TB
    subgraph "HTTP层"
        Router[Gin路由器]
        MW[中间件链]
        Router --> MW
    end
    
    subgraph "中间件层"
        MW --> CORS[CORS处理]
        MW --> Auth[认证中间件]
        MW --> Log[日志中间件]
        MW --> Error[错误处理]
    end
    
    subgraph "控制器层"
        Error --> UserHandler[用户控制器]
        Error --> ScanHandler[扫描控制器]
        Error --> AttachmentHandler[附件控制器]
        Error --> HealthHandler[健康检查]
    end
    
    subgraph "业务逻辑层"
        UserHandler --> UserFunc[用户业务逻辑]
        ScanHandler --> ScanFunc[扫描业务逻辑]
        AttachmentHandler --> AttachmentFunc[附件业务逻辑]
    end
    
    subgraph "服务层"
        UserFunc --> CacheService[缓存服务]
        ScanFunc --> ExcelService[Excel服务]
        AttachmentFunc --> S3Service[S3服务]
        AttachmentFunc --> LoggingService[日志服务]
    end
    
    subgraph "数据访问层"
        CacheService --> EntClient[Ent ORM客户端]
        ExcelService --> EntClient
        S3Service --> EntClient
        LoggingService --> EntClient
    end
    
    subgraph "外部依赖"
        EntClient --> DB[(数据库)]
        CacheService --> Redis[(Redis)]
        S3Service --> S3[(AWS S3)]
    end
    
    style Router fill:#e3f2fd
    style EntClient fill:#fff3e0
    style DB fill:#fff3e0
    style Redis fill:#ffebee
    style S3 fill:#f3e5f5
```

## 数据库设计

```mermaid
erDiagram
    User {
        uint64 id PK
        string name
        string email
        string password_hash
        time created_at
        time updated_at
        time deleted_at
        uint64 attachment_id FK
    }
    
    Scan {
        uint64 id PK
        string content
        time created_at
        time updated_at
        time deleted_at
        uint64 attachment_id FK
    }
    
    Attachment {
        uint64 id PK
        string filename
        string original_filename
        string content_type
        int64 size
        string s3_key
        string s3_bucket
        string s3_region
        time created_at
        time updated_at
        time deleted_at
    }
    
    Logging {
        uint64 id PK
        string level
        string message
        string context
        time created_at
        time updated_at
        time deleted_at
    }
    
    User ||--o| Attachment : "has profile image"
    Scan ||--o| Attachment : "has attachment"
```

## 目录结构设计

```mermaid
graph TD
    Root[go-backend/] --> Config[配置文件]
    Root --> Main[main.go]
    Root --> Database[database/]
    Root --> Internal[internal/]
    Root --> Pkg[pkg/]
    Root --> Shared[shared/]
    Root --> Docker[docker-compose/]
    
    Database --> Schema[schema/]
    Database --> Ent[ent/]
    Database --> Mixins[mixins/]
    
    Internal --> Handlers[handlers/]
    Internal --> Routes[routes/]
    Internal --> Middleware[middleware/]
    Internal --> Funcs[funcs/]
    
    Pkg --> Configs[configs/]
    Pkg --> DB[database/]
    Pkg --> Cache[caching/]
    Pkg --> S3[s3/]
    Pkg --> Excel[excel/]
    Pkg --> Utils[utils/]
    Pkg --> Logging[logging/]
    
    Shared --> Models[models/]
    
    style Root fill:#e1f5fe
    style Database fill:#fff3e0
    style Internal fill:#e8f5e8
    style Pkg fill:#fce4ec
    style Shared fill:#f3e5f5
```

## 请求处理流程

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant Router as Gin路由
    participant MW as 中间件
    participant Handler as 处理器
    participant Logic as 业务逻辑
    participant Cache as Redis缓存
    participant DB as 数据库
    participant S3 as S3存储
    
    Client->>Router: HTTP请求
    Router->>MW: 应用中间件
    MW->>MW: CORS处理
    MW->>MW: 认证验证
    MW->>MW: 日志记录
    MW->>Handler: 路由到处理器
    
    Handler->>Logic: 调用业务逻辑
    Logic->>Cache: 检查缓存
    
    alt 缓存命中
        Cache-->>Logic: 返回缓存数据
    else 缓存未命中
        Logic->>DB: 查询数据库
        DB-->>Logic: 返回数据
        Logic->>Cache: 更新缓存
    end
    
    opt 文件操作
        Logic->>S3: 上传/下载文件
        S3-->>Logic: 返回结果
    end
    
    Logic-->>Handler: 返回处理结果
    Handler-->>Router: HTTP响应
    Router-->>Client: 返回结果
```

## 部署架构

```mermaid
graph TB
    subgraph "生产环境"
        subgraph "DMZ区域"
            WAF[Web应用防火墙]
            LB[负载均衡器]
        end
        
        subgraph "应用区域"
            App1[Go服务实例1]
            App2[Go服务实例2]
            App3[Go服务实例3]
        end
        
        subgraph "数据区域"
            Master[(主数据库)]
            Slave[(从数据库)]
            Redis_Cluster[Redis集群]
        end
        
        subgraph "存储区域"
            S3_Primary[S3主存储]
            S3_Backup[S3备份存储]
        end
        
        subgraph "监控区域"
            Monitor[监控系统]
            Log_Server[日志服务器]
            Alert[告警系统]
        end
    end
    
    Internet[互联网] --> WAF
    WAF --> LB
    LB --> App1
    LB --> App2
    LB --> App3
    
    App1 --> Master
    App2 --> Master
    App3 --> Master
    
    App1 --> Slave
    App2 --> Slave
    App3 --> Slave
    
    App1 --> Redis_Cluster
    App2 --> Redis_Cluster
    App3 --> Redis_Cluster
    
    App1 --> S3_Primary
    App2 --> S3_Primary
    App3 --> S3_Primary
    
    S3_Primary --> S3_Backup
    
    App1 --> Monitor
    App2 --> Log_Server
    App3 --> Alert
    
    style Internet fill:#e1f5fe
    style Master fill:#fff3e0
    style Redis_Cluster fill:#ffebee
    style S3_Primary fill:#f3e5f5
```

## 安全架构

```mermaid
graph TB
    subgraph "安全防护层"
        Firewall[防火墙]
        WAF[Web应用防火墙]
        DDoS[DDoS防护]
    end
    
    subgraph "认证授权"
        JWT[JWT令牌]
        RBAC[角色权限控制]
        OAuth[OAuth认证]
    end
    
    subgraph "数据安全"
        Encrypt[数据加密]
        Hash[密码哈希]
        TLS[TLS传输]
    end
    
    subgraph "审计监控"
        AuditLog[审计日志]
        Monitor[安全监控]
        Alert[安全告警]
    end
    
    Internet[互联网请求] --> Firewall
    Firewall --> WAF
    WAF --> DDoS
    DDoS --> JWT
    JWT --> RBAC
    RBAC --> App[应用服务]
    
    App --> Encrypt
    App --> Hash
    App --> TLS
    
    App --> AuditLog
    AuditLog --> Monitor
    Monitor --> Alert
    
    style Internet fill:#ffcdd2
    style JWT fill:#c8e6c9
    style Encrypt fill:#fff3e0
    style Monitor fill:#e1f5fe
```

## 扩展性设计

系统采用分层架构和微服务设计原则，具备良好的扩展性：

1. **水平扩展**: 应用服务无状态设计，支持负载均衡
2. **垂直扩展**: 模块化设计，便于功能扩展
3. **存储扩展**: 支持主从复制、分库分表
4. **缓存扩展**: Redis集群支持
5. **文件存储**: 云存储服务，自动扩展

## 性能优化

1. **缓存策略**: 多级缓存设计
2. **数据库优化**: 读写分离、索引优化
3. **连接池**: 数据库连接池管理
4. **异步处理**: 文件上传、邮件发送等异步处理
5. **CDN加速**: 静态资源CDN分发

## 监控指标

- **应用指标**: 响应时间、吞吐量、错误率
- **系统指标**: CPU、内存、磁盘使用率
- **数据库指标**: 连接数、慢查询、锁等待
- **缓存指标**: 命中率、内存使用、连接数
- **业务指标**: 用户活跃度、功能使用情况
