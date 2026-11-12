# agentmesh

An open network that connects AI coding agents across platforms, enabling them to discover, communicate, and collaborate with each other through a unified protocol.

## 项目结构

项目使用根目录的单个 Go module，要求 Go 1.27.4 或以上。

```text
agentmesh/
├── cmd/
│   ├── server/
│   │   └── main.go             # HTTP 服务入口
│   └── agentmesh/
│       └── main.go             # CLI 入口
├── internal/
│   ├── server/
│   │   ├── server.go           # 服务初始化与启动
│   │   ├── logger.go           # Zap 日志初始化
│   │   ├── wire.go             # Wire 依赖注入声明
│   │   ├── config/             # 服务端 Viper 配置加载
│   │   ├── handler/            # Gin 路由与 HTTP 接口
│   │   ├── service/            # 业务逻辑
│   │   ├── repository/         # 数据访问层
│   │   └── middleware/         # 请求标识中间件
│   └── cli/
│       ├── command/            # Cobra 命令
│       ├── client/             # Resty HTTP 请求和 encoding/json 解析
│       └── config/             # CLI Viper 配置加载
├── pkg/
│   └── protocol/               # AgentMesh Protocol，两端共用
│       ├── ping.go             # PingResponse
│       └── hello.go            # HelloRequest、HelloResponse
├── migrations/                 # 数据库迁移预留目录
├── docs/                       # 文档预留目录
├── deployments/
│   ├── server.toml             # 服务端默认配置
│   └── cli.toml                # CLI 默认配置
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

`pkg/protocol` 只定义公共请求、响应结构体和 JSON 字段，不依赖 Gin 或 CLI 库。
Server 和 CLI 都直接引用 `github.com/alanjm0532/agentmesh/pkg/protocol`。

| 功能 | Go 库 |
| --- | --- |
| HTTP 服务 | Gin |
| CLI 命令 | Cobra |
| 配置管理 | Viper |
| HTTP 请求 | `github.com/go-resty/resty/v2` |
| JSON | `encoding/json` |

## 快速开始

在项目根目录启动 HTTP 服务：

```bash
make init
make run
```

服务默认监听 `127.0.0.1:8080`。另开终端运行 CLI：

```bash
go run ./cmd/agentmesh ping
# 或使用 Makefile
make run-cli ARGS="ping"
```

成功时输出 `pong`，退出码为 0。连接失败、超时、HTTP 状态异常或响应不符合预期时，
错误写入标准错误，退出码为 1。

## CLI 配置与参数

```bash
go run ./cmd/agentmesh --help
go run ./cmd/agentmesh ping --help
go run ./cmd/agentmesh ping --server http://127.0.0.1:8081 --timeout 3s
go run ./cmd/agentmesh --config /path/to/cli.toml ping
```

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `--config` | 可选 | 指定 CLI TOML 配置文件路径 |
| `--server` | `http://127.0.0.1:8080` | HTTP 服务地址，包含 `http://` 或 `https://` |
| `--timeout` | `5s` | 请求超时时间，必须大于 0 |

Viper 默认读取项目根目录下的 `deployments/cli.toml`：

```toml
server = "http://127.0.0.1:8080"
timeout = "5s"
```

默认配置文件可省略；使用 `--config` 显式指定的文件必须能够读取。
配置优先级为：命令行参数 > 环境变量 > TOML 配置 > 默认值。
支持 `AGENTMESH_SERVER` 和 `AGENTMESH_TIMEOUT` 环境变量：

```bash
AGENTMESH_SERVER=http://127.0.0.1:8081 AGENTMESH_TIMEOUT=3s go run ./cmd/agentmesh ping
```

CLI 的构建和运行不依赖服务端配置或 Wire 生成。

## HTTP 接口与服务端配置

服务保留以下接口：

- `GET /ping`：返回 `{"message":"pong"}`，CLI 的 `ping` 命令调用此接口。
- `GET /hello?name=Alan`：返回 `{"message":"Hello, Alan!"}`。

```bash
curl http://127.0.0.1:8080/ping
curl 'http://127.0.0.1:8080/hello?name=Alan'
```

服务端默认从当前工作目录读取 `deployments/server.toml`：

```toml
[server]
addr = "127.0.0.1:8080"
mode = "debug"
```

`server.mode` 支持 `debug`、`release` 和 `test`，配置由 `Config.Validate()` 校验。
修改监听地址后，通过 CLI 的 `--server` 参数或配置指定相应地址。

通过 `-config` 指定服务端配置文件：

```bash
make run ARGS="-config /path/to/server.toml"
# 构建后的服务端
./bin/agentmesh-server -config /path/to/server.toml
```

服务端日志使用 Zap，默认级别为 `info`，输出到 stdout。
`release` 模式使用 JSON 编码，`debug` 和 `test` 模式使用 console 编码。
Gin 请求和异常恢复日志使用 `gin-contrib/zap`，Wire 返回的清理函数负责刷新日志。

## 构建与开发命令

```bash
make build-cli    # 构建 CLI：bin/agentmesh
make build-server # 生成 Wire 代码并构建 HTTP 服务：bin/agentmesh-server
make build        # 构建 CLI 和 HTTP 服务
make run          # 生成 Wire 代码并启动 HTTP 服务
make run-cli ARGS="ping --timeout 3s"
make wire         # 生成服务端 wire_gen.go
make generate     # 生成项目代码（当前为 Wire）
make fmt          # 格式化代码
make lint         # 生成 Wire 代码后执行 go vet
make help         # 查看所有命令
```

构建后使用 CLI：

```bash
./bin/agentmesh ping
./bin/agentmesh ping --server http://127.0.0.1:8081 --timeout 3s
```

Wire 通过 `go.mod` 的 `tool` 声明管理，无需全局安装。
修改服务端依赖后，执行 `make wire` 生成 `internal/server/wire_gen.go`。
直接运行 `go run ./cmd/server` 前也需要先生成 Wire 代码。
`wire_gen.go` 和 `bin/` 构建产物已加入 `.gitignore`，不要手动编辑生成文件。
