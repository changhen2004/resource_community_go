#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9091}"
DURATION_SECONDS="${DURATION_SECONDS:-60}"
CONCURRENCY="${CONCURRENCY:-8}"
INCLUDE_ERROR_TRAFFIC="${INCLUDE_ERROR_TRAFFIC:-false}"
OUTPUT_DIR="${OUTPUT_DIR:-docs/evidence}"
REPORT_FILE="${REPORT_FILE:-}"

usage() {
  cat <<'USAGE'
Usage:
  scripts/observability_drill.sh [options]

Options:
  --base-url URL              Backend base URL. Default: http://localhost:8080
  --prometheus-url URL        Prometheus URL. Default: http://localhost:9091
  --duration SECONDS          Traffic duration. Default: 60
  --concurrency N             Concurrent workers. Default: 8
  --include-error-traffic     Send a small amount of 404 traffic for error-rate screenshots
  --output-dir DIR            Report output directory. Default: docs/evidence
  -h, --help                  Show help

Examples:
  scripts/observability_drill.sh --duration 90 --concurrency 12
  INCLUDE_ERROR_TRAFFIC=true scripts/observability_drill.sh
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --base-url)
      BASE_URL="$2"
      shift 2
      ;;
    --prometheus-url)
      PROMETHEUS_URL="$2"
      shift 2
      ;;
    --duration)
      DURATION_SECONDS="$2"
      shift 2
      ;;
    --concurrency)
      CONCURRENCY="$2"
      shift 2
      ;;
    --include-error-traffic)
      INCLUDE_ERROR_TRAFFIC="true"
      shift
      ;;
    --output-dir)
      OUTPUT_DIR="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

require_command curl
require_command date
require_command mktemp
require_command sort

BT='`'

if ! [[ "$DURATION_SECONDS" =~ ^[0-9]+$ ]] || [[ "$DURATION_SECONDS" -lt 10 ]]; then
  echo "--duration must be an integer >= 10" >&2
  exit 1
fi

if ! [[ "$CONCURRENCY" =~ ^[0-9]+$ ]] || [[ "$CONCURRENCY" -lt 1 ]]; then
  echo "--concurrency must be an integer >= 1" >&2
  exit 1
fi

check_url() {
  local url="$1"
  local label="$2"
  if ! curl -fsS --max-time 3 "$url" >/dev/null; then
    echo "$label is not reachable: $url" >&2
    exit 1
  fi
}

prom_query() {
  local query="$1"
  curl -fsS --get "${PROMETHEUS_URL}/api/v1/query" --data-urlencode "query=${query}"
}

prom_scalar() {
  local query="$1"
  prom_query "$query" | sed -n 's/.*"value":\[[^]]*,"\([^"]*\)".*/\1/p' | head -n 1
}

format_number() {
  local value="${1:-}"
  local digits="${2:-3}"
  if [[ -z "$value" || "$value" == "NaN" || "$value" == "+Inf" || "$value" == "-Inf" ]]; then
    printf "n/a"
    return
  fi
  awk -v value="$value" -v digits="$digits" 'BEGIN { printf "%.*f", digits, value }'
}

hit_endpoint() {
  local path="$1"
  local result
  if result="$(curl -sS -o /dev/null -w "%{http_code} %{time_total}" "${BASE_URL}${path}")"; then
    printf "%s %s\n" "$path" "$result"
  else
    printf "%s 000 0\n" "$path"
  fi
}

