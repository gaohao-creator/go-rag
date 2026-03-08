package service

import "context"

type IndexRequest struct {
	URI           string
	KnowledgeName string
}

type IndexedChunk struct {
	ChunkID string
	Content string
	Ext     string
}

type Indexer interface {
	Index(ctx context.Context, req IndexRequest) ([]IndexedChunk, error)
}
