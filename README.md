# Endpoint Probe

这是一个学习项目，用来练习 Go、Docker 和 Kubernetes。

不是为了解决复杂问题，核心功能：读取一组 HTTP/HTTPS 地址，并发发送 `GET` 请求，然后输出状态码和耗时。

```text
# 例如：
GitHub: status = 200 duration = 235ms
Go: status = 200 duration = 536ms
total duration: 536ms
```

只要有一个地址连接失败、超时或返回非 `2xx` 状态，程序就以失败状态退出。

仓库里还有一个很小的 `predictor` 服务，提供健康检查、版本和简单计算接口。它只是为了给探测程序准备一个可控目标，不是项目重点，也不是真正的 AI 推理服务。

需要说明的， `/version` 返回的并不是版本号，只是一个便于观察滚动更新的标记，目前统一为 `local`。

## K8s 帮它做了什么

若着眼功能实现，其实完全没必要使用 K8s，Systemd Timers 可以直接做到。

这里接入 K8s，把基础流程实际走一遍的同时，其实也补充了下述功能：

- 每分钟自动运行一次探测任务。
- 不重新制作镜像就能修改探测地址。
- 保持两个演示服务副本，其中一个被删除后自动补建。
- 给两个服务副本提供一个固定的访问名称。
- 根据程序退出码，把每次探测标记为成功或失败。
- 修改服务配置时逐步替换旧 Pod。

这些功能对于当前项目有些大材小用，但有助于我实际理解 Deployment、Service、ConfigMap、CronJob、健康检查和资源限制分别在做什么。

详细的过程和目前的理解记录在 [学习历程](LEARNING.md)。

## 本地运行

修改 `configs/targets.json`：

```json
[
    {
        "name": "Example",
        "url": "https://example.com",
        "timeout": 3
    }
]
```

然后运行：

```bash
go run .
```

运行检查：

```bash
go fmt ./...
go vet ./...
go test ./...
```

## Docker 运行

```bash
docker build -t endpoint-probe:local .
```

使用本地配置运行：

```bash
docker run --rm \
  --mount type=bind,source="$(pwd)/configs/targets.json",target=/app/configs/targets.json,readonly \
  endpoint-probe:local
```

## K8s 运行

使用 kind 在本地创建一个控制平面和两个 Worker：

```bash
docker build -f Dockerfile.predictor -t predictor:local .
docker build -t endpoint-probe:local .

kind create cluster \
  --name endpoint-lab \
  --config deploy/kind-config.yaml \
  --wait 120s

kind load docker-image predictor:local --name endpoint-lab
kind load docker-image endpoint-probe:local --name endpoint-lab

kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/predictor.yaml
kubectl apply -f deploy/k8s/probe-config.yaml
kubectl apply -f deploy/k8s/probe-cronjob.yaml
```

`:local` 只表示本机制作的镜像，无版本含义。修改代码后继续使用这个标签时，需要重新执行 `docker build` 和 `kind load docker-image`；如果 `predictor` 已经在运行，再执行：

```bash
kubectl rollout restart deployment/predictor -n endpoint-lab
```

查看定时任务和最近的执行结果：

```bash
kubectl get cronjobs,jobs -n endpoint-lab
kubectl logs job/<Job名称> -n endpoint-lab
```

手动运行一次：

```bash
JOB_NAME="endpoint-probe-manual-$(date +%s)"
kubectl create job --from=cronjob/endpoint-probe "$JOB_NAME" -n endpoint-lab
kubectl wait --for=condition=complete "job/$JOB_NAME" -n endpoint-lab --timeout=60s
kubectl logs "job/$JOB_NAME" -n endpoint-lab
```

删除集群：

```bash
kind delete cluster --name endpoint-lab
```

## 当前限制

- 只支持 HTTP/HTTPS `GET` 请求。
- 只判断状态码，不检查响应内容是否正确。
- 结果只写到日志，没有保存到数据库或监控系统。
- 没有涉及真实多机、GPU、推理框架或自定义调度器。
