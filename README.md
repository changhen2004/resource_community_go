# GinDemo ExchangeAPP

## 开发环境启动

项目提供 `Docker Compose` 开发环境，可一键启动：

- MySQL
- Redis
- Backend
- Frontend

启动命令：

```bash
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
