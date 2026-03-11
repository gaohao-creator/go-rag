package service

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
)

type IndexRequest struct {
	URI           string
	KnowledgeName string
}

type IndexedChunk struct {
	ChunkID string
	Content string
	Ext     string
}

// Indexer 定义文档索引能力，负责将输入资源切分并产出可检索的 chunk。
type Indexer interface {
	Index(ctx context.Context, req IndexRequest) ([]IndexedChunk, error)
}

type RetrieveRequest struct {
	Question      string
	TopK          int
	Score         float64
	KnowledgeName string
}

// Retriever 定义检索能力，负责根据问题从指定知识库召回相关 chunk。
type Retriever interface {
	Retrieve(ctx context.Context, req RetrieveRequest) ([]domainmodel.RetrievedChunk, error)
}

type ChunkStoreRequest struct {
	KnowledgeName string
	Chunks        []domainmodel.Chunk
	QAContents    map[string]string
}

// ChunkStore 定义 chunk 持久化能力，负责将 chunk 及其附加 QA 内容写入存储。
type ChunkStore interface {
	Store(ctx context.Context, req ChunkStoreRequest) error
}

type PromptGenerateInput struct {
	SystemPrompt string
	UserPrompt   string
}

// PromptModel 定义提示词生成能力，负责基于系统提示和用户提示生成文本结果。
type PromptModel interface {
	GeneratePrompt(ctx context.Context, in PromptGenerateInput) (string, error)
}

type QAGenerateInput struct {
	KnowledgeName string
	Content       string
}

// QAGenerator 定义问题生成能力，负责从内容中生成可用于检索增强的问题语料。
type QAGenerator interface {
	Generate(ctx context.Context, in QAGenerateInput) (string, error)
}

type RerankInput struct {
	Question string
	TopK     int
	Chunks   []domainmodel.RetrievedChunk
}

// Reranker 定义重排能力，负责对召回结果按相关性重新排序。
type Reranker interface {
	Rerank(ctx context.Context, in RerankInput) ([]domainmodel.RetrievedChunk, error)
}

type GradeInput struct {
	Question   string
	Answer     string
	References []domainmodel.RetrievedChunk
}

// Grader 定义答案评分能力，负责判断回答是否通过质量检查。
type Grader interface {
	Grade(ctx context.Context, in GradeInput) (bool, error)
}