endpoint_summary_table() {
  local log_file="$1"
  local code_tick="$BT"
  if [[ ! -s "$log_file" ]]; then
    printf "| 接口 | 请求数 | 非 2xx/失败数 | 错误率 | 平均耗时 | P95 |\n"
    printf "|------|--------|----------------|--------|----------|-----|\n"
    return
  fi

  printf "| 接口 | 请求数 | 非 2xx/失败数 | 错误率 | 平均耗时 | P95 |\n"
  printf "|------|--------|----------------|--------|----------|-----|\n"
  awk '{ print $1 }' "$log_file" | sort -u | while read -r path; do
    local count
    local errors
    local error_rate
    local avg_latency
    local p95_latency

    count="$(awk -v path="$path" '$1 == path { count++ } END { print count + 0 }' "$log_file")"
    errors="$(awk -v path="$path" '$1 == path && $2 !~ /^2/ { count++ } END { print count + 0 }' "$log_file")"
    error_rate="$(awk -v total="$count" -v errors="$errors" 'BEGIN { if (total == 0) printf "0.00"; else printf "%.2f", errors / total * 100 }')"
    avg_latency="$(awk -v path="$path" '$1 == path { sum += $3; count++ } END { if (count == 0) printf "0.0000"; else printf "%.4f", sum / count }' "$log_file")"
    p95_latency="$(
      awk -v path="$path" '$1 == path { print $3 }' "$log_file" |
        sort -n |
        awk '{
          values[NR] = $1
        }
        END {
          if (NR == 0) {
            printf "0.0000"
            exit
          }
          percentile_index = int(NR * 0.95)
          if (percentile_index < 1) {
            percentile_index = 1
          }
          if (percentile_index < NR * 0.95) {
            percentile_index++
          }
          if (percentile_index > NR) {
            percentile_index = NR
          }
          printf "%.4f", values[percentile_index]
        }'
    )"

    printf "| %s%s%s | %s | %s | %s%% | %ss | %ss |\n" "$code_tick" "$path" "$code_tick" "$count" "$errors" "$error_rate" "$avg_latency" "$p95_latency"
  done
}

worker() {
  local end_epoch="$1"
  local worker_id="$2"
  local log_file="$3"
  local paths=(
    "/healthz"
    "/api/articles?page=1&pageSize=10"
    "/api/articles/hot?limit=10"
    "/api/articles/1"
  )

  local i=0
  while [[ "$(date +%s)" -lt "$end_epoch" ]]; do
    local path="${paths[$(((i + worker_id) % ${#paths[@]}))]}"
    hit_endpoint "$path" >>"$log_file"

    if [[ "$INCLUDE_ERROR_TRAFFIC" == "true" && $((i % 20)) -eq 0 ]]; then
      hit_endpoint "/api/not-found-for-observability-drill" >>"$log_file"
    fi
    i=$((i + 1))
  done
}

check_url "${BASE_URL}/healthz" "backend"
check_url "${BASE_URL}/metrics" "backend metrics"
check_url "${PROMETHEUS_URL}/-/ready" "prometheus"

mkdir -p "$OUTPUT_DIR"
run_id="$(date +%Y%m%d-%H%M%S)"
if [[ -z "$REPORT_FILE" ]]; then
  REPORT_FILE="${OUTPUT_DIR}/observability-drill-${run_id}.md"
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT
traffic_log="${tmp_dir}/traffic.log"
touch "$traffic_log"

start_iso="$(date -Iseconds)"
end_epoch="$(($(date +%s) + DURATION_SECONDS))"

echo "starting traffic: duration=${DURATION_SECONDS}s concurrency=${CONCURRENCY} include_error_traffic=${INCLUDE_ERROR_TRAFFIC}"
for worker_id in $(seq 1 "$CONCURRENCY"); do
  worker "$end_epoch" "$worker_id" "$traffic_log" &
done
wait
end_iso="$(date -Iseconds)"

echo "waiting for prometheus scrape..."
sleep 8

total_requests="$(wc -l <"$traffic_log" | tr -d ' ')"
local_error_count="$(awk '$2 !~ /^2/ { count++ } END { print count + 0 }' "$traffic_log")"
local_error_rate="$(awk -v total="$total_requests" -v errors="$local_error_count" 'BEGIN { if (total == 0) print 0; else print errors / total * 100 }')"
endpoint_summary="$(endpoint_summary_table "$traffic_log")"

