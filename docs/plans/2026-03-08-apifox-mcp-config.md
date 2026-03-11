# Apifox MCP 配置 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为仓库补齐基于 Apifox 官方文档的 MCP 配置模板，使本项目的 OpenAPI 文档可被支持 MCP 的 IDE 直接接入。

**Architecture:** 复用现有 `backend/docs/openapi.yaml` 作为 OpenAPI 数据源，新增跨平台 Apifox MCP 配置模板，并在仓库文档中说明 Windows 与 macOS/Linux 的使用方式、替换项和验证步骤。避免修改用户本机全局 IDE 配置，只在仓库内提供可复用模板与说明。

**Tech Stack:** JSON 配置模板、Markdown 文档、Apifox MCP Server、OpenAPI 3.0。

---

### Task 1: 明确 OpenAPI 数据源与配置落点

**Files:**
- Modify: `backend/README.md`
- Create: `backend/docs/mcp/apifox-cline.windows.json`
- Create: `backend/docs/mcp/apifox-cline.unix.json`

**Step 1: 确认 OpenAPI 文档路径**

检查 `backend/docs/openapi.yaml` 是否为当前后端接口文档，并确认仓库内已有说明可以支撑该路径作为 `--oas` 参数输入。

**Step 2: 设计 Windows 配置模板**

按 Apifox 官方文档的 Windows 方案，使用：

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
        "--oas=<oas-url-or-path>"
      ]
    }
  }
}
```

并将 `<oas-url-or-path>` 替换为当前仓库 `backend/docs/openapi.yaml` 的说明占位符。

**Step 3: 设计 macOS / Linux 配置模板**

按 Apifox 官方文档的 Unix 方案，使用：

```json
{
  "mcpServers": {
    "go-rag API 文档": {
      "command": "npx",
      "args": [
        "-y",
        "apifox-mcp-server@latest",
        "--oas=<oas-url-or-path>"
      ]
    }
  }
}
```

**Step 4: 保持模板可复用**

模板文件中保留占位符，不将当前开发机绝对路径直接写入版本库；在文档中提供本机示例路径与替换说明。

### Task 2: 补充仓库使用说明

**Files:**
- Modify: `backend/README.md`

**Step 1: 新增 MCP 配置章节**

说明该配置基于 Apifox 官方文档，当前项目推荐使用 `backend/docs/openapi.yaml` 作为 `--oas` 数据源。

**Step 2: 给出 Windows 与 Unix 使用方式**

引用新增模板文件路径，并说明 Windows 下优先使用 `cmd /c npx ...` 方案。

**Step 3: 给出验证方法**

加入和官方文档一致的验证思路，例如让 AI 通过 MCP 读取 API 文档并统计接口数。

### Task 3: 验证变更自洽

**Files:**
- Test: `backend/docs/mcp/apifox-cline.windows.json`
- Test: `backend/docs/mcp/apifox-cline.unix.json`
- Test: `backend/README.md`

**Step 1: 校验 JSON 结构**

确保两个模板文件都是合法 JSON，且字段名与官方文档一致：`mcpServers`、`command`、`args`。

**Step 2: 校验文档引用路径**

确认 README 中引用的 `backend/docs/openapi.yaml` 与模板文件路径真实存在。

**Step 3: 校验与官方文档一致性**

核对 Windows / Unix 启动命令是否与 Apifox 官方文档保持一致，仅替换 `--oas` 参数值与服务名称。
