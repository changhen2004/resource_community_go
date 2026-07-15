# resource_community_go

基于 `Go + Gin + Gorm + Redis + RabbitMQ + Vue3` 的资源社区项目，补充了 Prometheus + Grafana 可观测性和本地压测演练脚本，用于记录 QPS、P50/P95、错误率等基础数据。

## 项目介绍

resource_community_go 是一个围绕“资源分享”场景构建的内容社区项目。  
项目当前包含用户系统、资源发布、资源浏览、点赞、评论、收藏、积分、签到、资源解锁、图片上传、热门资源流、异步 Worker、基础可观测性和压测演练报告等能力。

这个项目不是单纯堆 CRUD，而是把一个内容社区的主链路做完整：

- 用户可以注册登录、发布资源、浏览资源、点赞评论收藏，并通过积分解锁资源。
- 浏览、点赞、评论、收藏等互动行为会影响热榜，热度由 Redis ZSet 维护。
- 发布资源、浏览、点赞、评论、收藏等行为通过 RabbitMQ 投递异步任务，由 Worker 更新热度和积分。
- 后端暴露 Prometheus 指标，Grafana 展示 QPS、P50/P95、错误率和路由维度延迟。
- 提供本地演练脚本，生成可复盘的 Markdown 报告，作为压测和排障案例证据。

## 技术栈

### 后端

- Go
- Gin
- Gorm
- MySQL
- Redis
- RabbitMQ
- JWT
- Prometheus
- Grafana

### 前端

- Vue 3
- TypeScript
- Pinia
- Vue Router
- Element Plus
- Vite

## 功能

| 模块 | 功能 |
|------|------|
| 账号 | 注册、登录、Refresh Token、登出、Bearer Token 鉴权 |
| 资源 | 发布、列表、详情、分页、关键词搜索、标签筛选、最新/热度排序、资源解锁 |
| 点赞 | 点赞、点赞计数读取、RabbitMQ 异步落库、热度更新 |
| 评论 | 评论列表、发布、删除、RabbitMQ 异步积分发放、热度更新 |
| 收藏 | 收藏、取消收藏、我的收藏、热度更新 |
| 积分 | 积分余额、积分流水、每日签到、发布奖励、互动奖励、权益兑换 |
| 上传 | 封面图上传、正文配图上传 |
| 热榜 | Redis ZSet 热榜、浏览/点赞/评论/收藏参与热度计算、热榜缓存 |
| 可观测性 | `/metrics` 指标暴露、Prometheus 抓取、Grafana 仪表盘、P50/P95、QPS、错误率 |
| 演练 | 本地压测脚本、Prometheus 指标快照、Markdown 演练报告 |
| 工程 | Docker Compose、API/Worker 拆分运行、健康检查、pprof、GitHub Actions CI/CD |

## 项目结构

```text
resource_community_go/
├── backend/
│   ├── cmd/
│   │   └── worker/
│   ├── config/
│   ├── internal/
│   │   ├── app/
│   │   ├── article/
│   │   ├── asyncjob/
│   │   ├── auth/
│   │   ├── cachekey/
│   │   ├── comment/
│   │   ├── exchange/
│   │   ├── favorite/
│   │   ├── media/
│   │   ├── points/
│   │   └── worker/
│   ├── utils/
│   └── main.go
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   ├── router/
│   │   ├── store/
│   │   ├── types/
│   │   └── views/
├── docs/
│   ├── observability.md
│   └── evidence/
├── observability/
│   ├── prometheus/
│   └── grafana/
├── scripts/
│   └── observability_drill.sh
├── docker-compose.yml
└── README.md
```

## 后端模块说明

- `internal/auth`：认证与用户登录
- `internal/article`：资源发布、列表、详情、热榜
- `internal/comment`：评论相关能力
- `internal/favorite`：收藏相关能力
- `internal/points`：积分、签到、解锁、权益兑换
- `internal/media`：文件上传
- `internal/asyncjob`：异步任务定义与发布
- `internal/worker`：异步 Worker 消费处理
- `internal/app`：路由、鉴权、中间件、可观测性