qps="$(prom_scalar 'sum(rate(resource_community_http_requests_total[1m]))')"
p50="$(prom_scalar 'histogram_quantile(0.50, sum(rate(resource_community_http_request_duration_seconds_bucket[1m])) by (le))')"
p95="$(prom_scalar 'histogram_quantile(0.95, sum(rate(resource_community_http_request_duration_seconds_bucket[1m])) by (le))')"
non_2xx_error_rate="$(prom_scalar '(sum(rate(resource_community_http_requests_total{status!~"2.."}[1m])) / clamp_min(sum(rate(resource_community_http_requests_total[1m])), 0.001)) * 100')"
five_xx_error_rate="$(prom_scalar '(sum(rate(resource_community_http_requests_total{status=~"5.."}[1m])) / clamp_min(sum(rate(resource_community_http_requests_total[1m])), 0.001)) * 100')"

cat >"$REPORT_FILE" <<EOF_REPORT
# resource_community_go 可观测性演练报告

## 演练信息

- 开始时间：${start_iso}
- 结束时间：${end_iso}
- Backend：${BASE_URL}
- Prometheus：${PROMETHEUS_URL}
- 持续时间：${DURATION_SECONDS}s
- 并发数：${CONCURRENCY}
- 是否包含错误流量：${INCLUDE_ERROR_TRAFFIC}

## 本地请求统计

- 总请求数：${total_requests}
- 非 2xx/请求失败数：${local_error_count}
- 本地错误率：$(format_number "$local_error_rate" 2)%

## 接口维度本地统计

${endpoint_summary}

## Prometheus 指标快照

- QPS：$(format_number "$qps" 2)
- P50：$(format_number "$p50" 4)s
- P95：$(format_number "$p95" 4)s
- 非 2xx 错误率：$(format_number "$non_2xx_error_rate" 2)%
- 5xx 错误率：$(format_number "$five_xx_error_rate" 2)%

## 截图建议

在 Grafana 打开 ${BT}Resource Community / Resource Community API${BT}，时间范围选择 ${BT}Last 15 minutes${BT}，截取以下面板：

- QPS
- Non-2xx Error Rate
- P95 Latency
- Latency P50/P95
- QPS By Status
- P95 Latency By Route

## 排障案例记录

现象：

- 本次演练通过固定并发访问 ${BT}/healthz${BT}、${BT}/api/articles${BT}、${BT}/api/articles/hot${BT}、${BT}/api/articles/1${BT} 产生基础流量。
- 如果开启错误流量，会额外访问不存在路由，用于验证非 2xx 错误率面板。

判断：

- QPS 用于确认系统承压期间的请求吞吐。
- P50/P95 用于确认常规延迟和尾部延迟。
- 非 2xx 错误率用于确认客户端错误和服务端错误是否升高。
- 5xx 错误率用于确认服务端错误是否出现。

处理动作：

- 若 P95 升高，按 OnCallAgent 知识库文档 ${BT}resource-community-p95-latency.md${BT} 排查慢路由、Redis、MySQL、RabbitMQ 和后端日志。
- 若 5xx 升高，按 ${BT}resource-community-error-rate.md${BT} 定位错误路由和依赖异常。
- 若热榜不更新，按 ${BT}resource-community-hot-ranking.md${BT} 检查 Redis ZSet、缓存失效和 Worker 消费。
- 若异步更新延迟，按 ${BT}resource-community-rabbitmq-backlog.md${BT} 检查 RabbitMQ 队列和 Worker 日志。

## 可写入简历的数据表述模板

基于本地 Docker Compose 环境搭建 Prometheus + Grafana 可观测性平台，采集 Go 后端 HTTP 指标并完成压测演练，观测 QPS、P50/P95、非 2xx 错误率和 5xx 错误率；结合排障文档沉淀接口延迟、错误率升高、热榜不更新和 RabbitMQ 队列积压等场景的处理流程。
EOF_REPORT

echo "report written: ${REPORT_FILE}"
echo "qps=$(format_number "$qps" 2) p50=$(format_number "$p50" 4)s p95=$(format_number "$p95" 4)s non_2xx_error_rate=$(format_number "$non_2xx_error_rate" 2)% five_xx_error_rate=$(format_number "$five_xx_error_rate" 2)%"
