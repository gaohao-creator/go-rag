# Backend 向量检索迁移设计（ES 优先）

## 背景

`backend` 已完成主链路重写，但当前索引与检索仍以数据库方式为主，
未覆盖旧 `server` 中的向量索引、向量检索、重排与评分能力。
本设计在保持旧 ES 字段与 mapping 兼容的前提下，将向量能力迁移至
`backend`，并采用“ES 优先 + DB 降级”的运行策略。

## 目标

- 以 ES 为优先向量存储与检索后端
- 保持旧索引字段与 mapping 兼容（如 `content`、`content_vector`、`ext`、`knowledge_name`）
- 索引按知识库隔离（每知识库独立索引或 alias）
- 保留 DB 检索作为降级路径
- 保持对外接口与 DTO 语义不变

## 非目标

- 不重写现有 HTTP 接口与前端契约
- 不改动旧 `server` 控制器实现
- 不引入与 ES 无关的向量数据库

## 方案选择

采用“适配层 + 工厂”方案：
在 `backend` 内新增向量能力适配模块，
通过配置与工厂决定是否启用 ES 能力；
检索阶段优先 ES，失败或未配置时自动降级到 DB。

该方案兼顾兼容性与长期可维护性，避免直接移植旧层级结构导致耦合。

## 架构与模块边界

- `web`：仅处理 HTTP 适配与参数校验
- `service`：负责业务编排与降级策略
- `domain`：只保留实体与接口定义
- `repository/dao`：关系型持久化
- `vector`（新增）：ES 索引、检索与解析适配层

新增模块通过 `ioc` 注入，业务层不直接依赖 ES SDK。

## 核心组件

- `VectorIndexer`：将切分后的 chunk 写入 ES
- `VectorRetriever`：基于向量检索返回带分数的 chunk
- `IndexNameResolver`：将 `knowledge_name` 映射为索引名或 alias
- `VectorFactory`：根据配置返回 ES 或空实现

## 数据流

### 索引链路

1. `IndexerService` 调用 `IndexerEngine` 切分文档
2. 结果先写入 DB（保证最小可用与可追溯）
3. 同步调用 `VectorIndexer` 写入 ES
4. 索引不存在时自动创建并应用旧 mapping

### 检索链路

1. `RetrieverService` 优先调用 `VectorRetriever`
2. ES 不可用或失败则降级 DB 检索
3. 需要时接入 `Rerank` 与 `Grader` 做后处理

## 配置与索引策略

新增配置项（示意）：
- ES 地址、认证、超时
- 索引名前缀与命名策略
- 是否启用每知识库索引
- 向量字段名与维度
- Embedding 模型与 BaseURL

索引命名策略为“每知识库单独索引”。
`IndexNameResolver` 保证索引命名可预测与可追溯。

## 兼容性要求

- 字段名与 mapping 与旧版本一致
- 结果解析逻辑与旧 ES 命中文档一致
- 对外接口字段不变

## 错误处理与降级

- ES 初始化失败：记录错误，继续 DB 路径
- ES 写入失败：保证 DB 已落库，索引链路不中断
- 检索异常：自动降级 DB 检索
- 所有异常均返回统一错误语义

## 测试策略

- 单元测试：索引名映射、字段兼容、检索解析
- 集成测试：ES 启动下的索引/检索全链路
- 回归测试：接口响应结构与旧版本一致

## 迁移步骤

1. 配置与 IoC 先落地，ES 能力隐藏在开关下
2. 上线 `VectorIndexer`，先只写索引，不改检索
3. 切换 `VectorRetriever` 为主路径，保留 DB 降级

## 风险与假设

- 旧索引若已存在，需谨慎处理 mapping 变更
- Embedding 维度与模型不匹配会导致召回异常
- 旧接口调用方式多样时可能需要额外适配

## 里程碑

- M1：配置与 ES 客户端接入完成
- M2：索引链路写入 ES 通过集成测试
- M3：检索链路 ES 优先 + DB 降级稳定
- M4：重排与评分链路接入并完成回归