## 核心链路

### 资源互动与热榜

资源浏览、点赞、评论、收藏会影响文章热度：

1. API 接收到用户行为。
2. 主流程完成必要的同步响应。
3. 行为事件投递到 RabbitMQ。
4. Worker 消费任务，更新浏览量、点赞数、积分或 Redis ZSet 热榜分数。
5. 列表、详情和热榜缓存按需失效或重建。

这种设计让接口响应和后续统计更新解耦，同时保留消息队列异常时的同步兜底逻辑，避免核心流程完全依赖 RabbitMQ。

### 可观测性与演练

后端通过 `/metrics` 暴露 HTTP 指标：

- `resource_community_http_requests_total`
- `resource_community_http_request_duration_seconds`

Prometheus 根据这些指标计算：

- QPS
- P50 / P95 延迟
- 非 2xx 错误率
- 5xx 错误率
- 路由维度的 QPS 和 P95

Grafana 默认加载 `Resource Community API` 仪表盘。更多说明见 [docs/observability.md](docs/observability.md)。

## 接口清单

### 健康检查

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/healthz` | 否 | 服务健康检查 |

### 账号 `/api/auth`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/login` | 否 | 登录，返回 access_token + refresh_token |
| POST | `/register` | 否 | 注册 |
| POST | `/refresh` | 否 | 刷新 access_token |
| POST | `/logout` | JWT | 登出 |

### 资源 `/api/articles`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/articles` | 否 | 资源列表，支持分页、排序、关键词、标签 |
| GET | `/articles/hot` | 否 | 热门资源列表 |
| GET | `/articles/:id` | 否 | 资源详情 |
| POST | `/articles` | JWT | 发布资源 |
| POST | `/articles/:id/unlock` | JWT | 积分解锁资源 |

### 点赞 `/api/articles`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/articles/:id/like` | 否 | 获取点赞计数 |
| POST | `/articles/:id/like` | JWT | 点赞 |

### 评论 `/api`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/articles/:id/comments` | 否 | 评论列表 |
| POST | `/articles/:id/comments` | JWT | 发布评论 |
| DELETE | `/comments/:id` | JWT | 删除评论 |

### 收藏 `/api`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/articles/:id/favorite` | JWT | 收藏资源 |
| DELETE | `/articles/:id/favorite` | JWT | 取消收藏 |
| GET | `/me/favorites` | JWT | 我的收藏列表 |

### 积分 `/api/me`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/me/points` | JWT | 我的积分摘要 |
| GET | `/me/points/records` | JWT | 我的积分流水 |
| POST | `/me/check-in` | JWT | 每日签到 |
| POST | `/me/points/redeem` | JWT | 兑换权益 |

### 上传 `/api/uploads`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/uploads/cover` | JWT | 上传封面图 |
| POST | `/uploads/content-images` | JWT | 上传正文配图 |


## 本地运行

项目已提供 `.env.example`：

```bash
cp .env.example .env
```

### Docker Compose 启动

```bash
docker compose up --build
docker compose up -d
```

默认启动：

- MySQL
- Redis
- RabbitMQ
- Backend
- Worker
- Frontend
- Prometheus
- Grafana

访问地址：

- Frontend: `http://localhost:5173`
- Backend API: `http://localhost:8080/api`
- Health Check: `http://localhost:8080/healthz`
- Metrics: `http://localhost:8080/metrics`
- Prometheus: `http://localhost:9091`
- Grafana: `http://localhost:3001`，默认账号 `admin/admin`
- RabbitMQ 管理台: `http://localhost:15674`

停止：

```bash
docker compose down
```

### 分别启动

#### 启动后端

```bash
cd backend
go mod download
go run .
```

#### 启动 Worker

```bash
cd backend
go run ./cmd/worker
```

#### 启动前端

