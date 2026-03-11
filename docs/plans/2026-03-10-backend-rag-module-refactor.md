# Backend RAG Module Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 `backend/internal/vector` 与散落在 `internal/service`、`internal/domain/service` 中的 RAG 能力收拢到 `backend/internal/rag` 目录下，并让 `service` / `ioc` 直接面向 `rag` 根包暴露的能力工作。

**Architecture:** 新增 `internal/rag` 根包作为 RAG 能力门面，统一暴露索引、检索、写向量、重排、评分、PromptModel 等契约；具体实现按能力拆到 `rag/indexer`、`rag/retriever`、`rag/rerank`、`rag/grader`、`rag/es` 等子目录。`service` 只依赖 `rag` 根包，`ioc` 只负责调用 `rag` 根包的构造函数完成装配。

**Tech Stack:** Go 1.24、Gin、Gorm、`cloudwego/eino`、Elasticsearch v8 typed API、`httptest`。

---

### Task 1: 建立 rag 根包契约与门面

**Files:**
- Create: `backend/internal/rag/contracts/contracts.go`
- Create: `backend/internal/rag/rag.go`
- Modify: `backend/ioc/app_test.go`
- Modify: `backend/internal/service/indexer_test.go`
- Modify: `backend/internal/service/retriever_test.go`
- Modify: `backend/internal/service/chat_test.go`

**Step 1: Write the failing test**

新增/调整测试覆盖：
- `service` 测试中的 fake 依赖改为实现 `rag` 根包暴露的接口
- `ioc` 测试中新增对 `buildRAGComponents` 或等效门面装配函数的断言
- 断言 `service` 不再需要直接拼接 `VectorWriter/QAGenerator/Reranker/Grader` 等散接口

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service ./ioc -run 'TestIndexerService|TestRetrieverService|TestChatService|TestBuild' -count=1`
Expected: FAIL，因为 `rag` 根包和新的 service 依赖面还不存在

**Step 3: Write minimal implementation**

实现约束：
- `service` 和 `ioc` 只 import `internal/rag`
- `internal/domain/service` 仅保留 chat model 相关定义
- `rag` 根包通过 type alias / wrapper function 向外暴露稳定入口

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service ./ioc -run 'TestIndexerService|TestRetrieverService|TestChatService|TestBuild' -count=1`
Expected: PASS

### Task 2: 迁移索引能力到 rag/indexer

**Files:**
- Move: `backend/internal/service/indexer_engine.go` -> `backend/internal/rag/indexer/default_engine.go`
- Move: `backend/internal/service/indexer_engine_test.go` -> `backend/internal/rag/indexer/default_engine_test.go`
- Move: `backend/internal/service/qa_generator.go` -> `backend/internal/rag/indexer/qa_generator.go`
- Move: `backend/internal/service/qa_generator_test.go` -> `backend/internal/rag/indexer/qa_generator_test.go`
- Modify: `backend/internal/service/indexer.go`
- Modify: `backend/internal/service/indexer_test.go`
- Modify: `backend/ioc/app.go`

**Step 1: Write the failing test**

覆盖以下行为：
- `IndexerService` 只依赖 `rag.ChunkIndexer` 与 `rag.Writer`
- QA 生成由 `rag.Writer` 内部处理，`IndexerService` 不再单独持有 QA 依赖
- `ioc` 能通过 `rag` 根包创建默认索引引擎和向量写入器

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service ./ioc ./internal/rag/indexer -run 'TestIndexerService|TestLLMQAGenerator|TestDefaultIndexerEngine' -count=1`
Expected: FAIL，因为索引能力尚未迁移

**Step 3: Write minimal implementation**

实现约束：
- `rag/indexer` 承担“文档切块”和“QA 生成”两类职责
- 向量写入器内部可选启用 QA 生成，再将 `qa_content` 一并写入 ES
- service 层保留 DB 落库和状态流转职责

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service ./ioc ./internal/rag/indexer -run 'TestIndexerService|TestLLMQAGenerator|TestDefaultIndexerEngine' -count=1`
Expected: PASS

### Task 3: 迁移检索能力到 rag/retriever 与 rag/rerank

