# agentmesh
An open network that connects AI coding agents across platforms, enabling them to discover, communicate, and collaborate with each other through a unified protocol.

## Gin + MCP Server 最小示例

要求 Go 1.27.4 或以上，使用 Gin、官方 MCP Go SDK、Wire 依赖注入和 Viper TOML 配置。
启动入口为 `cmd/server/main.go`：

- `GET /ping`：普通 HTTP 接口，返回 `{"message":"pong"}`。
- `GET /hello?name=Alan`：调用业务层，返回 `{"message":"Hello, Alan!"}`。
- `/mcp`：MCP Streamable HTTP 接口，注册了 `hello` 工具，传入 `name` 返回问候语。

### 目录结构

```text
agentmesh/
├── cmd/
│   └── server/
│       ├── main.go          # 调用 Wire 初始化并启动服务
│       ├── server.go        # 服务实例与启动方法
│       ├── wire.go          # Wire 注入声明（wireinject）
│       └── wire_gen.go      # Wire 自动生成的组装代码
├── internal/
│   ├── handler/
│   │   ├── http.go          # Gin 路由与 HTTP 接口
│   │   ├── mcp.go           # MCP Server 与工具注册
│   │   └── provider.go      # handler.ProviderSet
│   ├── service/
│   │   ├── hello.go         # HTTP 和 MCP 共用的业务逻辑
│   │   └── provider.go      # service.ProviderSet
│   ├── repository/
│   │   └── doc.go           # 预留数据访问层
│   ├── model/
│   │   └── hello.go         # 输入、输出模型
│   └── middleware/
│       └── request_id.go    # 请求标识中间件
├── pkg/                    # 预留可复用的公共包
├── configs/
│   ├── config.go           # Viper 配置加载与结构体
│   └── config.toml         # TOML 配置文件
├── Makefile                # 生成、检查、构建和启动命令
├── go.mod
└── go.sum
```

HTTP 和 MCP 接口都调用 `HelloService.Hello`；业务层不依赖 Gin 或 MCP。
当前示例不涉及数据存储，`repository` 和 `pkg` 作为扩展目录保留。

### 启动

```bash
make init
make run
```

`wire_gen.go` 和 `bin/` 构建产物已加入 `.gitignore`。
`make build`、`make run` 和 `make lint` 会自动先生成 Wire 代码。
直接使用 `go run ./cmd/server` 前，请先执行 `make wire`。

服务监听 `127.0.0.1:8080`。另开终端调用：

```bash
curl http://127.0.0.1:8080/ping
curl 'http://127.0.0.1:8080/hello?name=Alan'
```

### TOML 配置

默认从当前工作目录读取 `configs/config.toml`：

```toml
[server]
addr = "127.0.0.1:8080"
mode = "release"
```

在项目根目录启动服务，或通过 `-config` 参数指定配置路径：

```bash
make wire
go run ./cmd/server -config /path/to/config.toml
# 编译后运行同样支持指定路径
./bin/agentmesh -config /path/to/config.toml
```

修改 TOML 中的 `server.addr` 即可调整监听地址，例如：

```toml
[server]
addr = "127.0.0.1:8081"
mode = "release"
```

修改端口后，客户端的 MCP URL 也需使用相应端口。
`server.addr` 和 `server.mode` 都必须在 TOML 中显式配置；配置文件缺失、格式错误或必填项缺失时，启动会报错退出。
`server.mode` 可选 `debug`、`release`、`test`。
本地调试时可以设置为 `debug`；`release` 会关闭 Gin 的调试提示，请求访问日志仍然保留。
配置通过 `Config.Validate()` 统一校验，规则写在字段的 `validate` 标签中；
`notblank` 校验非空白地址，`oneof` 限定运行模式。

### Wire 与构建命令

参考 `roomplay` 的组织方式，各层通过 `provider.go` 导出 `ProviderSet`，
在 `cmd/server/wire.go` 中由 `wireServer()` 汇总，生成实际的构造函数调用。
`configs.Load` 通过 Viper 读取 TOML 并返回配置或错误，`NewHelloService → NewHandler → NewRouter` 依次组装依赖，
最后由 `newServer` 接收配置和路由，构造服务实例。
配置加载错误由 Wire 生成的初始化函数传递到 `main`。

```bash
make wire       # 修改构造函数或 ProviderSet 后重新生成 wire_gen.go
make generate   # 生成项目代码（当前为 Wire）
make fmt        # 格式化代码
make lint       # go vet 检查
make build      # 构建 bin/agentmesh
make all        # 生成代码并构建
make help       # 查看所有命令
```

`make build` 默认使用 `sonic avx` 构建标签，启用 Gin 的 Sonic JSON 实现，实际命令为：

```bash
go build -tags="sonic avx" -o bin/agentmesh ./cmd/server
```

Wire v0.7.0 通过 `go.mod` 的 `tool` 声明管理，`make wire` 使用 `go tool wire`，
无需全局安装。请修改 `wire.go` 和各层 `provider.go`，通过命令生成 `wire_gen.go`，不要手动编辑生成文件。

### 调用 MCP

使用无状态模式并返回 JSON，方便通过 curl 演示，无需保存会话 ID。
以下命令按顺序执行。

1. 初始化 MCP 连接：

```bash
curl http://127.0.0.1:8080/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"curl-demo","version":"1.0.0"}}}'
```

2. 发送初始化完成通知（返回 HTTP 202，无响应正文）：

```bash
curl -i http://127.0.0.1:8080/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'MCP-Protocol-Version: 2025-11-25' \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}'
```

3. 查看工具列表：

```bash
curl http://127.0.0.1:8080/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'MCP-Protocol-Version: 2025-11-25' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
```

4. 调用 `hello`：

```bash
curl http://127.0.0.1:8080/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'MCP-Protocol-Version: 2025-11-25' \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"hello","arguments":{"name":"Alan"}}}'
```

响应的 `result.structuredContent` 为 `{"message":"Hello, Alan!"}`。
支持 Streamable HTTP 的 MCP 客户端可以直接连接 `http://127.0.0.1:8080/mcp`。
