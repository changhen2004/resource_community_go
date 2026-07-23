# resource_community_go 可观测性说明

本项目通过 Prometheus + Grafana 采集后端 HTTP 基础指标，用于观察 P50/P95、QPS、错误率和路由维度的延迟变化。

## 启动

```bash
docker compose up --build
```

访问地址：

- Backend: http://localhost:8080
- Metrics: http://localhost:8080/metrics
- Prometheus: http://localhost:9091
- Grafana: http://localhost:3001

Grafana 默认账号：

- 用户名：`admin`
- 密码：`admin`

默认仪表盘路径：

```text
Resource Community / Resource Community API
```

## 暴露指标

后端通过 `/metrics` 暴露以下指标：

- `resource_community_http_requests_total`：HTTP 请求总数，标签为 `method`、`path`、`status`
- `resource_community_http_request_duration_seconds`：HTTP 请求耗时直方图，标签为 `method`、`path`、`status`

`path` 使用 Gin 路由模板，例如 `/api/articles/:id`，避免将动态 ID 打成高基数标签。

## 常用 PromQL

QPS：

```promql
sum(rate(resource_community_http_requests_total[1m]))
```

5xx 错误率：

```promql
(
  sum(rate(resource_community_http_requests_total{status=~"5.."}[1m]))
  /
  clamp_min(sum(rate(resource_community_http_requests_total[1m])), 0.001)
) * 100
```

P50：

```promql
histogram_quantile(
  0.50,
  sum(rate(resource_community_http_request_duration_seconds_bucket[1m])) by (le)
)
```

P95：

```promql
histogram_quantile(
  0.95,
  sum(rate(resource_community_http_request_duration_seconds_bucket[1m])) by (le)
)
```

按路由查看 P95：

```promql
histogram_quantile(
  0.95,
  sum(rate(resource_community_http_request_duration_seconds_bucket[1m])) by (path, le)
)
```

## 本地制造基础流量

```bash
for i in $(seq 1 100); do
  curl -s http://localhost:8080/healthz >/dev/null
done
```

也可以访问资源相关接口观察路由维度指标：

```bash
curl -s "http://localhost:8080/api/articles?page=1&pageSize=10" >/dev/null
curl -s "http://localhost:8080/api/articles/hot?limit=10" >/dev/null
```

## 压测与演练报告

项目提供了一个轻量演练脚本，用于生成基础流量、从 Prometheus 查询指标，并在 `docs/evidence` 下生成 Markdown 报告。

默认运行：

```bash
scripts/observability_drill.sh
```

指定持续时间和并发：

```bash
scripts/observability_drill.sh --duration 90 --concurrency 12
```

生成少量 404 流量，用于截图验证非 2xx 错误率面板：

```bash
scripts/observability_drill.sh --duration 90 --concurrency 12 --include-error-traffic
```

报告内容包括：

- 本地请求数和错误率
- 接口维度请求数、非 2xx/失败数、错误率、平均耗时和 P95
- Prometheus 查询到的 QPS、P50、P95、非 2xx 错误率、5xx 错误率
- Grafana 截图建议
- 可关联到 OnCallAgent 知识库的排障案例记录

截图建议：

1. 打开 Grafana: http://localhost:3001
2. 进入 `Resource Community / Resource Community API`
3. 时间范围选择 `Last 15 minutes`
4. 截取 `QPS`、`Non-2xx Error Rate`、`P95 Latency`、`Latency P50/P95`、`QPS By Status`、`P95 Latency By Route`

## 当前告警规则

- `ResourceCommunityBackendDown`：Prometheus 无法抓取后端 `/metrics`
- `ResourceCommunityHighErrorRate`：1 分钟内 5xx 错误率超过 5%
- `ResourceCommunityHighP95Latency`：1 分钟内整体 P95 延迟超过 500ms
