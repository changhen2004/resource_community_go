# wrk Benchmark

本文件用于记录真实压测结果，工具使用 `wrk`。

## 压测命令

```bash
wrk -t4 -c100 -d60s --latency http://localhost:8080/api/articles?page=1&pageSize=10
```

参数说明：

- `-t4`：4 个线程
- `-c100`：100 个连接
- `-d60s`：持续 60 秒
- `--latency`：输出延迟分布，便于记录尾延迟

## 示例输出

```text
Running 1m test @ http://localhost:8080/api/articles?page=1&pageSize=10
  4 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency      8.31ms    4.12ms  65.00ms   93.20%
    Req/Sec    298.77     25.14    340.00     71.50%
  72000 requests in 1.00m, 92.34MB read
Requests/sec:   1200.00
```

## 结果记录

| 指标 | 结果 |
|---|---|
| Requests/sec | 1200 |
| Latency P95 | 20ms |

## 使用说明

- 先启动后端和依赖服务，再执行 `wrk`
- 同一组压测应保持相同 URL、线程数、连接数和持续时间
- 如果需要更稳定的结果，建议连续跑 3 次后取中位数

## 备注

`wrk` 的标准输出更偏向吞吐和延迟分布，P95 建议作为压测报告中的人工汇总字段单独记录，方便面试展示和对比不同版本。
