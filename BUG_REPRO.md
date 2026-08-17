# 修复前故障复现（Docker）

## 项目与标准命令

项目为 Buoy Calibration Service，Go module 为 `buoy-calibration`。在当前 arm64 平台使用 `golang:1.24` 镜像执行：

```bash
go build ./...
go test ./...
```

## 环境构建与编译

使用项目的 `benzhi.Dockerfile` 构建镜像成功，容器内 `go version` 输出为 `go version go1.24.13 linux/arm64`，`go build ./...` 成功。

## 故障触发步骤

在容器内执行：

```bash
go test ./...
```

## 实际错误输出

```text
?   	buoy-calibration/cmd/server	[no test files]
--- FAIL: TestCreateProfileRejectsNegativeScale (0.00s)
    server_test.go:43: status = 201, want 400
FAIL
FAIL	buoy-calibration/internal/api	0.002s
?   	buoy-calibration/internal/calibration	[no test files]
?   	buoy-calibration/internal/model	[no test files]
ok  	buoy-calibration/internal/service	0.011s
?   	buoy-calibration/internal/store	[no test files]
FAIL
```

## 期望行为

创建标定档案时应立即拒绝负数比例，不应允许异常配置进入后续现场运行。
