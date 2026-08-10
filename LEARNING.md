# 学习历程

## 1. Go 程序

最开始的 `endpoint-probe` 功能：

1. 从 JSON 文件读取目标名称、URL 和超时时间。
2. 并发发送 HTTP `GET` 请求。
3. 输出状态码、单次耗时、总耗时。
4. 请求失败或返回非 `2xx` 时，以退出码 `1` 结束。

主要练习了结构体、JSON、错误处理、HTTP 客户端、goroutine、channel 和单元测试。

## 2. Docker

先运行了现成的 Nginx 镜像，然后为两个 Go 程序分别写了 Dockerfile。

明白了：

- 镜像是程序和运行文件的 *Read-Only Template*。
- 每次 `docker run` 都会创建一个新容器。
- 容器中的程序仍然是 Linux 进程，不是完整虚拟机。
- `18080:80` 表示宿主机端口映射到容器端口。
- 当前工作目录不会限制 Docker 能看到哪些容器；它们都由同一个 Docker daemon 管理。
- bind mount 可以在不重建镜像的情况下替换配置文件。

两个容器有各自独立的网络空间。在 `endpoint-probe` 中，`localhost` 指向容器自己，而 `predictor` 在另一个容器里。把两者加入同一个网络后，Docker 会把名称 `predictor` 解析到对应容器，所以探测地址应写成 `http://predictor:8080`，不能写成 `http://localhost:8080`。

两个镜像目前统一使用 `:local` 标签。

## 3. kind 集群

我使用 kind 创建了一个控制平面和两个 Worker。

这些 Node 本身是 Docker 容器，内部又使用 containerd 运行 Pod。因此宿主机 Docker 中的镜像不会自动出现在 Node 中，需要使用 `kind load docker-image` 导入。

由于只有一台 Ubuntu-24.04，所以只能学习 K8s 的控制流程，真实多机性能和故障环境期待后续学习。

## 4. 部署服务

`predictor` 使用 Deployment 运行两个副本，Service 为它们提供固定的集群内地址。

我删除过一个 predictor Pod。旧 Pod 没有复活，而是 ReplicaSet 发现副本数不足，创建了一个名称不同的新 Pod，Scheduler 再为它选择 Node。

为了观察滚动更新，我曾修改 `APP_VERSION` 的显示值。由于这个环境变量属于 Pod 模板的一部分，修改它后 Deployment 创建了新的 ReplicaSet，等待新 Pod 就绪，再把旧 ReplicaSet 缩到零。这就是滚动更新的流程。

## 5. 定时探测

`endpoint-probe` 作为 CronJob 每分钟运行一次：

```text
CronJob -> Job -> Pod -> endpoint-probe
```

探测地址保存在 ConfigMap 中，并挂载到容器里的 `/app/configs/targets.json`。修改 ConfigMap 后，新 Job 可以直接使用新配置，而无需重新制作镜像。

## 6. 插曲二则

### 端口被旧容器占用

首次 `kubectl port-forward` 时，`18081` 已被旧容器占用，所以实际上端口转发没有启动。

之后 `curl` 仍然收到响应，但显然请求命中的是旧容器，不是 K8s Service。旧容器没有 `/predict`，因此返回了 `404`。

因此收到响应不代表请求到达了真正的目标，Debug 时要先确认端口由谁监听。

### 日志报错，但 Job 显示成功

服务缩容到零后，探测日志已经显示连接失败，但 Job 状态仍然是 `Complete`。

原因是 K8s 不会阅读日志来判断任务是否成功，只看容器主进程的退出码。原来的 Go 程序打印错误后正常结束，退出码仍然是 `0`。

因此主逻辑需要修改返回退出码：全部目标为 `2xx` 时返回 `0`，否则返回 `1`。修改后，失败任务重试一次并正确标记为 `Failed`，服务恢复后的任务则标记为 `Complete`。

## 7. 学到了什么

- Dockerfile、镜像、容器和进程的关系。
- Deployment 为什么适合长期服务，CronJob 为什么适合定时探测。
- Service 为什么能让不断变化的 Pod 使用固定名称。
- ConfigMap 如何把配置和镜像分开。
- readiness、liveness、资源请求和资源上限各自解决什么问题。
- Controller、Scheduler、kubelet 和容器运行时在 Pod 创建过程中分别做什么。
- 为什么应用退出码会影响 Job 的成功或失败状态。

## 8. 没做到什么

- 真实多机集群的网络和故障。
- GPU device plugin 和 GPU 资源申请。
- MPS、GPU共享和推理服务调度。
- Prometheus 指标、长期数据保存和告警。
- Helm、Operator 和自定义 Scheduler。