**Files:**
- Move: `backend/internal/service/retriever_engine.go` -> `backend/internal/rag/retriever/database.go`
- Move: `backend/internal/service/retriever_engine_test.go` -> `backend/internal/rag/retriever/database_test.go`
- Move: `backend/internal/service/retriever_fallback.go` -> `backend/internal/rag/retriever/fallback.go`
- Move: `backend/internal/service/retriever_fallback_test.go` -> `backend/internal/rag/retriever/fallback_test.go`
- Move: `backend/internal/service/reranker.go` -> `backend/internal/rag/rerank/http.go`
- Move: `backend/internal/service/reranker_test.go` -> `backend/internal/rag/rerank/http_test.go`
- Create: `backend/internal/rag/retriever/pipeline.go`
- Create: `backend/internal/rag/retriever/pipeline_test.go`
- Modify: `backend/internal/service/retriever.go`
- Modify: `backend/internal/service/retriever_test.go`
- Modify: `backend/ioc/app.go`

**Step 1: Write the failing test**

覆盖以下行为：
- `RetrieverService` 只依赖单个 `rag.Retriever`
- QA 检索并发、去重、rerank、DB 降级都在 `rag/retriever/pipeline.go` 中完成
- `ioc` 通过 `rag` 根包创建完整检索链，而不是 service 自己拼装

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service ./ioc ./internal/rag/retriever ./internal/rag/rerank -run 'TestRetrieverService|TestFallbackRetriever|TestHTTPReranker|TestPipeline' -count=1`
Expected: FAIL，因为检索编排还在 service 层

**Step 3: Write minimal implementation**

实现约束：
- service 只做入参校验、默认值处理和最终截断
- `rag/retriever` 内部负责内容检索、QA 检索、fallback、merge、rerank
- 根包只暴露最终 `rag.Retriever`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service ./ioc ./internal/rag/retriever ./internal/rag/rerank -run 'TestRetrieverService|TestFallbackRetriever|TestHTTPReranker|TestPipeline' -count=1`
Expected: PASS

### Task 4: 迁移评分与向量 ES 实现到 rag 目录

**Files:**
- Move: `backend/internal/service/grader.go` -> `backend/internal/rag/grader/llm.go`
- Move: `backend/internal/service/grader_test.go` -> `backend/internal/rag/grader/llm_test.go`
- Move: `backend/internal/vector/*` -> `backend/internal/rag/es/*` 以及 `backend/internal/rag/indexer/es_writer.go` / `backend/internal/rag/retriever/es_retriever.go`
- Modify: `backend/internal/service/chat.go`
- Modify: `backend/internal/service/chat_test.go`
- Modify: `backend/internal/vector` 相关测试迁移后的路径与 import
- Modify: `backend/ioc/app.go`

**Step 1: Write the failing test**

覆盖以下行为：
- `ChatService` 只依赖 `rag.Grader`
- `ioc` 使用 `rag` 根包创建 grader 和 ES 组件
- 原 `vector` 目录测试全部在 `rag` 新路径下仍然通过

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service ./ioc ./internal/rag/... -run 'TestChatService|TestLLMGrader|TestES' -count=1`
Expected: FAIL，因为 `vector` 尚未迁移到 `rag`

**Step 3: Write minimal implementation**

实现约束：
- `rag/es` 只放 ES / embedding / graph / mapping 等基础设施
- `rag/indexer` / `rag/retriever` 负责调用 `rag/es` 基础设施完成向量索引与检索
- 彻底删除 `internal/vector` 包

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service ./ioc ./internal/rag/... -count=1`
Expected: PASS

### Task 5: 收口 imports、删除旧接口并全量验证

**Files:**
- Delete: `backend/internal/domain/service/indexer.go`
- Delete: `backend/internal/domain/service/retriever.go`
- Delete: `backend/internal/domain/service/quality.go`
- Delete: `backend/internal/domain/service/vector.go`
- Modify: `backend/README.md`
- Modify: `backend/.env.example`
- Modify: `backend/docs/openapi.yaml`

**Step 1: Write the failing verification target**

确认：
- `service` 不再依赖旧 `internal/domain/service` 中的 RAG 契约
- `internal/vector` 目录已完全移除
- 文档路径和架构描述已更新为 `internal/rag`

**Step 2: Run targeted verification**

Run: `go test ./config ./ioc ./internal/service ./internal/rag/... -count=1`
Expected: PASS

**Step 3: Write minimal documentation updates**

文档更新内容：
- README 说明 `internal/rag` 为统一 RAG 模块目录
- `.env.example` 保持配置项不变，但描述统一改为 RAG 模块
- OpenAPI 说明检索质量链由 `rag` 模块承担

**Step 4: Run full verification**

Run: `go test ./... -count=1`
Expected: PASS
