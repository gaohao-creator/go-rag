# Backend Phase 1 Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在 `backend` 中完成第一阶段重写，使用 `gin + gorm + 依赖注入` 跑通“知识库 -> 文档 -> chunk -> indexer -> retriever”主链路，并与旧 `server` 隔离。

**Architecture:** 采用分层重组方案：`web` 负责 HTTP 适配，`service` 负责编排用例，`repository` 作为 DAO 上一层的仓储封装，`dao/model` 承担 Gorm 持久化细节，`domain` 只保留核心实体与接口。旧 `server/core` 中与 `gf` 无关的纯能力只做适配复用，不把旧控制器、旧 DAO、旧 DTO 原样迁入。

**Tech Stack:** Go 1.24、Gin、Gorm、MySQL、`httptest`、`testify`（如已引入）、旧 `server/core` 中可复用的 indexer/retriever 能力。

---

### Task 1: 固定目录骨架与模块入口

**Files:**
- Create: `backend/cmd/server/main.go`
- Create: `backend/api/dto/kb.go`
- Create: `backend/api/dto/document.go`
- Create: `backend/api/dto/chunk.go`
- Create: `backend/api/dto/indexer.go`
- Create: `backend/api/dto/retriever.go`
- Create: `backend/internal/domain/entity/knowledge_base.go`
- Create: `backend/internal/domain/entity/document.go`
- Create: `backend/internal/domain/entity/chunk.go`
- Create: `backend/internal/domain/repository/knowledge_base.go`
- Create: `backend/internal/domain/repository/document.go`
- Create: `backend/internal/domain/repository/chunk.go`
- Create: `backend/internal/domain/service/indexer.go`
- Create: `backend/internal/domain/service/retriever.go`
- Create: `backend/internal/web/router/router.go`
- Test: `backend/cmd/server/main_test.go`

**Step 1: 写失败测试，验证入口能完成最小装配**

```go
package main

import "testing"

func TestBuildApp_DoesNotPanic(t *testing.T) {
	_, err := buildApp()
	if err != nil {
		t.Fatalf("buildApp returned error: %v", err)
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./cmd/server -run TestBuildApp_DoesNotPanic -count=1`
Expected: FAIL with `undefined: buildApp`

**Step 3: 写最小实现，先把入口、目录、空接口立住**

```go
package main

import "github.com/gin-gonic/gin"

func buildApp() (*gin.Engine, error) {
	return gin.New(), nil
}

func main() {
	app, err := buildApp()
	if err != nil {
		panic(err)
	}
	_ = app.Run(":8080")
}
```

**Step 4: 再次运行测试确认通过**

Run: `go test ./cmd/server -run TestBuildApp_DoesNotPanic -count=1`
Expected: PASS

**Step 5: 提交**

```bash
git add backend/cmd/server/main.go backend/cmd/server/main_test.go backend/api/dto backend/internal/domain backend/internal/web/router
git commit -m "feat: scaffold backend phase1 skeleton"
```

### Task 2: 建立配置加载与数据库初始化

**Files:**
- Create: `backend/config/config.go`
- Create: `backend/config/config.yaml`
- Create: `backend/ioc/config.go`
- Create: `backend/ioc/db.go`
- Create: `backend/ioc/app.go`
- Test: `backend/ioc/config_test.go`
- Test: `backend/ioc/db_test.go`

**Step 1: 写失败测试，验证配置与 DB 初始化输出非空**

```go
package ioc

import "testing"

func TestNewConfig_LoadsDefaultConfig(t *testing.T) {
	conf, err := NewConfig("../config/config.yaml")
	if err != nil {
		t.Fatalf("NewConfig returned error: %v", err)
	}
	if conf.HTTP.Port == "" {
		t.Fatal("expected HTTP port")
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./ioc -run TestNewConfig_LoadsDefaultConfig -count=1`
Expected: FAIL with `undefined: NewConfig`

**Step 3: 写最小实现，显式定义配置结构与 Gorm 初始化**

