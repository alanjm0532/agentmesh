# agentmesh
An open network that connects AI coding agents across platforms, enabling them to discover, communicate, and collaborate with each other through a unified protocol.

## Gin + MCP Server 最小示例

要求 Go 1.27.4 或以上，使用 Gin 和官方 MCP Go SDK。
启动入口为 `cmd/server/main.go`：

- `GET /ping`：普通 HTTP 接口，返回 `{"message":"pong"}`。
- `GET /hello?name=Alan`：调用业务层，返回 `{"message":"Hello, Alan!"}`。
- `/mcp`：MCP Streamable HTTP 接口，注册了 `hello` 工具，传入 `name` 返回问候语。

### 目录结构

```text
agentmesh/
├── cmd/
│   └── server/
│       └── main.go          # 读取配置、组装依赖、启动服务
├── internal/
│   ├── handler/
│   │   ├── http.go          # Gin 路由与 HTTP 接口
│   │   └── mcp.go           # MCP Server 与工具注册
│   ├── service/
│   │   └── hello.go         # HTTP 和 MCP 共用的业务逻辑
│   ├── repository/
│   │   └── doc.go           # 预留数据访问层
│   ├── model/
│   │   └── hello.go         # 输入、输出模型
│   └── middleware/
│       └── request_id.go    # 请求标识中间件
├── pkg/                    # 预留可复用的公共包
├── configs/
│   └── config.go           # 环境变量配置
├── go.mod
└── go.sum
```

HTTP 和 MCP 接口都调用 `HelloService.Hello`；业务层不依赖 Gin 或 MCP。
当前示例不涉及数据存储，`repository` 和 `pkg` 作为扩展目录保留。

### 启动

```bash
go mod tidy
go run ./cmd/server
```

服务监听 `127.0.0.1:8080`。另开终端调用：

```bash
curl http://127.0.0.1:8080/ping
curl 'http://127.0.0.1:8080/hello?name=Alan'
```

可通过环境变量修改监听地址，例如：

```bash
SERVER_ADDR=127.0.0.1:8081 go run ./cmd/server
```

修改端口后，客户端的 MCP URL 也需使用相应端口。

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