```bash
cd frontend
npm ci
npm run dev
```

## 压测演练

启动 Docker Compose 后，可以运行演练脚本生成基础流量和指标报告：

```bash
scripts/observability_drill.sh --duration 90 --concurrency 12
```

如果需要让错误率面板有可截图数据，可以加入少量 404 请求：

```bash
scripts/observability_drill.sh --duration 90 --concurrency 12 --include-error-traffic
```

报告会生成到：

```text
docs/evidence/observability-drill-<timestamp>.md
```

报告包含：

- 本地请求数和错误率
- Prometheus 查询到的 QPS、P50、P95、非 2xx 错误率、5xx 错误率
- Grafana 截图建议
- 可关联到 OnCallAgent 的排障案例记录

## 与 OnCallAgent 联动

本项目的 Prometheus 指标和演练报告可作为 OnCallAgent 的实践数据来源：

1. 使用 Grafana/Prometheus 发现接口延迟、错误率或服务不可用问题。
2. 将社区项目排障文档上传到 OnCallAgent 知识库。
3. OnCallAgent 通过 Prometheus 工具查询当前活跃告警。
4. Agent 检索对应排障文档，生成处理建议。

当前适合导入 OnCallAgent 的排障文档包括：

- `resource-community-p95-latency.md`
- `resource-community-error-rate.md`
- `resource-community-hot-ranking.md`
- `resource-community-rabbitmq-backlog.md`

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `RESOURCE_COMMUNITY_GO_APP_PORT` | `8080` | 后端监听端口 |
| `RESOURCE_COMMUNITY_GO_ENABLE_PPROF` | `false` | 是否开启 pprof |
| `RESOURCE_COMMUNITY_GO_SLOW_REQUEST_THRESHOLD_MS` | `500` | 慢请求阈值，单位毫秒 |
| `RESOURCE_COMMUNITY_GO_DATABASE_DSN` | `resource_community_go:resource_community_go@tcp(mysql:3306)/resource_community_go?charset=utf8mb4&parseTime=True&loc=Local` | MySQL 连接串 |
| `RESOURCE_COMMUNITY_GO_REDIS_ADDR` | `redis:6379` | Redis 地址 |
| `RESOURCE_COMMUNITY_GO_REDIS_PASSWORD` | 空 | Redis 密码 |
| `RESOURCE_COMMUNITY_GO_REDIS_DB` | `0` | Redis DB |
| `RESOURCE_COMMUNITY_GO_RABBITMQ_URL` | `amqp://guest:guest@rabbitmq:5672/` | RabbitMQ 连接地址 |
| `RESOURCE_COMMUNITY_GO_RABBITMQ_EXCHANGE` | `resource_community_go.async` | RabbitMQ Exchange |
| `RESOURCE_COMMUNITY_GO_RABBITMQ_QUEUE` | `resource_community_go.async.jobs` | RabbitMQ Queue |
| `RESOURCE_COMMUNITY_GO_JWT_SECRET` | `change-me-in-production` | JWT 签名密钥 |
| `RESOURCE_COMMUNITY_GO_UPLOAD_DIR` | `uploads` | 上传目录 |
| `VITE_API_BASE_URL` | `http://localhost:3000/api` | 前端 API 基地址 |
| `GRAFANA_ADMIN_USER` | `admin` | Grafana 管理员用户名 |
| `GRAFANA_ADMIN_PASSWORD` | `admin` | Grafana 管理员密码 |

详见 `.env.example`。

## 测试

### 后端测试

```bash
cd backend
go test ./...
```

### 前端构建校验

```bash
cd frontend
npm run build
```

### 配置校验

```bash
docker compose config
python3 -m json.tool observability/grafana/provisioning/dashboards/resource-community-api.json
```

### Compose 配置校验

```bash
docker compose config
```


### CI

当前 CI 包含：

- 后端测试
- 前端构建
- Docker Compose 配置校验
- Docker 镜像构建校验
