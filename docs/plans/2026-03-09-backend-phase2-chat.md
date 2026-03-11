# Backend Phase 2 Chat Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在 `backend` 中完成第二阶段第一部分，先用可替换的假 ChatModel 适配器打通 `chat` 主链，并把会话历史持久化到当前数据库。

**Architecture:** 继续沿用 `handler -> service -> repository -> dao` 分层。`chat` 主链依赖 `retriever service` 获取参考资料，依赖 `message history repository` 读取/写入会话消息，依赖 `chat model adapter` 生成回答。第二阶段第一步只实现普通 `chat`，不实现 `chat stream`，但会把消息与模型边界设计成可复用结构，方便后续扩展流式输出。

**Tech Stack:** Go 1.24、Gin、Gorm、MySQL/SQLite、现有 `retriever` 服务、假 ChatModel 适配器、`httptest`。

---

### Task 1: 建立会话消息数据模型与持久化层

**Files:**
- Create: `backend/internal/domain/model/message.go`
- Create: `backend/internal/domain/model/message_filter.go`
- Create: `backend/internal/repository/dao/entity/message.go`
- Create: `backend/internal/repository/dao/message.go`
- Modify: `backend/internal/repository/dao/auto_migrate.go`
- Create: `backend/internal/repository/message.go`
- Create: `backend/internal/domain/repository/message.go`
- Test: `backend/internal/repository/dao/message_test.go`
- Test: `backend/internal/repository/message_test.go`

**Step 1: Write the failing test**

```go
func TestMessageDAO_CreateAndListByConversation(t *testing.T) {
    db := newTestDB(t)
    dao := NewMessageDAO(db)

    err := dao.Create(context.Background(), &daoentity.Message{
        ConvID:  "conv-1",
        Role:    "user",
        Content: "hello",
    })
    if err != nil {
        t.Fatalf("Create returned error: %v", err)
    }

    messages, err := dao.ListByConversation(context.Background(), "conv-1")
    if err != nil {
        t.Fatalf("ListByConversation returned error: %v", err)
    }
    if len(messages) != 1 {
        t.Fatalf("expected 1 message, got %d", len(messages))
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/repository/dao -run TestMessageDAO_CreateAndListByConversation -count=1`
Expected: FAIL with `undefined: NewMessageDAO`

**Step 3: Write minimal implementation**

实现约束：
- 消息表最小字段为 `id/conv_id/role/content/created_at`
- 查询按 `id asc` 返回，保证上下文顺序稳定
- 只支持 `system/user/assistant` 三类角色

**Step 4: Run test to verify it passes**

Run: `go test ./internal/repository/dao ./internal/repository -run TestMessage -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/domain/model backend/internal/domain/repository backend/internal/repository/dao backend/internal/repository
git commit -m "feat: add chat history persistence"
```

### Task 2: 建立假 ChatModel 适配器

**Files:**
- Create: `backend/internal/domain/service/chat_model.go`
- Create: `backend/internal/service/chat_model_fake.go`
- Test: `backend/internal/service/chat_model_fake_test.go`

**Step 1: Write the failing test**

```go
func TestFakeChatModel_GenerateAnswerIncludesQuestionAndReferences(t *testing.T) {
    model := NewFakeChatModel()
    answer, err := model.Generate(context.Background(), ChatGenerateInput{
        Question: "什么是 RAG",
        References: []domainmodel.RetrievedChunk{
            {ChunkID: "chunk-1", Content: "RAG 是检索增强生成。"},
        },
    })
    if err != nil {
        t.Fatalf("Generate returned error: %v", err)
    }
    if !strings.Contains(answer, "什么是 RAG") {
        t.Fatalf("expected answer to contain question, got %s", answer)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run TestFakeChatModel_GenerateAnswerIncludesQuestionAndReferences -count=1`
Expected: FAIL with `undefined: NewFakeChatModel`

**Step 3: Write minimal implementation**

