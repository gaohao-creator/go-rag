package rag

import (
	"context"
	"fmt"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

// RAG 统一封装索引、存储、检索和评分能力。
type RAG struct {
	indexer   domainservice.Indexer
	store     domainservice.ChunkStore
	retriever domainservice.Retriever
	grader    domainservice.Grader
}

// NewRAG 创建 RAG 根门面。
func NewRAG(
	indexer domainservice.Indexer,
	store domainservice.ChunkStore,
	retriever domainservice.Retriever,
	grader domainservice.Grader,
) *RAG {
	return &RAG{
		indexer:   indexer,
		store:     store,
		retriever: retriever,
		grader:    grader,
	}
}

// Index 执行索引流程并返回产出的 chunk。
func (r *RAG) Index(ctx context.Context, req domainservice.IndexRequest) ([]domainservice.IndexedChunk, error) {
	if r == nil || r.indexer == nil {
		return nil, fmt.Errorf("RAG 索引器未配置")
	}
	return r.indexer.Index(ctx, req)
}

// Store 将 chunk 写入底层检索存储。
func (r *RAG) Store(ctx context.Context, req domainservice.ChunkStoreRequest) error {
	if r == nil || r.store == nil {
		return nil
	}
	return r.store.Store(ctx, req)
}

// Retrieve 根据问题召回相关 chunk。
func (r *RAG) Retrieve(ctx context.Context, req domainservice.RetrieveRequest) ([]domainmodel.RetrievedChunk, error) {
	if r == nil || r.retriever == nil {
		return nil, fmt.Errorf("RAG 检索器未配置")
	}
	return r.retriever.Retrieve(ctx, req)
}

// Grade 对回答结果做质量判定。
func (r *RAG) Grade(ctx context.Context, req domainservice.GradeInput) (bool, error) {
	if r == nil || r.grader == nil {
		return true, nil
	}
	return r.grader.Grade(ctx, req)
}
