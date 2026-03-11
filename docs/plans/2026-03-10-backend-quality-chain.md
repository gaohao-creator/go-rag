# Backend Quality Chain Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在 `backend` 中补齐检索质量链能力，将旧 `server` 的 `QA + rerank + grader` 核心能力迁移到新的索引、检索与对话主链中。

**Architecture:** `indexer` 负责在切块后为内容生成 QA 语料，并在落库与向量写入时同时携带内容向量和 QA 向量；`retriever` 负责并行执行内容检索与 QA 检索、合并去重、再交给 `rerank` 精排；`grader` 作为可插拔后置判定器接入 chat 主链，默认关闭，仅在配置启用时参与答案质量判定。整个业务层继续面向 `domain/service` 抽象，避免把 provider 细节泄漏到 handler 和 router。

**Tech Stack:** Go 1.24、Gin、Gorm、`cloudwego/eino`、OpenAI compatible chat API、`httptest`。

---

### Task 1: 扩展配置与 IoC 装配

**Files:**
- Modify: `backend/config/config.go`
- Modify: `backend/config/config.yaml`
- Modify: `backend/config/config_test.go`
- Modify: `backend/ioc/app.go`
- Modify: `backend/ioc/app_test.go`

**Step 1: Write the failing test**

为配置与装配补回归用例，覆盖：
- `rerank` 配置默认关闭，支持 provider、base_url、api_key、model、top_n、min_score
- `quality.qa` 配置支持默认问题数、字段名、开关
- `quality.grader` 配置默认关闭，启用时要求可创建 grader 组件
- `NewApp` 在未启用质量链时仍保持当前行为

**Step 2: Run test to verify it fails**

Run: `go test ./config ./ioc -run 'TestConfig|TestNewApp' -count=1`
Expected: FAIL，因为质量链配置项和装配逻辑尚不存在

**Step 3: Write minimal implementation**

实现约束：
- 配置按“默认关闭、显式开启”设计
- QA 默认复用 chat model 作为生成模型，rerank/grader 默认复用 chat provider 凭据作为兜底
- IoC 仅负责构造质量链组件，不在装配层写业务逻辑

**Step 4: Run test to verify it passes**

Run: `go test ./config ./ioc -run 'TestConfig|TestNewApp' -count=1`
Expected: PASS

### Task 2: 先写 QA 索引链路测试并接入实现

**Files:**
- Modify: `backend/internal/domain/service/indexer.go`
- Create: `backend/internal/domain/service/quality.go`
- Modify: `backend/internal/service/indexer.go`
- Modify: `backend/internal/service/indexer_test.go`
- Create: `backend/internal/service/qa_generator.go`
- Create: `backend/internal/service/qa_generator_test.go`

**Step 1: Write the failing test**

新增测试覆盖：
- `IndexerService` 在 chunk 生成后会请求 QA 生成器，为每个 chunk 产出 QA 内容
- QA 生成失败时不阻断索引主链，但 chunk 仍正常落库
- 向量写入请求同时携带内容语料和 QA 语料

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run 'TestIndexerService|TestQAGenerator' -count=1`
Expected: FAIL，因为当前索引链尚无 QA 生成与扩展向量载荷

**Step 3: Write minimal implementation**

实现约束：
- 为向量写入请求增加 QA 文本字段，不改现有 HTTP DTO
- QA 生成器只依赖统一的 chat/model 抽象
- 对每个 chunk 最多生成固定数量问题，结果以稳定文本格式持久化到 `ext` 或独立字段中

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service -run 'TestIndexerService|TestQAGenerator' -count=1`
Expected: PASS

### Task 3: 先写双通道检索测试并接入实现

**Files:**
- Modify: `backend/internal/domain/service/retriever.go`
- Modify: `backend/internal/domain/service/vector.go`
- Modify: `backend/internal/service/retriever.go`
- Modify: `backend/internal/service/retriever_test.go`
- Create: `backend/internal/service/retrieval_pipeline.go`
- Create: `backend/internal/service/retrieval_pipeline_test.go`
- Modify: `backend/internal/vector/retriever.go`
- Modify: `backend/internal/vector/retriever_test.go`

