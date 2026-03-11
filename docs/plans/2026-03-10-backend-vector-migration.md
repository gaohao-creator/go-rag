# Backend Vector Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在 `backend` 中落地“ES 优先 + DB 降级”的向量索引与向量检索链路，保持现有 HTTP 接口和 DTO 语义不变。

**Architecture:** 在 `backend` 内新增独立 `vector` 模块承载 ES 客户端、索引名解析、索引写入和检索适配；`ioc` 只负责根据配置决定是否启用该模块；`service` 层只依赖抽象接口，索引阶段先写 DB 再尝试写 ES，检索阶段优先 ES，失败或未启用时降级到现有 DB 检索引擎。

**Tech Stack:** Go 1.24、Gin、Gorm、Elasticsearch v8 typed API、`cloudwego/eino` embedding/indexer/retriever 组件、`httptest`。

---

### Task 1: 扩展配置模型并接入向量 IoC

**Files:**
- Modify: `backend/config/config.go`
- Modify: `backend/config/config.yaml`
- Modify: `backend/ioc/config.go`
- Modify: `backend/ioc/app.go`
- Test: `backend/config/config_test.go`
- Test: `backend/ioc/app_test.go`

**Step 1: Write the failing test**

为配置加载增加向量场景断言：
- `Config.ApplyDefaults` 会为 `vector.enabled/backend/index_prefix/content_field/content_vector_field/knowledge_field/ext_field/dimensions/embedding_model` 注入默认值
- `ApplyEnvLookup` 能覆盖 `GO_RAG_VECTOR_*` 环境变量
- `NewApp` 在 `vector.enabled=false` 时保持可启动，不要求 ES 配置

**Step 2: Run test to verify it fails**

Run: `go test ./config ./ioc -run 'TestConfig|TestNewApp' -count=1`
Expected: FAIL because vector config fields / assertions do not exist yet

**Step 3: Write minimal implementation**

实现约束：
- 新增 `VectorConfig`
- 默认关闭向量能力
- 默认 ES 字段名兼容旧 `server/core/types/consts.go`
- embedding 相关配置支持沿用 chat 的 `api_key/base_url` 作为兜底

**Step 4: Run test to verify it passes**

Run: `go test ./config ./ioc -run 'TestConfig|TestNewApp' -count=1`
Expected: PASS

### Task 2: 新增向量模块基础设施

**Files:**
- Create: `backend/internal/vector/contracts.go`
- Create: `backend/internal/vector/index_name.go`
- Create: `backend/internal/vector/index_name_test.go`
- Create: `backend/internal/vector/es_client.go`
- Create: `backend/internal/vector/es_mapping.go`
- Create: `backend/internal/vector/es_mapping_test.go`
- Create: `backend/internal/vector/embedder.go`

**Step 1: Write the failing test**

覆盖两个核心约束：
- 索引名解析对知识库名做稳定、可预测的清洗，保证不同知识库不会冲突
- ES mapping 与旧字段兼容，至少包含 `content/content_vector/ext/_knowledge_name`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/vector -run 'TestResolveIndexName|TestBuildIndexMappings' -count=1`
Expected: FAIL because vector module does not exist yet

**Step 3: Write minimal implementation**

实现约束：
- 只实现 ES 路径，不实现 Qdrant
- 索引名格式：`<prefix><sanitized-knowledge-name>`
- `sanitized-knowledge-name` 仅保留小写字母、数字、连字符
- mapping 维度来自配置，默认 1024

**Step 4: Run test to verify it passes**

Run: `go test ./internal/vector -run 'TestResolveIndexName|TestBuildIndexMappings' -count=1`
Expected: PASS

### Task 3: 落地索引写入链路

**Files:**
- Create: `backend/internal/vector/indexer.go`
- Create: `backend/internal/vector/indexer_test.go`
- Modify: `backend/internal/domain/service/indexer.go`
- Modify: `backend/internal/service/indexer.go`
- Modify: `backend/internal/service/indexer_test.go`
- Modify: `backend/ioc/app.go`

**Step 1: Write the failing test**

新增服务层回归：
- `IndexerService` 在 DB chunk 落库成功后调用向量写入器
- 向量写入失败时不回滚 DB 成果，只返回成功并记录该错误为非阻断行为
- 向量关闭时不调用写入器

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run TestIndexerService -count=1`
Expected: FAIL because service has no vector writer hook

**Step 3: Write minimal implementation**

实现约束：
- `VectorIndexer` 按知识库维度自动建索引
- 将 chunk 内容、ext、knowledge_name 按旧字段写入 ES
- 先写 DB，后写 ES
- ES 失败不影响主链返回，但需要让服务可观测到错误

**Step 4: Run test to verify it passes**

Run: `go test ./internal/vector ./internal/service -run 'TestVectorIndexer|TestIndexerService' -count=1`
Expected: PASS

### Task 4: 落地检索主链与 DB 降级

**Files:**
- Create: `backend/internal/vector/retriever.go`
- Create: `backend/internal/vector/retriever_test.go`
- Create: `backend/internal/service/retriever_fallback.go`
- Create: `backend/internal/service/retriever_fallback_test.go`
- Modify: `backend/ioc/app.go`

**Step 1: Write the failing test**

覆盖以下行为：
- 已启用向量检索时，优先返回 ES 结果
- ES 检索报错时自动降级到 DB 检索
- 若 ES 返回空结果，也回退 DB，保证当前行为不退化

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service ./internal/vector -run 'TestFallbackRetriever|TestVectorRetriever' -count=1`
Expected: FAIL because fallback retriever and vector retriever do not exist

**Step 3: Write minimal implementation**

实现约束：
- 只暴露 `domainservice.Retriever` 抽象给业务层
- 结果解析保持 `chunk_id/content/ext/score` 语义
- 继续遵守现有 `top_k/score` 默认值逻辑

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service ./internal/vector -run 'TestFallbackRetriever|TestVectorRetriever|TestRetrieverService' -count=1`
Expected: PASS

### Task 5: 补齐集成验证与文档

**Files:**
- Modify: `backend/README.md`
- Modify: `backend/.env.example`
- Modify: `backend/docs/openapi.yaml`
- Modify: `backend/internal/web/router/integration_test.go`

**Step 1: Write the failing test**

补一条 IoC 集成测试，验证：
- 向量开关关闭时应用仍正常工作
- 检索接口在默认配置下依旧走 DB 路径且行为不变

**Step 2: Run test to verify it fails**

Run: `go test ./internal/web/router ./ioc -run 'TestIntegration|TestNewApp' -count=1`
Expected: FAIL if docs/config references are incomplete or route behavior regresses

**Step 3: Write minimal documentation updates**

文档更新内容：
- README 增加向量配置说明与降级语义
- `.env.example` 增加 `GO_RAG_VECTOR_*` 样例
- OpenAPI 增加检索实现说明，不改接口 schema

**Step 4: Run full verification**

Run: `go test ./... -count=1`
Expected: PASS