```go
type Config struct {
	HTTP struct {
		Port string `yaml:"port"`
	} `yaml:"http"`
	MySQL struct {
		DSN string `yaml:"dsn"`
	} `yaml:"mysql"`
}

func NewConfig(path string) (*config.Config, error) { /* 从 yaml 加载 */ }
func NewDB(conf *config.Config) (*gorm.DB, error) { /* gorm.Open(mysql.Open(conf.MySQL.DSN)) */ }
```

**Step 4: 运行针对性测试**

Run: `go test ./ioc -run 'TestNewConfig_LoadsDefaultConfig|TestNewDB' -count=1`
Expected: PASS

**Step 5: 提交**

```bash
git add backend/config backend/ioc
git commit -m "feat: add backend config and gorm bootstrap"
```

### Task 3: 定义统一响应、中间件与路由注册

**Files:**
- Create: `backend/internal/web/middleware/logger.go`
- Create: `backend/internal/web/middleware/recovery.go`
- Create: `backend/internal/web/middleware/cors.go`
- Create: `backend/internal/web/middleware/response.go`
- Modify: `backend/internal/web/router/router.go`
- Test: `backend/internal/web/router/router_test.go`

**Step 1: 写失败测试，验证 `/healthz` 返回统一 JSON 结构**

```go
func TestRouter_Healthz(t *testing.T) {
	engine := NewRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/web/router -run TestRouter_Healthz -count=1`
Expected: FAIL with `undefined: NewRouter`

**Step 3: 写最小实现，固定统一响应格式**

```go
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewRouter(handler *handler.Handler) *gin.Engine {
	r := gin.New()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: "pong"})
	})
	return r
}
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/web/router -run TestRouter_Healthz -count=1`
Expected: PASS

**Step 5: 提交**

```bash
git add backend/internal/web
git commit -m "feat: add gin router and middleware skeleton"
```

### Task 4: 建立 Gorm Model 与 DAO 层

**Files:**
- Create: `backend/internal/repository/dao/model/knowledge_base.go`
- Create: `backend/internal/repository/dao/model/document.go`
- Create: `backend/internal/repository/dao/model/chunk.go`
- Create: `backend/internal/repository/dao/knowledge_base.go`
- Create: `backend/internal/repository/dao/document.go`
- Create: `backend/internal/repository/dao/chunk.go`
- Create: `backend/internal/repository/dao/auto_migrate.go`
- Test: `backend/internal/repository/dao/knowledge_base_test.go`
- Test: `backend/internal/repository/dao/document_test.go`
- Test: `backend/internal/repository/dao/chunk_test.go`

**Step 1: 写失败测试，验证 DAO 可以基于 sqlite 内存库创建和查询记录**

```go
func TestKnowledgeBaseDAO_Create(t *testing.T) {
	db := newTestDB(t)
	dao := NewKnowledgeBaseDAO(db)
	id, err := dao.Create(context.Background(), &model.KnowledgeBase{Name: "demo"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected inserted id")
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/repository/dao -run TestKnowledgeBaseDAO_Create -count=1`
Expected: FAIL with `undefined: NewKnowledgeBaseDAO`

**Step 3: 写最小实现，映射旧表结构但去掉 gf 依赖**

```go
type KnowledgeBase struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string    `gorm:"column:name;size:50;not null"`
	Description string    `gorm:"column:description;size:200;not null"`
	Category    string    `gorm:"column:category;size:50"`
	Status      int       `gorm:"column:status;not null"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (KnowledgeBase) TableName() string { return "knowledge_base" }