实现约束：
- 假适配器输出稳定文本，便于测试
- 结果至少包含问题和命中的参考摘要
- 不引入任何外部模型依赖

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service -run TestFakeChatModel_GenerateAnswerIncludesQuestionAndReferences -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/domain/service backend/internal/service/chat_model_fake.go backend/internal/service/chat_model_fake_test.go
git commit -m "feat: add fake chat model adapter"
```

### Task 3: 建立 Chat 服务主链

**Files:**
- Create: `backend/internal/service/chat.go`
- Test: `backend/internal/service/chat_test.go`

**Step 1: Write the failing test**

```go
func TestChatService_ChatPersistsConversationAndReturnsAnswer(t *testing.T) {
    history := newFakeMessageRepository()
    retriever := newFakeRetrieverForChat()
    model := newFakeChatModelForService()
    service := NewChatService(history, retriever, model)

    result, err := service.Chat(context.Background(), ChatInput{
        ConvID:        "conv-1",
        Question:      "什么是 RAG",
        KnowledgeName: "demo",
    })
    if err != nil {
        t.Fatalf("Chat returned error: %v", err)
    }
    if result.Answer == "" {
        t.Fatal("expected answer")
    }
    if len(result.References) == 0 {
        t.Fatal("expected references")
    }
    if history.CountByConversation("conv-1") != 2 {
        t.Fatalf("expected 2 persisted messages, got %d", history.CountByConversation("conv-1"))
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run TestChatService_ChatPersistsConversationAndReturnsAnswer -count=1`
Expected: FAIL with `undefined: NewChatService`

**Step 3: Write minimal implementation**

实现约束：
- 每次调用先保存一条 `user` 消息
- 调用 `retriever service` 获得 references
- 调用 `chat model adapter` 生成 answer
- 保存一条 `assistant` 消息
- 返回 `answer + references`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service -run TestChatService_ChatPersistsConversationAndReturnsAnswer -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/internal/service/chat.go backend/internal/service/chat_test.go
git commit -m "feat: add non-stream chat service"
```

### Task 4: 暴露 `POST /api/v1/chat` 接口

**Files:**
- Create: `backend/api/dto/chat.go`
- Create: `backend/internal/web/handler/chat.go`
- Modify: `backend/internal/web/handler/handler.go`
- Modify: `backend/internal/web/router/router.go`
- Modify: `backend/ioc/app.go`
- Test: `backend/internal/web/handler/chat_test.go`
- Test: `backend/internal/web/router/integration_test.go`

**Step 1: Write the failing test**

```go
func TestChatHandler_Chat(t *testing.T) {
    handler := NewHandler(nil, nil, nil, nil, nil, &fakeChatService{})
    router := webrouter.NewRouter(handler)

    request := httptest.NewRequest(http.MethodPost, "/api/v1/chat", strings.NewReader(`{"conv_id":"conv-1","question":"什么是RAG","knowledge_name":"demo"}`))
    request.Header.Set("Content-Type", "application/json")
    response := httptest.NewRecorder()

    router.ServeHTTP(response, request)
    if response.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", response.Code)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/web/handler -run TestChatHandler_Chat -count=1`
Expected: FAIL with `undefined: fakeChatService` or route missing

**Step 3: Write minimal implementation**

实现约束：
- `POST /api/v1/chat`
- 请求字段：`conv_id/question/knowledge_name/top_k/score`
- 响应字段：`answer/references`
- 走统一响应包装

**Step 4: Run test to verify it passes**

Run: `go test ./internal/web/handler ./internal/web/router -run TestChat -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/api/dto/chat.go backend/internal/web/handler/chat.go backend/internal/web/handler/handler.go backend/internal/web/router/router.go backend/ioc/app.go
 git commit -m "feat: expose non-stream chat endpoint"
```

### Task 5: 更新数据库脚本与文档

**Files:**
- Modify: `backend/scripts/mysql/init.sql`
- Modify: `backend/README.md`
- Modify: `backend/docs/openapi.yaml`

**Step 1: Write the failing test**

使用现有集成测试验证 `/api/v1/chat` 路由和消息持久化路径；不额外引入新测试框架。

**Step 2: Run verification before docs change**

Run: `go test ./... -count=1`
Expected: PASS after Tasks 1-4 complete

**Step 3: Write minimal documentation updates**

文档更新内容：
- 增加消息表初始化 SQL
- README 增加 `POST /api/v1/chat`
- OpenAPI 增加 chat 接口与消息/引用模型说明

**Step 4: Run full verification**

Run: `go test ./... -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/scripts/mysql/init.sql backend/README.md backend/docs/openapi.yaml
 git commit -m "docs: add phase2 chat docs"
```

### Task 6: 明确延后项

**Files:**
- Modify: `docs/plans/2026-03-09-backend-phase2-chat.md`

**Step 1: Write the deferral checklist**

```text
- chat stream 延后到 phase2.2
- 真实大模型适配器延后到 phase2.1
- token usage / tool call / function calling 延后
- 多轮上下文裁剪与总结延后
```

**Step 2: Re-check boundaries**

Run: `git diff --stat`
Expected: 当前实现只涉及 chat 普通链路和消息历史

**Step 3: Commit**

```bash
git add docs/plans/2026-03-09-backend-phase2-chat.md
 git commit -m "docs: finalize phase2 chat plan"
```


### Task 6: ?????

**??????**
- ???????
- ? ChatModel ???
- ?? `POST /api/v1/chat`
- chat ?? SQL / README / OpenAPI

**????????**
- `chat stream`
- ????????
- token usage / tool call / function calling
- ??????????
- ??????????? chat
