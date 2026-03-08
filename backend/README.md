# go-rag backend

## 阶段一范围

当前 `backend` 已完成第一阶段主链：

- 知识库管理
- 文档管理
- chunk 管理
- 文档索引
- 基于已入库 chunk 的检索

当前未包含：

- chat / chat stream
- MCP
- 静态页面托管
- 向量数据库正式写入与向量检索

## 目录说明

- `cmd/server`：程序入口
- `config`：配置结构与示例配置
- `ioc`：依赖注入和应用装配
- `internal/service`：应用服务
- `internal/repository`：业务对象到 DAO 的存储封装
- `internal/repository/dao`：直接面向数据库的持久化访问
- `internal/web`：Gin 路由、Handler、中间件
- `scripts/mysql/init.sql`：MySQL 初始化脚本
- `docker-compose.yml`：本地开发 MySQL 容器编排

## 启动方式

在 `backend` 目录执行测试：

```bash
go test ./... -count=1
```

使用默认配置启动：

```bash
go run ./cmd/server
```

指定配置文件启动：

```bash
set GO_RAG_CONFIG=你的配置文件路径
go run ./cmd/server
```

## Docker Compose

当前提供了一个最小开发用 `docker-compose.yml`，只启动 MySQL。

在 `backend` 目录执行：

```bash
docker compose up -d
```

停止并清理：

```bash
docker compose down
```

查看配置是否正确展开：

```bash
docker compose config
```

默认数据库连接与 `config/config.yaml` 对齐：

- Host: `127.0.0.1`
- Port: `3306`
- User: `root`
- Password: `abc123`
- Database: `go_rag`

## 配置说明

默认读取 `config/config.yaml`。

测试默认使用 SQLite 内存库。
开发/部署可使用 MySQL，并配合执行：

```bash
mysql -u root -pabc123 < scripts/mysql/init.sql
```

`init.sql` 现在会先执行：

- `CREATE DATABASE IF NOT EXISTS go_rag`
- `USE go_rag`

因此既可以直接手工执行，也可以挂载到 MySQL 容器初始化目录。

## 当前接口

- `POST /api/v1/kb`
- `GET /api/v1/kb`
- `GET /api/v1/kb/:id`
- `PUT /api/v1/kb/:id`
- `DELETE /api/v1/kb/:id`
- `GET /api/v1/documents`
- `DELETE /api/v1/documents`
- `GET /api/v1/chunks`
- `DELETE /api/v1/chunks`
- `PUT /api/v1/chunks`
- `PUT /api/v1/chunks-content`
- `POST /api/v1/indexer`
- `POST /api/v1/retriever`

## 返回结构

统一返回结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

其中：

- `0` 表示成功
- `40001` 表示请求参数错误
- `50001` 表示服务内部错误
- `50003` 表示服务依赖未装配

## OpenAPI

- docs/openapi.yaml：根据当前 HTTP 路由生成的 OpenAPI 3.0 文档。

## MCP 配置（Apifox）

当前仓库已补充基于 Apifox 官方文档的 MCP 配置模板，
用于让支持 MCP 的 IDE 直接读取本项目的 OpenAPI 文档。

推荐把 `backend/docs/openapi.yaml` 作为 `--oas` 的输入。

模板文件：

- `backend/docs/mcp/apifox-cline.windows.json`
- `backend/docs/mcp/apifox-cline.unix.json`

如果你当前工作目录就是本仓库，Windows 下可将
`<oas-url-or-path>` 替换为：

```text
D:\project\a-ai-interview\go-rag\backend\docs\openapi.yaml
```

macOS / Linux 则替换为对应本机绝对路径，例如：

```text
/path/to/go-rag/backend/docs/openapi.yaml
```

Windows 推荐配置如下（与 Apifox 官方文档一致，使用 `cmd /c`）：

```json
{
  "mcpServers": {
    "go-rag API 文档": {
      "command": "cmd",
      "args": [
        "/c",
        "npx",
        "-y",
        "apifox-mcp-server@latest",
        "--oas=D:\\project\\a-ai-interview\\go-rag\\backend\\docs\\openapi.yaml"
      ]
    }
  }
}
```

macOS / Linux 推荐配置如下：

```json
{
  "mcpServers": {
    "go-rag API 文档": {
      "command": "npx",
      "args": [
        "-y",
        "apifox-mcp-server@latest",
        "--oas=/path/to/go-rag/backend/docs/openapi.yaml"
      ]
    }
  }
}
```

验证方式：

1. 启动或重载 IDE 中的 MCP 配置。
2. 让 AI 通过 MCP 读取 API 文档。
3. 示例提问：`请通过 MCP 获取 API 文档，并告诉我项目中有几个接口`。

注意事项：

- 首次运行会通过 `npx` 拉取 `apifox-mcp-server@latest`，本机需可访问 npm。
- 如果仓库路径变化，需要同步更新 `--oas` 指向的绝对路径。
- 当前配置使用本地 `openapi.yaml`，不依赖后端服务先启动。

