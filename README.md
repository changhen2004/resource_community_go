# GinDemo ExchangeAPP

## 环境变量

项目已提供 [`.env.example`](/home/chg/Go_Project/GinDemo_ExchangeAPP/.env.example)。

常见做法：

```bash
cp .env.example .env
```

可根据不同环境覆盖：

- 本地开发：使用本机 `localhost` 地址
- 测试环境：替换数据库、Redis、JWT 密钥
- 生产环境：替换数据库、Redis、JWT 密钥、上传目录、前端 API 地址
- 可观测性：可通过环境变量控制 `pprof` 和慢请求阈值

## 开发环境启动

项目提供 `Docker Compose` 开发环境，可一键启动：

- MySQL
- Redis
- Backend
- Frontend

启动命令：

```bash
cp .env.example .env
docker compose up --build
```

启动后默认访问地址：

- Frontend: `http://localhost:5173`
- Backend API: `http://localhost:3000/api`
- MySQL: `localhost:3306`
- Redis: `localhost:6379`

停止命令：

```bash
docker compose down
```

如需连同数据卷一起清理：

```bash
docker compose down -v
```
