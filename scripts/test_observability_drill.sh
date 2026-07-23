#!/usr/bin/env bash
set -euo pipefail

script_path="$(dirname "$0")/observability_drill.sh"

require_pattern() {
  local pattern="$1"
  local description="$2"
  if ! grep -Fq -- "$pattern" "$script_path"; then
    echo "missing: ${description}" >&2
    exit 1
  fi
}

require_pattern "--connect-timeout" "curl connect timeout for bounded client-side failures"
require_pattern "--max-time" "curl total timeout for bounded client-side failures"
require_pattern "客户端请求失败数" "separate client failure count in report"
require_pattern "HTTP 非 2xx 数" "separate HTTP non-2xx count in report"
require_pattern "服务端口径以 Prometheus 指标为准" "report explains server-side metric scope"
require_pattern '$2 == "000"' "client failures are counted separately from HTTP statuses"
