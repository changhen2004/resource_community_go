# Resource Community Platform

基于 Go 构建的高并发内容社区服务平台，模拟真实互联网社区业务场景，围绕内容分发、用户互动、异步任务处理、缓存治理和可观测性体系进行工程化设计。

## 项目背景

Resource Community Platform 是一个面向资源分享场景的内容社区系统。

项目参考互联网内容平台架构设计，实现用户从内容生产、内容分发到互动反馈的完整业务闭环。

### 系统覆盖

- 用户认证体系
- 内容发布与管理
- Feed 流分发
- 点赞评论收藏
- 用户关注关系
- 积分激励体系
- 热榜计算
- 异步任务处理
- 服务监控与故障分析

### 项目重点

- 高并发读场景优化
- 热点数据缓存治理
- 异步任务解耦
- 消息可靠性保证
- 服务可观测性建设

## 系统架构

```text
Client
  |
HTTP API
  |
+---------------+
|     Gin       |
+---------------+
  |
+-------------+-------------+
|                           |
Business Service        Async Service
|                           |
MySQL Redis             RabbitMQ
                                |
                              Worker
                                |
                   +----------+----------+
                   |                     |
              热度计算              积分处理

Prometheus
  |
Grafana
```

## 技术栈

### Backend

| 技术 | 用途 |
|---|---|
| Go | 后端服务开发 |
| Gin | HTTP API 框架 |
| Gorm | ORM 数据访问 |
| MySQL | 业务数据存储 |
| Redis | 缓存、排行榜、热点数据 |
| RabbitMQ | 异步任务消息队列 |
| JWT | 用户认证 |
| Prometheus | 指标采集 |
| Grafana | 可视化监控 |
| Docker Compose | 服务编排 |

### Frontend

| 技术 | 用途 |
|---|---|
| Vue3 | 前端框架 |
| TypeScript | 类型约束 |
| Pinia | 状态管理 |
| Vue Router | 路由管理 |
| Element Plus | UI 组件 |
| Vite | 构建工具 |

## 核心业务模块

### 用户认证

实现完整用户生命周期：

- 用户注册
- 登录
- Refresh Token
- JWT 鉴权
- 登出

认证流程：

```text
用户登录
  |
生成 Access Token + Refresh Token
  |
访问业务接口
  |
JWT Middleware 校验
```

### 内容服务

支持：

- 内容发布
- 内容列表
- 内容详情
- 分类筛选
- 标签搜索
- 资源解锁

核心接口：

```text
GET    /articles
GET    /articles/:id
POST   /articles
POST   /articles/:id/unlock
```

### Feed 分发系统

系统提供三类内容流。

#### 最新流

按照发布时间倒序返回：

```text
created_at DESC
```

#### 关注流

针对关注关系设计：

- 游标分页
- 用户级缓存

分页方式：

```text
(created_at, id)
instead of
offset pagination
```

避免大数据量情况下 offset 查询性能下降。

#### 热门流

基于 Redis ZSet 实现：

```text
article_id -> score

score =
浏览权重
+ 点赞权重
+ 评论权重
+ 收藏权重
```

## 高并发优化设计

### Redis 缓存治理

资源详情采用：

```text
Request
  |
Redis Cache
  |
Miss
  |
MySQL
```

针对缓存异常场景增加：

- 缓存穿透
- 缓存雪崩
- 缓存击穿

缓存穿透方案：

- 空值缓存
- 不存在资源 ID 写入 `article:not_found`
- TTL: `60s`

缓存雪崩方案：

- TTL 随机化
- `expire = 3600 + random()`

避免大量 Key 同时失效。

缓存击穿方案：

- `singleflight` 请求合并
- 1000 请求只触发 1 次 DB 查询

## 异步任务架构

用户行为不会直接阻塞主链路。

例如点赞流程：

```text
Client
  |
API
  |
更新点赞状态
  |
RabbitMQ
  |
Worker
  |
更新:
- 热榜
- 积分
- 统计数据
```

收益：

- 降低接口响应时间
- 提高系统吞吐能力
- 解耦业务模块

## RabbitMQ 可靠性设计

针对消息可靠性，实现：

### 消息重试

- `retry_count < 3` 时重新投递
- `retry_count >= 3` 时进入 DLQ

### 死信队列

异常消息进入 `xxx.dlq`，并保存：

- 原始消息
- 失败原因
- 重试次数

方便后续人工处理。

### 幂等消费

Worker 使用 Redis 维护：

```text
message_id
processing
done
```

避免：

- 重复积分
- 重复增加热度
- 重复统计

## 可观测性建设

系统集成：

```text
Go Service
  |
/metrics
  |
Prometheus
  |
Grafana
```

采集指标：

- 服务指标
- QPS
- P50 latency
- P95 latency
- HTTP error rate
- 路由指标

示例路由：

- `GET /articles`
- `GET /articles/hot`
- `GET /articles/:id`

分别统计：

- 请求量
- 延迟
- 错误率

## 压测验证

测试环境：

- Docker Compose
- CPU: 本地开发环境
- Duration: `90s`
- Concurrency: `12`

测试结果：

| 指标 | 结果 |
|---|---|
| 总请求 | 67955 |
| 服务端 QPS | 476 req/s |
| P50 | 2.5ms |
| P95 | 4.8ms |
| HTTP 非 2xx | 0% |
| 5xx 错误 | 0% |

说明：

系统在持续压力访问下保持稳定响应。

真实压测示例见 [docs/benchmark.md](docs/benchmark.md)。

## OnCallAgent 联动

项目提供真实业务监控数据：

```text
Resource Community
  |
Prometheus
  |
Alert
  |
OnCallAgent
  |
RAG + Agent
  |
生成排障方案
```

支持场景：

| 问题 | 排查文档 |
|---|---|
| 接口 P95 升高 | `resource-community-p95-latency` |
| 错误率升高 | `resource-community-error-rate` |
| 热榜异常 | `resource-community-hot-ranking` |
| RabbitMQ 积压 | `resource-community-rabbitmq-backlog` |

## 工程化能力

项目包含：

- Docker Compose 一键部署
- Backend / Worker 服务拆分
- Health Check
- pprof 性能分析
- GitHub Actions CI
- 自动化测试
- 配置管理

## 项目目录

```text
resource-community/
├── backend/
│   ├── cmd/
│   │   └── worker/
│   └── internal/
│       ├── auth/
│       ├── article/
│       ├── comment/
│       ├── social/
│       ├── points/
│       ├── asyncjob/
│       └── worker/
├── frontend/
├── observability/
│   ├── prometheus/
│   └── grafana/
├── scripts/
├── docs/
└── docker-compose.yml
```

## 本地运行

启动：

```bash
docker compose up --build
```

服务地址：

| 服务 | 地址 |
|---|---|
| Frontend | `http://localhost:5173` |
| Backend | `http://localhost:8080` |
| Prometheus | `http://localhost:9091` |
| Grafana | `http://localhost:3001` |
| RabbitMQ | `http://localhost:15674` |

## 后续优化方向

- Redis Cluster 支持
- MySQL 读写分离
- 分布式链路追踪 OpenTelemetry
- Kubernetes 部署
- 自动扩缩容 HPA
- Agent 自动故障恢复