```

**Step 4: 运行 DAO 层测试**

Run: `go test ./internal/repository/dao -run 'TestKnowledgeBaseDAO_Create|TestDocumentDAO|TestChunkDAO' -count=1`
Expected: PASS

**Step 5: 提交**

```bash
git add backend/internal/repository/dao
git commit -m "feat: add gorm models and dao layer"
```

### Task 5: 建立 Repository 层，作为 DAO 上一层应用封装

**Files:**
- Create: `backend/internal/repository/knowledge_base.go`
- Create: `backend/internal/repository/document.go`
- Create: `backend/internal/repository/chunk.go`
- Modify: `backend/internal/domain/repository/knowledge_base.go`
- Modify: `backend/internal/domain/repository/document.go`
- Modify: `backend/internal/domain/repository/chunk.go`
- Test: `backend/internal/repository/knowledge_base_test.go`
- Test: `backend/internal/repository/document_test.go`
- Test: `backend/internal/repository/chunk_test.go`

**Step 1: 写失败测试，验证仓储层返回领域对象而不是 Gorm Model**

```go
func TestKnowledgeBaseRepository_Create(t *testing.T) {
	repo := newTestKnowledgeBaseRepository(t)
	id, err := repo.Create(context.Background(), entity.KnowledgeBase{Name: "demo"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected inserted id")
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/repository -run TestKnowledgeBaseRepository_Create -count=1`
Expected: FAIL with `undefined: newTestKnowledgeBaseRepository`

**Step 3: 写最小实现，完成 model ↔ entity 转换**

```go
type KnowledgeBaseRepository struct {
	dao *dao.KnowledgeBaseDAO
}

func (r *KnowledgeBaseRepository) Create(ctx context.Context, kb entity.KnowledgeBase) (int64, error) {
	return r.dao.Create(ctx, &model.KnowledgeBase{
		Name:        kb.Name,
		Description: kb.Description,
		Category:    kb.Category,
		Status:      kb.Status,
	})
}
```

**Step 4: 运行仓储层测试**

Run: `go test ./internal/repository -run 'TestKnowledgeBaseRepository_Create|TestDocumentRepository|TestChunkRepository' -count=1`
Expected: PASS

**Step 5: 提交**

```bash
git add backend/internal/repository backend/internal/domain/repository
git commit -m "feat: add repository layer above dao"
```

### Task 6: 落地知识库、文档、Chunk 应用服务

**Files:**
- Create: `backend/internal/service/knowledge_base.go`
- Create: `backend/internal/service/document.go`
- Create: `backend/internal/service/chunk.go`
- Test: `backend/internal/service/knowledge_base_test.go`
- Test: `backend/internal/service/document_test.go`
- Test: `backend/internal/service/chunk_test.go`

**Step 1: 写失败测试，验证 service 只编排用例，不关心数据库细节**

```go
func TestKnowledgeBaseService_Create(t *testing.T) {
	repo := newFakeKnowledgeBaseRepository()
	svc := NewKnowledgeBaseService(repo)
	id, err := svc.Create(context.Background(), CreateKnowledgeBaseInput{Name: "demo", Description: "desc"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected inserted id")
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/service -run TestKnowledgeBaseService_Create -count=1`
Expected: FAIL with `undefined: NewKnowledgeBaseService`

**Step 3: 写最小实现，承接旧 `rag_v1_kb.go` 与 `logic/knowledge/*.go` 的语义**

```go
type KnowledgeBaseService struct {
	repo repository.KnowledgeBaseRepository
}

func (s *KnowledgeBaseService) Create(ctx context.Context, in CreateKnowledgeBaseInput) (int64, error) {
	if in.Name == "" {
		return 0, errors.New("知识库名称不能为空")
	}
	return s.repo.Create(ctx, entity.KnowledgeBase{
		Name:        in.Name,
		Description: in.Description,
		Category:    in.Category,
		Status:      1,
	})
}
```

**Step 4: 运行服务层测试**

Run: `go test ./internal/service -run 'TestKnowledgeBaseService|TestDocumentService|TestChunkService' -count=1`
Expected: PASS

**Step 5: 提交**

```bash
git add backend/internal/service
git commit -m "feat: add phase1 application services"
```

### Task 7: 适配旧 core 的 Indexer 能力

**Files:**
- Create: `backend/internal/service/indexer.go`
- Modify: `backend/internal/domain/service/indexer.go`
- Create: `backend/internal/service/indexer_test.go`
- Create: `backend/internal/service/testdata/index.txt`
- Reference: `server/core/indexer/indexer.go`
- Reference: `server/core/indexer/indexer_async.go`

**Step 1: 写失败测试，验证索引流程会创建文档记录并回写 chunk**

```go
func TestIndexerService_Index(t *testing.T) {
	svc := newTestIndexerService(t)
	ids, err := svc.Index(context.Background(), IndexInput{
		URI:           "./testdata/index.txt",
		KnowledgeName: "demo",
		FileName:      "index.txt",
	})
	if err != nil {
		t.Fatalf("Index returned error: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("expected chunk ids")
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/service -run TestIndexerService_Index -count=1`
Expected: FAIL with `undefined: IndexInput`

**Step 3: 写最小实现，保持第一阶段只做同步索引主流程**

```go
type Indexer interface {
	Index(ctx context.Context, req IndexRequest) ([]string, error)
}

type IndexerService struct {
	documentSvc DocumentService
	chunkSvc    ChunkService
	indexer     domainservice.Indexer
}
```

实现约束：
- 先创建 `knowledge_documents` 记录，状态置为 `pending`
- 调用底层 indexer 解析文档
- 成功后写入 chunk 记录并更新文档状态为 `active`
- 失败时更新文档状态为 `failed`
- 第一阶段不引入后台任务系统，只保留同步主链路

**Step 4: 运行测试确认通过**

Run: `go test ./internal/service -run TestIndexerService_Index -count=1`
Expected: PASS

**Step 5: 提交**

```bash
git add backend/internal/service/indexer.go backend/internal/domain/service/indexer.go backend/internal/service/indexer_test.go
git commit -m "feat: integrate phase1 indexer service"
```

### Task 8: 适配旧 core 的 Retriever 能力

**Files:**
- Create: `backend/internal/service/retriever.go`
- Modify: `backend/internal/domain/service/retriever.go`
- Create: `backend/internal/service/retriever_test.go`
- Reference: `server/core/retriever/retriever.go`
- Reference: `server/core/retriever/orchestration.go`

**Step 1: 写失败测试，验证检索结果按分数降序返回**

```go
func TestRetrieverService_Retrieve(t *testing.T) {
	svc := newTestRetrieverService(t)
	docs, err := svc.Retrieve(context.Background(), RetrieveInput{
		Question:      "什么是 RAG",
		TopK:          5,
		Score:         0.2,
		KnowledgeName: "demo",
	})
	if err != nil {
		t.Fatalf("Retrieve returned error: %v", err)
	}
	if len(docs) == 0 {
		t.Fatal("expected retrieval results")
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/service -run TestRetrieverService_Retrieve -count=1`
Expected: FAIL with `undefined: RetrieveInput`

**Step 3: 写最小实现，保留旧 `rag_v1_retriever.go` 的默认值与排序语义**

```go
type RetrieverService struct {
	retriever domainservice.Retriever
}

func (s *RetrieverService) Retrieve(ctx context.Context, in RetrieveInput) ([]*schema.Document, error) {
	if in.TopK == 0 {
		in.TopK = 5
	}
	if in.Score == 0 {
		in.Score = 0.2
	}
	return s.retriever.Retrieve(ctx, in)
}
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/service -run TestRetrieverService_Retrieve -count=1`
Expected: PASS

**Step 5: 提交**

```bash
git add backend/internal/service/retriever.go backend/internal/domain/service/retriever.go backend/internal/service/retriever_test.go
git commit -m "feat: integrate phase1 retriever service"
```

### Task 9: 暴露阶段一 HTTP 接口

**Files:**
- Create: `backend/internal/web/handler/handler.go`
- Create: `backend/internal/web/handler/knowledge_base.go`
- Create: `backend/internal/web/handler/document.go`
- Create: `backend/internal/web/handler/chunk.go`
- Create: `backend/internal/web/handler/indexer.go`
- Create: `backend/internal/web/handler/retriever.go`
- Modify: `backend/internal/web/router/router.go`
- Test: `backend/internal/web/handler/knowledge_base_test.go`
- Test: `backend/internal/web/handler/indexer_test.go`
- Test: `backend/internal/web/handler/retriever_test.go`

**Step 1: 写失败测试，验证 `POST /api/v1/kb` 与 `POST /api/v1/retriever` 可以命中 handler**

```go
func TestKnowledgeBaseHandler_Create(t *testing.T) {
	engine := newTestRouter(t)
	body := strings.NewReader(`{"name":"demo","description":"desc"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/kb", body)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/web/handler -run TestKnowledgeBaseHandler_Create -count=1`
Expected: FAIL with `404 page not found`

**Step 3: 写最小实现，先把阶段一所有路由注册完整**

路由清单：
- `POST /api/v1/kb`
- `PUT /api/v1/kb/:id`
- `DELETE /api/v1/kb/:id`
- `GET /api/v1/kb`
- `GET /api/v1/kb/:id`
- `GET /api/v1/documents`
- `DELETE /api/v1/documents`
- `GET /api/v1/chunks`
- `DELETE /api/v1/chunks`
- `PUT /api/v1/chunks`
- `PUT /api/v1/chunks-content`
- `POST /api/v1/indexer`
- `POST /api/v1/retriever`

**Step 4: 运行 Handler 与 Router 测试**

Run: `go test ./internal/web/... -count=1`
Expected: PASS

**Step 5: 提交**

```bash
git add backend/internal/web
git commit -m "feat: expose phase1 gin handlers and routes"
```

### Task 10: 装配 IOC、初始化 SQL 与阶段一总验证

**Files:**
- Modify: `backend/ioc/app.go`
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/scripts/mysql/init.sql`
- Test: `backend/ioc/app_test.go`
- Test: `backend/internal/web/router/integration_test.go`

**Step 1: 写失败测试，验证 `buildApp()` 返回的应用已具备阶段一完整依赖**

```go
func TestBuildApp_HasPhase1Routes(t *testing.T) {
	app, err := buildApp()
	if err != nil {
		t.Fatalf("buildApp returned error: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./cmd/server ./ioc ./internal/web/router -run TestBuildApp_HasPhase1Routes -count=1`
Expected: FAIL until IOC wiring is complete

**Step 3: 写最小实现，补齐依赖注入和初始化 SQL**

`backend/scripts/mysql/init.sql` 至少包含：
- `knowledge_base`
- `knowledge_documents`
- `knowledge_chunks`

IOC 装配顺序：
- 加载配置
- 初始化 DB
- AutoMigrate / 或显式校验表
- 构建 DAO
- 构建 Repository
- 构建 Service
- 构建 Handler
- 构建 Router

**Step 4: 运行阶段一总验证**

Run: `go test ./... -count=1`
Expected: PASS

如存在本地 MySQL 环境，再补一轮：

Run: `go test ./internal/web/router -run TestIntegration -count=1`
Expected: PASS

**Step 5: 提交**

```bash
git add backend/cmd/server backend/ioc backend/scripts/mysql/init.sql
git commit -m "feat: wire backend phase1 application"
```

### Task 11: 迁移规则与明确不做项

**Files:**
- Modify: `docs/plans/2026-03-08-backend-phase1-migration.md`
- Reference: `server/internal/controller/rag/rag_v1_chat.go`
- Reference: `server/internal/cmd/cmd.go`
- Reference: `server/internal/mcp/*.go`

**Step 1: 写检查清单，确保开发时不越界**

```text
- 不迁移 chat / chat_stream
- 不迁移 mcp
- 不迁移静态页面托管
- 不迁移 gf 命令体系
- 不在第一阶段重构文档与知识库的关联键模型
```

**Step 2: 在实现前核对范围**

Run: `rg -n "Chat|Mcp|StaticPath|chat_stream" server`
Expected: 仅作参考，不进入阶段一任务列表

**Step 3: 若实现中出现额外需求，先回到本计划补充任务，不直接编码**

```text
如果新增“后台异步索引”“文件管理”“知识库 ID 化重构”等需求，必须单列为阶段 1.1 或阶段 2 前置任务。
```

**Step 4: 人工复核范围没有漂移**

Run: `git diff --stat`
Expected: 仅涉及 `backend` 与 `docs/plans`

**Step 5: 提交**

```bash
git add docs/plans/2026-03-08-backend-phase1-migration.md
git commit -m "docs: finalize backend phase1 migration plan"
```