**Step 1: Write the failing test**

新增测试覆盖：
- 主检索会同时发起内容检索与 QA 检索
- 两路结果按 `chunk_id` 合并去重，并保留更高得分
- QA 路径报错时仅忽略该路，不影响内容检索结果
- 向量检索为空时继续维持既有 DB 降级逻辑

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service ./internal/vector -run 'TestRetrieverService|TestRetrievalPipeline|TestVectorRetriever' -count=1`
Expected: FAIL，因为当前只有单路检索

**Step 3: Write minimal implementation**

实现约束：
- 业务层只依赖统一 `Retriever` / `VectorRetriever` 抽象
- `vector` 层尽量复用 Eino retriever 接口和 graph，避免新增自定义协议
- 合并策略保持简单可预测：按 chunk 聚合，取最高分，最终统一排序截断

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service ./internal/vector -run 'TestRetrieverService|TestRetrievalPipeline|TestVectorRetriever' -count=1`
Expected: PASS

### Task 4: 先写 rerank 测试并接入检索主链

**Files:**
- Create: `backend/internal/service/reranker.go`
- Create: `backend/internal/service/reranker_test.go`
- Modify: `backend/internal/service/retriever.go`
- Modify: `backend/internal/service/retriever_test.go`
- Modify: `backend/ioc/app.go`

**Step 1: Write the failing test**

新增测试覆盖：
- 检索服务会在合并结果后调用 reranker 精排
- rerank 失败时回退为原始检索排序，不中断查询
- rerank 会按配置 `top_n` 截断并按 `min_score` 过滤

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run 'TestRetrieverService|TestReranker' -count=1`
Expected: FAIL，因为当前没有 rerank 组件

**Step 3: Write minimal implementation**

实现约束：
- reranker 先复用 OpenAI compatible HTTP 调用方式，与旧 `server` 行为对齐
- 组件接口保持最小，只暴露“输入 query + chunks，输出排序后 chunks”
- 失败按非阻断处理，方便线上逐步启用

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service -run 'TestRetrieverService|TestReranker' -count=1`
Expected: PASS

### Task 5: 先写 grader 测试并以可插拔方式接入 chat

**Files:**
- Create: `backend/internal/service/grader.go`
- Create: `backend/internal/service/grader_test.go`
- Modify: `backend/internal/service/chat.go`
- Modify: `backend/internal/service/chat_test.go`
- Modify: `backend/ioc/app.go`

**Step 1: Write the failing test**

新增测试覆盖：
- chat 在 grader 启用时会基于问题、答案、references 调用 grader
- grader 判定失败时返回明确错误或空答案策略
- grader 关闭时 chat 行为与当前保持一致

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run 'TestChatService|TestGrader' -count=1`
Expected: FAIL，因为当前 chat 不包含 grader 钩子

**Step 3: Write minimal implementation**

实现约束：
- grader 默认关闭
- grader 只做结果判定，不承担重写答案职责
- chat 服务仅在最终落库前应用 grader 结果，避免污染历史消息

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service -run 'TestChatService|TestGrader' -count=1`
Expected: PASS

### Task 6: 文档更新与全量验证

**Files:**
- Modify: `backend/README.md`
- Modify: `backend/.env.example`
- Modify: `backend/docs/openapi.yaml`

**Step 1: Write the failing verification target**

梳理本轮对外可见变化：
- 新增质量链配置项
- 检索会走“内容 + QA + rerank”
- grader 为可选能力，默认关闭

**Step 2: Run targeted verification**

Run: `go test ./config ./ioc ./internal/service ./internal/vector -count=1`
Expected: PASS

**Step 3: Write minimal documentation updates**

文档更新内容：
- README 增加质量链架构说明和开关
- `.env.example` 补充 `GO_RAG_RERANK_*`、`GO_RAG_QUALITY_QA_*`、`GO_RAG_QUALITY_GRADER_*`
- OpenAPI 仅补充行为说明，不修改现有 schema

**Step 4: Run full verification**

Run: `go test ./... -count=1`
Expected: PASS
